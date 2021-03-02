# Registry

Package registry provides a simple container for values in the application space.

## Usage

The registry name `string` and registry value `interface{}` argument must have a non-zero value, and the registered name
must be unique, otherwise an error will return.

If a validator is registered, and the registry name matches any `Validate.Prefix`, it will be checked
against `Validate.Fn` before the value will be added to the registry.

### Set

```go 
import "github.com/patrickascher/gofer/registry"

err := registry.Set("Version","0.1")
```

### Get

```go 
import "github.com/patrickascher/gofer/registry"

value,err := registry.Get("Version")
// Output value
// 0.1
```

### Prefix

Prefix will return all entries which name starts with the given prefix.

```go 
import "github.com/patrickascher/gofer/registry"
err := registry.Set("export_json","")

values := registry.Prefix("export_")
//...
```

### Validator

Because of the value type `interface{}`, any type can be registered as value. Sometimes it makes sense to check the
value before its getting added (e.g. against a type).

The `Validator` function requires a `Validate` struct as argument. It must be defined with `Validate.Prefix string`
and  `Validate.Fn func(name string, value interface{}) error`.

The `Validate.Prefix` must be unique, otherwise an error will return. The `Validate.Fn` will receive the registry name
and registry value as arguments.

Now before any value will be added to the registry, it will be checked against the `Validate.Fn` if the registry name
matches the `Validate.Prefix`.

```go 
import "github.com/patrickascher/gofer/registry"

err = registry.Validator(registry.Validate{Prefix: "test_", Fn: func(name string, value interface{}) error {
		if reflect.TypeOf(value).Kind() != reflect.String {
			return errors.New("wrong type")
		}
		return nil
	}})
//...

// test_foo matches the registeres prefix test_. 
// The custom type check is ok.
err := registry.Set("test_foo","ok")
//... 

// test_bar matches the registeres prefix test_.
// The custom type check throws an error, because its no string.
err := registry.Set("test_bar",false)
// ...
```

## Examples

### Type casting

You can add any type as value, as long as it is not nil. In the following example we are going to add a function as
value.

!!! tip 

    If you are adding a function as reference (without braces), the variables/objects of the function will only be allocated on function call. Like this, the memory will only be allocated, when needed!

```go 
import "github.com/patrickascher/gofer/registry"

type Config struct{
    Debug bool
}

// The function we are going to add
func Debug(cfg Config) bool {
	return cfg.Debug
}

// set the new registry "dummyFunc" with the function as reference
err := registry.Set("dummyFunc",New)
//...

// getting the "dummyFunc" registry
fn,err := registry.Get("dummyFunc")
//...

// casting the function and call it with the config argument
output = fn.(func(Config) bool)(Config{Debug:true})
// output: true
```
