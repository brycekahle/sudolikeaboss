package onepass

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type websocketClient struct {
	conn *websocket.Conn
}

func NewWebsocketConnection(url string, origin string) (OnePasswordConnection, error) {
	headers := http.Header{}
	headers.Add("Origin", origin)

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	conn, _, err := dialer.Dial(url, headers)
	return &websocketClient{conn}, err
}

func (w *websocketClient) SendCommand(c *Command) (*Response, error) {
	log.Debugf("send: %+v", c)
	if err := w.conn.WriteJSON(c); err != nil {
		return nil, err
	}
	r := Response{}
	if err := w.conn.ReadJSON(&r); err != nil {
		return nil, err
	}
	log.Debugf("recv: %+v", r)
	return &r, nil
}

func (w *websocketClient) Close() error {
	return w.conn.Close()
}
