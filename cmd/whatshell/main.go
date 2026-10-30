package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sepisotoni/whatshell-go/internal/config"
	"github.com/sepisotoni/whatshell-go/internal/store"
	"github.com/sepisotoni/whatshell-go/internal/ui"
	"github.com/sepisotoni/whatshell-go/internal/wa"
)

func main() {
	if err := run(); err != nil {
		log.Println("whatshell:", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	paths, err := config.Resolve()
	if err != nil {
		return err
	}
	for _, path := range []string{paths.History, paths.Session} {
		if err := config.ProtectFile(path); err != nil {
			return fmt.Errorf("protect private file: %w", err)
		}
	}
	archive, err := store.Open(ctx, paths.History)
	if err != nil {
		return fmt.Errorf("open message history: %w", err)
	}
	defer func() { _ = archive.Close() }()
	client, err := wa.New(ctx, paths.Session, archive)
	if err != nil {
		return fmt.Errorf("create whatsapp client: %w", err)
	}
	defer func() { _ = client.Close() }()
	if err := ui.New(archive, client, paths).Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}
