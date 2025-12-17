package conn

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Conn struct {
	conn        *websocket.Conn
	connReadMu  sync.Mutex
	connWriteMu sync.Mutex
}

func NewConn(websocket *websocket.Conn) *Conn {
	return &Conn{
		conn: websocket,
	}
}

func (c *Conn) ReadMessage() (int, []byte, error) {
	c.connReadMu.Lock()
	defer c.connReadMu.Unlock()

	return c.conn.ReadMessage()
}

func (c *Conn) WriteJSON(v any) error {
	c.connWriteMu.Lock()
	defer c.connWriteMu.Unlock()

	return c.conn.WriteJSON(v)
}

func (c *Conn) Close() error {
	c.connWriteMu.Lock()
	defer c.connWriteMu.Unlock()

	c.connReadMu.Lock()
	defer c.connReadMu.Unlock()

	return c.conn.Close()
}
