package main

import (
	"context"
	"express/pkg/types/location"
	"fmt"
)

func main() {
	ctx := context.Background()
	lc, err := location.NewLocation(ctx, "location-1", "192.168.0.250:10002", "/express")
	if err != nil {
		fmt.Printf("err %v", err)
	}

	lc.Run()

	for {
		//var input string
		//_, err := fmt.Scanf("%s", &input)
		//if err != nil {
		//	fmt.Printf("[client] fails to read input, err %v", err)
		//}
		//
		//pkg := _package.Package{
		//	ID:               "",
		//	PackageType:      _package.PackageCommon,
		//	SourceLocationID: "location-1",
		//	TargetLocationID: "location-1",
		//	Content:          []byte(input),
		//}
		//
		//lc.SendPackage(&pkg)
		fmt.Printf("received pacakge %v", lc.ReceivePackage())
	}
}
