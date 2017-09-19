package onepass

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type EmbeddedResponse struct {
	Action  ResponseAction `json:"action"`
	Version string         `json:"version"`
}

type EncryptedResponse struct {
	EmbeddedResponse
	Payload struct {
		Algorithm string `json:"alg"`
		Iv        string `json:"iv"`
		Hmac      string `json:"hmac"`
		Data      string `json:"data"`
	} `json:"payload"`
}

type IntermediateResponse struct {
	EmbeddedResponse
	Payload *json.RawMessage `json:"payload"`
}

func (response *FillItemResponse) GetPassword() (string, error) {
	itemBytes := []byte(*response.Payload.Item)
	var item Item

	switch response.Payload.Action {
	case FillLogin:
		var loginItem LoginItem
		if err := json.Unmarshal(itemBytes, &loginItem); err != nil {
			return "", errors.Wrap(err, "error unmarshaling LoginItem")
		}
		item = loginItem

	case FillPassword:
		var passwordItem PasswordItem
		if err := json.Unmarshal(itemBytes, &passwordItem); err != nil {
			return "", errors.Wrap(err, "error unmarshaling PasswordItem")
		}
		item = passwordItem

	default:
		return "", fmt.Errorf("Payload action \"%s\" does not have a password", response.Payload.Action)
	}

	return item.GetPassword()
}

type Item interface {
	GetPassword() (string, error)
}

type LoginField struct {
	Value       string `json:"value"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Designation string `json:"designation"`
}

type LoginItem struct {
	UUID           string            `json:"uuid"`
	NakedDomains   []string          `json:"nakedDomains"`
	Overview       map[string]string `json:"overview"`
	SecureContents struct {
		HTMLForm map[string]interface{} `json:"htmlForm"`
		Fields   []LoginField           `json:"fields"`
	} `json:"secureContents"`
}

func (item LoginItem) GetPassword() (string, error) {
	for _, fieldObj := range item.SecureContents.Fields {
		if fieldObj.Designation == "password" {
			return fieldObj.Value, nil
		}
	}

	return "", errors.New("no password found in the item")
}

type PasswordItem struct {
	UUID           string                 `json:"uuid"`
	Overview       map[string]interface{} `json:"overview"`
	SecureContents struct {
		Password string `json:"password"`
	} `json:"secureContents"`
}

func (item PasswordItem) GetPassword() (string, error) {
	return item.SecureContents.Password, nil
}
