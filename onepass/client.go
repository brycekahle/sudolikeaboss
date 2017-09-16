package onepass

import (
	"bytes"
	"encoding/base64"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/satori/go.uuid"
)

type Command struct {
	Action   string  `json:"action"`
	Number   int     `json:"number,omitempty"`
	Version  string  `json:"version,omitempty"`
	BundleID string  `json:"bundleId,omitempty"`
	Payload  Payload `json:"payload"`
}

type Payload struct {
	Version      string            `json:"version,omitempty"`
	ExtID        string            `json:"extId,omitempty"`
	Method       string            `json:"method,omitempty"`
	Secret       string            `json:"secret,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty"`
	URL          string            `json:"url,omitempty"`
	Options      map[string]string `json:"options,omitempty"`
	CC           string            `json:"cc,omitempty"`
	CS           string            `json:"cs,omitempty"`
	M4           string            `json:"M4,omitempty"`
	M3           string            `json:"M3,omitempty"`
	Algorithm    string            `json:"alg,omitempty"`
	Iv           string            `json:"iv,omitempty"`
	Data         string            `json:"data,omitempty"`
	Hmac         string            `json:"hmac,omitempty"`
}

// Configuration struct
type Configuration struct {
	WebsocketURI    string `json:"websocketUri"`
	WebsocketOrigin string `json:"websocketOrigin"`
	DefaultHost     string `json:"defaultHost"`
	StateDirectory  string `json:"stateDirectory"`
}

type OnePasswordClient struct {
	DefaultHost             string
	conn                    OnePasswordConnection
	encryptedConn           *EncryptedOnePasswordConnection
	StateDirectory          string
	number                  int
	extID                   string
	secret                  []byte
	base64urlWithoutPadding *b64.Encoding
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

	base64urlWithoutPadding := b64.URLEncoding.WithPadding(b64.NoPadding)
	client.base64urlWithoutPadding = base64urlWithoutPadding

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

func (client *OnePasswordClient) SendShowPopupCommand() (*Response, error) {
	payload := Payload{
		URL:     client.DefaultHost,
		Options: map[string]string{"source": "toolbar-button"},
	}

	command := client.createCommand("showPopup", payload)
	return client.encryptedConn.SendCommand(command)
}

func (client *OnePasswordClient) createCommand(action string, payload Payload) *Command {
	command := Command{
		Action: action,
		//Number:   client.number,
		Version: "4.6.2.90",
		//BundleID: "com.sudolikeaboss.sudolikeaboss",
		Payload: payload,
	}

	// Increment the number (it's a 1password thing that I saw whilst listening
	// to their commands
	client.number++
	return &command
}

func (client *OnePasswordClient) SendHelloCommand() (*Response, error) {
	capabilities := make([]string, 2)
	capabilities[0] = "auth-sma-hmac256"
	capabilities[1] = "aead-cbchmac-256"

	payload := Payload{
		Version:      "4.6.2.90",
		ExtID:        client.extID,
		Capabilities: capabilities,
	}

	command := client.createCommand("hello", payload)

	response, err := client.conn.SendCommand(command)
	if err != nil {
		return nil, err
	}

	if response.Action != "authNew" && response.Action != "authBegin" {
		errorMsg := fmt.Sprintf("Unexpected response: %s", response.Action)
		err = errors.New(errorMsg)
		return nil, err
	}

	return response, nil
}

func (client *OnePasswordClient) authRegister() (*Response, error) {
	secretB64 := base64.URLEncoding.EncodeToString(client.secret)

	authRegisterPayload := Payload{
		ExtID:  client.extID,
		Method: "auth-sma-hmac256",
		Secret: secretB64,
	}

	authRegisterCommand := client.createCommand("authRegister", authRegisterPayload)

	registerResponse, err := client.conn.SendCommand(authRegisterCommand)
	if err != nil {
		return nil, err
	}

	if registerResponse.Action != "authRegistered" {
		errorMsg := fmt.Sprintf("Unexpected response: %s", registerResponse.Action)
		err = errors.New(errorMsg)
		return nil, err
	}

	return registerResponse, nil
}

func (client *OnePasswordClient) authBegin(cc []byte) (*Response, error) {
	ccB64 := client.base64urlWithoutPadding.EncodeToString(cc)

	authBeginPayload := Payload{
		Method: "auth-sma-hmac256",
		ExtID:  client.extID,
		CC:     ccB64,
	}

	authBeginCommand := client.createCommand("authBegin", authBeginPayload)

	authBeginResponse, err := client.conn.SendCommand(authBeginCommand)
	if err != nil {
		return nil, err
	}

	if authBeginResponse.Action != "authContinue" {
		errorMsg := fmt.Sprintf("Unexpected response: %s", authBeginResponse.Action)
		err = errors.New(errorMsg)
		return nil, err
	}

	return authBeginResponse, nil
}

func (client *OnePasswordClient) Register(code string) (*Response, error) {
	fmt.Printf("The 1password helper will request registration of code: %s\n", code)
	fmt.Println("To complete registration. You must accept that code from the helper.")
	_, err := client.authRegister()

	if err != nil {
		fmt.Printf("Registration failed with %s\n", err)
		return nil, err
	}

	return nil, nil
}

func (client *OnePasswordClient) Authenticate(register bool) (*Response, error) {
	helloResponse, err := client.SendHelloCommand()
	if err != nil {
		return nil, err
	}

	if register {
		if helloResponse.Action != "authNew" {
			fmt.Println("sudolikeaboss is already registered.")
			os.Exit(0)
		}

		_, err = client.Register(helloResponse.Payload.Code)
		if err != nil {
			return nil, err
		}
	}

	cc, err := GenerateRandomBytes(16)
	if err != nil {
		return nil, err
	}

	authBeginResponse, err := client.authBegin(cc)
	if err != nil {
		return nil, err
	}

	m3, err := base64.RawURLEncoding.DecodeString(authBeginResponse.Payload.M3)
	if err != nil {
		return nil, err
	}

	// Verify M3
	cs, _ := base64.RawURLEncoding.DecodeString(authBeginResponse.Payload.CS)

	expectedM3Bytes := generateM3(client.secret, cs, cc)

	if !bytes.Equal(expectedM3Bytes, m3) {
		errorMsg := fmt.Sprintf("M3 not expected value")
		err = errors.New(errorMsg)
		return nil, err
	}

	m4 := HmacSha256(client.secret, m3)
	m4B64 := base64.RawURLEncoding.EncodeToString(m4)

	authVerifyPayload := Payload{
		Method: "auth-sma-hmac256",
		M4:     m4B64,
		ExtID:  client.extID,
	}

	authVerifyCommand := client.createCommand("authVerify", authVerifyPayload)

	authVerifyResponse, err := client.conn.SendCommand(authVerifyCommand)
	if err != nil {
		return nil, err
	}

	if authVerifyResponse.Action != "welcome" {
		errorMsg := fmt.Sprintf("Unexpected response: %s", authVerifyResponse.Action)
		err = errors.New(errorMsg)
		return nil, err
	}

	// Generate the keys
	//
	// encK = HMAC-SHA256(secret, M3|M4|"encryption")
	sessionEncK := HmacSha256(client.secret, m3, m4, []byte("encryption"))

	// hmacK = HMAC-SHA256(secret, M4|M3|"hmac")
	sessionHmacK := HmacSha256(client.secret, m4, m3, []byte("hmac"))

	client.encryptedConn, err = NewEncryptedConnection(client.conn, sessionEncK, sessionHmacK, client.secret)
	if err != nil {
		return nil, err
	}

	_, err = client.encryptedConn.decryptResponse(&authVerifyResponse.Payload)
	if err != nil {
		return nil, err
	}
	return authVerifyResponse, nil
}
