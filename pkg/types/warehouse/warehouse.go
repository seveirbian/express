package warehouse

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	uuid2 "github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/seveirbian/express/pkg/log"
	"github.com/seveirbian/express/pkg/types/courier"
	_package "github.com/seveirbian/express/pkg/types/package"
)

const (
	defatulStoreSize     = 10000
	defaultSendStoreSize = 1000

	defaultSorterNum = 3
)

type WareHouse struct {
	ctx context.Context

	wareHouseID string

	wareHouseAddr string
	wareHousePort int

	Store chan *_package.Package

	wareHouseCouriers sync.Map
}

func NewWareHouse(ctx context.Context, addr string, port int) (*WareHouse, error) {
	uuid, err := uuid2.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("[warehouse] failed to create a new warehouse")
	}

	return &WareHouse{
		ctx:               ctx,
		wareHouseID:       uuid.String(),
		wareHouseAddr:     addr,
		wareHousePort:     port,
		Store:             make(chan *_package.Package, defatulStoreSize),
		wareHouseCouriers: sync.Map{},
	}, nil
}

func (w *WareHouse) Work() {
	router := mux.NewRouter()
	router.HandleFunc("/express", w.doWork)

	server := http.Server{
		Addr:        fmt.Sprintf("%s:%d", w.wareHouseAddr, w.wareHousePort),
		Handler:     router,
		TLSConfig:   nil,
		BaseContext: nil,
	}

	for i := 0; i < defaultSorterNum; i++ {
		go func(i int) {
			log.SugarLogger.Infof("[sorter] sorter %d is working", i)
			for {
				pkg := <-w.Store
				log.SugarLogger.Infof("[warehouse] sorter %d get a package", i)

				wc, err := w.ReadWareHouseCourier(pkg.TargetLocationID)
				if err != nil {
					log.SugarLogger.Errorf("[warehouse] sorter %d dropped a package %v for not finding a warehouse"+
						" courier %s", i, pkg, pkg.TargetLocationID)
					continue
				}

				select {
				case wc.SendStore <- pkg:
					log.SugarLogger.Infof("[warehouse] sorter %d succeed to send a package to wc %s",
						i, pkg.TargetLocationID)
				default:
					log.SugarLogger.Infof("[warehouse] sorter %d failed to send a package %v to warehouse courier %s", i, pkg, pkg.TargetLocationID)
				}
			}
		}(i)
	}

	go func() {
		log.SugarLogger.Infof("[warehouse] start warehouse")
		err := server.ListenAndServe()
		log.SugarLogger.Errorf("[warehouse] failed to listen and serve for %v", err)
	}()
}

func (w *WareHouse) doWork(res http.ResponseWriter, req *http.Request) {
	log.SugarLogger.Infof("[warehouse] accept a new connection from %s", req.RemoteAddr)

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.SugarLogger.Errorf("[warehouse] failed to upgrade http to ws for %v", err)
		return
	}

	_, message, err := conn.ReadMessage()
	if err != nil {
		log.SugarLogger.Errorf("[warehouse] failed to read first register package for %v", err)
		conn.Close()
		return
	}

	pkg, err := _package.MessageToPackage(message)
	if err != nil {
		log.SugarLogger.Errorf("[warehouse] failed to parse message to package for %v", err)
		conn.Close()
		return
	}

	switch pkg.PackageType {
	case _package.PackageRegister:
		var wc *courier.WareHouseCourier

		wc, err := w.ReadWareHouseCourier(pkg.SourceLocationID)
		if err != nil {
			wc = courier.NewWareHouseCourier(pkg.SourceLocationID, conn, w.Store, make(chan *_package.Package, defaultSendStoreSize),
				make(chan *websocket.Conn))

			err := w.AddWareHouseCourier(pkg.SourceLocationID, wc)
			if err != nil {
				log.SugarLogger.Errorf("[warehouse] failed to add a warehouse courier to warehouse for %v", err)
				conn.Close()
				return
			}

			wc.Work()
			log.SugarLogger.Infof("[warehouse] succeed to register a new conn with id %s", pkg.SourceLocationID)
		} else {
			wc.UpdateConn(conn)
			log.SugarLogger.Infof("[warehouse] update warehouse courier from location %s with a new conn", pkg.SourceLocationID)
		}
	default:
		conn.Close()
		log.SugarLogger.Infof("[warehouse] receive a invalid package and dropped it")
	}
}

func (w *WareHouse) AddWareHouseCourier(id string, wc *courier.WareHouseCourier) error {
	_, ok := w.wareHouseCouriers.Load(id)
	if !ok {
		w.wareHouseCouriers.Store(id, wc)
		return nil
	}

	return fmt.Errorf("[warehouse courier manager] exists same id courier")
}

func (w *WareHouse) DeleteWareHouseCourier(id string) error {
	_, ok := w.wareHouseCouriers.Load(id)
	if !ok {
		log.SugarLogger.Infof("[warehouse] exists no warehouse courier with id %s", id)
		return nil
	}

	w.wareHouseCouriers.Delete(id)
	log.SugarLogger.Infof("[warehouse] delete warehouse courier with id %s", id)
	return nil
}

func (w *WareHouse) ReadWareHouseCourier(id string) (*courier.WareHouseCourier, error) {
	value, ok := w.wareHouseCouriers.Load(id)
	if !ok {
		return nil, fmt.Errorf("[warehouse courier manager] no exists courier with id %s", id)
	}

	wc, ok := value.(*courier.WareHouseCourier)
	if !ok {
		return nil, fmt.Errorf("[warehouse] value is not a valid warehouse courier")
	}

	return wc, nil
}
