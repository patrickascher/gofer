// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"context"
	"fmt"
	"reflect"
	"time"

	valid "github.com/go-playground/validator/v10"
	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/slicer"
)

// predefined constants.
const (
	prefixCache = "orm_"
	MODEL       = "orm.model" // MODEL is the ctx key for validator.
	ID          = "ID"        // ID is the field name for primary keys.
	RootStruct  = ""          // RootStruct is used for the configuration to identify the root Struct.
)

// Error messages
var (
	ErrInit      = "orm: forgot to call Init() on %s"
	ErrMandatory = "orm: %s is mandatory but has zero-value in %s"
	ErrResultPtr = "orm: result variable must be a ptr in %s (All)"
)

// Interface of the orm model.
type Interface interface {
	Init(Interface) error
	Scope() (Scope, error)

	First(c ...condition.Condition) error
	All(result interface{}, c ...condition.Condition) error
	Count(c ...condition.Condition) (int, error)
	Create() error
	Update() error
	Delete() error

	// Permissions
	Permissions() (p int, fields []string)
	SetPermissions(p int, fields ...string)

	// Defaults
	DefaultBuilder() query.Builder
	DefaultTableName() string
	DefaultDatabaseName() string
	DefaultSoftDelete() SoftDelete
	DefaultCache() (cache.Manager, time.Duration)
	DefaultStrategy() string

	// internal
	setParent(*Model)
	model() *Model
}

// Model struct has to get embedded to enable the orm features.
type Model struct {
	// allows to access the root struct from relations.
	parentModel *Model

	// struct information.
	name   string
	db     string
	table  string
	caller Interface

	// scope options.
	scope scope

	// defined strategy.
	strategy Strategy

	// defined fields and relations.
	fields    []Field
	relations []Relation

	// builder and tx.
	// autoTx will be set on root level if there is no tx defined yet.
	// TODO: at the moment its not possible to define a tx as user but orm is already build for it.
	builder query.Builder
	tx      query.Tx
	autoTx  bool

	// cache settings for the struct.
	cache      cache.Manager
	cacheTTL   time.Duration
	softDelete *SoftDelete

	permissionList *permissionList
	snapshot       bool
	changedValues  []ChangedValue

	config        map[string]config
	loopDetection map[string][]string

	TimeFields
}

// PreInit is a function to pre-initialize models.
// This can be used on application start to boost the performance.
func PreInit(orm Interface) error {
	return orm.Init(orm)
}

// Count the existing rows by the given condition.
// TODO move the logic to provider for none db orm in the future?
func (m *Model) Count(c ...condition.Condition) (int, error) {
	// check if model is init.
	if err := m.isInit(); err != nil {
		return 0, err
	}

	// create sql condition
	var cond condition.Condition
	if len(c) == 0 {
		cond = condition.New()
	} else {
		cond = c[0]
	}

	// create query
	row, err := m.builder.Query(m.tx).Select(m.scope.FqdnTable()).Condition(cond).Columns(query.DbExpr("COUNT(*)")).First()
	if err != nil {
		return 0, err
	}

	var count int
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// First will return the first found row.
// The condition is optional, if set the first argument will be used.
// A sql.ErrNoRows will return if no result was found.
func (m *Model) First(c ...condition.Condition) error {

	// check if model is init.
	if err := m.isInit(); err != nil {
		return err
	}

	// TODO Callbacks before

	// create sql condition
	var cond condition.Condition
	if len(c) == 0 {
		cond = condition.New()
	} else {
		cond = c[0]
	}

	err := m.scope.setFieldPermission()
	if err != nil {
		return err
	}

	err = m.scope.checkLoopMap(cond)
	if err != nil {
		return err
	}

	err = m.strategy.First(&m.scope, cond, Permission{Read: true})
	if err != nil {
		return err
	}

	// TODO Callbacks after

	return nil
}

// All will return all results found by the condition.
// The result argument must be as ptr slice to the struct.
// The condition is optional, if set the first argument will be used.
// No error will return if no result was found.
func (m *Model) All(result interface{}, c ...condition.Condition) error {

	// check if model is init.
	if err := m.isInit(); err != nil {
		return err
	}

	// checking if the res is a ptr to an Interface slice.
	if result == nil || reflect.TypeOf(result).Kind() != reflect.Ptr ||
		(reflect.TypeOf(result).Elem().Kind() != reflect.Slice && reflect.TypeOf(result).Elem().Kind() != reflect.Ptr) ||
		!implementsInterface(reflect.ValueOf(result).Elem()) {
		return fmt.Errorf(ErrResultPtr, m.scope.Name(true))
	}

	// TODO Callbacks before

	// create sql condition
	var cond condition.Condition
	if len(c) == 0 {
		cond = condition.New()
	} else {
		cond = c[0]
	}

	err := m.scope.setFieldPermission()
	if err != nil {
		return err
	}

	err = m.scope.checkLoopMap(cond)
	if err != nil {
		return err
	}

	err = m.strategy.All(result, &m.scope, cond)
	if err != nil {
		return err
	}

	// TODO Callbacks after

	return nil
}

// Create the given orm model.
// A transaction will be created in the background for all relations and a rollback will be triggered if an error happens.
// The orm model will be checked if its valid by tags.
// TODO tx on different database drivers.
func (m *Model) Create() (err error) {
	defer func() { modelDefer(m, err) }()

	// check if model is init.
	if err = m.isInit(); err != nil {
		return
	}

	// TODO callback before

	// set the CreatedAt info if exists
	// it only gets saved if the field exists in the db (permission is set)
	m.CreatedAt = query.NewNullTime(time.Now(), true)

	// if the model is empty no need for creating.
	if m.scope.IsEmpty(Permission{Write: true}) {
		return nil
	}

	// setFieldPermission must be called before isValid.
	err = m.scope.setFieldPermission()
	if err != nil {
		return
	}

	err = m.IsValid()
	if err != nil {
		return
	}

	err = m.addAutoTx()
	if err != nil {
		return
	}

	err = m.strategy.Create(&m.scope)
	if err != nil {
		return
	}

	err = m.commitAutoTx()
	if err != nil {
		return
	}

	// TODO callback after
	return
}

// Update the given orm model.
// A transaction will be created in the background for all relations and a rollback will be triggered if an error happens.
// The orm model will be checked if its valid by tags.
// A snapshot is taken and only changed values will trigger a sql query.
// TODO tx on different database drivers.
func (m *Model) Update() (err error) {
	defer func() { modelDefer(m, err) }()

	// check if model is init.
	if err := m.isInit(); err != nil {
		return err
	}

	// check primary keys
	if !m.scope.PrimaryKeysSet() {
		err = fmt.Errorf("PKEY Err for %s", m.name)
		return
	}

	// must be called before isValid
	err = m.scope.setFieldPermission()
	if err != nil {
		return
	}

	err = m.IsValid()
	if err != nil {
		return
	}

	// create where condition
	pKeys, err := m.scope.PrimaryKeys()
	if err != nil {
		return
	}
	c := condition.New()
	for _, pkey := range pKeys {
		c.SetWhere(m.scope.Builder().QuoteIdentifier(pkey.Information.Name)+" = ?", m.scope.FieldValue(pkey.Name).Interface())
	}

	// TODO callback before

	// call delete on strategy
	err = m.addAutoTx()
	if err != nil {
		return
	}

	if m.parentModel == nil {
		m.scope.TakeSnapshot(true)
	}

	// snapshot
	if m.snapshot {
		// reset condition loop
		m.loopDetection = nil

		// init snapshot
		snapshot := newValueInstanceFromType(reflect.TypeOf(m.caller)).Addr().Interface().(Interface)
		err = m.scope.InitRelation(snapshot, "")
		if err != nil {
			return
		}

		err = m.strategy.First(&snapshot.model().scope, c, Permission{Write: true})
		if err != nil {
			return
		}

		m.changedValues, err = m.scope.EqualWith(snapshot)

		if err != nil {
			return
		}
	}

	// if no data was changed
	if m.changedValues == nil {
		// needed to avoid a not closed tx.
		return m.commitAutoTx()
	}

	// set the UpdatedAt info if exists
	// it only gets saved if the field exists in the db (permission is set)
	m.UpdatedAt = query.NewNullTime(time.Now(), true)

	err = m.strategy.Update(&m.scope, c)
	if err != nil {
		return
	}

	err = m.commitAutoTx()
	if err != nil {
		return
	}
	// TODO callback after

	return nil
}

// Delete the orm model by its primary keys.
// A transaction will be created in the background for all relations and a rollback will be triggered if an error happens.
func (m *Model) Delete() (err error) {
	defer func() { modelDefer(m, err) }()

	// check if model is init.
	if err := m.isInit(); err != nil {
		return err
	}

	// check primary keys
	if !m.scope.PrimaryKeysSet() {
		err = fmt.Errorf("PKEY Err for %s", m.name)
		return
	}

	err = m.scope.setFieldPermission()
	if err != nil {
		return
	}

	// create where condition
	pKeys, err := m.scope.PrimaryKeys()
	if err != nil {
		return
	}
	c := condition.New()
	for _, pkey := range pKeys {
		c.SetWhere(m.scope.Builder().QuoteIdentifier(pkey.Information.Name)+" = ?", m.scope.FieldValue(pkey.Name).Interface())
	}

	// check if its a soft delete
	if m.softDelete != nil {
		_, err = m.scope.Builder().Query(m.tx).Update(m.scope.FqdnTable()).Columns(m.softDelete.Field).Set(map[string]interface{}{m.softDelete.Field: m.softDelete.Value}).Condition(c).Exec()
		return err
	}

	// TODO callback before

	// call delete on strategy
	err = m.addAutoTx()
	if err != nil {
		return
	}

	err = m.strategy.Delete(&m.scope, c)
	if err != nil {
		return
	}

	err = m.commitAutoTx()
	if err != nil {
		return
	}

	// TODO callback after

	return nil
}

// Init the orm mode.
func (m *Model) Init(caller Interface) error {
	// set caller
	m.caller = caller

	// check if a cache is set.
	if m.cache, m.cacheTTL = m.caller.DefaultCache(); m.cache == nil {
		return fmt.Errorf(ErrMandatory, "cache", reflectName(caller))
	}

	// set scope
	m.scope.model = m

	// check if a cache exists
	if m.cache.Exist(prefixCache, m.scope.Name(true)) {
		item, err := m.cache.Get(prefixCache, m.scope.Name(true))
		if err != nil {
			return err
		}
		*m = item.Value().(Model)
		m.caller = caller
		m.parentModel = nil       // needed - otherwise parent will be set and will cause problems if the child model is shared between more models.
		m.scope = scope{model: m} // needed
	} else {
		// setting the default values, builder, table name, database name.
		if m.builder = m.caller.DefaultBuilder(); m.builder == nil {
			return fmt.Errorf(ErrMandatory, "builder", m.scope.Name(true))
		}
		if m.db = m.caller.DefaultDatabaseName(); m.db == "" {
			return fmt.Errorf(ErrMandatory, "db-name", m.scope.Name(true))
		}
		if m.table = m.caller.DefaultTableName(); m.table == "" {
			return fmt.Errorf(ErrMandatory, "table-name", m.scope.Name(true))
		}

		// parse complete struct to fetch all fields and relations.
		fields, relations, err := m.scope.parseStruct(m.caller)
		if err != nil {
			return err
		}

		// create fields.
		err = m.createFields(fields)
		if err != nil {
			return err
		}

		// create relations.
		err = m.createRelations(relations)
		if err != nil {
			return err
		}

		// add db validations
		err = m.addDBValidation()
		if err != nil {
			return err
		}

		// add strategy
		m.strategy, err = strategy(m.DefaultStrategy())
		if err != nil {
			return fmt.Errorf("orm: %w", err)
		}

		// set default config
		m.scope.model.config = make(map[string]config, 1)
		m.scope.SetConfig(&config{allowHasOneZero: true, showDeletedRows: false})

		// set cache
		err = m.cache.Set(prefixCache, m.scope.Name(true), *m, m.cacheTTL)
		if err != nil {
			return fmt.Errorf("orm: %w", err)
		}
	}

	// fields and relations are copied so that no change will reflect the cache.
	m.copyFieldRelationSlices()

	return nil
}

// Scope will return the models scope with some helper functions.
// Error will return if the model was not initialized yet.
func (m *Model) Scope() (Scope, error) {
	if err := m.isInit(); err != nil {
		return nil, fmt.Errorf(ErrInit, reflectName(m.caller))
	}
	return &m.scope, nil
}

// SetPermissions allows to set a policy (White/Blacklist), Fields or Relations.
func (m *Model) SetPermissions(p int, fields ...string) {
	fields = slicer.StringUnique(fields)
	m.permissionList = newPermissionList(p, fields)
}

// Permissions returns the permission policy and defined fields.
func (m *Model) Permissions() (p int, fields []string) {
	if m.permissionList == nil {
		return WHITELIST, nil
	}
	return m.permissionList.policy, m.permissionList.fields
}

// IsValid checks if a custom validation was added and runs it.
// After that the struct will be validated by tag, if set.
func (m Model) IsValid() error {

	// custom validation
	for _, field := range m.scope.SQLFields(Permission{Write: true}) {
		if config := field.Validator.Config(); config != "" {
			err := errorMessage(m, field.Name, validate.VarCtx(newCtx(m), m.scope.FieldValue(field.Name).Interface(), config))
			if err != nil {
				return err
			}
		}
	}

	// struct tag validation
	// TODO: this will end in a loop on Animal - Address - *Animal backref.
	err := errorMessage(m, "", validate.StructCtx(newCtx(m), m.caller))
	if err != nil {
		return err
	}

	return nil
}

// isInit is a helper to identify if the orm.Model was already initialized.
func (m *Model) isInit() error {
	if m.scope.model == nil {
		return fmt.Errorf(ErrInit, reflectName(m.caller))
	}
	return nil
}

// modelDefer is a helper to check if an error happens on Create, Update or Delete and rollback the sql transaction.
func modelDefer(m *Model, err error) {
	// rollback happens already in query on a sql error that's why we have to exp
	if err != nil && m.isInit() == nil && m.autoTx && m.tx != nil && m.tx.HasTx() {
		// Rollback already happens in the query package on tx queries.
		// This is only here if the logic will change some day.
		rErr := m.tx.Rollback()
		m.tx = nil
		if rErr != nil {
			panic(rErr)
		}
	}
}

// addAutoTx is a helper to add a sql transaction if its the root orm model, and if there are relations and none tx was set manually before.
func (m *Model) addAutoTx() error {
	if m.parentModel == nil && m.tx == nil && len(m.relations) > 0 {
		var err error
		m.autoTx = true
		m.tx, err = m.builder.Query().Tx()
		if err != nil {
			return err
		}
	}
	return nil
}

// commitAutoTx is a helper to commit the sql transaction if its the root orm model.
func (m *Model) commitAutoTx() error {
	if m.autoTx && m.parentModel == nil {
		m.autoTx = false
		err := m.tx.Commit()
		m.tx = nil
		if err != nil {
			return err
		}
	}
	return nil
}

// setParent is a helper to set a parent.
func (m *Model) setParent(caller *Model) {
	m.parentModel = caller
}

// model will return the orm.Model.
func (m *Model) model() *Model {
	return m
}

// copyFieldRelationSlices is needed that the cached fields and relations of the orm model are not getting changed.
func (m *Model) copyFieldRelationSlices() {
	cFields := make([]Field, len(m.fields))
	copy(cFields, m.fields)
	m.fields = cFields

	cRelations := make([]Relation, len(m.relations))
	copy(cRelations, m.relations)
	m.relations = cRelations

	// set a new reference.
	cConfig := make(map[string]config, len(m.config))
	cConfig[RootStruct] = m.config[RootStruct]
	m.config = cConfig
}

// errorMessage is a helper to render the valid error messages.
// TODO at the moment only the first will return. change return all errors at once. This will be helpful for the frontend?
func errorMessage(m Model, field string, err error) error {
	if err != nil {
		for _, vErr := range err.(valid.ValidationErrors) {
			f := vErr.StructField()
			if f == "" {
				f = field
			}
			return fmt.Errorf(ErrValidation, m.scope.Name(true), f, vErr.ActualTag(), vErr.Value())
		}
	}
	return nil
}

// newCtx is a helper to create a new context with the key "orm.MODEL" and the orm.Interface as value.
func newCtx(m Model) context.Context {
	return context.WithValue(context.Background(), MODEL, m.caller)
}
