package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/gobuffalo/packr"
	"github.com/papertrail/go-tail/follower"
	"mxmz.it/wtailf/util"
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

type Service struct {
	ID       string    `json:"id,omitempty"`
	Endpoint string    `json:"endpoint,omitempty"`
	Hostname string    `json:"hostname,omitempty"`
	When     time.Time `json:"when,omitempty"`
}

func main() {

	// go test(1)
	// go test(2)
	// go test(3)

	// <-time.Tick(60 * time.Minute)

	var bindAddrStr = os.Args[1]
	var sources = os.Args[2:]

	var bindAddr, _ = net.ResolveTCPAddr("tcp", bindAddrStr)
	var myIfaces = util.GetNetInterfaceAddresses()
	var announceCh = make(chan *Service)
	hostname, _ := os.Hostname()

	for _, i := range myIfaces {
		log.Printf("%s\n", i)
		first, last := cidr.AddressRange(i.Net)
		log.Printf("%s %s %s\n", i, first, last)
		var svcURL = fmt.Sprintf("http://%s:%d", i.IP, bindAddr.Port)
		var svcID = fmt.Sprintf("%s-%d", hostname, bindAddr.Port)
		go serviceAnnouncer(svcID, svcURL, last)
	}

	go serviceListener(announceCh)

	var peersLock sync.RWMutex
	var peers = map[string]*Service{}

	go func(ch <-chan *Service) {

		for {
			select {
			case message := <-ch:
				{
					log.Printf("%v\n", message)
					peersLock.Lock()
					peers[message.Endpoint] = message
					peersLock.Unlock()
				}
			}
		}

	}(announceCh)

	var _ = myIfaces

	fs := &fsAdapted{http.FileServer(packr.NewBox("./dist"))}
	http.Handle("/", fs)
	http.HandleFunc("/sources", func(w http.ResponseWriter, r *http.Request) {
		var m = getSourceMap(sources)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		enc := json.NewEncoder(w)
		enc.Encode(m)
	})
	http.HandleFunc("/peers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var rv []*Service
		peersLock.RLock()
		defer peersLock.RUnlock()
		for _, s := range peers {
			rv = append(rv, s)
		}
		w.WriteHeader(200)
		enc := json.NewEncoder(w)
		enc.Encode(rv)
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

		log.Printf("follower: %s  %d [%d]\n", file, firstBlock, size)
		t, err := follower.New(file, follower.Config{
			Whence: io.SeekEnd,
			Offset: -firstBlock,
			Reopen: true,
		})
		if err != nil {
			panic(err)
		}
		// for line := range t.Lines {
		// 	log.Println(line.Text)
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
		flusher.Flush()
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

	log.Print("Listening on " + bindAddrStr)
	log.Fatal(http.ListenAndServe(bindAddrStr, nil))
}

func serviceAnnouncer(serviceID string, serviceURL string, broadcast net.IP) {
	// listenAddr, err := net.ResolveUDPAddr("udp4", ":18081")
	// if err != nil {
	// 	panic(err)
	// }
	//list, err := net.ListenUDP("udp4", listenAddr)
	addr := &net.UDPAddr{broadcast, 18081, ""}
	connection, err := net.DialUDP("udp", nil, addr)

	if err != nil {
		log.Println(err)
		return
	}
	defer connection.Close()
	hostname, _ := os.Hostname()
	svc := Service{Endpoint: serviceURL, ID: serviceID, Hostname: hostname, When: time.Now()}
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.Encode(svc)
	for {
		delay := 10 * time.Second
		_, err := connection.Write(buffer.Bytes())
		if err != nil {
			log.Println(err)
			delay = 60 * time.Second
		}
		time.Sleep(delay)
	}
}

func serviceListener(ch chan<- *Service) {
	listenAddr, err := net.ResolveUDPAddr("udp4", ":18081")
	if err != nil {
		panic(err)
	}
	list, err := net.ListenUDP("udp4", listenAddr)
	if err != nil {
		panic(err)
	}
	defer list.Close()

	for {
		var message Service
		inputBytes := make([]byte, 4096)
		log.Printf("Waiting...\n")
		length, _, _ := list.ReadFromUDP(inputBytes)
		buffer := bytes.NewBuffer(inputBytes[:length])
		decoder := json.NewDecoder(buffer)
		err := decoder.Decode(&message)
		if err != nil {
			log.Printf("Ignoring malformed message: %s\n", string(inputBytes))
			continue
		}
		log.Printf("[%v]\n", message)
		ch <- &message
	}
}
