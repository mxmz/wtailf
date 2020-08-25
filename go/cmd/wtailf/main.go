package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	"github.com/gobuffalo/packr/v2"
	"github.com/hpcloud/tail"
	"github.com/libp2p/go-reuseport"
	"mxmz.it/wtailf/util"
)

//

var filePattern = regexp.MustCompile(`\.log$|\.stderr$|\.stdout$`)

func getSourceList(sources []string) []string {
	var m = []string{}
	for _, s := range sources {
		var info, err = os.Stat(s)
		if err != nil {
			log.Println(err)
			continue
		}
		if info.Mode().IsRegular() {
			m = append(m, s)
		}
		if info.Mode().IsDir() {
			var dir = s
			var ii, err = ioutil.ReadDir(dir)
			if err != nil {
				log.Println(err)
				continue
			}
			var sublist = []string{}
			for _, i := range ii {
				if i.Mode().IsRegular() && filePattern.MatchString(i.Name()) {
					sublist = append(sublist, path.Join(dir, i.Name()))
				} else if i.Mode().IsDir() {
					sublist = append(sublist, path.Join(dir, i.Name()))
				}
			}
			subsources := getSourceList(sublist)
			for _, f := range subsources {
				m = append(m, f)
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

var acl = util.NewACL()

func aclWrap(hndlr func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var host, _, _ = net.SplitHostPort(r.RemoteAddr)
		var ip = net.ParseIP(host)
		if acl.IsAllowed(ip) {
			hndlr(w, r)
		} else {
			w.WriteHeader(403)
			w.Write([]byte(ip.String() + " not allowed\n"))
		}

	}
}
func fixFs(Handler http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Index(r.URL.Path, ".") == -1 {
			r.URL.Path = "/"
		}
		Handler.ServeHTTP(w, r)
	}
}

func getSourcesHandler(sources []string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var m = getSourceList(sources)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		enc := json.NewEncoder(w)
		enc.Encode(m)
	}
}

func getPeersHandler(peers *sync.Map) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var rv []*Service
		peers.Range(func(k interface{}, v interface{}) bool {
			var s = v.(*Service)
			rv = append(rv, s)
			return true
		})

		w.WriteHeader(200)
		enc := json.NewEncoder(w)
		enc.Encode(rv)
	}
}

func eventHandler(sources []string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		source := r.URL.Query().Get("source")
		file := ""
		var m = getSourceList(sources)
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

		log.Printf("follower: %s  %d\n", file, firstBlock)
		// t, err := follower.New(file, follower.Config{
		// 	Whence: io.SeekEnd,
		// 	Offset: -firstBlock,
		// 	Reopen: true,
		// })
		t, err := tail.TailFile(file, tail.Config{Follow: true, ReOpen: true, Location: &tail.SeekInfo{Offset: -firstBlock, Whence: os.SEEK_END}})
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
		//var lines = t.Lines()
		var lines = t.Lines
		flusher.Flush()
		defer t.Stop()
	LOOP:
		for {
			select {
			case line := <-lines:
				{
					// err := t.Err()
					// if err != nil {
					// 	panic(err)
					// }
					//log.Printf("%s | %s | %v", r.RemoteAddr, line.String(), t.Err())
					//log.Printf("%s | %s ", r.RemoteAddr, line)
					//fmt.Fprintf(w, "event: log\ndata: %s\n\n", line.String())
					fmt.Fprintf(w, "event: log\ndata: %s\n\n", line.Text)
					flusher.Flush() // Trigger "chunked" encoding and send a chunk...
				}
			case <-r.Context().Done():
				{
					log.Printf("%s | aborted", r.RemoteAddr)
					break LOOP
				}
			}

		}
		//t.Close()
	}
}

func main() {

	var bindAddrStr = os.Args[1]
	var sources = os.Args[2:]

	var bindAddr, _ = net.ResolveTCPAddr("tcp", bindAddrStr)
	var myIfaces = util.GetNetInterfaceAddresses()
	var announceCh = make(chan *Service)
	hostname, _ := os.Hostname()

	var defaultACL = []util.ACLEntry{util.LocalhostAllow()}

	if true {
		for _, i := range myIfaces {
			log.Printf("%s\n", i)
			first, last := cidr.AddressRange(i.Net)
			defaultACL = append(defaultACL, util.NewACLEntry(i.Net, true))
			log.Printf("%s %s %s\n", i, first, last)
			var svcURL = fmt.Sprintf("http://%s:%d", i.IP, bindAddr.Port)
			var svcID = fmt.Sprintf("%s-%d", hostname, bindAddr.Port)
			go serviceAnnouncer(svcID, svcURL, last)
		}

		go serviceListener(announceCh)
	}

	envACL := os.Getenv("WTAILF_ACL")
	if len(envACL) == 0 {
		acl = util.NewACL(defaultACL...)
	} else {
		envACLParsed, err := util.ParseACL(envACL)
		if err != nil {
			panic(err)
		}
		acl = util.NewACL(envACLParsed...)
	}

	//var peersLock sync.RWMutex
	//var peers = map[string]*Service{}
	var peers sync.Map

	go func(ch <-chan *Service) {

		for {
			select {
			case message := <-ch:
				{
					peers.Store(message.Endpoint, message)
				}
			}
		}

	}(announceCh)

	var _ = myIfaces

	http.HandleFunc("/", aclWrap(fixFs(http.FileServer(packr.New("dist", "./dist")))))
	http.HandleFunc("/sources", aclWrap(getSourcesHandler(sources)))
	http.HandleFunc("/peers", aclWrap(getPeersHandler(&peers)))
	http.HandleFunc("/events", aclWrap(eventHandler(sources)))
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

	l1, err := reuseport.ListenPacket("udp4", ":18081")
	if err != nil {
		panic(err)
	}
	defer l1.Close()
	for {
		var message Service
		inputBytes := make([]byte, 4096)
		//		log.Printf("Waiting...\n")
		length, _, _ := l1.ReadFrom(inputBytes)
		buffer := bytes.NewBuffer(inputBytes[:length])
		decoder := json.NewDecoder(buffer)
		err := decoder.Decode(&message)
		if err != nil {
			log.Printf("Ignoring malformed message: %s\n", string(inputBytes))
			continue
		}
		//log.Printf("[%v]\n", message)
		message.When = time.Now()
		ch <- &message
	}

}

//func test(c int) {

// 	t, _ := follower.New("../../.temp", follower.Config{
// 		Whence: io.SeekEnd,
// 		Offset: 0,
// 		Reopen: true,
// 	})
// 	var lines = t.Lines()
// 	for {
// 		select {
// 		case line := <-lines:
// 			{
// 				log.Println(c)
// 				log.Println(line)
// 			}
// 		}
// 	}

// }
