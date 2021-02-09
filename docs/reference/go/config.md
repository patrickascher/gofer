# Config

Package config provides a config manager for any type that implements the `config.Interface`. It will load the parsed
values into a configuration struct.

Supports JSON, TOML, YAML, HCL, INI, envfile and Java properties config files (viper provider). Every provider has its
own options, please see the [providers section](config.md#providers) for more details.

It also offers automatically config reload if the underlaying file changes and offers a callback function.

## Usage

Inspired by the `database/sql`, this module is based on providers. You have to import the needed provider with a dash in
front:
This will only call the init function(s) of the package(s) and will register itself. For a full list of all available
providers, see the [providers section](config.md#providers).

```go 
import "github.com/patrickascher/gofer/config"
import _ "github.com/patrickascher/gofer/config/viper" // example for the viper provider
```

### Load

``` go
type Config struct{
    Host string
    User string
    //....
}

// provider options
options := viper.Options{FileName:"config.json",FilePath:".",FileType:"json"}
// config struct
cfg := Config{}

// Load the configuration.
err = config.Load(config.VIPER,&cfg,options)

// cfg will have the loaded values

```

## Providers

All pre-defined providers:

### Viper

A wrapper for [viper](https://github.com/spf13/viper).

This provider can load JSON, TOML, YAML, HCL, INI, envfile and Java properties config files.

A watcher can be added to reload the configuration file on changes. Additional callback can be set.

!!! Info By default, the configuration struct will be updated if the config file changes.

Name:

`config.VIPER`

Options:

| Option      | Description                          |
| ----------- | ------------------------------------ |
| `FileName`       | `string` Name of the config file. **Mandatory**  |
| `FileType`       | `string` Type of the config file. **Mandatory**  |
| `FilePath`       | `string` Path of the config file. Use `.` for the working directory. **Mandatory**  |
| `Watch`       | `bool` Watch the config file for changes.  |
| `WatchCallback`       | `func(cfg interface{}, viper *viper.Viper, e fsnotify.Event)` Callback function which will be triggered if the watcher is activated and a file change happens.  |
| `EnvPrefix`       | `string` Env prefix.  |
| `EnvAutomatic`       | `bool` All env variables will be mapped if the config struct key exists.  |
| `EnvBind`       | `[]string` Only the defined env variables will be mapped if they exist in the config struct.  |

Usage:

```go 
import "github.com/patrickascher/gofer/config"
import _ "github.com/patrickascher/gofer/config/viper"

type Config struct{
    Host string
    User string
    //....
}

// options
options := viper.Options{
    FileName:"config.json",
    FilePath:".",
    FileType:"json",
    Watch:true,
    WatchCallback:fileChangedCallback
}

func fileChangedCallback(cfg interface{}, viper *viper.Viper, e fsnotify.Event){
    // on change, do something
}

// config struct
cfg := Config{}

// Load the configuration.
err = config.Load(config.VIPER,&cfg,options)
// cfg will have the loaded values


```

## Create your own provider

To create your own provider, you have to implement the `config.Interface`.

```go
type Interface interface {
	Parse(config interface{}, options interface{}) error
}
```

Use the `init` function to register your provider.

The registered value must be of the type `config.Interface`.

```go 
// init register your config provider
func init() {
	err := registry.Set("your-config-provider", new(yourProvider))
    if err != nil {
    	log.Fatal(err)
    }
}

type yourProvider struct{}
	
// New creates a super-fast-memory type which implements the cache.Interface.
func (p *yourProvider) Parse(cfg interface{}, opt interface{}) error {
    // parse something
}
```

**Usage**

```go 
import "github.com/patrickascher/gofer/cache"
import _ "your/repo/config/yourProvider"

type Config struct{
    Host string
    User string
    //...
}

// configuration struct
cfg := Config{}

// your provider options, if you dont need any, just add nil as argument.
options := YourProviderOptions{}

err := config.Load("your-config-provider", &cfg, options)
```

