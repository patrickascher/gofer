# Structer

Package structer providers some util functions for structs.

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
