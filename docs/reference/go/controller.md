# Controller

Package controller provides a controller / action based `http.Handler` for the [router](router.md).

A controller can have different renderer and is easy to extend.

Data, redirects and errors can be set directly in the controller. A context with some helpers for the response and
request is provided.

## Usage

Controller can be easy extended and added as route. You only need to embed the `controller.Base`.

```go 
type MyController struct{
    constroller.Base
}

func (c MyController) User(){
    // logic
} 

// add controller to router with the mapping HTTP GET calls MyController.User()
myController := MyController{}
err = r.AddSecureRoute(router.NewRoute("/user", &myController, router.NewMapping([]string{http.MethodGet}, myController.User, nil)))
```

## Initialize

Initialize is used to set the struct reference. This is called automatically by the router.

```go
c := MyController{}
c.Initialize(&c)
```

## Action

Return the requested Action name.

## Name

Returns the controller name incl. the package name.

## RenderType

Returns the render type. Default `json`.

## SetRenderType

Set a custom render type.

## Set

Is a helper to set controller variables. It wraps the [context.Response.SetValue](controller.md#setvalue).

## Error

Is a helper to set an error. If an error ist set, all defined values will be deleted. It wraps
the [context.Response.Error](controller.md#error_1).

## Redirect

Redirect to an URL. On a redirect the old controller data will be lost.

## ServeHTTP

Implements the `http.Handler` interface.

It creates a new instance of the requested controller and creates a new [context](controller.md#context).

It checks if a action name is available and checks if a function with that name exists and calls it. If not, an error
will return.

Between the `Action` and `Render call, it checks if the Browser is still connected. Otherwise the request will be
cancelled.

If the controller has no custom set [Error](controller.md#error), the [Render](controller.md#render) function will be
called.

## SetContext

Set a context to the controller.

## Context

Context provides some useful helper for the `Request` and `Response`.

### Request

#### Body

Returns the raw body data.

#### SetBody

Can be used to manipulate the body data.

#### Localizer

Returns the controller localizer.

#### Pattern

Returns the router url pattern.

Example: `http://example.com/user/1` will return `/user/:id`

#### HTTPRequest

Returns the original `*http.Request`.

#### JWTClaim

A helper to return the claim set by the request context.

#### Method

Returns the HTTP method in uppercase.

#### Is

Compares the given method with the request HTTP method.

```go
if r.Is(http.MethodGet){
//...
}
```

#### IsSecure

Returns true if the request is `https`.

#### IsGet

Checks if the request is a `http.MethodGet`.

#### IsPost

Checks if the request is a `http.MethodPost`.

#### IsPatch

Checks if the request is a `http.MethodPatch`.

#### IsPut

Checks if the request is a `http.MethodPut`.

#### IsDelete

Checks if the request is a `http.MethodDelete`.

#### Param

Returns a parameter by key. It returns a `[]string` because the underlying HTML input field could be an array. Error
will return on internal error or if the key does not exist.

```go
p, err := r.Param("user")
```

#### Params

Returns all existing parameters. It returns a `map[string][]string` because the underlying HTML input field could be an
array. Error will return on internal error.

#### IP

IP of the request.

#### Proxy

Return all IPs which are in the X-Forwarded-For header.

#### Scheme

Scheme (http/https) checks the `X-Forwarded-Proto` header. If that one is empty the URL.Scheme gets checked. If that is
also empty the request TLS will be checked.

#### Host

Returns the host name. Port number will be removed. If no host info is available, localhost will return.

`https://example.com:8080/user?id=12#test` will return `example.com`.

#### Protocol

Returns the protocol name, such as HTTP/1.1.

#### URI

Returns full request url with query string fragment.

`https://example.com:8080/user?id=12#test` will return `/user?id=12#test`.

#### URL

Returns request url path without the query string and fragment.

`https://example.com:8080/user?id=12#test` will return `/user`

#### FullURL

Returns the schema,host,port,uri.

`https://example.com:8080/user?id=12#test` will return `https://example.com:8080/user?id=12#test`.

#### Site

Returns base site url as `scheme://domain` type without the port.

`https://example.com:8080/user?id=12#test` will return `https://example.com`.

#### Domain

Is an alias for [Host](controller.md#host).

#### Port

Will return the port of the request. If it is empty, 80 will return as default.

#### Referer

Returns the Referer Header.

### Response

Is a helper to set data and to render the content. A custom
render [provider](controller.md#create-your-own-render-provider) can be created, simply implement the `Renderer`
interface.

#### SetValue

Set a value by key/value pair.

#### Value

Value by the key. If the key does not exist, nil will return.

#### Values

Returns all defined values.

#### ResetValues

Reset all defined values.

#### Writer

Returns the `*http.ResponseWriter`.

#### Render

Render will render the content by the given render type. An error will return if the render provider does not exist or
the renders write function returns one.

#### Error

Error will render the error message by the given render type. An error will return if the render provider does not exist
or the renders error function returns one.

## Create your own render provider

To create your own render provider, you have to implement the `controller.Renderer` interface.

### CheckBrowserCancellation, CallAction, HasError

TODO docu, just added because of ther unexported interface and mocking problem.

```go
// Renderer interface for the render providers.
type Renderer interface {
Name() string
Icon() string
Write(response *Response) error
Error(response *Response, code int, err error) error
}
```

Use the `init` function to register your provider.

```go 
// init register your render provider
func init() {
    //...
	err := controller.RegisterRenderer("xml",newXmlRenderer)
    if err != nil {
    	log.Fatal(err)
    }
}

type xmlRenderer struct{}

func newXmlRenderer() (Renderer, error) {
	return &xmlRenderer{}, nil
}

// ... render functions
```

**Usage**

```go 
import "github.com/patrickascher/gofer/controller"
import _ "your/repo/controller/renderer"

// somewhere in your controller
c.SetRenderType("xml")
```

