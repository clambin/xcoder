package main

import (
	"context"
	"fmt"
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

	if err := videoConvertor.Run(ctx, os.Stderr); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "failed to run. error:", err.Error())
		os.Exit(1)
	}

}
