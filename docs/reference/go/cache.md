# Cache

Package cache provides a cache manager for any type that implements the `cache.Interface`. It is developer friendly,
offers additional prefixing and functions for hit/miss statistics.

## Usage

Inspired by the `database/sql`, this module is based on providers. You have to import the needed provider with a dash in
front:
This will only call the init function(s) of the package(s) and will register itself. For a full list of all available
providers, see the [providers section](cache.md#providers).

```go 
import "github.com/patrickascher/gofer/cache"
import _ "github.com/patrickascher/gofer/cache/memory"
```

### New

The `New` function requires two arguments. First the name of the registered provider and a provider configuration. Each
provider will have different configuration settings, please see the [providers section](cache.md#providers) for more
details.

Error will return if the provider was not found or registered before.

```go 
// Example of the memory cache provider
mem, err := cache.New(cache.MEMORY, nil)
```

### SetDefaultPrefix

Set a default prefix for all cache items which will be set with `cache.DefaultPrefix`.

```go 
mem.SetDefaultPrefix("session")
```

### SetDefaultExpiration

Set a default expiration for all cache items which will be set with `cache.DefaultExpiration`.

```go 
mem.SetDefaultExpiration(5*time.Hour)
```

### Exist

Exist wraps the `Get()` function and will return a boolean instead of an error.

```go 
exists := mem.Exist("name")  // Boolean
```

### Get, Prefix, All

Get one or more cached items. Error will return if the cache item / prefix does not exist.

```go 
// get a single cached item.
item,err := mem.Get(cache.DefaultPrefix,"name")  // (Item, error)

// get all cached items with the prefix "session"
item,err := mem.Prefix("session") // ([]Item, error)

// get all cached items
item,err := mem.All()    // ([]Item, error)
```

### Set

Set a cache item by prefix, name, value and expiration.

!!! info "Infinity live time"
If you need no expiration for a cache item, use `cache.NoExpiration`.

```go 
// the managers default expiration time. 
err := mem.Set(cache.DefaultPrefix,"name","value",cache.DefaultExpiration)

// custom expiration time and prefix.
err := mem.Set("session","name","value",5*time.Hour)

// no expiration.
err := mem.Set(cache.DefaultPrefix,"name","value",cache.NoExpiration)
```

### Delete, DeletePrefix, DeleteAll

Delete one or more cached items. Error will return if the cache item / prefix does not exist.

```go 
// delete the item by key
err := mem.Delete(cache.DefaultPrefix,"name")

// delete all items with the prefix "session"
err := mem.DeletePrefix("session")

// delete all cached items
err := mem.DeleteAll()
```

### HitCount

Statistics how often the cache item was hit.

```go 
err := mem.HitCount(cache.DefaultPrefix,"user")
```

### MissCount

Statistics how often the cache item was missed.

```go 
err := mem.MissCount(cache.DefaultPrefix,"user")
```

## Providers

All pre-defined providers:

### Memory

A simple in memory cache.

Name:

`cache.MEMORY`

Options:

| Option      | Description                          |
| ----------- | ------------------------------------ |
| `GCInterval`       | time.Duration how often the GC should run in a loop.  |

Usage:

```go 
import "github.com/patrickascher/gofer/cache"
import _ "github.com/patrickascher/gofer/cache/memory"

mem, err := cache.Manager(cache.MEMORY, nil)

```

## Create your own provider

To create your own provider, you have to implement the `cache.Interface`.

```go
type Interface interface {
// Get returns an Item by its name.
// Error must returns if it does not exist.
Get(name string) (Item, error)
// All cached items.
// Must returns nil if the cache is empty.
All() ([]Item, error)
// Set an item by its name, value and lifetime.
// If cache.NoExpiration is set, the item should not get deleted.
Set(name string, value interface{}, exp time.Duration) error
// Delete a value by its name.
// Error must return if it does not exist.
Delete(name string) error
// DeleteAll items.
DeleteAll() error
// GC will be called once as goroutine.
// If the cache backend has its own garbage collector (redis, memcached, ...) just return void in this method.
GC()
} 
```

Use the `init` function to register your provider.

The registered value must be of the type `func(interface {}) (cache.Interface, error)`.

```go 
// init register the superFastMemory provider.
func init() {
	err := cache.Register("superFastMemory", New)
	if err != nil{
		log.Fatal(err)
	}
}
	
// New creates a super-fast-memory type which implements the cache.Interface.
func New(opt interface{}) (cm.Interface,error) {
    //...
    return &superFastMemory{},nil
}
```

**Usage**

```go 
import "github.com/patrickascher/gofer/cache"
import _ "your/repo/cache/superFastMemory"

memoryProvider, err := cache.Manager("superFastMemory", nil)
```
