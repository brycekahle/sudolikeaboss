package onepass

import (
	"encoding/json"
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
	if err != nil {
		return nil, err
	}
	return &websocketClient{conn}, nil
}

func (w *websocketClient) SendCommand(c *Command) error {
	log.Debugf("send: %+v", c)
	return w.conn.WriteJSON(c)
}

func (w *websocketClient) ReadResponse(r interface{}) error {
	_, buf, err := w.conn.ReadMessage()
	if err != nil {
		return err
	}
	log.WithField("raw", true).Debugf("recv: %s", buf)
	err = json.Unmarshal(buf, &r)
	if err != nil {
		return err
	}
	// if err := w.conn.ReadJSON(&r); err != nil {
	// 	return nil, err
	// }
	log.Debugf("recv: %+v", r)
	return nil
}

func (w *websocketClient) Close() error {
	return w.conn.Close()
}
