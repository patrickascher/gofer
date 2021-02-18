# Auth

Package auth provides a standard auth for your website. Custom auth providers can be added. Simply implement
the `auth.Interface`

## New

The `New` function requires the provider name as argument.

Error will return if the provider was not registered or configure before.

```go 
// Example of the memory cache provider
provider, err := auth.New("native")
```

## ConfigureProvider

Must be used to configure the provider on webserver start. Because there can be different providers with different
configurations, a map is passed as second argument. Please see the provider section for the different configuration
parameters.

```go
err := ConfigureProvider("native", nil)
//...
```

## Config

The config is predefined in `Server.Configuration.Auth`.

| Name                 | Description                                                                                                  |
|----------------------|--------------------------------------------------------------------------------------------------------------|
| BcryptCost           | bcrypt cost                                                                                                  |
| AllowedFailedLogin   | number of failed login attempts before the user gets locked.                                                 |
| LockDuration         | defines how long a user should be locked before the next login attempt is allowed. (RFC3339 duration string) |
| InactiveDuration     | the allowed duration between the last login and now (RFC3339  duration string)                               |
| TokenDuration        | JWT token live time (RFC3339  duration string)                                                               |
| RefreshTokenDuration | Refresh token live time (RFC3339  duration string)                                                           |


## Controller

The `auth.Controller` can be used out of the box. All required routes can be added with the helper
function `auth.AddRoutes`.

If you need to extend the controller, simply embed it into your controller and extend the functions.


## Protocol

A Protocol is added which loggs the following user actions:

* Login
* Enter wrong password
* Locked user
* Inactive user
* Refresh Token
* Refresh Token failed
* Logout
* Reset password

If you need to add additional protocols, use the helper function `AddProtocol(userID,key,value)`

## Claim

By default, the user claim which will be included in the jwt token looks like this. It can be fully customized but be
aware to re-implement the `jwt.CallbackGenerate` and `jwt.CallbackRefresh` for your requirements.

```go
type Claim struct {
jwt.Claim

Provider string `json:"provider"`
UserID   int    `json:"id"`

Name    string   `json:"name"`
Surname string   `json:"surname"`
Login   string   `json:"login"`
Roles   []string `json:"Roles"`

Options map[string]string `json:"options"`
}
```

## JWT middleware

There are two pre-defined callbacks for the jwt middleware `JWTRefreshCallback` and `JWTGenerateCallback`. Which can be
used out of the box.

## RBAC middleware

A pre-defined rbac middleware can be used out of the box.

## Providers

### Native

The native provider will connect to the existing user database. It will add an user option `password` for the
authentication.

## Create your own provider

To create your own provider, you have to implement the `auth.Interface`.

```go
type Interface interface {
Login(p controller.Interface) (Schema, error)
Logout(p controller.Interface) error
RecoverPassword(p controller.Interface) error
}
```

Use the `init` function to register your provider.

The registered value must be of the type `func(opt map[string]interface{}) (Interface, error)`.

```go 
// init register the superFastMemory provider.
func init() {
	err := auth.Register("yourProvider", func(options map[string]interface{}) (auth.Interface, error) { return &yourProvider{}, nil })
	if err != nil {
		panic(err)
	}
}
	
type yourProvider struct {
}

func (n *Native) Login(c controller.Interface) (auth.Schema, error) {
//...
}
```

### Schema

A Schema is defined which must be used from all providers. Options will be saved in the user database (not yet
implemented). The Login will be mapped with the local user database.

```go
type Schema struct {
Provider string
UID      string

Login      string
Name       string
Surname    string
Salutation string

Options []Option
}
```
