package courier

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"

	"github.com/seveirbian/express/pkg/log"
	_package "github.com/seveirbian/express/pkg/types/package"
)

type Courier struct {
	courierID string

	receiveStore chan<- *_package.Package
	sendStore    <-chan *_package.Package

	conn       *websocket.Conn
	connSwitch bool

	checkInterval time.Duration
	readDeadline  time.Duration
	writeDeadline time.Duration

	waitTime time.Duration
}

func NewCourier(id string, conn *websocket.Conn, receiveStore chan<- *_package.Package,
	sendStore <-chan *_package.Package) *Courier {
	return &Courier{
		courierID: id,

		receiveStore: receiveStore,
		sendStore:    sendStore,

		conn:       conn,
		connSwitch: false,

		checkInterval: defaultConnCheckInterval,
		readDeadline:  defaultReadDeadline,
		writeDeadline: defaultWriteDeadline,

		waitTime: defaultWaitTime,
	}
}

func (c *Courier) Update(conn *websocket.Conn) {
	c.conn = conn
}

func (c *Courier) ReadLoop(stopCh chan struct{}) {
	for {
		if c.connSwitch {

			log.SugarLogger.Infof("[location courier] id %s read loop wait for conn switch", c.courierID)
			time.Sleep(c.waitTime)

			// make next wait time longer
			func() {
				if c.waitTime < maxWaitTime {
					c.waitTime = c.waitTime * 2
				}
			}()

			continue
		}

		_, data, err := c.conn.ReadMessage()
		if err != nil {
			stopCh <- struct{}{}
			log.SugarLogger.Errorf("[location courier] id %s failed to read message for %v\n", c.courierID, err)
			continue
		}

		var pkg _package.Package
		err = json.Unmarshal(data, &pkg)
		if err != nil {
			log.SugarLogger.Errorf("[location courier] id %s failed to unmarshal data to package for %v\n", c.courierID, err)
			continue
		}

		c.receiveStore <- &pkg
	}
}

func (c *Courier) WriteLoop(stopCh chan struct{}) {
	for {
		if c.connSwitch {
			log.SugarLogger.Infof("[location courier] id %s write loop wait for conn switch", c.courierID)
			time.Sleep(time.Second)
			continue
		}

		pkg := <-c.sendStore

		log.SugarLogger.Infof("[courier] id %s receive a pkg", c.courierID)

		data, err := json.Marshal(pkg)
		if err != nil {
			log.SugarLogger.Errorf("[location courier] id %s failed to marshal package for %v\n", c.courierID, err)
			continue
		}

		err = c.conn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			stopCh <- struct{}{}
			log.SugarLogger.Errorf("[location courier] id %s failed to write message for %v\n", c.courierID, err)
			continue
		}

		log.SugarLogger.Infof("[courier] id %s succeed to send a pkg to conn", c.courierID)
	}
}
