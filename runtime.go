package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Runtime struct {
	DevServerAddr string
	UserServerURL string
	Command       []string
	Watcher       *Watcher
	Throttle      time.Duration

	refreshUI chan struct{}

	// shutdown is called when an error occurs
	shutdown func()
	// wg is used for gracefull shutdown
	wg sync.WaitGroup
	// err is availible after wg is done
	err     error
	errOnce sync.Once
}

func NewRuntime() (*Runtime, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	watcher, err := NewWatcher(wd)
	if err != nil {
		return nil, err
	}

	return &Runtime{
		DevServerAddr: ":8081",
		UserServerURL: "http://localhost:8080",
		Watcher:       watcher,
		Command:       []string{"go", "run", "."},
		Throttle:      5 * time.Millisecond,

		refreshUI: make(chan struct{}),
	}, nil
}

func (r *Runtime) Run(ctx context.Context) error {
	ctx, r.shutdown = context.WithCancel(ctx)

	go r.start(ctx, r.server)
	go r.start(ctx, r.Watcher.Start)

	executor := NewExecutor(r.Command)
	err := executor.Start(ctx)
	if err != nil {
		r.notify(err)
	}

	refresh := throttle(r.Throttle, r.Watcher.Events())
	for {
		select {
		case <-ctx.Done():
			executor.Wait(context.Background())
			r.wg.Wait()
			return r.err
		case <-refresh:
			err := executor.Restart(ctx)
			if err != nil {
				r.notify(err)
			}
			r.refreshUI <- struct{}{}

		}
	}
}

func (r *Runtime) notify(msg ...any) {
	fmt.Println(append([]any{"notify"}, msg...))
}

func (r *Runtime) server(ctx context.Context) error {
	fmt.Println("dev-server listening on", r.DevServerAddr)
	proxy, err := NewProxy(r.UserServerURL)
	exitOnError(err)
	proxy.Inject = `
	<script>
		const source = new EventSource("/__dev-server__")
		window.addEventListener("beforeunload", () => {
			source.close()
	   	})
		source.onmessage = (event) => {
			console.log('refresh')
			window.location.reload()
		}
  	</script>`
	proxy.Do = func(req *http.Request) (*http.Response, error) {
		return http.DefaultClient.Do(req)
	}

	listeners := []func() (stop bool){}
	subscribe := make(chan func() bool)
	go func() {
		for {
			select {
			case f := <-subscribe:
				listeners = append(listeners, f)
			case <-r.refreshUI:
				for i := 0; i < len(listeners); {
					stop := listeners[i]()
					if stop {
						end := len(listeners) - 1
						listeners[i] = listeners[end]
						listeners = listeners[:end]
						continue
					}
					i++
				}
			}
		}
	}()

	// nooplog := log.New(io.Discard, "", 0)
	server := http.Server{
		// ErrorLog: nooplog,
		Addr: r.DevServerAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/__dev-server__" {
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				w.Header().Set("Transfer-Encoding", "chunked")

				subscribe <- func() (stop bool) {
					if r.Context().Err() != nil {
						return true
					}
					go func() {
						w.Write([]byte("data: refresh\n\n"))
						if err != nil {
							log.Println("from write", err)
						}
						if f, ok := w.(http.Flusher); ok {
							f.Flush()
						}
					}()
					return false
				}
				<-r.Context().Done()
				return
			}
			proxy.ServeHTTP(w, r)
		})}
	go func() {
		<-ctx.Done()
		server.Close()
	}()

	err = server.ListenAndServe()
	if err == http.ErrServerClosed {
		err = nil
	}
	return err
}

func (r *Runtime) start(ctx context.Context, f func(ctx context.Context) error) {
	r.wg.Add(1)
	defer r.wg.Done()
	err := f(ctx)
	if err == nil {
		return
	}
	r.errOnce.Do(func() { r.err = err })
	r.shutdown()
}

func throttle[T any](dur time.Duration, src <-chan T) <-chan T {
	dst := make(chan T, cap(src))
	go func() {
		defer close(dst)
		first := <-src
		dst <- first
		emit := time.After(dur)
		next, ok := <-src
		if !ok {
			return
		}
		for v := range src {
			select {
			case <-emit:
				dst <- next
				emit = time.After(dur)
				next = v
			default:
			}
		}
		dst <- next
	}()

	return dst
}
