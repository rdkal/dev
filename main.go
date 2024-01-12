package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := GetConfig()
	exitOnError(err)

	flag.CommandLine.Usage = func() {
		printLn(`Dev helps ya do go development stuff

Usage:

  dev [OPTION] <command>

Commands:
  init		create .dev.toml with default settings in current directory 
`)
		flag.CommandLine.PrintDefaults()
	}
	flag.Parse()
	switch flag.Arg(0) {
	case "init":
		err := InitConifg()
		exitOnError(err)
		printLn(".dev.toml created with default settings")
		return
	default:
	}

	runtime, err := NewRuntime()
	exitOnError(err)

	runtime.DevServerAddr = fmt.Sprintf(":%d", cfg.DevServerPort)
	runtime.UserServerURL = cfg.FowardToURL
	runtime.Watcher.ExcludeFiles = cfg.ExcludeFiles
	runtime.Command = cfg.Command
	printLn("cmd:", runtime.Command)

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-ctx.Done()
		time.Sleep(10 * time.Second)
		panic("gracefull exit failed")
	}()
	exitOnError(runtime.Run(ctx))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
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

func printJSON(v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	check(err)
	fmt.Println(string(b))

}
