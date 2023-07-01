package main

import (
	"context"
	"embed"
	"fmt"

	"github.com/titpetric/task-ui/server"
)

var (
	//go:embed templates/*.tpl public_html/static/*
	files embed.FS
)

func start(ctx context.Context) error {
	svc, err := server.New(&files)
	if err != nil {
		return err
	}
	return svc.Start(ctx)
}

func main() {
	ctx := context.Background()
	if err := start(ctx); err != nil {
		fmt.Println("Got error:", err)
	}
	fmt.Println("Exiting")
}
