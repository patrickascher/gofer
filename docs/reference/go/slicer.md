# Slicer

Package slicer providers some util functions for slices.

## InterfaceExists

Checks if the given interface exists in a slice. If it exists, a the position and a boolean `true` will return

```go

slice := []interface{}{1, 2}
pos, exists := slicer.InterfaceExists(slice, 1)
// pos: 0, exists:true
```

## StringPrefixExists

Checks if the given prefix exists in the string slice. If it exists, a slice with all matched results will return.

```go

cache := []string{"orm_User", "orm_Address"}
result := slicer.StringPrefixExists(cache, "orm_")
// result: []string{"orm_User", "orm_Address"}
```

## StringExists

Checks if the given string exists in the string slice. If it exists, the position and a boolean `true` will return

```go

cache := []string{"orm_User", "orm_Address"}
result := slicer.StringPrefixExists(cache, "orm_User")
// result: 0,true
```

## StringUnique

Will unique all strings in the given slice.

```go

cache := []string{"orm_User", "orm_User","orm_Address"}
result := slicer.StringUnique(cache
// result:[]string{"orm_User", "orm_Address"}
```
