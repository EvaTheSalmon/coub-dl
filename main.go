package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) < 2 {
		usage()
		return 64
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	hc := &http.Client{Timeout: 60 * time.Second}

	switch os.Args[1] {
	case "download":
		return cmdDownload(ctx, hc, os.Args[2:])
	case "sync":
		return cmdSync(ctx, hc, os.Args[2:])
	default:
		usage()
		return 64
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  coub-dl download [-out dir] [-name file] <link|permalink>")
	fmt.Fprintln(os.Stderr, "  coub-dl sync [-out dir] [-workers n]")
}
