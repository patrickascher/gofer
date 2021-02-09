# Router

Package router provides a manager to add public and secure routes based on a `http.Handler` or `http.HandlerFunc`.

Specific Action<->HTTP Method mapping can be defined.

Middleware can be added for each route or globally for all secured routes.

Files or directories can be added.

The `PATTERN`, `PARAMS`, `ACTION` and (`ALLOWED` HTTP Methods - only on OPTIONS) will be added as request context.

## Usage

Inspired by the `database/sql`, this module is based on providers. You have to import the needed provider with a dash in
front:
This will only call the init function(s) of the package(s) and will register itself. For a full list of all available
providers, see the [providers section](router.md#providers).

```go 
import "github.com/patrickascher/gofer/router"
import _ "github.com/patrickascher/gofer/router/httprouter" // example for the julienschmidt http router
```

### New

The New function requires two arguments. First the name of the registered provider and a provider configuration. Each
provider will have different configuration settings, please see the [providers section](router.md#providers) for more
details.

```go 
// get a new router instance.
routerManager,err := router.New(router.JSROUTER,nil)
```

### AllowHTTPMethod

Can be used to disable one or more HTTP methods globally. By Default: `TRACE` and `CONNECT` are disabled.

```go 
// will disable globally HTTP GET for any routes.
err = routerManager.AllowHTTPMethod(http.MethodGet,false)
```

### SetSecureMiddleware

Middleware(s) can be added. They will automatically apply to the `AddSecureRoute`(s).

```go 
routerManager.SetSecureMiddleware(mw)
```

### SetFavicon

Sets the fav icon. The pattern is `/favicon.ico`. If the source does not exist, an error will return.

```go 
err := routerManager.SetFavicon("assets/img/favicon.ico")
```

### AddPublicFile

The first argument is the pattern, and the second one is the source. The pattern must begin with a `/`. If the pattern
already exists, or the source does not exist, an error will return.

```go 
err := routerManager.AddPublicFile("/robot.txt","assets/static/robot.txt")
```

### AddPublicDir

The first argument is the pattern, and the second one is the source. Directories are not allowed on pattern root
level `/`. The pattern must begin with a `/`. If the pattern already exists, or the source does not exist, an error will
return.

```go 
err := routerManager.AddPublicDir("/images","assets/img")
```

### AddPublicRoute

A route can be added to the router. Please see the [route section](router.md#route) for more details.

```go 
err := routerManager.AddPublicRoute(router.NewRoute("/login", handleFunc))
```

### AddSecureRoute

A secure route can be added to the router. Please see the [route section](router.md#route) for more details.

If no secure middleware(s) are defined, an error will return.

```go 
err := routerManager.AddSecureRoute(router.NewRoute("/admin", handleFunc))
```

### Routes

All defined routes of the router will return.

```go 
routes := routerManager.Routes()
```

### RouteByPattern

The route by the given pattern will return. If the pattern does not exist, an error will return.

```go 
routes := routerManager.RouteByPattern("/favicon.ico")
```

### Handler

Returns the `http.Handler`.

```go 
handler := routerManager.Handler()
```

### SetNotFound

Set a custom handler for all routes which can not be found.

```go 
routerManager.SetNotFound(hanlder)
```

## Route

A new route can be created with `router.NewRoute(pattern string, handler interface{}, mapping ...Mapping)`.

`patter`: If the pattern already exists, an error will return.

`handler` can be of type `http.Handler` or `http.HandlerFunc`. If it is `nil` or any other type, an error will return.

A action name mapping is required on `http.Handler`. Mappings can be defined optionally on `http.HandlerFunc`. By
default, all allowed HTTP methods of the router, will be mapped.

A Mapping instance can be created with `router.NewMapping(methods []string, action interface{}, mw *middleware)`.

The `methods` are any HTTP methods which should be mapped to the pattern. If its nil, all allowed HTTP methods of the
router manager will be added.

The `action` can be of the type `string` or `func`. If the type is `func`, the function name will be set as string on
runtime. The action string will be added as request context.

If set, the `middlewares` will be added to the route.

For each pattern, any HTTP method must be unique, otherwise an error will return.

```go 
route := router.NewRoute("/public2", handleFunc, router.NewMapping([]string{http.MethodGet}, "View", nil), router.NewMapping([]string{http.MethodPUt}, "Create", nil))
```

## Middleware

All pre-defined middleware:

### Logger

Provides a middleware for the `logger.Manager`. The logged information is remoteAddr, HTTP Method, URL, Proto, HTTP
Status, Response size and requested time. On HTTP status < 400 a `log.Info()` will be called otherwise `log.Error()`.

!!!Info The logger middleware should used before all other middlewares, otherwise the request time will be incorrect.

```go 
// the middleware
mw := router.NewMiddleware(middleware.NewLogger(logManager).MW)
```    

### JWT

Provides a middleware to check against a JWT token. If the JWT token is invalid a `http.StatusUnauthorized` will return.
If the JWT token is expired it will be re-generated if allowed.

There are two callback functions. `CallbackGenerate` for manipulating the claim before its signed. `CallbackRefresh` to
check if the refresh token is still valid, against a custom logic.

The claim will be set as request context with the key `jwt.CLAIM`.

A claim struct is provided and can be embedded into a custom struct.

```go 
cfg := jwt.Config{
    Alg: jwt.HS512, 
    Issuer: "authserver", 
    Audience: "client", 
    Subject: "auth", 
    Expiration: 5*time.Minute, 
    SignKey: "secret",
    RefreshToken: jwt.RefreshConfig{Expiration: 30*24*time.Hour}
}

claim := jwt.Claim{}
	
jwt := jwt.New(cfg,claim);
jwt.CallbackRefresh = func(http.ResponseWriter, *http.Request, Claimer){return nil}  // your logic
jwt.CallbackGenerate = func(http.ResponseWriter, *http.Request, Claimer){return nil} // your logic

// the middleware
mw := router.NewMiddleware(jwt.MW)
```  

### RBAC

Provides a role based access control list. It is build on top of the JWT middleware.

A RoleService must be set, to check against the custom logic. Simply implement the `RoleService` interface. The
arguments `pattern` `HTTP method` and `claim` will be passed to the `Allowed` function.

!!!Info The JWT middleware must be set before the RBAC middleware.

```go 
roleService := CustomService{};
rbac := middleware.NewRbac(roleService)

// the middleware
mw := router.NewMiddleware(jwt.MW, rbac.MW)
```

## Providers

All pre-defined providers:

### JSROUTER

A wrapper for [httprouter](https://github.com/julienschmidt/httprouter).

Name:

`router.JSROUTER`

Options:

no options are available at the moment.

Usage:

```go 
import "github.com/patrickascher/gofer/router"
import _ "github.com/patrickascher/gofer/router/jsrouter"

r,err := router.New(router.JSROUTER,nil)
```

## Create your own provider

To create your own provider, you have to implement the `router.Provider` interface.

```go
// Provider interface.
type Provider interface {
// Handler must return the mux for http/server.
Handler() http.Handler
// custom NotFound handler can be set.
SetNotFound(http.Handler)
// AddRoute to the router.
AddRoute(Route) error
// AddPublicDir to the router.
// The source is already checked if it exists.
AddPublicDir(url string, path string) error
// AddPublicFile to the router
// The source is already checked if it exists.
AddPublicFile(url string, path string) error
}
```

Use the `init` function to register your provider.

The registered value must be of the type `func(m Manager, options interface{}) (router.Provider, error)`.

```go 
// init register your config provider
func init() {
    //...
	err := router.Register("my-provider",New)
    if err != nil {
    	log.Fatal(err)
    }
}

func New(m Manager, options interface{}) (router.Provider, error){
    //...
    return provider,nil
}
```

**Usage**

```go 
import "github.com/patrickascher/gofer/router"
import _ "your/repo/router/yourProvider"

err := router.New("my-provider", options)
```

