package auth

import (
	"fmt"

	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/registry"
)

const registryPrefix = "auth_"
const Login = "login"
const Password = "password"

type providerFn func(opt interface{}) (Interface, error)

type Interface interface {
	Login(p controller.Interface) (Schema, error)
	Logout(p controller.Interface) error
	RecoverPassword(p controller.Interface) error
}

// Schema auth schema
type Schema struct {
	Provider string
	UID      string

	Name      string
	Email     string
	FirstName string
	LastName  string
	Location  string
	Image     string
	Phone     string
	URL       string

	RawInfo interface{}
}

// Register a new cache provider by name.
func Register(name string, provider providerFn) error {
	return registry.Set(registryPrefix+name, provider)
}

func New(provider string, options interface{}) (Interface, error) {
	provider = registryPrefix + provider

	// get the registry entry.
	instanceFn, err := registry.Get(provider)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	return instanceFn.(providerFn)(options)
}
