package connection

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Client is a websocket client meant to be used with the 1password
type Client struct {
	*websocket.Conn
}

func NewWebsocketClient(url string, origin string) (OnePasswordConnection, error) {
	headers := http.Header{}
	headers.Add("Origin", origin)

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		return nil, err
	}

	c := Client{conn}
	return c, nil
}
