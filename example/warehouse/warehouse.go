package main

import (
	"context"
	"express/pkg/types/warehouse"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	ctx := context.Background()

	wh, err := warehouse.NewWareHouse(ctx, "0.0.0.0", 10002)
	if err != nil {
		fmt.Printf("failed to new a warehouse for %v", err)
		return
	}

	wh.Work()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
}
