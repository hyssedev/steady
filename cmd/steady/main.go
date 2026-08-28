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

	"github.com/hyssedev/steady/internal/config"
)

type App struct {
	HttpServer *http.Server
}

func NewApp(cfg *config.Config) App {
	server := &http.Server{
		Addr: cfg.Listen,
	}

	return App{
		HttpServer: server,
	}
}

func Run() error {
	cfg, err := config.ReadConfig("config.yaml")
	if err != nil {
		return err
	}

	app := NewApp(&cfg)

	go func() {
		log.Printf("listening on %s\n", app.HttpServer.Addr)

		if err := app.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error listening and serving: %s\n", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Print("shutting down server ...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.HttpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shut down server: %w", err)
	}

	log.Print("server shut down")

	return nil
}

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}
