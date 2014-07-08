package main

import (
	"flag"
	"github.com/Topface/diplodocus"
	"github.com/howeyc/fsnotify"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var listen = flag.String("listen", "", "addreess to listen: 127.0.0.1:8000")
var root = flag.String("root", "", "root dir for logs")

func main() {
	flag.Parse()

	if *listen == "*" || *root == "" {
		flag.PrintDefaults()
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	m := diplodocus.NewFileManager(watcher)
	go func() {
		err := m.Watch()
		if err != nil {
			log.Fatal(err)
		}
	}()

	err = filepath.Walk(*root, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if f.IsDir() {
			log.Println("watching", path)
			return watcher.Watch(path)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		path := *root + req.URL.Path

		file, err := m.GetFile(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		listener := make(diplodocus.Listener)
		file.AddListener(listener)

		log.Println("listener added for:", path)

		var flusher http.Flusher
		if f, ok := w.(http.Flusher); ok {
			flusher = f
		} else {
			log.Println("damn, no flushing")
		}

		for ev := range listener {
			if ev.Error != nil {
				log.Println("error for "+path, ev.Error)
				file.RemoveListener(listener)
				return
			}

			_, err := w.Write(*ev.Buffer)
			if err != nil {
				log.Println("error for "+path, err)
				file.RemoveListener(listener)
				return
			}

			flusher.Flush()
		}
	})

	log.Println("listening", *listen)

	err = http.ListenAndServe(*listen, mux)
	if err != nil {
		log.Fatal(err)
	}
}