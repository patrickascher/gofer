package native

import (
	"github.com/patrickascher/gofer/auth"
	"github.com/patrickascher/gofer/controller"
)

func init() {
	err := auth.Register("native", func(options interface{}) (auth.Interface, error) { return &native{}, nil })
	if err != nil {
		panic(err)
	}
}

type native struct {
}

func (n *native) Login(c controller.Interface) (auth.Schema, error) {
	user := auth.Schema{}
	user.Email = "pat@fullhouse-productions.com"
	return user, nil
}

func (n *native) Logout(c controller.Interface) error {
	return nil
}

func (n *native) RecoverPassword(c controller.Interface) error {
	return nil
}
