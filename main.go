package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	FlagUserServerURL = flag.String("forward-to-url", "http://localhost:8080", "forwards request to the given url")
	FlagPort          = flag.Int("port", 8081, "port of the dev server")
)

func main() {
	flag.CommandLine.Usage = func() {
		printLn(`Dev helps ya do go development stuff
		
Usage:

  dev [OPTION]... [CMD] [ARG]...
`)
		flag.CommandLine.PrintDefaults()
	}
	flag.Parse()
	runtime, err := NewRuntime()
	exitOnError(err)

	runtime.DevServerAddr = fmt.Sprintf(":%d", *FlagPort)
	runtime.UserServerURL = *FlagUserServerURL
	runtime.Command = []string{"go", "run", "."}
	if args := flag.Args(); len(args) > 0 {
		runtime.Command = args
	}
	printLn("cmd:", runtime.Command)

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-ctx.Done()
		time.Sleep(10 * time.Second)
		panic("gracefull exit failed")
	}()
	exitOnError(runtime.Run(ctx))
}

func printLn(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
