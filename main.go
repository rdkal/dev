package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
// FlagServer = flag.String("proxy", "http://localhost:8080", "start a dev server that forwards request to the given url")
// FlagPort   = flag.String("port", ":8081", "port of the dev server")
)

func main() {
	// flag.Parse()
	runtime, err := NewRuntime()
	exitOnError(err)
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-ctx.Done()
		time.Sleep(10 * time.Second)
		panic("gracefull exit failed")
	}()
	exitOnError(runtime.Run(ctx))
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
