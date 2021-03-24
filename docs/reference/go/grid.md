# Grid

The grid package converts any `grid.Source` into a CRUD backend.

* All primary-, foreign, references and polymorphic fields are removed by default
* Relation will be displayed depth max 1 at the moment.
* `belongsTo` and `manyToMany` will be select boxes on the frontend.
* Errors will be set as controller errors.
* Field validation will happen automatically on front- and backend.
* Frontend fields will be rendered automatically by field type.
* Developer friendly. Every source which implements the `grid.Source` can be used.

## Usage

```go 
func (c MyController) User(){
    g := grid.New(c,grid.Orm(model),nil)
    //...
    g.Render()
} 
```

## New

Creates a new grid instance. The first argument must be the `controller`, the second is the `grid.Source` and the third
is the `grid.config` which is optional.

The grid will be cached, to avoid re-creating the grid fields. The cache key will be the configured grid ID. Be aware
that the config will be cached. If you need dynamic configuration, use `grid.Scope().Config` for it, after init.

```go
g := grid.New(c, grid.Orm(model))
```

## Config

The grid can be fully configured. If the configuration should be changed after init dynamically, the `Scope.Config()`
can be used.

| Name        | Default |Description                                                       | 
|-------------|----|-------------------------------------------------------------------|
| ID          | `controller:action`    | Unique name for the grid. This is used as cache key.              |
| Title       | `{ID}-title`   |  Title of the grid.                 |
| Description | `{ID}-description`   | Description of the grid.              |
| Policy      | `orm.WHITELIST`   | If the Policy is `WHITELIST`, the fields have to be set explicit. |
| Exports     | `csv`  | Slice of names of render types.                                      |
| Action      |    | see ACTION                                                        |
| Filter      |    | see FILTER                                                        |
| History      |    | see HISTORY                                                  |

**Action**

| Name        |  Default |  Description                                                       |
|-------------|-----|----------------------------------------------------------------|
| PositionLeft          | `false`   | Defines where the action (details,edit,delete) column should be displayed on the grid table.            |
| DisableDetails       | `true` |  Disables the details mode. It is disabled because its not implemented yet!                   |
| DisableCreate | `false`  |  Disables the create mode.                 |
| DisableUpdate      | `false`  |  Disables the update mode. |
| DisableDelete      | `false`  |  Disables the delete mode.                                                       |
| CreateLinks      | `nil`  |  You can add params to the grid Add button. IF multiple entries exist, a menu will be generated.                                                      |

**Filter**

| Name        |  Default |  Description                                                       |
|-------------|-----|----------------------------------------------------------------|
| Disable          | `false`   | Disable filter.            |
| DisableQuickFilter       | `false` |  Disable the quick filter.                    |
| DisableCustomFilter | `false`  |  Disable the custom filter.                 |
| OpenQuickFilter      | `false` |  The quick filter will be opened by default |
| AllowedRowsPerPage | `-1`,`5`,`10`,`15`,`25`,`50`  | The allowed rows per page. | 
| RowsPerPage | `15`  | Default rows per page. |

**History**

| Name        |  Default |  Description                                                       |
|-------------|-----|----------------------------------------------------------------|
| Disable          | `false`   | Disable the history.            |
| AdditionalIDs          | `[]string{}`   | Additional grid IDs can be added to show in the history content.          |

## Mode

The grid mode is defined by the `http.Method` and `controller.Params`.

| Mode        |  http.Method |  Param                                                       |
|-------------|-----|----------------------------------------------------------------|
| `grid.FeTable`          | `GET`   |             |
| `grid.FeHistory`          | `GET`   | `mode=history`            |
| `grid.FeFilter`          | `GET`   | `mode=filter`            |
| `grid.SrcCallback`          | `GET`   | `mode=callback`            |
| `grid.FeDetails`          | `GET`   | `mode=details`            |
| `grid.FeCreate`          | `GET`   | `mode=create`            |
| `grid.FeUpdate`          | `GET`   | `mode=update`            |
| `grid.FeExport`          | `GET`   | `mode=export`            |
| `grid.SrcCreate`          | `POST`   |           |
| `grid.SrcUpdate`          | `PUT`   |           |
| `grid.SrcDelete`          | `DELETE`   |           |

## Field

Will return the grid field by name. If the field does not exist, an empty field with an error will return.

```go
field := grid.Field("ID")
```

A field can be configured by the following functions. Each function returns itself, this way it can be chained. If an
error occures, the fields error will be set. Error will be handled in `grid.Render`.

The configuration for `SetPosition`, `SetTitle`, `SetDescription`, `SetRemove`, `SetHidden` and `SetView` must be set
with `grid.NewValue()` or native go type.

* `string` for `SetTitle`, `SetDescription`, `SetView`
* `bool` for `SetRemove`, `SetHidden`
* `int` for `SetPosition`

If the native type is different, an error will be set.

```go
grid.Field("ID").SetTitle(grid.NewValue("ID").SetDetails("Identifier"))
// grid mode: table, update, create will have the title "ID" 
// and details will have the title "Identifier".
```

| Function        |  available frontend |  Description | 
|-------------|-----|-----|
| `Name`, `SetName`   | x | Will set the fields name. The name is used in the frontend as id.   |
| `Primary`, `SetPrimary`  | x | Will define if the field is a primary key.  |
| `Type`, `SetType`   | x | Defines the field type.  |
| `Title`, `SetTitle`   |x | Will set the fields title.  |
| `Description`, `SetDescription`  |x | Will set the fields description.  |
| `Position`, `SetPosition`  |x | Will set the fields position.  |
| `Removed`, `SetRemove`  |x | Will flag the field as removed.  |
| `Hidden`, `SetHidden`  | x| Will set the field as hidden.  |
| `View`, `SetView`   |x| Will set a custom frontend view component for the field.  |
| `ReadOnly`, `SetReadOnly`  | x| Will set the field as read only.  |
| `Sort`, `SetSort`   |x| Will allow the sorting of the field and set the condition field name. |
| `Filter`, `SetFilter`   |x| Will allow the filtering of the field and set the condition operator and field name. |
| `GroupAble`, `SetGroupAble`   |x| Will set the field as group able.  |
| `Options`, `Option`, `SetOption`  | x| Will add a option for the field.  |
| `Relation`, `SetRelation`   |x| Will define the field as relation  |
| `Field`   || Will return a field by name. If the field was not found, an field error will be set. (relation)  |
| `Fields`, `SetFields`   |x| Will return all child fields. (relation)  |
| `Error`  || Will return the field error. |

**Field types**

| Name        |  implemented in frontend |  Description |
|-------------|-----|-----|
| `Bool`   |  | Checkbox   |
| `Integer`   |  | Input-Integer   |
| `Float`   |  | Input-Numeric   |
| `Text`   |  | Input-Text   |
| `TextArea`   |  | TextArea   |
| `Time`   |  | Input   |
| `Date`   |  | Datepicker   |
| `DateTime`   |  | Datepicker+Input   |
| `Select`   |  | Select   |
| `MultiSelect`   |  | Select   |
| `belongsTo`   |  | Select   |
| `hasOne`   |  | Inputs   |
| `hasMany`   |  | Inputs   |
| `m2m`   |  | Select   |

**Options**

| Name        |  value |  Description |
|-------------|-----|-----|
| `DecoratorOption`   | `string`,`string` | a field name can be used {{Name}}. As second param a separator can be set - if set the FE escaping will be disabled.  |

**Callbacks**

| Name        |  value |  Description |
|-------------|-----|-----|
| `Select`   | `?` |   |

TODO: Validate

## Scope

The scope will return some helper functions.

### Source

Will return the grid source.

```go
src := scrope.Source()
```

### Config

Will return a pointer to the grid config. For dynamically configuration of the grid.

```go
cfg := scrope.Config()
```

### Fields

Will return all configured grid fields.

```go
fields := scrope.Fields()
```

### PrimaryFields

Will return all defined primary fields of the grid.

```go
primaryFields := scrope.PrimaryFields()
```

### Controller

Will return the grid controller instance.

```go
ctrl := scrope.Controller()
```

## Render

Will render the grid by the actual grid mode.

| Mode        | set in frontend data |  Description | 
|-------------|-----|-----|
| `grid.SrcCallback`   | `data`  | The source callback function is called. as first param the requested callback will be set as string. |
| `grid.SrcCreate`   |   | The source create function is called. | 
| `grid.SrcUpdate`     || The source update function is called. | 
| `grid.SrcDelete`     || The condition first will be called to ensure the correct primary key. The source delete function is called.| 
| `grid.FeTable`    | `pagination`, `head`, `data`, `config`| ConditionAll is called to create the condition. Add header/pagination if its not excluded by param. The source all function is called. Add config and result to the controller. call the defined render type.| 
| `grid.FeExport`     | `head`, `data`, `config`| Same as FeTable but without the pagination and limit.|
| `grid.FeCreate`    |`head` | add header data. | 
| `grid.FeDetails`,`grid.FeUpdate`    | `head`, `data`| add header data. call conditionFirst. fetch the entry by the given id and set the controller data. | 
| `grid.FeFilter`    | | TODO | 
| `grid.FeHistory`    |`histories`, `users` | all history entries and user data to the given sourceID will be fetched. | 

## Orm

With the orm function an `orm.Interface` will be converted to a `grid.Source` and can be used out of the box.

History is implemented.

```go
g := grid.New(ctrl, grid.Orm(model), nil)
```

## History

!!! info

    Must be implmented by the source.

The data will be saved in the `histories` table by the `grid.Histroy` struct. The following Fields are available
defined:

*grid.History* saves the entries in the database with all the needed information.

| Field        | value |  Description | 
|-------------|-----|-----|
| GridID | `string`|The grid id. There can be multiple IDs set. The will get separated through `,`|
| UserID | `string`|The users id as a string.|
| SrcID | `string`|The ID of the source struct.|
| Type | `enum`|Can have the value `Created`, `Updated` or `Deleted`|
| Value | `text`| `orm.ChangeValue` as json.|
| CratedAt | `datetime`| The current datetime when it was created.|

!!! info

    If the `UserID` is `0`, it will be displayed as a SYSTEM user. This can be used for cronjobs or other automated changes.

*orm.ChangeValue* will be used to describe the source changes.

| Field        | value |  Description | 
|-------------|-----|-----|
| Field| `string`|The name of the struct field.|
| Operation| `string`| Value of `create`, `update` or `delete`|
| New| `string`|The new value of the field. Can be empty if zero value.|
| Old| `string`|The old value of the field. Can be empty if zero value. |
| Index| `int`| Only used for `hasMany` relations. |
| Children| `[]orm.ChangeValue`| Same fields as described before in a deeper level. |

*Create:* Fields will only be added if they have no zero value.

| Type        |  Description | 
|-------------|-----|
|normal field|  `New` will be the value of the field.|
|belongsTo, m2m| `New` field will be the value of the select `TextValue` field. To guarantee the correct value in the future, also if the ID got deleted.|
|HasOne|  Every field will be in the `Childeren` slice if the value is not zero.|
|HasMany|  Same as `hasOne` but a the `Index` field will be set.|

*Update* Only changed values will be added.

| Type        |  Description | 
|-------------|-----|
|normal field|  `New` and `Old` will have the fields value. If one of it has a zero value, it will be omitted.|
|belongsTo, m2m| `New` and `Old` will be the value of the select `TextValue` field. To guarantee the correct value in the future, also if the ID got deleted. If one of it has a zero value, it will be omitted.|
|HasOne|  Every field will be in the `Childeren` slice if the value is not zero.|
|HasMany|  Can have the following state `create`, `update` or `delete`. On `create` only the new value will be set, on `delete` only the old value.|

*Delete*

A `orm.History` entry will be added with the `Type: DELETED`.

### manually add history

```go

err := grid.NewHistory("gridID", "userID", "srcID", grid.HistoryCreated, "New data received.")
//...

```

## Source interface

To create your own source, you have to implement the `grid.Source`.

```go
type Source interface {
Cache() cache.Manager

PreInit(Grid) error
Init(Grid) error
Fields(Grid) ([]Field, error)
UpdatedFields(Grid) error

Callback(string, Grid) (interface{}, error)
First(condition.Condition, Grid) (interface{}, error)
All(condition.Condition, Grid) (interface{}, error)
Create(Grid) (interface{}, error)
Update(Grid) error
Delete(condition.Condition, Grid) error
Count(condition.Condition, Grid) (int, error)

//Interface() interface{}
}
```
