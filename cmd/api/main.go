package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/knstch/course/internal/app"
	"github.com/knstch/course/internal/app/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}

	if err := config.InitENV(dir); err != nil {
		return err
	}

	config := config.GetConfig()

	_, err = app.InitContainer(config)
	if err != nil {
		return err
	}

	srv := http.Server{
		Addr:    config.Port,
		Handler: nil,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Print(err)
		}
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Print(err)
		return err
	}

	<-idleConnsClosed
	return nil
}
