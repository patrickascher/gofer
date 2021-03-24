# Orm

Package orm transfers a struct into an ORM by simple embedding the `orm.Model`. Relations `hasOne`, `belongsTo`
, `hasMany` and `m2m` will be defined automatically / manually.

## Usage

```go
type User struct{
orm.Model

ID      int
Name    string
Surname string
}

user := User{}

// initialize orm model
err = user.Init(&user)
if err!=nil{
//...
}

// scope for some helper - if needed
scope, err := user.Scope()
if err!=nil{
//...
}

// set data
user.Name = "John"
user.Surname = "Doe"

// create entry
err = user.Create()
//..
```

Requirements / defaults:

* Database name, Table name, Builder and Cache must be set. [[see Default]](orm.md#defaults)
* Model requires one or more primary keys. If the field `ID` exists, it will automatically be taken as primary key.
  Primary keys can be set manually via Tags. [[see Tags]](orm.md#tags)
* All fields and relations must be available on the database backend, or they must be defined as custom.
* Unique field names must be provided. If an embedded struct overwrites a field name or relation, an error will return.
* Fields are allowed with the following type `string`, `bool` `uint` `int` `float` and any type which implements
  the `sql.Scanner` and `driver.Valuer` interface.
* Relations are only set if they implement the `orm.Interface`, except it's defined as custom.

## First

Will return the first found row.

Error `sql.ErrNoRows` will return if no result was found.

```go

user := User{}
err := user.Init(&orm)
// ...

// first without any condition
err = user.First()

// first with a condition (id=1
err = user.First(condition.New().SetWhere("id = ?", 1))
```

## All

Will return all rows by the given condition. For more details about the relation
handling, [see Strategy](orm.md#strategy).

No error will return if no result was found (TODO CHANGE? same logic as First?)

```go

user := User{}
err := user.Init(&orm)
// ...

// all without any condition
var users []User
err = user.All(&user)

// all with a condition (id>10)
err = user.All(&user, condition.New().SetWhere("id > ?", 10))
```

## Count

Count the existing rows by the given condition.

```go

user := User{}
err := user.Init(&orm)
// ...

// count without any condition
rows, err = user.Count()

// count with a condition (id>10)
rows, err = user.Count(condition.New().SetWhere("id > ?", 10))
```

## Create

Will create an entry. For more details about the relation handling, [see Strategy](orm.md#strategy).

```go

user := User{}
err := user.Init(&orm)
// ...

user.Name = "John"
user.Surname = "Doe"
user.Phonenumbers = append(user.Phonenumbers, "000-111-222") //has m

err = user.Create()
```

## Update

Will update an entry. For more details about the relation handling, [see Strategy](orm.md#strategy).

A Snapshot will be taken and only changed values will be updated.

```go

user := User{}
err := user.Init(&orm)
// ...

user.Name = "Foo"
user.Surname = "Bar"

err = user.Update()
```

## Delete

Will delete an entry. For more details about the relation handling, [see Strategy](orm.md#strategy).

```go

user := User{}
err := user.Init(&orm)
// ...
user.ID = 4
err = user.Delete()
```

## Permissions

Like the permission tag [see Tags](orm.md#tags), it's sometimes useful to dynamic set the policy and fields.

`Permissions` sets the read/write permission for the given fields. This means you can allow or disallow single fields
for saving / fetching. The field setting will overwrite the configured permission tag.

!!! info

    Primary-, foreign-, reference and polykeys are always added. This means if an ID, which is a primary key gets blacklisted, the ID field will be removed from the blacklist automatically.

```go

// set field permission - only Name, Surname and all mandatory keys will be loaded.
user.SetPermissions(orm.WHITELIST, "Name", "Surname")

// read the configured permissions
policy, fields := user.Permissions()
```

## Tags

The orm struct fields can be simple configured by tags. The tag must be defined under the key `orm`

For more details about the tags, [see ParseTag](structer.md#parsetag)

| Tag               | Description                                                                                                     | Values         | Example               |   |
|-------------------|-----------------------------------------------------------------------------------------------------------------|----------------|-----------------------|---|
| `-`        | Skips the complete struct field.                                                                         |                | `orm:"-"`    |   |
| custom        | Defines a field as a none sql field.                                                                         |                | `orm:"custom"`    |   |
| column            | Set a custom table column name                                                                                  | name           | `orm:"column:name"`   |   |
| permission        | A field can be defined as Write or Read only. If the permission is empty read and write will be set to `false`. If a read permission is false, it the column will not be fetched by first and all. If a write permission is false, the column will not be saved on create or update. | r,w or empty.    | `orm:"permission:rw"` |   |
| sql         | Set a custom select for the column. Only supported for `First` and `All`.                                       | string         | `orm:"sql:CONCAT(name,surname)"`    |   |
| primary           | Defines a column as primary.                                                                                    |          | `orm:"primary"`       |   |
| relation           | Defines a relation   | `hasOne`, `belongsTo`, `hasMany`, `m2m`          | `orm:"relation:belongsTo"`       |   |
| fk           | Defines a custom foreign key | string         | `orm:"fk:CustomID"`       |   |
| refs           | Defines a custom references key.                                                                                    |string          | `orm:"refs:UserID"`       |   |
| join_table           | Defines a custom  join table name.                                                                                    | string         | `orm:"join_table:user_mapping"`       |   |
| join_fk           | Defines a custom foreign column name for the junction table.                                                                                |string          | `orm:"join_fk:CustomID"`       |   |
| join_refs           | Defines a custom references column name for the junction table.                                                                                     | string         | `orm:"join_refs:UserID"`       |   |
| poly           | Defines a custom poly name.                                                                             | string         | `orm:"poly:Toy"`       |   |
| poly_value           | Defines a custom poly value.                                                                                 | string         | `orm:"poly:User"`       |   |

## Validation

Validation for struct fields can be configured by tags.

Under the hood the package [validator](https://github.com/go-playground/validator) is used. Struct fields validation can
be defined by the tag `validate`. Please check out the validator documentation for all available tags.

All [query.NullTypes](query.md#null-types) are registered and can be validated.

Custom validation `tags` can be registered
by `orm.RegisterValidation(tag string, fn func(ctx context.Context, fl valid.FieldLevel) bool, callValidationEvenIfZero ...bool)`
. As context the `orm.Interface` will be set under the name `orm.MODEL`.

!!! info

    The validation happens on `orm.Create` and `orm.Update`. Only on struct fields with write permission.

```go
type User struct{
orm.Model

Name `validate:"required"`
Country `validate:"country"`
}

err := orm.RegisterValidation("country", countryValidation)
if err != nil{
// ...
}

func countryValidation(ctx ctx.Context, fl valid.FieldLevel) bool {
model := ctx.Value(orm.MODEL).(orm.Interface)
// ... some checks
return true
}
```

## Defaults

### Struct defaults

The orm can be simple customized by struct functions.

```go

func (u User) DefaultTableName(){
return "users"
}

```

| Function            | Description |  Default           | Return Value       |   |
|---------------------|-------------|------------------------------|---|---|
| DefaultCache        |  A `cache.Manager` and `time.Duration` must be set. The `time.Duration` indicates how log the `orm.Model` should be cached. [[see Cache]](cache.md)           | -  | `cache.Manager`, `time.Duration`  |   |
| DefaultBuilder      |  A `query.Builder` must be set for the sql handling. [[see Query]](query.md)          | -              | `query.Builder`    |   |
| DefaultTableName    |  The struct table name.            | Plural name of the struct in snake_case.                       | `string`  |   |
| DefaultDatabaseName |  The struct database name.           | The `query.Builder.Config().Database` value.                       | `string`  |   |
| DefaultStrategy     |  The data fetching strategy. [see Strategy]           | eager                      | `string`  |   |
| DefaultSoftDelete   |  Soft deletion instead of deleting the complete db entry.  [[see SoftDelete]](orm.md?#softdelete)               | `DeletedAt`               | `orm.SoftDelete`  |   |

### SoftDelete

By default, a db row will not get deleted, when a column `deleted_at` is available. The default value will be the actual
timestamp.

To change this behaviour, simple overwrite the `DefaultSoftDelete` function. In the example the db field `status` will
be set with the value `1` and all rows with the value `0` are active.

```go
func (y YourModel) DefaultSoftDelete() SoftDelete {
SoftDelete{Field: "Status", Value: "1", ActiveValues: []interface{}{"0"}}
}
```

!!! info

    If the soft delete field does not exist in the struct, an error will return on `orm.Init()`.

### Relations

!!! warn

    Everything in Relations will be developer information and you can probably skip it.

!!! info

    All default settings can be overwritten by tag.

By default, relations will be defined by the struct type.

* struct will be by default a `hasOne` relation.
* slice will be by default a `hasMany` relation.
* slice self referencing will be by default a `m2m` relation

#### HasOne, HasMany

`fk` The foreign key will be the primary key of the orm model.

`refs` The references will be the orm model name + ID on the relation model.

`poly` The polymorphic is by default the relation orm name + ID (will be set as Refs) and name + Type. The value will be
the orm model name.

```go
user User{
ID int
Adr Address{
ID      int
UserID  int
Street  string
}
}
// fk   = ID
// refs = UserID or AddressID if poly is set.
// poly = Address
// poly_value = User
```

#### BelongsTo

`fk` The foreign key will be the relation orm model name + ID on the orm model.

`refs` The relation orm model's primary key.

`poly` The polymorphic is by default the relation orm name + ID (will be set as FK) and name + Type. The value will be
the orm model name.

```go
user User{
ID int
AddressID
Adr Address{
ID      int
UserID  int
Street  string
}
}
// fk   = AddressID
// refs = ID or AddressID if poly is set.
// poly = Address
// poly_value = User
```

#### ManyToMany

`fk` The foreign key will be the primary key of the orm model.

`refs` The references will be the primary key of the orm relation model.

`poly` The polymorphic is by default the relation orm name + ID (will be set as Refs) and name + Type. The value will be
the orm model name.

`join_table` The orm model name + orm relation name in snake style and plural.

`join_fk` The foreign key will be the orm model name + ID of the orm model.

`join_refs` The references key will be the orm relation model name + ID of the orm relation model. It will be `child_id`
on self referencing.

```go
user User{
ID int
AddressID
Adr []Address{
ID      int
UserID  int
Street  string
}
}
// fk   = ID
// refs = ID 
// poly = Address
// poly_value = User
// join_table = user_addresses
// join_fk = user_id
// join_refs = address_id , child_id - on self referencing
```

## Scope

The scope includes some helper functions for the orm model.

Error will return if the orm model was not initialized yet.

```go 
// ...
scope,err := model.Scope()
if err!=nil{
    // ...
}
```

### SetConfig

Can be used to customize the relation or root orm model configuration. If no name is given, the scopes root will be set.

```go 
// customizing a relation sql condition
config := scope.SetConfig(orm.NewConfig().SetCondition(condition.New().SetWhere("id>?",10),true),"Address")
```

| Field               | Default |Description                                                                                                      |
|-------------------|---|---------------------------------------------------------------------------------------------------------------|
|  SetAllowHasOneZero       | `true`  | will trigger an error if a `hasOne` relation has no rows and its set to false.              |         
|  SetShowDeletedRows       | `false`  | will show/hide the deleted rows by the soft delete definitions.            |                      
|  SetUpdateReferenceOnly       | `false`  | will only update the reference on `belongsTo` and `m2m` relations instead of updating the relation model.         |                               
|  SetCondition       |  | add a sql condition. the condition can be merged with the defaults or replace them.                           |                                      
|  Condition       |   | will return the defined condition                                                                   |         

### Config

Will return the defined orm model configuration. If no name is given, the scopes root configuration will be taken.

```go 
config := scope.Config()
```

### Name

Will return the name of the struct, with or without the package prefix.

```go 
// with package name
name := scope.Name(true)

// without package name
name = scope.Name(true)
```

### Builder

Builder will return the model builder.

```go 
builder := scope.Builder()
```

### FqdnTable

Is a helper to display the models database and table name.

```go 
table := scope.FqdnTable()
```

### FqdnModel

Is a helper to display the model name and the field name.

```go 
field := scope.FqdnModel("Name")
// field: orm.User:Name
```

### Model

Will return the scopes orm model.

```go 
model := scope.Model()
```

### Caller

Will return the orm model caller.

```go 
caller := scope.Caller()
```

### Cache, SetCache

Set or get the model cache. At the moment not in use because of the DefaultCache logic. TODO: Delete?

### SQLFields

Will return all struct fields by permission as slice string.

```go 
fields := scope.SQLFields(Permission{Read:true})
```

### SQLScanFields

SQLScanFields is a helper for row.scan. It will scan the struct fields by the given permission.

```go 
fields := scope.SQLScanFields(Permission{Read:true})
```

### SQLColumns

Will return all struct fields by permission as slice string.

```go 
cols := scope.SQLColumns(Permission{Read:true})
```

### Field

Returns a ptr to the struct field by name. Error will return if the field does not exist.

```go 
field,err := scope.Field("Name")
```

### FieldValue

Returns a reflect.Value of the orm caller struct field. It returns the zero Value if no field was found.

```go 
rv := scope.FieldValue("Name")
```

### SQLRelation

Will return teh requested relation by permission. Relations(s) which are defined as "custom" or have not the required
Permission will not be returned. Error will return if the relation does not exist or has not the required permission.

```go 
relation,err := scope.SQLRelation("Address",Permission{Read:true})
```

### SQLRelations

SQLRelations will return all sql relations by the given Permission. Relation(s) which are defined as "custom" or have
not the required Permission will not be returned.

```go 
relations := scope.SQLRelations(Permission{Read:true})
```

### PrimaryKeysSet

Checks if all primaries have a non zero value.

```go 
valid := scope.PrimaryKeysSet()
```

### PrimaryKeys

Will return all defined primary keys of the struct. Error will return if none was defined.

```go 
primaryFields,err := scope.PrimaryKeys()
if err!=nil{
  // ...
}
```

### SoftDelete

Will return the soft deleting struct.

```go 
sd := scope.SoftDelete()
```

### Parent

Parent returns the parent model by name or the root model if the name is empty. The name must be the orm struct name
incl. namespace. Error will return if no parent exists or the given name does not exist. The max search depth is limited
to 20.

```go 
model,err := scope.Parent("User")
```

### SetParent

```go 
scope.SetParent(model)
```

### IsEmpty

Checks if all the orm model fields and relations are empty.

```go 
valid := scope.IsEmpty(Permission{Read:true})
```

### IsSelfReferenceLoop

IsSelfReferenceLoop checks if the model has a self reference loop.

Animal (hasOne) -> Address (belongsTo) -> *Animal

```go 
valid := scope.IsSelfReferenceLoop(relation)
```

### IsSelfReferencing

IsSelfReferencing is a helper to check if the model caller has the same type as the given field type.

Role.Roles (m2m) -> Role

```go 
valid := scope.IsSelfReferenceLoop(relation)
```

### TakeSnapshot

TakeSnapshot will define if a snapshot of the orm model will be taken. This is used mainly in update.

### AppendChangedValue

AppendChangedValue adds the changedValue if it does not exist yet by the given field name.

### SetChangedValues

SetChangedValues sets the changedValues field of the scope. This is used to pass the values to a child orm model.

### ChangedValueByFieldName

ChangedValueByFieldName returns a *changedValue by the field name. Nil will return if it does not exist.

### InitRelationByField

InitRelationByField will return the orm.Interface of the field.

ptr * = if the value was nil, a new orm.Interface gets set, if its not nil, the value will be taken.

struct = * of that struct

ptr *[], [] = new orm.Interface

### InitRelation

InitRelation initialize the given relation. The orm model parent will be set, config, permission list and tx will be
passed.

### SetBackReference

SetBackReference will set a backreference if detected.

### NewScopeFromType

Will return a new scope of the given type.

## Strategy

Every orm model can have its own loading strategy. By default `eager` is defined.

### Eager

The eager loading strategy will load the root orm and all relations at once.

It is possible to only load required data by setting the field/relation
permissions.  [see Permissions](orm.md#permissions).

The following will describe the internal logic of the eager strategy. This information is only interesting for the
framework contributors.

#### First

First will return one row by the given condition. If a soft delete field is defined, by default only the "not soft
deleted" rows will be shown. This can be changed by config. If a HasOne relation returns no result, an error will
return. This can be changed by config. Only fields with the read permission will be read. Error (sql.ErrNoRows) returns
if First finds no rows.

***HasOne, BelongsTo:*** will call orm First().

***HasMany, ManyToMany*** will call orm All().

#### All

All rows by the given condition will be fetched. All foreign keys are collected after the main select, all relations are
handled by one request to minimize the db queries. m2m has actual 3 selects to ensure a different db builder could be
used. The data is mapped automatically afterwards. Only fields with the read permission will be read.

*** TODO***  Back-Reference only works for First -> All calls at the moment.

#### Create

Create a new entry.

***BelongsTo:*** will be skipped on empty or if a self reference loop is detected. Otherwise the entry will be created
and the reference field will be set. If the belongsTo primary key(s) are already set, it will update the entry instead
of creating it (if the pkey exists in the db). There is an option to only update the reference field without creating or
updating the linked entry. (belongsTo, manyToMany)
Only fields with the write permission will be written.

***Field(s):*** will be created and the last inserted ID will be set to the model.

***HasOne:***

If the value is zero it will be skipped.

The reference keys (fk and poly) will be set to the child orm and will be created.

***HasMany:***

If the value is zero it will be skipped.

If the relations has no sub relations, a batch insert is made to limit the db queries.

If relations exists, a normal Create will happen.

In both cases, the reference keys (fk and poly) will be set to the child orm and will be created.

***ManyToMany:***

If the value is zero it will be skipped.

If the primary key(s) are already set, it will update the entry instead of creating it (if the pkey exists in the db).

The junction table will be filled automatically.

There is an option to only update the reference field without creating or updating the linked entry.

#### Update

Update entry by the given condition. Only fields with the wrote permission will be written. There is an option to only
update the reference field without creating or updating the linked entry. (BelongsTo, ManyToMany)
Only changed values will be updated. A Snapshot over the whole orm is taken before.

***BelongsTo*** :

* CREATE: create or update (pk exist) the orm model.
* UPDATE: Update the parent orm model.
* DELETE: Only the reference is deleted at the moment.

***Field(s):*** gets updated if the value changed.

***HasOne:***

* CREATE: set reference IDs to the child orm, delete old references (can happen if a user add manually a struct), create
  the new entry.
* UPDATE: set reference IDs to the child orm, update the child orm.
* DELETE: delete the child orm. (query deletion happens - performance, TODO: call orm.Delete() to ensure soft delete?)

***HasMany:***

* CREATE: create the entries.
* UPDATE: the changed value entry is defined in the following categories.

  		- CREATE: slice entries gets created.
  		- UPDATE: slice entries gets updates.
  		- DELETE: all IDs gets collected and will be deleted as batch to minimize the db queries.
* DELETE: entries will get deleted by query.(query deletion happens - performance, TODO: call orm.Delete() to ensure
  soft delete?

***ManyToMany:***

* CREATE: Create or update (if pk is set and exists in db) the slice entry. the junction table will be batched to
  minimize the db queries.
* UPDATE: the changed value entry is defined in the following categories.

  		- CREATE: slice entries gets created or updated (if pk is set and exists in db). the junction table will be batched.
  		- UPDATE: the slice entry.
  		- DELETE: collect all deleted entries. delete only in the junction table at the moment. the junction table will be batched. TODO: think about a strategy.
* DELETE: entries are only deleted by the junction table at the moment. TODO: think about a strategy.

### Create your own

To create your own strategy, you have to implement the `Strategy` interface.

```go
type Strategy interface {
First(scope Scope, c condition.Condition, permission Permission) error
All(res interface{}, scope Scope, c condition.Condition) error
Create(scope Scope) error
Update(scope Scope, c condition.Condition) error
Delete(scope Scope, c condition.Condition) error
Load(interface{}) Strategy
}
```

Use the `init` function to register your strategy by name

The registered value must be of the type `func(Strategy, error)`.

````go
func init() {
err := Register("yourStrategy", newStrategy)
if err != nil {
log.Fatal(err)
}
}

// newEager returns the orm.Strategy.
func newStrategy() (Strategy, error) {
return &something{}, nil
}
````

Now you can can access the strategy from your orm model by defining the `DefaultStrategy` function with the required
strategy.
