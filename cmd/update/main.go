package main

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-spatial-hierarchy/app/update"
	"log"
)

func main() {

	ctx := context.Background()
	logger := log.Default()

	err := update.Run(ctx, logger)

	if err != nil {
		logger.Fatalf("Failed to run PIP application, %v", err)
	}

}
