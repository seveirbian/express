package courier

import (
	"encoding/json"
	"express/pkg/log"
	"express/pkg/types/package"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"time"
)

type LocationCourier struct {
	wareHouseAddr string
	wareHousePath string

	*Courier
}

func registerToWareHouse(locationID string, wareHouseAddr string, wareHousePath string) (*websocket.Conn, error) {
	pkg := &_package.Package{
		ID:               "",
		PackageType:      _package.PackageRegister,
		SourceLocationID: locationID,
		TargetLocationID: "",
		Content:          nil,
	}

	return connectToWareHouse(wareHouseAddr, wareHousePath, pkg)
}

func connectToWareHouse(wareHouseAddr string, wareHousePath string, pkg *_package.Package) (*websocket.Conn, error) {
	wsURL := url.URL{Scheme: "ws", Host: wareHouseAddr, Path: wareHousePath}

	// DefaultDialer do the handshake in it self
	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), http.Header{})
	if err != nil {
		return nil, fmt.Errorf("[location courier] failed to connect to warehouse for %v\n", err)
	}

	// TODO: send a register package to warehouse
	data, err := json.Marshal(pkg)
	if err != nil {
		return nil, fmt.Errorf("[location courier] failed to marshal register pacakge for %v", err)
	}

	err = conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		return nil, fmt.Errorf("[location courier] failed to send register pacakge for %v", err)
	}

	log.Logger.Sugar().Infof("[location courier] succeed to connect to warehouse addr %s path %s\n", wareHouseAddr, wareHousePath)

	return conn, nil
}

func NewLocationCourier(id string, wareHouseAddr string, wareHousePath string,
	receiveStore chan<- *_package.Package, sendStore <-chan *_package.Package) (*LocationCourier, error) {
	c, err := registerToWareHouse(id, wareHouseAddr, wareHousePath)
	if err != nil {
		return nil, err
	}

	courier := NewCourier(id, c, receiveStore, sendStore)

	return &LocationCourier{
		wareHouseAddr: wareHouseAddr,
		wareHousePath: wareHousePath,
		Courier:       courier,
	}, nil
}

func (l *LocationCourier) connManage(stopCh chan struct{}) {
	ticker := time.NewTicker(l.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			l.connSwitch = true
			l.conn.Close()
			for {
				c, err := registerToWareHouse(l.courierID, l.wareHouseAddr, l.wareHousePath)
				if err != nil {
					log.SugarLogger.Errorf("[location courier] check conn, find a disconnect conn, "+
						"but failed to create a new conn for %v\n", err)
					continue
				}
				l.Update(c)
				break
			}
			l.connSwitch = false
		case <-ticker.C:
			l.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(l.writeDeadline))
			l.conn.SetWriteDeadline(time.Now().Add(l.writeDeadline))
			l.conn.SetReadDeadline(time.Now().Add(l.readDeadline))
		}
	}
}

func (l *LocationCourier) Work() {
	stopCh := make(chan struct{})

	go l.connManage(stopCh)
	go l.ReadLoop(stopCh)
	go l.WriteLoop(stopCh)
}
