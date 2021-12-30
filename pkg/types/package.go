package types

const (
	PackageCommon = iota
	PackageAck
	PackageHandbookUpdate
)

type Package struct {
	id               string
	packageType      string
	sourceLocationID string
	targetLocationID string
	content          []byte
}
