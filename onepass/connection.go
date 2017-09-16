package onepass

type OnePasswordConnection interface {
	SendCommand(c *Command) (*Response, error)
	Close() error
}

func NewConnection(configuration *Configuration) (OnePasswordConnection, error) {
	if configuration.WebsocketURI != "" {
		return NewWebsocketConnection(configuration.WebsocketURI, configuration.WebsocketOrigin)
	}
	return NewNativeMessagingConnection()
}
