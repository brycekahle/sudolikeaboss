package onepass

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func (client *OnePasswordClient) Authenticate(register bool) (*EncryptedClient, error) {
	helloResponse, err := client.SendHelloCommand()
	if err != nil {
		return nil, err
	}

	if register {
		if helloResponse.Action != ResponseAuthNew {
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
	cs, err := base64.RawURLEncoding.DecodeString(authBeginResponse.Payload.CS)
	if err != nil {
		return nil, err
	}

	expectedM3Bytes := generateM3(client.secret, cs, cc)
	if !bytes.Equal(expectedM3Bytes, m3) {
		return nil, fmt.Errorf("invalid M3: %s (actual) != %s (expected)", authBeginResponse.Payload.M3, base64.RawURLEncoding.EncodeToString(expectedM3Bytes))
	}

	m4 := HmacSha256(client.secret, m3)
	authVerifyResponse, err := client.authVerify(m4)
	if err != nil {
		return nil, err
	}

	keys := EncryptionKeys{
		// enc = HMAC-SHA256(secret, M3|M4|"encryption")
		encryption: HmacSha256(client.secret, m3, m4, []byte("encryption")),
		// hmac = HMAC-SHA256(secret, M4|M3|"hmac")
		hmac:   HmacSha256(client.secret, m4, m3, []byte("hmac")),
		secret: client.secret,
	}

	encClient, err := NewEncryptedClient(client, &keys)
	if err != nil {
		return nil, err
	}

	rp, err := encClient.conn.decryptResponse(&authVerifyResponse.Payload)
	if err != nil {
		return nil, err
	}
	authVerifyResponse.Payload = *rp
	log.WithField("encrypted", true).Debugf("recv: %+v", authVerifyResponse)
	return encClient, nil
}

func (client *OnePasswordClient) authRegister() (*Response, error) {
	authRegisterPayload := Payload{
		ExtID:  client.extID,
		Method: methodSmaHmac256,
		Secret: base64.URLEncoding.EncodeToString(client.secret),
	}

	authRegisterCommand := NewCommand(SendAuthRegister, authRegisterPayload)
	registerResponse, err := client.conn.SendCommand(authRegisterCommand)
	if err != nil {
		return nil, err
	}

	if registerResponse.Action != ResponseAuthRegistered {
		return nil, fmt.Errorf("Unexpected response: %s", registerResponse.Action)
	}
	return registerResponse, nil
}

func (client *OnePasswordClient) authBegin(cc []byte) (*Response, error) {
	authBeginPayload := Payload{
		ExtID:  client.extID,
		Method: methodSmaHmac256,
		CC:     base64.URLEncoding.EncodeToString(cc),
	}

	authBeginCommand := NewCommand(SendAuthBegin, authBeginPayload)
	authBeginResponse, err := client.conn.SendCommand(authBeginCommand)
	if err != nil {
		return nil, err
	}

	if authBeginResponse.Action != ResponseAuthContinue {
		return nil, fmt.Errorf("Unexpected response: %s", authBeginResponse.Action)
	}

	return authBeginResponse, nil
}

func (client *OnePasswordClient) authVerify(m4 []byte) (*Response, error) {
	authVerifyPayload := Payload{
		ExtID:  client.extID,
		Method: methodSmaHmac256,
		M4:     base64.RawURLEncoding.EncodeToString(m4),
	}

	authVerifyCommand := NewCommand(SendAuthVerify, authVerifyPayload)
	authVerifyResponse, err := client.conn.SendCommand(authVerifyCommand)
	if err != nil {
		return nil, err
	}

	if authVerifyResponse.Action != ResponseWelcome {
		return nil, fmt.Errorf("Unexpected response: %s", authVerifyResponse.Action)
	}
	return authVerifyResponse, nil
}
