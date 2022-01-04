package _package

import (
	"encoding/json"
	"fmt"
)

const (
	PackageCommon = iota
	PackageRegister
)

type Package struct {
	ID               string `json:"id"`
	PackageType      int    `json:"package_type"`
	SourceLocationID string `json:"source_location_id"`
	TargetLocationID string `json:"target_location_id"`
	Content          []byte `json:"content"`
}

func MessageToPackage(data []byte) (*Package, error) {
	var pkg Package

	err := json.Unmarshal(data, &pkg)
	if err != nil {
		return nil, fmt.Errorf("[package] failed to unmarshal data to package for %v", err)
	}

	return &pkg, nil
}
