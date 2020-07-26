package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hpcloud/tail"
)

type fsAdapted struct {
	Handler http.Handler
}

func (h *fsAdapted) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Index(r.URL.Path, ".") == -1 {
		r.URL.Path = "/"
	}
	h.Handler.ServeHTTP(w, r)
}

func main() {
	var file = os.Args[1]
	fs := &fsAdapted{http.FileServer(http.Dir("./dist"))}
	http.Handle("/", fs)
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		t, _ := tail.TailFile(file, tail.Config{Follow: true})
		// for line := range t.Lines {
		// 	fmt.Println(line.Text)
		// }
		flusher, ok := w.(http.Flusher)
		if !ok {
			panic("expected http.ResponseWriter to be an http.Flusher")
		}
		//		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

	LOOP:
		for {
			select {
			case line := <-t.Lines:
				{
					fmt.Fprintf(w, "event: log\ndata: %s\n\n", line.Text)
					flusher.Flush() // Trigger "chunked" encoding and send a chunk...
				}
			case <-r.Context().Done():
				{
					t.Stop()
					log.Println("Aborted.")
					break LOOP
				}
			}

		}

	})

	log.Print("Listening on localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
