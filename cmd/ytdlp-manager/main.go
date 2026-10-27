package main

import (
	"context"
	_ "embed"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/shinwookim/ytdlp-manager/internal/app"
	"github.com/shinwookim/ytdlp-manager/internal/store"
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	dbPath := env("DB_PATH", "./data/ytdlp-manager.db")
	downloadPath := env("DOWNLOAD_PATH", "./downloads")
	addr := env("HTTP_ADDR", "127.0.0.1:8080")

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(downloadPath, 0755); err != nil {
		log.Fatal(err)
	}

	conn, err := store.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	if err := store.Init(conn); err != nil {
		log.Fatal(err)
	}

	application, err := app.New(conn, downloadPath)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go application.Worker.Run(ctx)

	server := &http.Server{
		Addr:              addr,
		Handler:           application.HTTP.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(), 10*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP shutdown: %v", err)
		}
	}()

	log.Printf("Listening on %s", addr)

	if err := server.Serve(listener); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
