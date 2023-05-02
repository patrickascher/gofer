// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/patrickascher/gofer/locale/translation"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/grid/options"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/query/types"
	"github.com/patrickascher/gofer/slicer"
)

// Error messages.
var (
	ErrCallback    = "requested callback %s is not implemented"
	ErrRequestBody = "request body is empty in %s"
	ErrJSONInvalid = "json is invalid in %s"
	ErrConfig      = "no fields are configured"
	ErrConfigSrc   = fmt.Errorf("config the source over grid.Scope().Source() after the grid instance was created")
)

type gridSource struct {
	orm orm.Interface
}

// Orm converts an orm model to a grid source.
func Orm(orm orm.Interface) *gridSource {
	g := &gridSource{}
	g.orm = orm
	return g
}

// Init is called when the source is added to the grid.
// The orm gets initialized and the "update reference only" config is set.
// (needed for belongsTo,M2M because they are only select boxes with an id and not the full entry. This would be recognized as a change on the orm).
func (g *gridSource) Init(grid Grid) error {
	// check if the source got already init.
	_, err := g.orm.Scope()
	if err == nil {
		return ErrConfigSrc
	}

	// init orm
	err = g.orm.Init(g.orm)
	if err != nil {
		return err
	}
	scope, err := g.orm.Scope()
	if err != nil {
		return err
	}
	scope.SetConfig(orm.NewConfig().SetUpdateReferenceOnly(true))
	return nil
}

func (g *gridSource) PreInit(grid Grid) error {
	return nil
}

func (g *gridSource) Callback(cbk string, gr Grid) (interface{}, error) {

	switch cbk {
	case "select":
		selectField, err := gr.Scope().Controller().Context().Request.Param("f")
		if err != nil {
			return nil, err
		}
		return selectCallback(gr, selectField[0])
	case "upload":
		return uploadCallback(gr)
	}

	return nil, fmt.Errorf(ErrCallback, cbk)
}

func (g *gridSource) Cache() cache.Manager {
	c, _ := g.orm.DefaultCache()
	return c
}

// UpdatedFields is called before the grid render starts.
// The permissions will be set.
func (g *gridSource) UpdatedFields(grid Grid) error {

	if grid.Mode() != SrcCallback && grid.Mode() != SrcDelete {
		// get user defined field config
		whitelist, err := whitelistFields(g, grid.Scope().Fields(), "")
		if err != nil {
			return err
		}

		if len(whitelist) > 0 {
			g.orm.SetPermissions(orm.WHITELIST, whitelist...)
			s, err := g.orm.Scope()
			if err != nil {
				return err
			}
			cfg := s.Config()
			s.SetConfig(cfg.SetUpdateReferenceOnly(true).SetPermissionsExplicit(true))
		} else {
			return fmt.Errorf(ErrConfig)
		}
	}

	return nil
}

// whitelistFields is a helper to set a whitelist of added fields.
// Only fields will be displayed which are not removed by the grid.
// If a readOnly is defined, the permission will be set to READ.
// This function will be called recursively for relations.
func whitelistFields(g *gridSource, fields []Field, parent string) ([]string, error) {
	var whitelist []string

	if parent != "" {
		parent += "."
	}
	for k, f := range fields {

		// check if its a normal field
		if !f.Relation() {

			// add if its not removed
			if !f.Removed() {
				if _, exists := slicer.StringExists(whitelist, parent+f.referenceName); !exists {
					whitelist = append(whitelist, parent+f.referenceName)
				}

				// set write permission to false if its read only.
				// only working on root level.
				if f.ReadOnly() && parent == "" { // because relation fields can not be fetched
					scope, err := g.orm.Scope()
					if err != nil {
						return nil, err
					}
					ormField, err := scope.Field(f.referenceName)
					if err != nil {
						return nil, err
					}
					ormField.Permission.Write = false
				}
			}
		} else {

			// If sub notations are allowed of the relation, allow the relation but dont whitelist the whole.
			bList, err := whitelistFields(g, f.Fields(), parent+f.referenceName)

			if err != nil {
				return nil, err
			}
			// if there is a dot notation, add it.
			if len(bList) > 0 {
				for _, tmp := range bList {
					if _, exists := slicer.StringExists(whitelist, tmp); !exists {
						whitelist = append(whitelist, tmp)
					}
				}
				fields[k].SetRemove(NewValue(false))
			} else {
				// If the whole relation is allowed.
				if !f.Removed() {
					// add the complete relation if there were no fields set.
					if _, exists := slicer.StringExists(whitelist, parent+f.referenceName); !exists {
						whitelist = append(whitelist, parent+f.referenceName)
					}
				}
			}

		}
	}

	return whitelist, nil
}

// Will return the correct source as interface for casting.
func (g *gridSource) Interface() interface{} {
	return g.orm
}

// Fields will generate grid.Fields from the []orm.Field and []orm.Relation.
func (g *gridSource) Fields(grid Grid) ([]Field, error) {
	scope, err := g.orm.Scope()
	if err != nil {
		return nil, err
	}
	return gridFields(scope, grid, "")
}

// First returns one row.
// Used for the grid details and edit view.
func (g *gridSource) First(c condition.Condition, grid Grid) (interface{}, error) {

	// fetch data
	err := g.orm.First(c)
	if err != nil {
		return nil, err
	}

	return g.orm, nil
}

// All returns all rows by the given condition.
func (g *gridSource) All(c condition.Condition, grid Grid) (interface{}, error) {
	// creating result slice
	model := reflect.New(reflect.ValueOf(g.orm).Elem().Type()).Elem() //new type
	resultSlice := reflect.New(reflect.MakeSlice(reflect.SliceOf(model.Type()), 0, 0).Type())

	// fetch data
	err := g.orm.All(resultSlice.Interface(), c)
	if err != nil {
		return nil, err
	}

	return reflect.Indirect(resultSlice).Interface(), nil
}

// Create a new entry.
// The primary key will be returned.
func (g *gridSource) Create(grid Grid) (interface{}, error) {

	err := g.unmarshalModel(grid)
	if err != nil {
		return nil, err
	}

	err = g.orm.Create()
	if err != nil {
		return 0, err
	}

	//response with the pkey values, because the frontend is reloading the view by id.
	pkeys := make(map[string]interface{}, 0)
	scope, err := g.orm.Scope()
	if err != nil {
		return 0, err
	}
	for _, f := range scope.SQLFields(orm.Permission{Read: true}) {
		if f.Information.PrimaryKey {
			pkeys[f.Name] = scope.FieldValue(f.Name).Interface()
		}
	}

	// return with the pkey(s) value
	return pkeys, historyGridHelper(grid)
}

// Update the entry
func (g *gridSource) Update(grid Grid) error {
	err := g.unmarshalModel(grid)
	if err != nil {
		return err
	}

	err = g.orm.Update()
	if err != nil {
		return err
	}

	return historyGridHelper(grid)
}

// Delete an entry.
// First the whole row will be fetched and deleted afterwards. To guarantee that relations are deleted as well.
func (g *gridSource) Delete(c condition.Condition, grid Grid) error {
	err := g.orm.First(c)
	if err != nil {
		return err
	}
	err = g.orm.Delete()
	if err != nil {
		return err
	}

	return historyGridHelper(grid)
}

// Count returns a number of existing rows.
func (g *gridSource) Count(c condition.Condition, grid Grid) (int, error) {
	return g.orm.Count(c)
}

// unmarshalModel is needed for create and update an orm model.
// It checks if the request json is correct.
// Only struct fields are allowed.
// Empty struct is not allowed.
// Errors will return if one of the rules are not satisfied.
func (g *gridSource) unmarshalModel(grid Grid) error {
	body := grid.Scope().Controller().Context().Request.Body()
	if body == nil {
		return fmt.Errorf(ErrRequestBody, grid.Scope().Config().ID)
	}

	// check if the json is valid
	if !json.Valid(body) {
		return fmt.Errorf(ErrJSONInvalid, grid.Scope().Config().ID)
	}

	// unmarshal the request to the model struct
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	for dec.More() {
		err := dec.Decode(g.orm)
		if err != nil {
			return err
		}
	}

	// check if the model has any data
	//if g.src.IsEmpty() {
	//	return ErrModelIsEmpty
	//}

	return nil
}

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

// gridFields is recursively adding the orm fields/relations to the grid.
//
// orm.Fields:
//   - add referenceID, referenceName. The referenceID will be the same as referenceName if its a no sql field.
//   - add Name. Will be the orm field name or the json name if defined.
//   - set primary, type, position.
//   - set title, description. By default is the orm model name + field name + -title or -description.
//   - set sort. By default its allowed and the field value is the orm column name.
//   - set filter. By default its allowed and the condition operator equal and the field value is the orm column name.
//   - set groupAble. By default allowed.
//   - validator config is added as option by the key "validate".
//   - if the type is SELECT or MULTISELCET, the select is added as option by the key "select".
//   - if its a primary-, fk-, refs-, polymorphic key the field is getting removed by default.
//
// orm.Relations:
//   - If its a BelongsTo and its not the FeTable, the relation will be skipped.
//     The Foreign-Key will be un-removed because this will be the dropdown field on the frontend.
//   - referenceName will be set as orm field name.
//   - add Name. Will be the orm field name or the json name if defined.
//   - set relation, type, position
//   - set title, description. By default is the orm model name + field name + -title or -description.
//   - TODO sort and filter for relations depth 1
//   - validator config is added as option by the key "validate".
//   - IF its belongsTo or M2M relation a Select is added. TextField = field index 2(experimental orm,id,->name)  and ValueField will be the Mapping.References.Name.
//   - recursively add all relation fields.
//   - if its a primary-, fk-, refs-, polymorphic key the field is getting removed by default.
func gridFields(scope orm.Scope, g Grid, parent string) ([]Field, error) {

	var rv []Field
	i := 0

	// normal fields
	for _, f := range scope.Fields(orm.Permission{}) {
		// only allow read permission (and TimeFields) fields to be shown.
		if !f.Permission.Read &&
			f.Name != orm.CreatedAt &&
			f.Name != orm.UpdatedAt &&
			f.Name != orm.DeletedAt {
			continue
		}

		field := Field{}

		field.referenceID = f.Name
		field.referenceName = f.Name
		if !f.NoSQLColumn {
			field.referenceID = f.Information.Name
		}

		field.SetName(f.Name)
		if skipJSONByTagOrSetJSONName(scope, &field) {
			continue
		}
		field.SetPrimary(f.Information.PrimaryKey)
		field.SetType(f.Information.Type.Kind())

		field.SetTitle(NewValue(translation.ORM + scope.Name(true) + "." + f.Name))
		// TODO translate desc
		//field.SetDescription(NewValue(translation.ORM+scope.Name(true) + "." + f.Name + "-Description"))

		field.SetPosition(NewValue(i))
		// field.SetRemove(NewValue(false))
		// field.SetHidden(NewValue(false))
		// field.SetView(g.NewValue(""))
		field.SetSort(true, f.Information.Name)
		field.SetFilter(true, query.LIKE, f.Information.Name)
		if f.Information.Type.Kind() == types.DATE || f.Information.Type.Kind() == types.DATETIME {
			field.SetFilter(true, query.MYSQLDATE, f.Information.Name)
		}
		field.SetGroupAble(true)
		// set validation tag
		if f.Validator.Config() != "" {
			field.SetOption(orm.TagValidate, f.Validator.Config())
		}

		if f.Information.Type.Kind() == types.SELECT || f.Information.Type.Kind() == types.MULTISELECT {
			var items []options.SelectItem
			sel := f.Information.Type.(types.Items)
			for _, i := range sel.Items() {
				items = append(items, options.SelectItem{Text: translation.ORM + scope.Name(true) + "." + f.Name + "." + i, Value: i})
			}
			var multiple bool
			if f.Information.Type.Kind() == types.MULTISELECT {
				multiple = true
			}
			field.SetOption(options.SELECT, options.Select{ReturnValue: true, TextField: "text", ValueField: "value", Items: items, Multiple: multiple})
		}

		// field manipulations
		// Primary keys are not shown on frontend header.
		// In the backend they are loaded anyway because they are mandatory for relations.
		if f.Information.PrimaryKey {
			field.SetRemove(NewValue(true))
		}

		// belongsTo references key is getting removed in grid table and details.
		for _, relation := range scope.Relations(orm.Permission{Read: true}) {
			if relation.Kind == orm.BelongsTo {
				if f.Name == relation.Mapping.ForeignKey.Name {
					field.SetRemove(NewValue(true).SetCreate(false).SetUpdate(false))
				}
			}
		}

		rv = append(rv, field)
		i++
	}

	// relation fields
	for _, relation := range scope.Relations(orm.Permission{Read: true}) {

		// skip on self referencing orm model loops.
		// skips everything > depth 1 because the root scope has no parent yet.
		if _, err := scope.Parent(relation.Type.String()); err == nil {
			continue
		}

		// if its a belongsTo relation and the view is no grid table, only the association field is loaded because this is
		// a dropdown on the frontend and only the ID is needed.
		// TODO if this changes to a combobox in the future, this code has to get changed.
		if relation.Kind == orm.BelongsTo {
			for k := range rv {
				if rv[k].referenceName == relation.Mapping.ForeignKey.Name {
					if parent != "" {
						parent += "."
					}
					rv[k].SetType(relation.Kind)

					// TODO better solution.
					// Idea is to automatic find out the TextField of the database. at the moment the 2nd ord 3rd field will be taken.
					// Soluation id - the field with the next "TEXT" type orm description.
					// 3 fields = id, parendid, text
					// 2 fields = id, value

					var t reflect.Type
					if relation.Type.Kind() == reflect.Ptr {
						t = relation.Type.Elem()
					} else {
						t = relation.Type
					}

					if t.NumField() == 2 {
						rv[k].SetOption(options.SELECT, options.Select{ReturnValue: true, OrmField: parent + relation.Field, TextField: t.Field(1).Name, ValueField: relation.Mapping.References.Name})
					} else {
						rv[k].SetOption(options.SELECT, options.Select{ReturnValue: true, OrmField: parent + relation.Field, TextField: t.Field(2).Name, ValueField: relation.Mapping.References.Name})
					}
				}
			}

			// the whole relation has to get removed. otherwise the single foreign key will not work with the existing logic.
			// why: the relation object is not changing in the frontend, only the afk on the root model. But the afk is manipulated automatically of the orm on create.
			// another reason why this is handled like this, the whole belongsTo object has to get loaded for every select item - performance risk.
			field := Field{}
			field.referenceName = relation.Field
			field.SetTitle(NewValue(translation.ORM + scope.Name(true) + "." + relation.Field))
			field.SetName(relation.Field)
			field.SetRemove(NewValue(true).SetTable(false)).SetRelation(true)

			if relation.Kind == orm.BelongsTo {
				// TODO see above
				// display only the text field... experimental
				var t reflect.Type
				if relation.Type.Kind() == reflect.Ptr {
					t = relation.Type.Elem()
				} else {
					t = relation.Type
				}
				if t.NumField() == 2 {
					field.SetOption(options.DECORATOR, "{{"+t.Field(1).Name+"}}")
				} else {
					field.SetOption(options.DECORATOR, "{{"+t.Field(2).Name+"}}")
				}

			}

			// recursively add fields
			// this is needed because otherwise all fields are loaded on a belongsTo relation
			// TODO investigate to find a better solution
			// TODO CreatedAt,UpdatedAt are also loaded by default - even if they are on removed.
			rScope, err := scope.NewScopeFromType(relation.Type)
			if err != nil {
				return nil, err
			}
			rScope.SetParent(scope.Model()) // adding parent to avoid loops
			rField, err := gridFields(rScope, g, relation.Field)
			if err != nil {
				return nil, err
			}
			if len(rField) > 0 {
				field.SetFields(rField)
			}

			rv = append(rv, field)

			continue
		}

		// relation field
		field := Field{}
		field.referenceName = relation.Field
		field.SetName(relation.Field)
		field.SetRelation(true)
		if skipJSONByTagOrSetJSONName(scope, &field) {
			continue
		}
		field.SetType(relation.Kind)
		field.SetTitle(NewValue(translation.ORM + scope.Name(true) + "." + relation.Field))
		// TODO desc translation
		//	field.SetDescription(NewValue(scope.Name(true) + ":" + relation.Field + "-description"))

		field.SetPosition(NewValue(i))
		//field.SetRemove(NewValue(false))
		//field.SetHidden(false)
		//field.SetView(g.NewValue(""))
		// set validation tag
		if relation.Validator.Config() != "" {
			field.SetOption(orm.TagValidate, relation.Validator.Config())
		}
		// TODO
		//field.SetSort(false)
		//field.SetFilter(false)
		//field.SetGroupAble(false)

		// add options for BelongsTo and ManyToMany relations.
		if relation.Kind == orm.BelongsTo || relation.Kind == orm.ManyToMany {
			// experimental (model,id, title...) by default always the third field is taken.
			var name string
			if relation.Kind == orm.BelongsTo {
				name = relation.Type.Field(2).Name
			} else {
				name = relation.Type.Elem().Field(2).Name
			}
			field.SetOption(options.SELECT, options.Select{TextField: name, ValueField: relation.Mapping.References.Name})

		}

		// recursively add fields
		rScope, err := scope.NewScopeFromType(relation.Type)
		if err != nil {
			return nil, err
		}
		rScope.SetParent(scope.Model()) // adding parent to avoid loops
		rField, err := gridFields(rScope, g, relation.Field)
		if err != nil {
			return nil, err
		}

		// field manipulations - FK,AFK,Poly are removed.
		for k := range rv {
			if rv[k].referenceName == relation.Mapping.ForeignKey.Name {
				rv[k].SetRemove(NewValue(true))
			}
		}
		for k, relField := range rField {
			if (relField.referenceName == relation.Mapping.References.Name) ||
				(relation.IsPolymorphic() && relation.Kind != orm.ManyToMany && relField.referenceName == relation.Mapping.Polymorphic.TypeField.Name) {
				rField[k].SetRemove(NewValue(true))
			}
		}

		// adding relation fields
		if len(rField) > 0 {
			field.SetFields(rField)
		}

		rv = append(rv, field)
		i++
	}

	return rv, nil
}

// skipJSONByTagOrSetJSONName will return true if the json skip tag exists.
// Otherwise it checks if a json name is set, and sets the field Name.
func skipJSONByTagOrSetJSONName(scope orm.Scope, f *Field) bool {

	rField, ok := reflect.TypeOf(scope.Caller()).Elem().FieldByName(f.name)
	if ok {
		jsonTag := rField.Tag.Get("json")
		if jsonTag == "-" {
			return true
		}

		if jsonTag != "" && jsonTag != "-" {
			if commaIdx := strings.Index(jsonTag, ","); commaIdx != 0 {
				if commaIdx > 0 {
					f.SetName(jsonTag[:commaIdx])
				} else {
					f.SetName(jsonTag)
				}
			}
		}
	}

	return false
}

func uploadCallback(g Grid) (interface{}, error) {
	// get field name and options
	selectField, err := g.Scope().Controller().Context().Request.Param("f")
	if err != nil {
		return nil, err
	}
	path := g.Field(selectField[0]).Option(options.FILEPATH)
	if len(path) != 1 || path[0].(string) == "" {
		return nil, errors.New("path is not defined")
	}

	switch g.Scope().Controller().Context().Request.Method() {
	case http.MethodPost:
		filesTags, err := g.Scope().Controller().Context().Request.Files()
		if err != nil {
			return nil, err
		}
		for _, files := range filesTags {
			for _, file := range files {
				rv, err := saveFile(path[0].(string), file)
				if err != nil {
					return nil, err
				}
				return rv, nil
			}
		}
	case http.MethodDelete:
		// TODO create a secure version over file id.
		type File struct {
			File string
		}

		file := File{}
		err = json.Unmarshal(g.Scope().Controller().Context().Request.Body(), &file)
		if err != nil {
			return nil, err
		}

		err = os.Remove(file.File)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func saveFile(filepath string, header *multipart.FileHeader) (interface{}, error) {
	file, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// create folder
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	path_ := path.Dir(ex) + "/" + filepath
	err = os.MkdirAll(path_, os.ModePerm)
	if err != nil {
		return nil, err
	}
	tempFile, err := ioutil.TempFile(path_, "upload-*.png")
	if err != nil {
		return nil, err
	}

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	// write this byte array to our temporary file
	_, err = tempFile.Write(fileBytes)

	// return that we have successfully uploaded our file!
	rv := map[string]interface{}{}
	rv["Name"] = tempFile.Name()
	rv["Size"] = header.Size
	rv["Type"] = header.Header.Get("content-type")
	return rv, err
}

func selectCallbackHistory(g Grid, selectField string, cond condition.Condition) string {
	v, err := selectCallback(g, selectField, cond)
	if err != nil {
		return ""
	}

	var rv []string
	for i := 0; i < reflect.ValueOf(v).Len(); i++ {
		rv = append(rv, reflect.ValueOf(v).Index(i).FieldByName(g.Field(selectField).option[options.SELECT][0].(options.Select).TextField).String())
	}
	return strings.Join(rv, defaultSeparator)
}

// TODO docu
// cond is used for additional condition, needed in history.
func selectCallback(g Grid, selectField string, cond ...condition.Condition) (interface{}, error) {

	// get the defined grid object
	selField := g.Field(selectField)
	if selField.Error() != nil {
		return nil, selField.Error()
	}
	selX := selField.Option(options.SELECT)
	sel := selX[0].(options.Select)

	var relScope orm.Scope
	fields := strings.Split(selectField, ".")
	if sel.OrmField != "" {
		fields = strings.Split(sel.OrmField, ".")
	}

	// get relation field of the orm
	src := g.Scope().Source().(orm.Interface)
	scope, err := src.Scope()
	relation, err := scope.SQLRelation(fields[0], orm.Permission{})

	if err != nil {
		return nil, err
	}

	if len(fields) > 1 {
		// create a new orm object
		relScope, err := scope.NewScopeFromType(relation.Type)
		if err != nil {
			return nil, err
		}
		// get relation field of the orm
		relation, err = relScope.SQLRelation(fields[1], orm.Permission{Read: true})
		if err != nil {
			return nil, err
		}
	}

	// create a new orm object
	relScope, err = scope.NewScopeFromType(relation.Type)
	if err != nil {
		return nil, err
	}

	// create the result slice
	var rRes reflect.Value
	// TODO check if slice then
	if relation.Type.Kind() == reflect.Slice {
		rRes = reflect.New(relation.Type)
	} else {
		rRes = reflect.New(reflect.MakeSlice(reflect.SliceOf(relation.Type), 0, 0).Type())
	}

	// set whitelist fields
	reqFields := []string{sel.ValueField}
	textFields := strings.Split(sel.TextField, ",")
	for k, tf := range textFields {
		textFields[k] = strings.Trim(tf, " ")
		reqFields = append(reqFields, strings.Trim(tf, " "))
	}

	_, err = g.Scope().Controller().Context().Request.Param("allFields")
	if err != nil {
		relScope.Model().SetPermissions(orm.WHITELIST, reqFields...)
	} else {
		//fmt.Println(g.Scope().Controller().Context().Request.Param("allFields"))
	}

	// request the data
	var c condition.Condition
	if len(cond) == 1 {
		c = cond[0]
	} else {
		c = condition.New()
		if sel.Condition != "" {
			c.SetWhere(sel.Condition)
		}
	}

	err = relScope.Model().All(rRes.Interface(), c)
	if err != nil {
		return nil, err
	}

	return reflect.Indirect(rRes).Interface(), nil
}
