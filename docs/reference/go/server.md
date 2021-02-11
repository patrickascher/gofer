# Router

Package server is a configurable webserver with pre-defined hooks.

## New

New will create a new webserver instance. Only one webserver instance can exist at the moment.

```go 
cfg := server.Configuration{}
//...

err := server.New(cfg)
// ...
```

### Config

The server config can be embedded in your application configuration which can be passed to the `New` function.

```go
type MyAppConfig struct {
server.Configuration `mapstructure:",squash"`
Name string
}
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

## Config

Will return the server configuration. Error will return if this function is called before a server instance exists.

```go
cfg, err := server.Config()
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




