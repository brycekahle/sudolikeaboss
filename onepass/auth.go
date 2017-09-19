package onepass

import (
	"bytes"
	"encoding/base64"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func (client *OnePasswordClient) Authenticate() (*EncryptedClient, error) {
	cc, err := generateRandomBytes(16)
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
	cs, err := base64.RawURLEncoding.DecodeString(authBeginResponse.Payload.CS)
	if err != nil {
		return nil, err
	}

	expectedM3Bytes := generateM3(client.secret, cs, cc)
	if !bytes.Equal(expectedM3Bytes, m3) {
		return nil, fmt.Errorf("invalid M3: %s (actual) != %s (expected)", authBeginResponse.Payload.M3, base64.RawURLEncoding.EncodeToString(expectedM3Bytes))
	}

	m4 := hmacSha256(client.secret, m3)
	er, err := client.authVerify(m4)
	if err != nil {
		return nil, err
	}

	keys := EncryptionKeys{
		// enc = HMAC-SHA256(secret, M3|M4|"encryption")
		encryption: hmacSha256(client.secret, m3, m4, []byte("encryption")),
		// hmac = HMAC-SHA256(secret, M4|M3|"hmac")
		hmac:   hmacSha256(client.secret, m4, m3, []byte("hmac")),
		secret: client.secret,
	}

	encClient, err := NewEncryptedClient(client, &keys)
	if err != nil {
		return nil, err
	}

	authVerifyResponse := WelcomeResponse{}
	err = encClient.conn.Decrypt(er, &authVerifyResponse)
	if err != nil {
		return nil, err
	}
	log.WithField("encrypted", true).Debugf("recv: %+v", authVerifyResponse)
	return encClient, nil
}

func (client *OnePasswordClient) authRegister() error {
	authRegisterPayload := Payload{
		ExtID:  client.extID,
		Method: methodSmaHmac256,
		Secret: base64.URLEncoding.EncodeToString(client.secret),
	}

	authRegisterCommand := NewCommand(SendAuthRegister, authRegisterPayload)
	err := client.conn.SendCommand(authRegisterCommand)
	if err != nil {
		return err
	}
	registerResponse := EmbeddedResponse{}
	err = client.conn.ReadResponse(&registerResponse)
	if err != nil {
		return err
	}

	if registerResponse.Action != ResponseAuthRegistered {
		return fmt.Errorf("Unexpected response: %s", registerResponse.Action)
	}
	return nil
}

type AuthContinueResponse struct {
	EmbeddedResponse
	Payload struct {
		M3 string `json:"m3"`
		CS string `json:"cs"`
	} `json:"payload"`
}

func (client *OnePasswordClient) authBegin(cc []byte) (*AuthContinueResponse, error) {
	authBeginPayload := Payload{
		ExtID:  client.extID,
		Method: methodSmaHmac256,
		CC:     base64.URLEncoding.EncodeToString(cc),
	}

	authBeginCommand := NewCommand(SendAuthBegin, authBeginPayload)
	err := client.conn.SendCommand(authBeginCommand)
	if err != nil {
		return nil, err
	}
	authBeginResponse := AuthContinueResponse{}
	err = client.conn.ReadResponse(&authBeginResponse)
	if err != nil {
		return nil, err
	}

	if authBeginResponse.Action != ResponseAuthContinue {
		return nil, fmt.Errorf("Unexpected response: %s", authBeginResponse.Action)
	}

	return &authBeginResponse, nil
}

type WelcomeResponse struct {
	EmbeddedResponse
	Payload struct {
		Capabilities []string `json:"capabilities"`
	} `json:"payload"`
}

func (client *OnePasswordClient) authVerify(m4 []byte) (*EncryptedResponse, error) {
	authVerifyPayload := Payload{
		ExtID:  client.extID,
		Method: methodSmaHmac256,
		M4:     base64.RawURLEncoding.EncodeToString(m4),
	}

	authVerifyCommand := NewCommand(SendAuthVerify, authVerifyPayload)
	err := client.conn.SendCommand(authVerifyCommand)
	if err != nil {
		return nil, err
	}
	authVerifyResponse := EncryptedResponse{}
	err = client.conn.ReadResponse(&authVerifyResponse)
	if err != nil {
		return nil, err
	}

	if authVerifyResponse.Action != ResponseWelcome {
		return nil, fmt.Errorf("Unexpected response: %s", authVerifyResponse.Action)
	}
	return &authVerifyResponse, nil
}
