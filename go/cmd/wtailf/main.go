package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
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

var filePattern = regexp.MustCompile(`\.log$|\.stderr|\.stdout$`)

func getSourceMap(sources []string) []string {
	var m = []string{}
	for _, s := range sources {
		var info, err = os.Stat(s)
		if err != nil {
			panic(err)
		}
		if info.Mode().IsRegular() {
			m = append(m, s)
		}
		if info.Mode().IsDir() {
			var dir = s
			var ii, err = ioutil.ReadDir(dir)
			if err != nil {
				panic(err)
			}
			for _, i := range ii {
				if filePattern.MatchString(i.Name()) {
					m = append(m, path.Join(dir, i.Name()))
				}
			}
		}
	}
	return m

}

func main() {

	// go test(1)
	// go test(2)
	// go test(3)

	// <-time.Tick(60 * time.Minute)

	//var file = os.Args[1]
	var sources = os.Args[2:]
	fs := &fsAdapted{http.FileServer(http.Dir("./dist"))}
	http.Handle("/", fs)
	http.HandleFunc("/sources", func(w http.ResponseWriter, r *http.Request) {
		var m = getSourceMap(sources)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		enc := json.NewEncoder(w)
		enc.Encode(m)
	})
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		source := r.URL.Query().Get("source")
		file := ""
		var m = getSourceMap(sources)
		for _, s := range m {
			if s == source {
				file = s
			}
		}

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
