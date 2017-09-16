package onepass

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

const methodSmaHmac256 = "auth-sma-hmac256"
const algAeadCbcHmac256 = "aead-cbchmac-256"

// Configuration struct
type Configuration struct {
	WebsocketURI    string `json:"websocketUri"`
	WebsocketOrigin string `json:"websocketOrigin"`
	DefaultHost     string `json:"defaultHost"`
	StateDirectory  string `json:"stateDirectory"`
}

type OnePasswordClient struct {
	DefaultHost    string
	conn           OnePasswordConnection
	StateDirectory string
	number         int
	extID          string
	secret         []byte
}

type StateFileConfig struct {
	Secret string `json:"secret"`
	ExtID  string `json:"extID"`
}

func NewClientWithConfig(configuration *Configuration) (*OnePasswordClient, error) {
	opClient, err := NewConnection(configuration)
	if err != nil {
		return nil, err
	}
	return NewCustomClient(opClient, configuration.DefaultHost, configuration.StateDirectory)
}

func NewCustomClient(opc OnePasswordConnection, defaultHost string, stateDirectory string) (*OnePasswordClient, error) {
	client := OnePasswordClient{
		conn:           opc,
		DefaultHost:    defaultHost,
		StateDirectory: stateDirectory,
	}

	// Load the state directory if stuff is in there
	err := client.LoadOrSetupState()
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func (client *OnePasswordClient) LoadOrSetupState() error {
	stateFilePath := path.Join(client.StateDirectory, "state.json")
	var stateFileConfig StateFileConfig
	var secret []byte
	stateFileBytes, err := ioutil.ReadFile(stateFilePath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(client.StateDirectory, 0700)
		if err != nil {
			return err
		}

		secret, err = GenerateRandomBytes(32)
		if err != nil {
			return err
		}

		stateFileConfig = StateFileConfig{
			ExtID:  uuid.NewV4().String(),
			Secret: base64.RawURLEncoding.EncodeToString(secret),
		}

		stateFileBytes, err = json.Marshal(&stateFileConfig)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(stateFilePath, stateFileBytes, 0700)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		err = json.Unmarshal(stateFileBytes, &stateFileConfig)
		if err != nil {
			return err
		}

		secret, err = base64.RawURLEncoding.DecodeString(stateFileConfig.Secret)
		if err != nil {
			return err
		}
	}

	client.extID = stateFileConfig.ExtID
	client.secret = secret
	return nil
}

func (client *OnePasswordClient) SendHelloCommand() (*Response, error) {
	payload := Payload{
		Version:      clientVersion,
		ExtID:        client.extID,
		Capabilities: []string{methodSmaHmac256, algAeadCbcHmac256},
	}

	command := NewCommand(SendHello, payload)
	response, err := client.conn.SendCommand(command)
	if err != nil {
		return nil, err
	}

	if response.Action != ResponseAuthNew && response.Action != ResponseAuthBegin {
		return nil, fmt.Errorf("Unexpected response: %s", response.Action)
	}
	return response, nil
}

func (client *OnePasswordClient) Register(code string) (*Response, error) {
	fmt.Printf("The 1password helper will request registration of code: %s\n", code)
	fmt.Println("To complete registration. You must accept that code from the helper.")
	_, err := client.authRegister()

	if err != nil {
		return nil, errors.Wrap(err, "registration failed")
	}

	return nil, nil
}
