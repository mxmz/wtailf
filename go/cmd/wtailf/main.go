package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/papertrail/go-tail/follower"
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

func test(c int) {

	t, _ := follower.New("../../.temp", follower.Config{
		Whence: io.SeekEnd,
		Offset: 0,
		Reopen: true,
	})
	var lines = t.Lines()
	for {
		select {
		case line := <-lines:
			{
				log.Println(c)
				log.Println(line)
			}
		}
	}

}

func main() {

	// go test(1)
	// go test(2)
	// go test(3)

	// <-time.Tick(60 * time.Minute)

	var file = os.Args[1]
	fs := &fsAdapted{http.FileServer(http.Dir("./dist"))}
	http.Handle("/", fs)
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		file := file
		info, err := os.Stat(file)
		if err != nil {
			panic(err)
		}
		var firstBlock int64 = 100 * 1024
		var size = info.Size()
		if size < firstBlock {
			firstBlock = size
		}
		t, err := follower.New(file, follower.Config{
			Whence: io.SeekEnd,
			Offset: -firstBlock,
			Reopen: true,
		})
		if err != nil {
			panic(err)
		}
		// for line := range t.Lines {
		// 	fmt.Println(line.Text)
		// }
		flusher, ok := w.(http.Flusher)
		if !ok {
			panic("expected http.ResponseWriter to be an http.Flusher")
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		var lines = t.Lines()
	LOOP:
		for {
			select {
			case line := <-lines:
				{
					err := t.Err()
					if err != nil {
						panic(err)
					}
					log.Printf("%s | %s | %v", r.RemoteAddr, line.String(), t.Err())
					fmt.Fprintf(w, "event: log\ndata: %s\n\n", line.String())
					flusher.Flush() // Trigger "chunked" encoding and send a chunk...
				}
			case <-r.Context().Done():
				{
					log.Printf("%s | aborted", r.RemoteAddr)
					break LOOP
				}
			}

		}
		t.Close()
	})

	log.Print("Listening on localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
