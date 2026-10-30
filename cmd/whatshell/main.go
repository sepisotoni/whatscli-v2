package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/sepisotoni/whatshell-go/internal/config"
	"github.com/sepisotoni/whatshell-go/internal/doctor"
	"github.com/sepisotoni/whatshell-go/internal/store"
	"github.com/sepisotoni/whatshell-go/internal/ui"
	"github.com/sepisotoni/whatshell-go/internal/wa"
)

// These values are injected into tagged release builds.
var version = "dev"
var commit = "unknown"

func main() {
	if err := run(); err != nil {
		log.Println("whatshell:", err)
		os.Exit(1)
	}
}

func run() error {
	versionMode := flag.Bool("version", false, "print the WhatsCLI binary version and exit")
	doctorMode := flag.Bool("doctor", false, "check local storage and SQLite integrity without connecting to WhatsApp")
	demoMode := flag.Bool("demo", false, "run an isolated UI demo with fictional messages (no WhatsApp connection)")
	previewMode := flag.Bool("demo-onboarding", false, "preview first-run welcome and login without a WhatsApp connection")
	flag.Parse()
	if *versionMode {
		fmt.Printf("WhatsCLI %s (%s; %s)\n", version, commit, runtime.Version())
		return nil
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if *demoMode || *previewMode {
		return runDemo(ctx, *previewMode)
	}
	paths, err := config.Resolve()
	if err != nil {
		return err
	}
	if *doctorMode {
		return doctor.Run(ctx, paths, os.Stdout)
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
