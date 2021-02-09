# Stringer

Package stringer providers some util functions for strings.

## CamelToSnake

```go
stringer.CamelToSnake("GoTestExample")
// returns go_test_example
```

## SnakeToCamel

```go
stringer.CamelToSnake("go_test_example")
// returns GoTestExample
```

## Singular

```go
stringer.Singular("users")
// returns user
```

## Plural

```go
stringer.Plural("user")
// returns users
```
