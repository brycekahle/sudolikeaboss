package websocketclient

import (
	ws "golang.org/x/net/websocket"
)

type Codec interface {
	Receive(*ws.Conn, interface{}) error
	Send(*ws.Conn, interface{}) error
}

// Client is a websocket client meant to be used with the 1password
type Client struct {
	WebsocketURI      string
	WebsocketProtocol string
	WebsocketOrigin   string
	conn              *ws.Conn
	dial              func(string, string, string) (*ws.Conn, error)
	codec             Codec
}

func NewClient(websocketURI string, websocketProtocol string, websocketOrigin string) *Client {
	return NewCustomClient(websocketURI, websocketProtocol, websocketOrigin, ws.Dial, ws.Message)
}

func NewCustomClient(websocketURI string, websocketProtocol string, websocketOrigin string,
	dial func(string, string, string) (*ws.Conn, error), codec Codec) *Client {

	client := Client{
		WebsocketURI:      websocketURI,
		WebsocketProtocol: websocketProtocol,
		WebsocketOrigin:   websocketOrigin,
		dial:              dial,
		codec:             codec,
	}

	return &client
}

func (client *Client) Connect() error {
	conn, err := client.dial(client.WebsocketURI, client.WebsocketProtocol, client.WebsocketOrigin)
	if err != nil {
		return err
	}

	client.conn = conn

	return nil
}

func (client *Client) Close() {
	client.conn.Close()
}

func (client *Client) Receive() (string, error) {
	var s string
	err := client.codec.Receive(client.conn, &s)
	return s, err
}

func (client *Client) Send(v string) error {
	return client.codec.Send(client.conn, v)
}
