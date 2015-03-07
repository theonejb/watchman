/*
Package pkgwatcher provides a utility to watch a package and all it's non-std-lib imported packages
via fsnotify.

Inspired by github.com/codegangsta/gin. While gin is amazing, one issue I had was that if I changed
a file in a package that is *imported* by the package that gin is watching, gin wouldn't notice that.

This isn't meant to be a replacement for gin. It's supposed to be plugged into it so that the pkg watching
in gin can be improved.
*/
package watchman

import (
	"fmt"
	"gopkg.in/fsnotify.v1"
	"path/filepath"
	"time"
)

func WatchPackageAndReturnOnChange(path string) {
	importPaths, err := getImportPaths(path)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	nonRootPkgs := getPathsForNonRootPkgs(importPaths)
	// Also watch the current path
	nonRootPkgs = append(nonRootPkgs, path)
	watcher := addWatcherToPackages(nonRootPkgs)

	// We need a label here (makes me sad as well) because we need to break out of the loop from inside the select
EventWatcherLoop:
	for {
		select {
		case ev := <-watcher.Events:
			// We only break if we get an event on a .go file
			if filepath.Ext(ev.Name) != ".go" {
				continue
			}
			break EventWatcherLoop
		case <-watcher.Errors:
			break EventWatcherLoop
		}
	}

	// Wait 100ms to get any other events related to the first one and then drain them.
	// We need this because during testing, vim wrote to the file twice and then changed the mod.
	// Waiting and draining makes sure we only get 1 event in that specific case. Other editors
	// will probably need a different way of handling this
	time.Sleep(100 * time.Millisecond)
	for {
		select {
		case <-watcher.Events:
		case <-watcher.Errors:
			continue
		default:
			return
		}
	}

	watcher.Close()
}

func addWatcherToPackages(pkgPaths []string) *fsnotify.Watcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err.Error())
	}

	for _, path := range pkgPaths {
		err := watcher.Add(path)
		if err != nil {
			fmt.Printf("Unable to watch package %s. Error: %s\n", path, err.Error())
		}
	}

	return watcher
}
