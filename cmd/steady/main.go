package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	HttpServer *http.Server
}

func NewApp() (App, error) {
	server := &http.Server{
		Addr: ":8080",
	}

	return App{
		HttpServer: server,
	}, nil
}

func Run() error {
	app, err := NewApp()
	if err != nil {
		panic("error setting up app")
	}

	go func() {
		log.Printf("listening on %s\n", app.HttpServer.Addr)

		if err := app.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error listening and serving: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Print("shutting down server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.HttpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("error when shutting down server: %v", err)
	}

	log.Print("server shut down")

	return nil
}
