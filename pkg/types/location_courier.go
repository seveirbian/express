package types

import "github.com/gorilla/websocket"

type LocationCourier struct {
	courierID     string
	warehouseAddr string
	conn          *websocket.Conn
}

func NewLocationCourier(id string, addr string) (*LocationCourier, error) {
	return &LocationCourier{
		courierID:     id,
		warehouseAddr: addr,
		conn:          nil,
	}, nil
}
