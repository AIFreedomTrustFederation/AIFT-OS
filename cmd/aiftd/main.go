package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/uxi"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/uxihttp"
)

func main() {
	addr := flag.String("addr", envOr("AIFT_UXI_ADDR", "127.0.0.1:8787"), "loopback listen address")
	dataRoot := flag.String("data", defaultDataRoot(), "UXI data directory")
	modelURL := flag.String("model-url", envOr("AIFT_MODEL_URL", "http://127.0.0.1:8080/v1"), "OpenAI-compatible local model endpoint")
	modelName := flag.String("model", envOr("AIFT_MODEL_NAME", "local"), "local model name")
	flag.Parse()

	if err := requireLoopback(*addr); err != nil {
		log.Fatal(err)
	}
	aiftRoot, err := uxi.ResolveAIFTRoot()
	if err != nil {
		log.Fatal(err)
	}
	store, err := uxi.NewStore(*dataRoot)
	if err != nil {
		log.Fatal(err)
	}
	engine, err := uxi.NewEngine(store, aiftRoot, uxi.NewLocalInferenceClient(*modelURL, *modelName))
	if err != nil {
		log.Fatal(err)
	}
	uxiServer, err := uxihttp.New(engine)
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:              *addr,
		Handler:           uxiServer.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("MoBox UXI listening on http://%s", *addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func requireLoopback(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid listen address: %w", err)
	}
	if host == "localhost" || host == "" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("refusing non-loopback address %q; MoBox UXI is local-first", addr)
	}
	return nil
}

func defaultDataRoot() string {
	if value := strings.TrimSpace(os.Getenv("AIFT_UXI_HOME")); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".aift", "uxi")
	}
	return filepath.Join(home, ".aift", "uxi")
}

func envOr(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
