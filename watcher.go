package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	Dir          string
	ExcludeGlobs []string

	watcher *fsnotify.Watcher
	out     chan fsnotify.Event
}

func NewWatcher(dir string) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		Dir:     dir,
		watcher: watcher,
		out:     make(chan fsnotify.Event),
	}, nil
}

func (w *Watcher) Start(ctx context.Context) error {
	defer w.watcher.Close()
	err := w.watch(ctx, w.Dir)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-w.watcher.Events:
			if event.Op.Has(fsnotify.Create) {
				err := w.watchIfDirectory(ctx, event)
				if err != nil {
					return err
				}
			}
			shouldMute, err := w.shouldMute(event)
			if err != nil {
				return err
			}
			if shouldMute {
				continue
			}
			w.out <- event
		case err := <-w.watcher.Errors:
			return err
		}
	}
}

func (w *Watcher) Events() <-chan fsnotify.Event {
	return w.out
}

func (w *Watcher) shouldMute(event fsnotify.Event) (bool, error) {
	for _, glob := range w.ExcludeGlobs {
		rel, err := filepath.Rel(w.Dir, event.Name)
		if err != nil {
			return false, err
		}
		match, err := filepath.Match(glob, rel)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

func (w *Watcher) watchIfDirectory(ctx context.Context, event fsnotify.Event) error {
	stat, err := os.Stat(event.Name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !stat.IsDir() {
		return nil
	}
	return w.watch(ctx, event.Name)
}

func (w *Watcher) watch(ctx context.Context, dir string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	err := w.watcher.Add(dir)
	if err != nil {
		return err
	}
	entires, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entires {
		if !entry.IsDir() {
			continue
		}
		err := w.watch(ctx, filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}
