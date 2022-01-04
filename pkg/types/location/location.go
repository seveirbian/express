package location

import (
	"context"
	"express/pkg/log"
	"express/pkg/types/courier"
	"express/pkg/types/package"
	"fmt"
)

const defaultStoreSize = 1000

type Location struct {
	ctx context.Context

	locationID string

	sendStore    chan *_package.Package
	receiveStore chan *_package.Package

	locationCourier *courier.LocationCourier
}

func NewLocation(ctx context.Context, id string, wareHouseAddr string, wareHousePath string) (*Location, error) {
	rStore := make(chan *_package.Package, defaultStoreSize)
	sStore := make(chan *_package.Package, defaultStoreSize)

	lc, err := courier.NewLocationCourier(id, wareHouseAddr, wareHousePath, rStore, sStore)
	if err != nil {
		return nil, fmt.Errorf("[location] failed to new a location courier for %v\n", err)
	}

	return &Location{
		ctx:             ctx,
		locationID:      id,
		receiveStore:    rStore,
		sendStore:       sStore,
		locationCourier: lc,
	}, nil
}

func (l *Location) SendPackage(pkg *_package.Package) {
	select {
	case l.sendStore <- pkg:
		log.SugarLogger.Infof("[location] succeed to give a package to location courier")
	default:
		log.SugarLogger.Errorf("[location] failed to give package to location courier for it is full, "+
			"and this package %v will be discarded", pkg)
	}
}

func (l *Location) ReceivePackage() *_package.Package {
	pkg := <-l.receiveStore
	return pkg
}

func (l *Location) Run() {
	l.locationCourier.Work()
}
