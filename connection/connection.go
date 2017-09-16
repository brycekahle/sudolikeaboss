package connection

type OnePasswordConnection interface {
	ReadJSON(v interface{}) error
	WriteJSON(v interface{}) error
	Close() error
}

func NewClient(url string, origin string) (OnePasswordConnection, error) {
	return NewWebsocketClient(url, origin)
}
