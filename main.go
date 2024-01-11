package main

import (
	"context"
	"fmt"

	"github.com/SiddyP/gotibber/tibber"
)

func main() {
	ctx := context.Background()
	fmt.Println("starting..")

	t := tibber.Client{}

	t.Init()

	t.Run(ctx)

	t.Query(ctx)
}
