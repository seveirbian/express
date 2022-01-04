package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/seveirbian/express/pkg/types/warehouse"
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
