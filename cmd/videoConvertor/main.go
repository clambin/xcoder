package main

import (
	"context"
	"github.com/clambin/videoConvertor/internal/cmd/videoConvertor"
	"net/http"
	"os"
	"os/signal"

	_ "net/http/pprof"
)

func main() {
	go func() {
		_ = http.ListenAndServe(":6060", nil)
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	_ = videoConvertor.Run(ctx, os.Stderr)
}
