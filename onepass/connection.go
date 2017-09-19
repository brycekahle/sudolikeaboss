package onepass

type OnePasswordConnection interface {
	SendCommand(c *Command) error
	ReadResponse(r interface{}) error
	Close() error
}

type EncryptedConnection interface {
	OnePasswordConnection
	Decrypt(*EncryptedResponse, interface{}) error
}

func NewConnection(configuration *Configuration) (OnePasswordConnection, error) {
	if configuration.WebsocketURI != "" {
		return NewWebsocketConnection(configuration.WebsocketURI, configuration.WebsocketOrigin)
	}
	return NewNativeMessagingConnection()
}
