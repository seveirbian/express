package courier

import (
	"express/pkg/types/package"
	"github.com/gorilla/websocket"
	"time"
)

type WareHouseCourier struct {
	connCh    chan *websocket.Conn
	SendStore chan *_package.Package
	*Courier
}

func NewWareHouseCourier(id string, c *websocket.Conn, receiveStore chan *_package.Package,
	sendStore chan *_package.Package, connCh chan *websocket.Conn) *WareHouseCourier {
	courier := NewCourier(id, c, receiveStore, sendStore)
	return &WareHouseCourier{
		connCh:    connCh,
		SendStore: sendStore,
		Courier:   courier,
	}
}

func (w *WareHouseCourier) UpdateConn(conn *websocket.Conn) {
	w.connCh <- conn
}

func (w *WareHouseCourier) connManage(stopCh chan struct{}) {
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			w.connSwitch = true
			w.conn.Close()
		case c := <-w.connCh:
			w.Update(c)
			w.connSwitch = false
		case <-ticker.C:
			w.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(w.writeDeadline))
			w.conn.SetWriteDeadline(time.Now().Add(w.writeDeadline))
			w.conn.SetReadDeadline(time.Now().Add(w.readDeadline))
		}
	}
}

func (w *WareHouseCourier) Work() {
	stopCh := make(chan struct{})

	go w.connManage(stopCh)
	go w.ReadLoop(stopCh)
	go w.WriteLoop(stopCh)
}
