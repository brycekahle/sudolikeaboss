package onepass

import (
	"github.com/pkg/errors"
)

func (client *OnePasswordClient) Register() error {
	err := client.authRegister()
	return errors.Wrap(err, "registration failed")
}
