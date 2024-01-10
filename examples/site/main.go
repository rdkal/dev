package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

//go:embed *.html
var FS embed.FS

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<-ctx.Done()
		os.Exit(1)
	}()
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(os.Getpid())
		http.FileServer(http.FS(FS)).ServeHTTP(w, r)
	})))
}
