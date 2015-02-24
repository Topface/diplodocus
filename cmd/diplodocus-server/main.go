package main

import (
	"flag"
	"github.com/Topface/diplodocus"
	"github.com/fsnotify/fsnotify"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	listen := flag.String("listen", "", "addreess to listen, example: 127.0.0.1:8000")
	root := flag.String("root", "", "root dir for logs")
	flag.Parse()

	if *listen == "" || *root == "" {
		flag.PrintDefaults()
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("watcher creation error:", err)
	}

	m := diplodocus.NewFileManager(watcher)
	go func() {
		err := m.Watch()
		if err != nil {
			log.Fatal("watch error:", err)
		}
	}()

	err = filepath.Walk(*root, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if f.IsDir() {
			log.Println("watching", path)
			return watcher.Add(path)
		}

		return nil
	})

	if err != nil {
		log.Fatal("watch path adding error:", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		path := *root + req.URL.Path

		file, err := m.GetFile(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		listener := make(diplodocus.Listener, 10)
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
				log.Println("listener error for", path, ev.Error)
				file.RemoveListener(listener)
				return
			}

			_, err := w.Write(*ev.Buffer)
			if err != nil {
				log.Println("write error for", path, err)
				file.RemoveListener(listener)
				return
			}

			flusher.Flush()
		}

		log.Println("listener finished (probably timeout) for", path)
	})

	log.Println("listening", *listen)

	err = http.ListenAndServe(*listen, mux)
	if err != nil {
		log.Fatal("listen error:", err)
	}
}
