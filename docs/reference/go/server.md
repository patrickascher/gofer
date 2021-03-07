# Server

Package server is a configurable webserver with pre-defined hooks.

## New

New will create a new webserver instance. Only one webserver instance can exist.

The following hooks will be called:

| Name      | Description                                                                                                                                             |
|-----------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| router    | If a router provider is defined by config, a new router manager will be created. The Favicon, PublicDir(s) and PublicFile(s)s will be added if defined. |
| databases | All defined databases will be saved globally and opened.                                                                                                |
| caches    | All defined caches will be created and saved globally.|
| translation    | If a translation provider is defined by config, a new translation manger will be created. By default the translation raw messages for all registered modules, navigations, and controller will be generated.|

If the configuration `Router.CreateDBRoutes` is set to `true`, for each route an db entry will be made.

```go 
cfg := server.Configuration{}
//...

err := server.New(cfg)
// ...
```

### Config

The server config can be embedded in your application configuration which can be passed to the `New` function. Error
will return if this function is called before a server instance exists.

```go
type MyAppConfig struct {
server.Configuration `mapstructure:",squash"`
Name string
}
```

Within your application you can access your Config by `server.Config()`. You have to cast the interface to your actual
type.

```go
config, err := server.Config()
myConfig := config.(MyConfig) // cast to ...
```

### ServerConfig

By this function you will only receive the `server.Configuration` struct. Error will return if this function is called
before a server instance exists.

```go
srvConfig, err := server.ServerConfig() 
```

## Start

Will start the webserver. Error will return if this function is called before a server instance exists.

```go
err := server.Start()
// ...
```

## Stop

Will stop the webserver. Error will return if this function is called before a server instance exists.

```go
err := server.Stop()
// ...
```

## JWT

Will return the `*jwt.Token` of the webserver. Error will return if this function is called before a server instance
exists.

```go
jwt, err := server.JWT()
// ...
```

## SetJWT

Set the server jwt token.. Error will return if this function is called before a server instance exists.

```go
err := server.SetJWT(jwt)
// ...
```

## Router

Will return the defined router. Error will return if this function is called before a server instance exists.

```go
router, err := server.Router()
// ...
```

## Caches

Will return all defined cache managers. Error will return if this function is called before a server instance exists.

```go
caches, err := server.Databases()
// ...
```

## Databases

Will return all defined databases. The database is globally used and already opened. Error will return if this function
is called before a server instance exists.

```go
dbs, err := server.Databases()
// ...
```

## Translation

Will return a `translation.Manager`.

```go
i18n, err := server.Translation()
// ...
```





