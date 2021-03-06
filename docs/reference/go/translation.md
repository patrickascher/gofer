# Translation

Package translation provides an i18n implementation for the back- and frontend.
For the backend [go-i18n](https://github.com/nicksnyder/go-i18n) is used in the back and for the frontend [vue-i18n](https://vue-i18n.intlify.dev/).

See the [controller section](controller.md#translation-t-tp) how to use it in the backend or the [frontend section](./../frontend/translate.md) for the frontend.

## Usage

Inspired by the `database/sql`, this module is based on providers. You have to import the needed provider with a dash in
front:
This will only call the init function(s) of the package(s) and will register itself. For a full list of all available
providers, see the [providers section](translation.md#providers).

```go 
import "github.com/patrickascher/gofer/locale/translation"
import _ "github.com/patrickascher/gofer/locale/translation/db" // example for the database provider
```

If you are using the skeleton app, the translation will be added automatically if it's configured in the `config.json`.

Otherwise, you can create a new manager instance like this:

```go 
manager, err = translation.New(translation.DB, nil, translation.Confgi{DefaultLanguage:"en",Controller:true})
//..
```

## Add raw messages

To add a raw message, simply use the `translation.AddRawMessage` function.
This must be added at an early stage of the application. One way is to add it by the `init` function or before the server is started.

```go 
func init() {
    translation.AddRawMessage(i18n.Message{ID: "Title", Description: "The application title"})
}
```

## Config

```go 
// Config for the translation.
type Config struct {
	// Controller - if enabled, translations will be available in the controller.
	Controller bool
	// JSONFilePath - if not zero, JSON files will be generated for each defined language.
	JSONFilePath string
	// DefaultLanguage - Default language of the application.
	DefaultLanguage string
}
```

### Controller

A controller with a full functional CRUD for the translation is available `locale.Controller`.

#### AddRoutes

For an easy migration you can use the `AddRoutes` function to add the translation routes to your router. Two routes will
be added:

* the CRUD router for the frontend.
* the added JSON directory (if enabled by config)

```go 
locale.AddRoutes(yourRouter)
```

## Providers

All pre-defined providers:

### DB

A wrapper for [httprouter](https://github.com/julienschmidt/httprouter).

Name:

`translation.DB`

Options:

no options are available at the moment.

Usage:

```go 
import "github.com/patrickascher/gofer/locale/translation"
import _ "github.com/patrickascher/gofer/locale/translation/db"

err := translation.New(translation.DB, nil,translation.Config{})
```

## Create your own provider

To create your own provider, you have to implement the `translation.Provider` interface.

```go
// Provider interface.
type Provider interface {
Bundle() (*i18n.Bundle, error)
Languages() ([]language.Tag, error)
JSON(path string) error
AddRawMessage([]i18n.Message) error
DefaultMessage(id string) *i18n.Message
SetDefaultLanguage(language.Tag)
}
```

Use the `init` function to register your provider.

The registered value must be of the type `func(m Manager, options interface{}) (router.Provider, error)`.

```go 
// init register your config provider
func init() {
    //...
	err := translation.Register("my-provider",New)
    if err != nil {
    	log.Fatal(err)
    }
}

func New(m Manager, options interface{}) (translation.Provider, error){
    //...
    return provider,nil
}
```

**Usage**

```go 
import "github.com/patrickascher/gofer/router"
import _ "your/repo/router/yourProvider"

err := translation.New("my-provider", nil,translation.Config{})
```

