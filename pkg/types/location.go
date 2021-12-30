package types

import (
	"sync"
)

type Location struct {
	locationID string

	sendStore    sync.Map
	receiveStore sync.Map

	courier *LocationCourier
}

func NewLocation(id string) *Location {
	return &Location{
		locationID:   id,
		receiveStore: sync.Map{},
		sendStore:    sync.Map{},
	}
}

func (l *Location) SendPackage(pkg *Package) {

}

func (l *Location) ReceivePackage() *Package {
	return nil
}
