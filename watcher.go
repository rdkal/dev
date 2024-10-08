package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	Dir          string
	ExcludeFiles []string
	ExcludeDirs  []string
	IncludeFiles []string
	Debug        bool

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
	err := w.validateOptions()
	if err != nil {
		return err
	}
	defer w.watcher.Close()
	err = w.watch(ctx, w.Dir)
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
			if w.Debug {
				log.Println("event", event)
			}
			match, err := w.shouldExcludeFile(event)
			if err != nil {
				return err
			}
			if match {
				if w.Debug {
					log.Println("excluded", event)
				}
				continue
			}
			match, err = w.shouldIncludeFile(event)
			if err != nil {
				return err
			}
			if !match {
				if w.Debug {
					log.Println("not included", event)
				}
				continue
			}
			if w.Debug {
				log.Println("send", event)
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

func (w *Watcher) validateOptions() error {
	for _, pattern := range w.ExcludeDirs {
		_, err := filepath.Match(pattern, "")
		if err != nil {
			return err
		}
		if filepath.IsAbs(pattern) {
			return fmt.Errorf("exclude dir %q must be relative", pattern)
		}
	}
	err := validateFilePattern(w.ExcludeFiles...)
	if err != nil {
		return err
	}
	err = validateFilePattern(w.IncludeFiles...)
	if err != nil {
		return err
	}
	return nil
}

func validateFilePattern(patterns ...string) error {
	for _, pattern := range patterns {
		_, err := filepath.Match(pattern, "")
		if err != nil {
			return err
		}
		if filepath.Dir(pattern) != "." {
			return fmt.Errorf("exclude file %q must be a file name without a directory: %s", pattern, filepath.Dir(pattern))
		}
	}
	return nil
}

func (w *Watcher) shouldExcludeFile(event fsnotify.Event) (bool, error) {
	for _, glob := range w.ExcludeFiles {
		match, err := filepath.Match(glob, filepath.Base(event.Name))
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

func (w *Watcher) shouldIncludeFile(event fsnotify.Event) (bool, error) {
	if len(w.IncludeFiles) == 0 {
		return true, nil
	}
	for _, glob := range w.IncludeFiles {
		match, err := filepath.Match(glob, filepath.Base(event.Name))
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
	reldir, err := filepath.Rel(w.Dir, dir)
	if err != nil {
		return err
	}
	for _, pattern := range w.ExcludeDirs {
		match, err := filepath.Match(pattern, reldir)
		if err != nil {
			return err
		}
		if w.Debug {
			fmt.Fprintln(os.Stderr, "dir", reldir, "pattern:", pattern, "match:", match)
		}
		if match {
			return nil
		}
	}
	err = w.watcher.Add(dir)
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
