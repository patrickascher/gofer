# Query

The package query provides a simple programmatically sql query builder. The idea was to create a unique query builder
which can be used with any database driver in go - with minimal effort.

Features:

* Unique Placeholder for all database drivers
* Batching function for large Inserts
* Whitelist
* automatic quote of column and table names.
* SQL queries and durations log debugging

## Usage

Inspired by the `database/sql`, this module is also based on providers. You have to import the needed provider with a
dash in front:
This will only call the init function(s) of the package(s) and will register itself. For a full list of all available
providers, see the [providers section](query.md#providers).

```go 
import "github.com/patrickascher/gofer/query"
import _ "github.com/patrickascher/gofer/query/mysql"
```

### New

To create a builder instance, you have to call `New()` with the needed provider name and configuration. Please see
the [providers section](query.md#providers) for a full list.

```go
builder, err := query.New("mysql", query.Config{})
```

### SetLogger

A `logger.Manager` [logger](logger.md) can be added to the builder. If one is defined, all queries will be logged
on `DEBUG` level.

```go
builder.SetLogger(logManager)
```

### Config

Will return the `query.Config`.

```go
builder.Config()
```

### QuoteIdentifier

Will quote an identifier by the providers quote tags.

```go
name := builder.QuoteIdentifier("test")
// on mysql name will be `test`
```

### Null types

The following null type are defined for dealing with nullable SQL and JSON values.

Following helper functions are defined `NewNullString(string,valid)`, `NewNullBool(bool,valid)`
,`NewNullInt(int64,valid)`,`NewNullFloat(float64,valid)` and `NewNullTime(time.Time,valid)`.

!!! info It' s a type alias for [https://pkg.go.dev/github.com/guregu/null](https://pkg.go.dev/github.com/guregu/null)

* query.NullString
* query.NullBool
* query.NullInt
* query.NullFloat
* query.NullTime

### SanitizeValue

SanitizeValue is a helper which converts:

* `int`,`int8`,`int16`,`int32`,`int64`,`uint`,`uint8`,`uint16`,`uint32`,`uint64`,`query.NullInt` to `int64`
* `string`,`query.NullString` to `string`

Error will return if the argument is not of the described types, or a NullType is not valid.

```go

value,err := query.SanitizeValue(1) 
// value will be int64(1) and err will be nil

```

### Query

To create any query you have to call the query function

```go
builder.Query()
```

#### Select

##### Columns

Columns define a fixed column order for the insert. If the columns are not set manually, `*` will be used. Only Values
will be inserted which are defined here. This means, you can use Columns as a whitelist.

```go
b.Query().Select("test").Column("name", "surname")
```

##### Condition

Condition adds your own condition to the query.

```go
c := condition.New().SetWhere("id = ?", 1)
b.Query().Select("test").Condition(c)
```

##### Join

Join wraps the `condition.SetJoin()` function. For more details see [condition section](query.md#setjoin-join).

```go
b.Query().Select("test").Join(condition.LEFT, "test_relation", "test.id = test_relation")
```

##### Where

Where wraps the `condition.SetWhere()` function. For more details see [condition section](query.md#setwhere-where).

```go
b.Query().Select("test").Where("id = ?", 1)
```

##### Group

Group wraps the `condition.SetGroup()` function. For more details see [condition section](query.md#setgroup-group).

```go
b.Query().Select("test").Group("name", "surname")
```

##### Having

Having wraps the `condition.SetHaving()` function. For more details see [condition section](query.md#sethaving-having).

```go
b.Query().Select("test").Having("id = ?", 1)
```

##### Order

Order wraps the `condition.SetOrder()` function. For more details see [condition section](query.md#setorder-order).

```go
b.Query().Select("test").Order("name", "-surname")
```

##### Limit

Limit wraps the `condition.SetLimit()` function. For more details see [condition section](query.md#setlimit-limit).

```go
b.Query().Select("test").Limit(10)
```

##### Offset

Offset wraps the `condition.SetOffset()` function. For more details see [condition section](query.md#setoffset-offset).

```go
b.Query().Select("test").Offset(20)
```

##### String

String returns the rendered statement and arguments.

```go
sql, args, err := b.Query().Select("test").String()
```

##### First

First will return a sql.Row.

```go
row, err := b.Query().Select("test").Column("name", "surname").First()
// SELECT `name`, `surname` FROM test
```

##### All

All will return sql.Rows.

```go
res, err := b.Query().Select("test").Column("name", "surname").All()
// SELECT `name`, `surname` FROM test
```

#### Insert

##### Batch

Batch sets the batching size. Default batching size is 50.

```go
b.Query().Insert("test").Batch(20)
```

##### Columns

Columns define a fixed column order for the insert. If the columns are not set manually, all keys of the Values will be
added. Only Values will be inserted which are defined here. This means, you can use Columns as a whitelist.

```go
b.Query().Insert("test").Columns("name")
```

##### Values

Values sets the insert data.

```go
values := []map[string]interface{}{{"name": "John"}}
b.Query().Insert("test").Columns("name").Values(values)
```

##### LastInsertedID

LastInsertedID gets the last id over different drivers. The first argument must be a ptr to the value field. The second
argument should be the name of the ID column - if needed.

```go
var id int
b.Query().Insert("test").Columns("name").LastInsertedID(&id)
```

##### String

String returns the rendered statement and arguments.

```go
values := []map[string]interface{}{{"name": "John"}}
b.Query().Insert("test").Columns("name").Values(values).String()
```

##### Exec

Exec the statement. It will return a slice of `[]sql.Result` because it could have been batched.

```go
values := []map[string]interface{}{{"name": "John"}}
res, err := b.Query().Insert("test").Columns("name").Values(values).Exec()
```

#### Update

##### Columns

Columns define a fixed column order for the insert. If the columns are not set manually, all keys of the Values will be
added. Only Values will be inserted which are defined here. This means, you can use Columns as a whitelist.

```go
b.Query().Update("test").Columns("name")
```

##### Set

Set the values.

```go
b.Query().Update("test").Set(map[string]interface{}{"name": "John"})
```

##### Where

Where wraps the `condition.SetWhere()` function. For more details see [condition section](query.md#setwhere-where).

```go
b.Query().Update("test").Set(map[string]interface{}{"name": "John"}).Where("id = ?", 1)
```

##### Condition

Condition adds your own condition to the query.

```go
c := condition.New().SetWhere("id = ?", 1)
b.Query().Update("test").Set(map[string]interface{}{"name": "John"}).Condition(c)
```

##### String

String returns the rendered statement and arguments.

```go
sql, args, err := b.Query().Update("test").Set(map[string]interface{}{"name": "John"}).String()
```

##### Exec

Exec the statement. It will return a `sql.Result`.

```go
res, err := b.Query().Update("test").Set(map[string]interface{}{"name": "John"}).Exec()
```

#### Delete

##### Where

Where wraps the `condition.SetWhere()` function. For more details see [condition section](query.md#setwhere-where).

```go
b.Query().Delete("test").Where("id = ?", 1)
```

##### Condition

Condition adds your own condition to the query.

```go
c := condition.New().SetWhere("id = ?", 1)
b.Query().Delete("test").Condition(c)
```

##### String

String returns the rendered statement and arguments.

```go
sql, args, err := b.Query().Delete("test").String()
```

##### Exec

Exec the statement. It will return a `sql.Result`.

```go
res, err := b.Query().Delete("test").Exec()
```

#### Information

##### Describe

Describe the defined table.

```go
// all columns
cols, err := b.Query().Information("test").Describe()

// only some columns
cols, err := b.Query().Information("test").Describe("name", "surname")
```

##### ForeignKey

ForeignKey will return the foreign keys for the defined table.

```go
fks, err := b.Query().Information("test").ForeignKey()
```

## DbExpr

DBExpr is a helper to avoid quoting. Every string which is wrapped in `query.DbExrp("test")` will not get quoted by the
builder.

## Condition

Condition provides a sql condition builder. The placeholder `?` must be used and will automatically replaced with the
driver placeholder later on.

### New

Will create a new condition instance.

```go
c := condition.New()
```

### SetJoin, Join

SetJoin will create a sql JOIN condition.
`LEFT`, `RIGHT`, `INNER` and `CROSS` are supported. SQL USING() is not supported at the moment. If the join type is
unknown or the table is empty, an error will be set.

```go 
c.SetJoin(condition.LEFT,"users","users.id = ?",1)
```

Join will return the defined conditions as `condition.Clause`. On a clause, you can receive the condition and passed
arguments.

```go
clauses := c.Join()
clause[0].Condition() // would return `JOIN LEFT users ON users.id = ?`
clause[0].Arguments() // would return `[]int{1}`
```

### SetWhere, Where

SetWhere will create a sql WHERE condition. When called multiple times, its getting chained by AND operator.

Arrays and slices can be passed as argument.

```go 
c.SetWhere("id = ? AND name = ?",1,"John")
c.SetWhere("id IN (?)",[]int{10,11,12}) // will render the condition into `id IN (?, ?, ?)`
```

Where will return the defined conditions as `condition.Clause`. On a clause, you can receive the condition and passed
arguments.

```go
clauses := c.Where()
clause[1].Condition() // would return `id IN (?, ?, ?)`
clause[1].Arguments() // would return `[]int{10,11,12}`
```

### SetGroup, Group

SetGroup should only be called once. If it's called more often, the last values are set.

```go 
c.SetGroup("id","name")
```

Group will return a slice of a string with all the added values.

### SetHaving, Having

`SetHaving` and `Having` have the same functionality as `SetWhere` and `Where`.

### SetOrder, Order

SetOrder should only be called once. If a column has a `-` prefix, DESC order will get set. If it's called more often,
the last values are set.

```go
c.SetOrder("name", "-surname") // rendered into `ORDER BY name ASC, surname DESC`
```

### SetLimit, Limit

Set or get the sql `LIMIT`.

```go
c.SetLimit(10)
```

### SetOffset, Offset

Set or get the sql `OFFSET`.

```go
c.SetOffset(2)
```

### Reset

Reset the complete condition or only single parts.

```go
// complete condition will be reset.
c.Reset()

// only the WHERE conditions will be reset.
c.Reset(condition.Where)
```

### Merge

Merge two conditions.
`Group`, `Offset`, `Limit` and `Order` will be set if they have a none zero value instead of merged, because they should
only be used once.
`Where`, `Having` and `Join` will be merged, if exist.

```go
a := conditon.New()
// ...
b := condition.New()
// ...

a.Merge(b)
```

### ReplacePlaceholders

ReplacePlaceholders is a helper function to replace the `condition.Placholder` `?` with any other placeholder.

```go

condition.ReplacePlaceholders("id = ? AND name = ?", Placeholder{Char:"$", Numeric:true})
// will return `id = $1 AND name = $2`
```

## Config

The basic sql config is required. If a provider needs some additional configuration, its no problem to embed this struct
but the providers function `Config()` must return this struct.

```go
type Config struct {
Username string
Password string
Host     string
Port     int
Database string

MaxIdleConnections int
MaxOpenConnections int
MaxConnLifetime    time.Duration
Timeout            string

PreQuery []string
}

```

## Providers

### Mysql

Mysql Provider which uses `github.com/go-sql-driver/mysql` under the hood.

Time structs will be parsed and a timeout limit is set to 30s by default.

#### Usage:

```go
import "github.com/patrickascher/gofer/query"
import _ "github.com/patrickascher/gofer/query/mysql"

query.New("mysql", config)
```
