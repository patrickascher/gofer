// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/structer"
)

// maxSearchDepth defines how many parent models will be searched.
const maxSearchDepth = 20

// Error messages:
var (
	ErrFieldType      = "orm: field type %s is not allowed (%s). Fields must include default go types or implement the sql.Scanner/driver.Valuer or relations must implement the orm.Interface or set as custom"
	ErrFieldName      = "orm: field/relation (%s) does not exist or does not have the required permission"
	ErrFieldUnique    = "orm: field name (%s) is not unique"
	ErrPrimaryKey     = "orm: no primary key is defined in %s"
	ErrMaxSearchDepth = "orm: the max parent search depth of %d was reached (%s)"
	ErrInfinityLoop   = "orm: 🎉 congratulation you created an infinity loop (%s)"
)

// Scope provide some useful helper functions for the orm.Model.
type Scope interface {
	Name(bool) string

	Builder() query.Builder
	FqdnTable() string
	FqdnModel(string) string
	Model() *Model
	Caller() Interface

	Cache() cache.Manager
	SetCache(mgr cache.Manager) error

	SQLFields(permission Permission) []Field
	Fields(permission Permission) []Field
	SQLScanFields(permission Permission) []interface{}
	SQLColumns(permission Permission) []string

	Field(name string) (*Field, error)
	FieldValue(name string) reflect.Value

	SQLRelation(relation string, permission Permission) (Relation, error)
	SQLRelations(permission Permission) []Relation
	Relations(permission Permission) []Relation

	PrimaryKeys() ([]Field, error)
	PrimaryKeysSet() bool

	// internals
	foreignKey(tag string, tags map[string]string) (Field, error)
	references(tag string, tags map[string]string, rel Scope) (Field, error)
	polymorphic(relScope Scope, tags map[string]string, fk *Field) (Polymorphic, error)

	// helper
	InitRelationByField(field string, setParent bool) (Interface, error)
	InitRelation(relation Interface, field string) error
	SetBackReference(Relation) error
	NewScopeFromType(reflect.Type) (Scope, error)

	// experimental
	Config(...string) config
	SetConfig(*config, ...string)
	SoftDelete() *SoftDelete
	Parent(name string) (*Model, error) // needed for back reference
	SetParent(model *Model)             // needed to avoid loops when called from outside of the orm package (grid).
	IsEmpty(Permission) bool            // needed in create
	IsSelfReferenceLoop(relation Relation) bool
	IsSelfReferencing(relation Relation) bool

	TakeSnapshot(bool)
	Snapshot() Interface

	ChangedValues(withoutUpdatedAt bool) []ChangedValue
	AppendChangedValue(c ChangedValue)
	SetChangedValues(c []ChangedValue)
	ChangedValueByFieldName(field string) *ChangedValue
}

// scope struct.
type scope struct {
	model *Model
}

// Config will return the defined orm model configuration.
// If no name is given, the scopes root configuration will be taken.
func (s scope) Config(name ...string) config {
	n := RootStruct
	if len(name) > 0 {
		n = name[0]
	}
	return s.model.config[n]
}

// SetConfig can be used to customize the relation or root orm model configuration.
// If no name is given, the scopes root will be set.
func (s scope) SetConfig(c *config, name ...string) {
	if len(name) > 0 {
		s.model.config[name[0]] = *c
		return
	}
	s.model.config[RootStruct] = *c
}

// Model will return the scopes orm model.
func (s scope) Model() *Model {
	return s.model
}

// TakeSnapshot will define if a snapshot of the orm model will be taken.
// This is used mainly in update.
func (s scope) TakeSnapshot(snapshot bool) {
	s.model.snapshot = snapshot
	return
}

// Snapshot will be returned if taken.
func (s scope) Snapshot() Interface {
	return s.model.snapshotCaller
}

// IsSelfReferenceLoop checks if the model has a self reference loop.
// Animal (hasOne) -> Address (belongsTo) -> *Animal
func (s scope) IsSelfReferenceLoop(relation Relation) bool {
	if relation.Kind == BelongsTo && relation.Type.Kind() == reflect.Ptr {
		_, err := s.Parent(relation.Type.String())
		if err == nil {
			return true
		}
	}
	return false
}

func (s *scope) IsSelfReferencing(relation Relation) bool {
	return s.Model().isSelfReferencing(relation.Type)
}

// IsEmpty checks if all the orm model fields and relations are empty.
func (s scope) IsEmpty(perm Permission) bool {
	for _, field := range s.SQLFields(perm) {
		if !s.FieldValue(field.Name).IsZero() {
			if field.Name == CreatedAt || field.Name == UpdatedAt || field.Name == DeletedAt {
				continue
			}
			return false
		}
	}

	for _, rel := range s.SQLRelations(perm) {

		relField := s.FieldValue(rel.Field)
		// for slice and ptr.
		if relField.IsZero() ||
			(relField.Type().Kind() == reflect.Slice && relField.Len() == 0) {
			continue
		}

		// check if the orm.model struct isEmpty (recursive).
		if relField.Type().Kind() == reflect.Struct || relField.Type().Kind() == reflect.Ptr {
			var relOrm Interface
			if relField.Type().Kind() == reflect.Struct {
				relOrm = relField.Addr().Interface().(Interface)
			} else {
				// ptr
				relOrm = relField.Interface().(Interface)
			}
			_ = s.InitRelation(relOrm, rel.Field)
			if !relOrm.model().scope.IsEmpty(perm) {
				return false
			}
			relOrm = nil
		}

	}
	return true
}

// Caller will return the orm model caller.
func (s scope) Caller() Interface {
	return s.model.caller
}

// NewScopeFromType will return a new scope of the given type.
func (s scope) NewScopeFromType(p reflect.Type) (Scope, error) {
	v := newValueInstanceFromType(p)
	model := v.Addr().Interface().(Interface)

	// init the orm instance.
	err := model.Init(model)
	if err != nil {
		return nil, err
	}

	// copy fields/relation permission from parent, if its of the same type.
	if s.Name(true) == strings.Replace(p.String(), "*", "", -1) {
		copy(model.model().fields, s.model.fields)
		copy(model.model().relations, s.model.relations)
	}

	// add loop detection
	s.copyLoopDetection(model)

	return model.Scope()
}

// SoftDelete will return the soft deleting struct.
func (s *scope) SoftDelete() *SoftDelete {
	return s.model.softDelete
}

// Builder will return the model builder.
func (s *scope) Builder() query.Builder {
	return s.model.builder
}

// Cache will return the callers cache.
// TODO at the moment not in use because the logic changed to DefaultCache. Delete?
func (s *scope) Cache() cache.Manager {
	return s.model.cache
}

// SetCache allows to manipulate the default cache.
// TODO at the moment not in use because the logic changed to DefaultCache. Delete?
func (s *scope) SetCache(cache cache.Manager) error {
	if cache == nil {
		return fmt.Errorf(ErrMandatory, "cache", s.Name(true))
	}
	s.model.cache = cache
	return nil
}

// Name will return the orm caller name with or without package prefix.
func (s scope) Name(ns bool) string {
	name := s.model.name
	if s.model.name == "" {
		name = reflectName(s.model.caller)
	}
	if idx := strings.Index(name, "."); !ns && idx != -1 {
		return strings.Title(name[idx+1:])
	}
	s.model.name = name
	return name
}

// SQLRelations will return all sql relations by the given Permission.
// Relation(s) which are defined as "custom" or have not the required Permission will not be returned.
func (s scope) SQLRelations(p Permission) []Relation {
	var rv []Relation
	for _, rel := range s.Relations(p) {
		if rel.NoSQLColumn || (p.Read && !rel.Permission.Read) || (p.Write && !rel.Permission.Write) {
			continue
		}
		rv = append(rv, rel)
	}
	return rv
}

func (s scope) Relations(p Permission) []Relation {
	var rv []Relation
	for _, rel := range s.model.relations {
		if (p.Read && !rel.Permission.Read) || (p.Write && !rel.Permission.Write) {
			continue
		}
		rv = append(rv, rel)
	}
	return rv
}

// SQLRelation will return the requested relation by permission.
// Relations(s) which are defined as "custom" or have not the required Permission will not be returned.
// Error will return if the relation does not exist or has not the required permission.
func (s scope) SQLRelation(relation string, p Permission) (Relation, error) {
	for _, rel := range s.SQLRelations(p) {
		if rel.Field == relation {
			return rel, nil
		}
	}
	return Relation{}, fmt.Errorf(ErrFieldName, relation)
}

// SQLColumns will return all struct fields by permission as slice string.
func (s scope) SQLColumns(p Permission) []string {
	fields := s.SQLFields(p)
	cols := make([]string, len(fields))
	for i, field := range fields {
		cols[i] = field.Information.Name
	}
	return cols
}

// SQLScanFields is a helper for row.scan.
// It will scan the struct fields by the given permission.
func (s scope) SQLScanFields(p Permission) []interface{} {
	fields := s.SQLFields(p)
	cols := make([]interface{}, len(fields))
	for i, field := range fields {
		cols[i] = s.FieldValue(field.Name).Addr().Interface()
	}
	return cols
}

// SQLFields will return all sql column fields by the given Permission.
// Field(s) which are defined as "custom" or have not the required Permission will not be returned.
func (s scope) SQLFields(p Permission) []Field {
	var rv []Field
	for _, field := range s.Fields(p) {
		// skip if no real sql field or permission is not permitted.
		if field.NoSQLColumn {
			continue
		}
		rv = append(rv, field)
	}
	return rv
}

func (s scope) Fields(p Permission) []Field {
	var rv []Field
	for _, field := range s.model.fields {
		// skip if permission is not permitted.
		if (p.Read && !field.Permission.Read) || (p.Write && !field.Permission.Write) {
			continue
		}
		rv = append(rv, field)
	}
	return rv
}

// Field returns a ptr to the struct field by name.
// Error will return if the field does not exist.
func (s scope) Field(name string) (*Field, error) {
	for i, field := range s.model.fields {
		if field.Name == name {
			return &s.model.fields[i], nil
		}
	}
	return nil, fmt.Errorf(ErrFieldName, s.FqdnModel(name))
}

// FieldValue returns a reflect.Value of the orm caller struct field.
// It returns the zero Value if no field was found.
func (s scope) FieldValue(name string) reflect.Value {
	return reflect.ValueOf(s.model.caller).Elem().FieldByName(name)
}

// PrimaryKeys will return all defined primary key fields.
// Error will return if no primary key was found.
func (s scope) PrimaryKeys() ([]Field, error) {
	var rv []Field
	for _, f := range s.SQLFields(Permission{}) {
		if f.Information.PrimaryKey {
			rv = append(rv, f)
		}
	}
	if len(rv) == 0 {
		return nil, fmt.Errorf(ErrPrimaryKey, s.Name(true))
	}
	return rv, nil
}

// PrimaryKeysSet checks if all primaries have a non zero value.
func (s scope) PrimaryKeysSet() bool {
	for _, field := range s.model.fields {
		if field.Information.PrimaryKey {
			if s.FieldValue(field.Name).IsZero() {
				return false
			}
		}
	}
	return true
}

func (s scope) SetParent(m *Model) {
	s.model.parentModel = m
}

// Parent returns the parent model by name or the root model if the name is empty.
// The name must be the orm struct name incl. namespace.
// Error will return if no parent exists or the given name does not exist.
// The max search depth is limited to 20.
func (s scope) Parent(name string) (*Model, error) {

	// removing ptr or slice from name.
	name = strings.Replace(strings.Replace(name, "*", "", -1), "[]", "", -1)

	p := s.model.parentModel
	i := 0
	for p != nil {
		if i > maxSearchDepth {
			return nil, fmt.Errorf(ErrMaxSearchDepth, maxSearchDepth, name)
		}
		// return root parent
		if name == "" && p.parentModel == nil {
			return p, nil
		}
		// return named parent
		if p.name == name {
			return p, nil
		}
		p = p.parentModel
		i++
	}
	return nil, fmt.Errorf("was not found %s", name)
}

// FqdnTable is a helper to display the models database and table name.
func (s scope) FqdnTable() string {
	return s.model.db + "." + s.model.table
}

// FqdnModel is a helper to display the model name and the field name.
func (s scope) FqdnModel(field string) string {
	return s.Name(true) + ":" + field
}

// InitRelation initialize the given relation.
// The orm model parent will be set, config, permission list and tx will be passed.
func (s scope) InitRelation(relation Interface, field string) error {

	// initialize model or only replace the caller if its already initialized.
	if err := relation.model().isInit(); err != nil {
		err := relation.Init(relation)
		if err != nil {
			return err
		}
	} else {
		relation.model().caller = relation                      //needed
		relation.model().scope = scope{model: relation.model()} // needed
	}

	// if no string is given, its no relations - its the root element which should have no parent.
	if field != "" {
		relation.model().parentModel = s.Model() // set the correct parent
	}

	// pass custom configuration, if exist
	s.model.scope.passConfig(field, relation)

	// passing parent fields
	s.addParentPermissionsList(relation, field)

	// copy the loopDetection, map is referenced by. so all the changes would also be in the parent models.
	s.copyLoopDetection(relation)

	// TODO better solution - this is breaking on different providers.
	// TODO mode addAutoTx and commitAutoTx must be rewritten.
	// add tx if the parent scope has one and its the same builder
	relation.model().tx = s.model.tx

	return nil
}

// InitRelationByField will return the orm.Interface of the field.
// * = if the value was nil, a new orm.Interface gets set, if its not nil, the value will be taken.
// struct = * of that struct
// *[], [] = new orm.Interface
func (s scope) InitRelationByField(field string, setParent bool) (Interface, error) {
	rvRel := s.FieldValue(field)

	var relInterface Interface
	switch rvRel.Type().Kind() {
	case reflect.Ptr:
		if rvRel.Type().Elem().Kind() == reflect.Slice {
			relInterface = newValueInstanceFromType(rvRel.Type().Elem().Elem()).Addr().Interface().(Interface)
		} else {
			if reflect.ValueOf(rvRel.Interface().(Interface)).IsNil() {
				rvRel.Set(newValueInstanceFromType(rvRel.Type()).Addr())
			}
			relInterface = rvRel.Interface().(Interface)
		}
	case reflect.Struct:
		relInterface = rvRel.Addr().Interface().(Interface)
	case reflect.Slice:
		relInterface = newValueInstanceFromType(rvRel.Type()).Addr().Interface().(Interface)
	}

	err := s.InitRelation(relInterface, field)
	if err != nil {
		return nil, err
	}

	return relInterface, nil
}

// SetBackReference will set a backreference if detected.
func (s scope) SetBackReference(relation Relation) error {
	c, err := s.Parent(relation.Type.String())
	if err != nil {
		return err
	}
	f := s.FieldValue(relation.Field)
	return SetReflectValue(f, reflect.ValueOf(c.caller))
}

// SetReflectValue is a helper to set the fields value.
// It checks the field type and casts the value to it.
func SetReflectValue(field reflect.Value, value reflect.Value) error {

	switch field.Kind() {
	case reflect.Ptr:
		if field.Type().Elem().Kind() == reflect.Slice {
			// create a zero value if its nil.
			if field.IsZero() {
				field.Set(reflect.New(reflect.MakeSlice(reflect.SliceOf(field.Type().Elem().Elem()), 0, 0).Type()))
			}
			if field.Type().Elem().Elem().Kind() == reflect.Ptr {
				field.Elem().Set(reflect.Append(field.Elem(), reflect.Indirect(value).Addr()))
			} else {
				field.Elem().Set(reflect.Append(field.Elem(), value))
			}
		} else {
			if value.CanAddr() {
				field.Set(value.Addr())
			} else {
				field.Set(value)
			}
		}
	case reflect.Struct:
		scannerI := reflect.TypeOf((*sql.Scanner)(nil)).Elem()
		if field.Addr().Type().Implements(scannerI) && field.Type() != value.Type() {
			f := field.Addr().Interface().(sql.Scanner)
			err := f.Scan(value.Interface())
			if err != nil {
				return err
			}

		} else {
			field.Set(value)
		}
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.Ptr {
			field.Set(reflect.Append(field, reflect.Indirect(value).Addr()))
		} else {
			field.Set(reflect.Append(field, value))
		}
	default:
		if field.Type() != value.Type() {
			if v, ok := value.Interface().(driver.Valuer); ok {
				vv, err := v.Value()
				if err != nil {
					return err
				}
				value = reflect.ValueOf(vv)
			}
		}
		// int mapping, create a better solution
		if field.Kind() == reflect.Int && value.Kind() == reflect.String {
			v, err := strconv.Atoi(value.Interface().(string))
			if err != nil {
				return err
			}
			value = reflect.ValueOf(v)
		}
		// int mapping, create a better solution
		if field.Kind() == reflect.Int && value.Kind() == reflect.Int64 {
			value = reflect.ValueOf(int(value.Interface().(int64)))
		}
		if field.Kind() == reflect.Int64 && value.Kind() == reflect.Int {
			value = reflect.ValueOf(value.Int())
		}

		field.Set(value)
	}
	return nil
}

// IsValueZero checks if a reflect.Value is zero or if its a ptr if the value is zero.
func IsValueZero(value reflect.Value) bool {
	if value.IsZero() || (value.Kind() == reflect.Ptr && value.Elem().IsZero()) {
		return true
	}

	return false
}

// reflectName will return the callers struct name.
// If the caller was not defined yet, the file and line number will return as debug information.
func reflectName(p Interface) string {
	// if the caller was not defined yet, return some debug information.
	if p == nil {
		var _, file, line, _ = runtime.Caller(2)
		return file + ":" + strconv.Itoa(line)
	}
	// return calle name
	return reflect.ValueOf(p).Type().Elem().String()
}

// implementsInterface is a helper to identify if the given reflect.Value implements the orm.Interface.
func implementsInterface(value reflect.Value) bool {
	i := reflect.TypeOf((*Interface)(nil)).Elem()
	v := newValueInstanceFromType(value.Type())
	return v.Addr().Type().Implements(i)
}

// implementsScannerValuer is a helper to identify if the given reflect.Value implements the sql.Scanner and driver.Valuer interface.
func implementsScannerValuer(value reflect.Value) bool {
	v := newValueInstanceFromType(value.Type())
	if v.Addr().Type().Implements(reflect.TypeOf((*sql.Scanner)(nil)).Elem()) && v.Addr().Type().Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) {
		return true
	}
	return false
}

// hasCustomTag is a helper to identify if the struct, slice or ptr was defined with a custom tag.
func hasCustomTag(field reflect.StructField) bool {
	tag := field.Tag.Get(TagKey)
	tags := structer.ParseTag(tag)

	if _, ok := tags[tagNoSQLField]; ok {
		switch field.Type.Kind() {
		case reflect.Struct, reflect.Slice, reflect.Ptr:
			return true
		}
	}

	return false
}

// addParentPermissionsList passes the permission list from the parent to the child orm model, if a dot notation exists.
// If the field is empty the root permission list is copied. This for example needed by new slice orm models.
// For example Animal.ID: the field ID will be added to the child animal orm model.
func (s scope) addParentPermissionsList(relation Interface, field string) {

	// zero value
	if s.model.permissionList == nil {
		return
	}

	// on self reference, add the same wb list of parent
	// also used for snapshot
	if s.model.name == relation.model().name && !s.model.config[RootStruct].permissionsExplicit {
		relation.SetPermissions(s.model.permissionList.policy, s.model.permissionList.fields...)
		return
	}

	// otherwise add only if the relation is defined
	var fields []string
	for _, a := range s.model.permissionList.fields {
		if strings.HasPrefix(a, field+".") {
			fields = append(fields, strings.Replace(a, field+".", "", 1))
		}
	}

	if len(fields) > 0 {
		relation.SetPermissions(s.model.permissionList.policy, fields...)
	}
}

// copyLoopDetection will copy the scope loopDetection to the relation.
// This is needed to identify loops over relations.
func (s scope) copyLoopDetection(relation Interface) {
	relation.model().loopDetection = make(map[string][]string, len(s.model.loopDetection))
	for index := range s.model.loopDetection {
		relation.model().loopDetection[index] = make([]string, len(s.model.loopDetection[index]))
		copy(relation.model().loopDetection[index], s.model.loopDetection[index])
	}
}

// passConfig is a helper to set the relation configuration - if defined.
func (s scope) passConfig(name string, relation Interface) {
	// add root config
	if c, ok := s.model.config[name]; ok {
		relation.model().scope.SetConfig(&c)
	}
	// add child configs
	for n, c := range s.model.config {
		if strings.HasPrefix(n, name+".") {
			relation.model().scope.SetConfig(&c, strings.Replace(n, name+".", "", -1))
		}
	}
}

// checkLoopMap is checking if the relation model was already asked before with the same where condition.
// Error will return if the same condition was already asked before.
func (s scope) checkLoopMap(c condition.Condition) error {

	// reset loop detection on root struct
	if s.model.parentModel == nil {
		s.model.loopDetection = nil
	}

	whereString := ""
	for _, where := range c.Where() {
		whereString += fmt.Sprint(where.Condition(), where.Arguments())
	}

	rel := s.Name(true)
	counter := 0
	for _, b := range s.model.loopDetection[rel] {
		if b == whereString {
			counter++
		}
		if counter >= 1 {
			return fmt.Errorf(ErrInfinityLoop, rel)
		}
	}
	if s.model.loopDetection == nil {
		s.model.loopDetection = map[string][]string{}
	}
	s.model.loopDetection[rel] = append(s.model.loopDetection[rel], whereString)

	return nil
}

// newValueInstanceFromType is a helper to creates a new reflect.Value of the given type.
// It ensures that the returned reflect.Value is the value a ptr or slice points to.
// []type 	= type as value
// []*type 	= type as value
// *type 	= type as value
// *[]type 	= type as value
// *[]*type = type as value
// type 	= type as value
func newValueInstanceFromType(field reflect.Type) reflect.Value {

	// convert slice to single element
	var v reflect.Value
	if field.Kind() == reflect.Slice {
		//handles ptr
		if field.Elem().Kind() == reflect.Ptr {
			v = reflect.New(field.Elem().Elem())
		} else {
			v = reflect.New(field.Elem())
		}
	} else {
		if field.Kind() == reflect.Ptr {
			if field.Elem().Kind() == reflect.Slice { // if its a *[]
				v = newValueInstanceFromType(field.Elem())
			} else {
				v = reflect.New(field.Elem())
			}
		} else {
			v = reflect.New(field)
		}
	}

	// convert from ptr to value
	return reflect.Indirect(v)
}

// parseStruct is a helper to collect all fields and relations of the orm model.
// Only exported fields will be parsed. Embedded structs will be rendered as well.
// It checks if the field should get skipped by tag (-).
// It checks if the given Field(s) and Relation(s) are unique by name.
// Field will be counted if it's a normal go type or it implements the sql.Scanner or driver.Valuer.
// Relation will only be counted if it implements the orm.Interface or is defined as "custom" by tag.
func (s scope) parseStruct(_struct interface{}) ([]reflect.StructField, []reflect.StructField, error) {
	var fields []reflect.StructField
	var relations []reflect.StructField
	var err error

	v := reflect.Indirect(reflect.ValueOf(_struct))
	if v.IsValid() {
		for i := 0; i < v.Type().NumField(); i++ {
			field := v.Type().Field(i)

			// skip if unexported or skip tag is defined.
			if field.PkgPath != "" || field.Tag.Get(TagKey) == tagSkip {
				continue
			}

			// embedded struct
			if field.Anonymous {
				embeddedFields, embeddedRelations, err := s.parseStruct(v.FieldByName(field.Name).Interface())
				if err != nil {
					return nil, nil, err
				}
				fields, err = s.addField(fields, embeddedFields)
				if err != nil {
					return nil, nil, err
				}
				relations, err = s.addField(relations, embeddedRelations)
				if err != nil {
					return nil, nil, err
				}
				continue
			}

			// checks if it's a orm.Interface or custom struct/slice/ptr.
			if implementsInterface(v.FieldByName(field.Name)) || hasCustomTag(field) {
				relations, err = s.addField(relations, []reflect.StructField{field})
				if err != nil {
					return nil, nil, err
				}
				continue
			}

			// check if the field type is allowed.
			err = s.allowedFieldType(v.FieldByName(field.Name))
			if err != nil {
				return nil, nil, err
			}

			// add field to model.
			fields, err = s.addField(fields, []reflect.StructField{field})
			if err != nil {
				return nil, nil, err
			}
		}

	}

	// sort fields that the time fields (createdAt, updatedAt, deletedAt) are at the end.
	var timeFields []reflect.StructField
	for i := 0; i < len(fields); i++ {
		if fields[i].Name == CreatedAt || fields[i].Name == UpdatedAt || fields[i].Name == DeletedAt {
			timeFields = append(timeFields, fields[i])
			fields = append(fields[:i], fields[i+1:]...)
			i--
		}
	}

	return append(fields, timeFields...), relations, nil
}

// allowedFieldType checks if the struct field has a normal type.
// (string, bool, uint, int and float) are allowed or any type that implements the sql.Scanner and driver.Valuer interface.
// Error will return otherwise.
func (s scope) allowedFieldType(field reflect.Value) error {
	rv := newValueInstanceFromType(field.Type())

	switch rv.Kind() {
	case reflect.Struct:
		if implementsScannerValuer(rv) {
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return nil
	case reflect.String:
		return nil
	case reflect.Float32, reflect.Float64:
		return nil
	case reflect.Bool:
		return nil
	}

	return fmt.Errorf(ErrFieldType, rv.Type(), s.Name(true))
}

// addField is a helper to only add fields if they dont exist yet.
// Error will return if field name already exists, except CreatedAt, UpdatedAt, DeletedAt - they will be skipped (reason: embedded fields can have the orm.Model as well)
func (s scope) addField(a []reflect.StructField, b []reflect.StructField) ([]reflect.StructField, error) {
	for _, fieldB := range b {
		var exist bool
		for _, fieldA := range a {
			if fieldA.Name == fieldB.Name {
				switch fieldA.Name {
				case CreatedAt, UpdatedAt, DeletedAt:
					exist = true
				default:
					return nil, fmt.Errorf(ErrFieldUnique, s.FqdnModel(fieldA.Name))
				}
				exist = true
			}
		}
		if !exist {
			a = append(a, fieldB)
		}
	}
	return a, nil
}

// polymorphic will return a Polymorphic struct if its defined by tag.
// If only the poly tag is defined without any value by default the values will be set:
// - the ForeignKey will be set as the relation struct name + ID
// - the TypeField is relation struct name + Type.
// - the Value will be the model struct name.
// If the poly tag is set with a value, the IDField, TypeField will be the tag {value}+ID, {value}+Type.
// If the poly_value tag is set, the Polymorphic value can be customized.
// Error will return if the field does not exists in the struct.
func (s scope) polymorphic(relScope Scope, tags map[string]string, fk *Field) (Polymorphic, error) {

	if v, ok := tags[tagPolymorphic]; ok {
		name := relScope.Name(false)
		if v != "" {
			name = v
		}

		// check if field exists.
		var fieldID *Field
		var err error
		if v, ok := tags[tagReferences]; ok {
			fieldID, err = relScope.Field(v)
		} else {
			fieldID, err = relScope.Field(name + "ID")
		}
		if err != nil {
			return Polymorphic{}, err
		}
		*fk = *fieldID

		// check if field exists.
		fieldType, err := relScope.Field(name + "Type")
		if err != nil {
			return Polymorphic{}, err
		}
		fieldValue := s.Name(false)
		if v, ok = tags[tagPolymorphicValue]; ok {
			fieldValue = v
		}
		return Polymorphic{TypeField: *fieldType, Value: fieldValue}, nil
	}
	return Polymorphic{}, nil
}

// references will return by default {model name}+ID or it's set by tag.
// Error will return if the field name does not exist in the relation model.
func (s scope) references(tag string, tags map[string]string, rel Scope) (Field, error) {

	name := s.Name(false) + "ID"
	if v, ok := tags[tag]; ok {
		name = v
	}

	// check if field exists.
	field, err := rel.Field(name)
	if err != nil {
		return Field{}, err
	}

	return *field, err
}

// foreignKey will return by default the first primary key of the model or it's set by tag.
// Error will return if the field does not exist or no primary key is defined.
func (s scope) foreignKey(tag string, tags map[string]string) (Field, error) {
	if v, ok := tags[tag]; ok {
		field, err := s.Field(v)
		if err != nil {
			return Field{}, err
		}
		return *field, nil
	}

	// if fk was not set by tag
	pkeys, err := s.PrimaryKeys()
	if err != nil {
		return Field{}, err
	}
	return pkeys[0], nil
}
