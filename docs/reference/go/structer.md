# Structer

Package structer providers some util functions for structs.


## Merge, MergeByMap
Are functions to merge a s struct by another struct or by a map.
It`s a wrapper for [mergo](https://github.com/imdario/mergo). For more the options check out the repository.

!!! info

    Use `mergo.WithOverride` options to overwrite dest values.


```go

type Foo struct {
    A string
    B int
}

src := Foo{
    A: "one",
    B: 2,
}
dest := Foo{
    A: "two",
}

structer.Merge(&dest,src)
// result: A:two, B:2
```


## ParseTag

ParseTag will parse the tag to key/value pairs `map[string]string`.

Multiple entries can be separated by `;`
Values are optionally.

```go

// `orm:primary`            // will be parsed to: map[primary]""
// `orm:primary; column:id` // will be parsed to: map[column]"id"

sfield := reflect.StructField{}
structer.ParseTag(field.Tag("orm"))
```
