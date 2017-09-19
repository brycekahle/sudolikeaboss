package onepass

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type EncryptedClient struct {
	client *OnePasswordClient
	conn   EncryptedConnection
}

type EncryptedOnePasswordConnection struct {
	conn  OnePasswordConnection
	b64   *base64.Encoding
	block cipher.Block
	keys  *EncryptionKeys
}

type EncryptionKeys struct {
	encryption []byte
	hmac       []byte
	secret     []byte
}

func NewEncryptedClient(client *OnePasswordClient, keys *EncryptionKeys) (*EncryptedClient, error) {
	econn, err := NewEncryptedConnection(client.conn, keys)
	if err != nil {
		return nil, err
	}

	return &EncryptedClient{
		conn:   econn,
		client: client,
	}, nil
}

type FillItemResponse struct {
	EmbeddedResponse
	Payload struct {
		Action  PayloadAction    `json:"action"`
		Options map[string]bool  `json:"options"`
		Item    *json.RawMessage `json:"item"`
	} `json:"payload"`
}

func (c *EncryptedClient) ShowPopup() (*FillItemResponse, error) {
	payload := Payload{
		URL:     c.client.DefaultHost,
		Options: map[string]string{"source": "toolbar-button"},
	}

	command := NewCommand(ShowPopup, payload)
	err := c.conn.SendCommand(command)
	if err != nil {
		return nil, err
	}
	response := FillItemResponse{}
	err = c.conn.ReadResponse(&response)
	if err != nil {
		return nil, err
	}
	return &response, err
}

func (c *EncryptedClient) Close() error {
	return c.conn.Close()
}

func NewEncryptedConnection(conn OnePasswordConnection, keys *EncryptionKeys) (EncryptedConnection, error) {
	block, err := aes.NewCipher(keys.encryption)
	if err != nil {
		return nil, err
	}

	return &EncryptedOnePasswordConnection{
		conn:  conn,
		b64:   base64.RawURLEncoding,
		block: block,
		keys:  keys,
	}, nil
}

func (e *EncryptedOnePasswordConnection) SendCommand(command *Command) error {
	log.WithField("encrypted", true).Debugf("send: %+v", command)
	encryptedPayload, err := e.encryptPayload(&(command.Payload))
	if err != nil {
		return err
	}

	command.Payload = *encryptedPayload
	return e.conn.SendCommand(command)
}

func (e *EncryptedOnePasswordConnection) ReadResponse(r interface{}) error {
	er := EncryptedResponse{}
	err := e.conn.ReadResponse(&er)
	if err != nil {
		return err
	}

	err = e.Decrypt(&er, r)
	if err != nil {
		return err
	}

	log.WithField("encrypted", true).Debugf("recv: %+v", r)
	return nil
}

func (e *EncryptedOnePasswordConnection) Close() error {
	return e.conn.Close()
}

func (e *EncryptedOnePasswordConnection) encryptPayload(payload *Payload) (*Payload, error) {
	iv, err := generateRandomBytes(16)
	if err != nil {
		return nil, err
	}

	// Encrypt the payload
	payloadJSONStr, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Encrypt the payload
	encryptedPayload, err := e.encrypt(iv, payloadJSONStr)
	if err != nil {
		return nil, err
	}

	newPayload := Payload{
		Iv:        e.b64.EncodeToString(iv),
		Data:      e.b64.EncodeToString(encryptedPayload),
		Algorithm: algAeadCbcHmac256,
	}

	newPayload.Hmac = e.b64.EncodeToString(hmacSha256(e.keys.hmac, []byte(newPayload.Iv), []byte(newPayload.Data)))
	return &newPayload, nil
}

func (e *EncryptedOnePasswordConnection) Decrypt(er *EncryptedResponse, r interface{}) error {
	iv, err := e.b64.DecodeString(er.Payload.Iv)
	if err != nil {
		return err
	}

	data, err := e.b64.DecodeString(er.Payload.Data)
	if err != nil {
		return err
	}

	hmac, err := e.b64.DecodeString(er.Payload.Hmac)
	if err != nil {
		return err
	}

	// Verify hmac
	expectedHmac := hmacSha256(e.keys.hmac, []byte(er.Payload.Iv), []byte(er.Payload.Data))
	if !bytes.Equal(expectedHmac, hmac) {
		return fmt.Errorf("invalid HMAC: %s (actual) != %s (expected)", e.b64.EncodeToString(hmac), e.b64.EncodeToString(expectedHmac))
	}

	payload, err := e.decrypt(iv, data)
	if err != nil {
		return err
	}
	log.WithField("raw_payload", true).Debugf("recv: %s", payload)

	ip := json.RawMessage(payload)
	ir := IntermediateResponse{
		EmbeddedResponse: EmbeddedResponse{
			Action:  er.Action,
			Version: er.Version,
		},
		Payload: &ip,
	}
	ibuf, err := json.Marshal(ir)
	if err != nil {
		return err
	}
	err = json.Unmarshal(ibuf, r)
	if err != nil {
		return err
	}
	return nil
}

func generateM3(secret []byte, cs []byte, cc []byte) []byte {
	csAndCc := append(cs[:], cc[:]...)

	csAndCcSha := sha256.New()
	_, _ = csAndCcSha.Write(csAndCc)

	h := hmac.New(sha256.New, secret)
	_, _ = h.Write(csAndCcSha.Sum(nil))
	return h.Sum(nil)
}

func hmacSha256(key []byte, dataToSign ...[]byte) []byte {
	h := hmac.New(sha256.New, key)
	for _, data := range dataToSign {
		_, _ = h.Write(data)
	}
	return h.Sum(nil)
}

// generateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// pkcs7Pad Appends padding.
func pkcs7Pad(data []byte, blocklen int) ([]byte, error) {
	if blocklen <= 0 {
		return nil, fmt.Errorf("invalid blocklen %d", blocklen)
	}
	padlen := 1
	for ((len(data) + padlen) % blocklen) != 0 {
		padlen = padlen + 1
	}

	pad := bytes.Repeat([]byte{byte(padlen)}, padlen)
	return append(data, pad...), nil
}

// pkcs7Unpad Returns slice of the original data without padding.
func pkcs7Unpad(data []byte, blocklen int) ([]byte, error) {
	if blocklen <= 0 {
		return nil, fmt.Errorf("invalid blocklen %d", blocklen)
	}
	if len(data)%blocklen != 0 || len(data) == 0 {
		return nil, fmt.Errorf("invalid data len %d", len(data))
	}
	padlen := int(data[len(data)-1])
	if padlen > blocklen || padlen == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	// check padding
	pad := data[len(data)-padlen:]
	for i := 0; i < padlen; i++ {
		if pad[i] != byte(padlen) {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	return data[:len(data)-padlen], nil
}

func (e *EncryptedOnePasswordConnection) encrypt(iv []byte, plaintext []byte) ([]byte, error) {
	paddedPlaintext, err := pkcs7Pad(plaintext, aes.BlockSize)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, len(paddedPlaintext))

	cbc := cipher.NewCBCEncrypter(e.block, iv)
	cbc.CryptBlocks(ciphertext, paddedPlaintext)

	return ciphertext, nil
}

func (e *EncryptedOnePasswordConnection) decrypt(iv []byte, ciphertext []byte) ([]byte, error) {
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("Ciphertext is not a multiple of the AES blocksize")
	}

	cbc := cipher.NewCBCDecrypter(e.block, iv)
	cbc.CryptBlocks(ciphertext, ciphertext)

	return pkcs7Unpad(ciphertext, aes.BlockSize)
}
