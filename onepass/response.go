package onepass

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Response struct {
	Action  string          `json:"action"`
	Version string          `json:"version"`
	Payload ResponsePayload `json:"payload"`
}

func (response *Response) GetPassword() (string, error) {
	if response.Action != "fillItem" {
		errorMsg := fmt.Sprintf("Response action \"%s\" does not have a password", response.Action)
		return "", errors.New(errorMsg)
	}

	itemBytes := []byte(*response.Payload.Item)
	var item Item

	switch response.Payload.Action {
	case "fillLogin":
		var loginItem LoginItem
		if err := json.Unmarshal(itemBytes, &loginItem); err != nil {
			return "", err
		}
		item = loginItem

	case "fillPassword":
		var passwordItem PasswordItem
		if err := json.Unmarshal(itemBytes, &passwordItem); err != nil {
			return "", err
		}
		item = passwordItem

	default:
		errorMsg := fmt.Sprintf("Payload action \"%s\" does not have a password", response.Payload.Action)
		return "", errors.New(errorMsg)
	}

	return item.GetPassword()
}

type ResponsePayload struct {
	Item           *json.RawMessage       `json:"item"`
	Algorithm      string                 `json:"alg"`
	Method         string                 `json:"method"`
	Code           string                 `json:"code"`
	Data           string                 `json:"data"`
	Hmac           string                 `json:"hmac"`
	Iv             string                 `json:"iv"`
	M3             string                 `json:"m3"`
	CS             string                 `json:"cs"`
	AssociatedData string                 `json:"adata"`
	Options        map[string]interface{} `json:"options"`
	OpenInTabMode  string                 `json:"openInTabMode"`
	Action         string                 `json:"action"`
}

type Item interface {
	GetPassword() (string, error)
}

type LoginItem struct {
	UUID           string                  `json:"uuid"`
	NakedDomains   []string                `json:"nakedDomains"`
	Overview       map[string]interface{}  `json:"overview"`
	SecureContents LoginItemSecureContents `json:"secureContents"`
}

func (item LoginItem) GetPassword() (string, error) {
	for _, fieldObj := range item.SecureContents.Fields {
		if fieldObj["designation"] == "password" {
			return fieldObj["value"], nil
		}
	}

	return "", errors.New("no password found in the item")
}

type LoginItemSecureContents struct {
	HTMLForm map[string]interface{} `json:"htmlForm"`
	Fields   []map[string]string    `json:"fields"`
}

type PasswordItem struct {
	UUID           string                     `json:"uuid"`
	Overview       map[string]interface{}     `json:"overview"`
	SecureContents PasswordItemSecureContents `json:"secureContents"`
}

func (item PasswordItem) GetPassword() (string, error) {
	return item.SecureContents.Password, nil
}

type PasswordItemSecureContents struct {
	Password string `json:"password"`
}
