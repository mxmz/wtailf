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
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/gobuffalo/packr/v2"
	"github.com/hpcloud/tail"
	"github.com/libp2p/go-reuseport"
	"github.com/pkg/errors"
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
	OS       string    `json:"os,omitempty"`
}

func jwtAuthorizer(a *util.PubKeyJwtAuthorizer, validate func(d *util.JwtData) bool) util.AuthFunc {
	return func(c util.AuthContext, r *http.Request) (util.AuthContext, error) {
		if !strings.Contains(r.URL.Path, ".") {
			var data, err = a.Authorize(r)
			if err != nil || !validate(data) {
				return nil, errors.New("Invalid JWT or identity not authorized")
			} else {
				c["sub"] = data.Sub
				c["iss"] = data.Iss
			}
		}
		return c, nil
	}
}

func aclAuthorizer(acl *util.ACL) util.AuthFunc {
	return func(c util.AuthContext, r *http.Request) (util.AuthContext, error) {
		var host, _, _ = net.SplitHostPort(r.RemoteAddr)
		var ip = net.ParseIP(host)
		if acl.IsAllowed(ip) {
			c["ip"] = ip.String()
			return c, nil
		} else {
			return nil, errors.New(ip.String() + " not allowed\n")
		}
	}
}

func fixFs(hndlr http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Index(r.URL.Path, ".") == -1 {
			r.URL.Path = "/"
		}
		hndlr.ServeHTTP(w, r)
	}
}

func getSourcesHandler(sources []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m = getSourceList(sources)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		enc := json.NewEncoder(w)
		enc.Encode(m)
	}
}

func getNeighboursHandler(neighbours *sync.Map) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var rv []*Service
		var ac = r.Context().Value(util.AuthContextKey)
		var _ = ac
		neighbours.Range(func(k interface{}, v interface{}) bool {
			var s = v.(*Service)
			rv = append(rv, s)
			return true
		})

		w.WriteHeader(200)
		enc := json.NewEncoder(w)
		enc.Encode(rv)
	}
}

func eventHandler(sources []string) http.HandlerFunc {
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
		t, err := tail.TailFile(file, tail.Config{Follow: true, ReOpen: true, Location: &tail.SeekInfo{Offset: -firstBlock, Whence: os.SEEK_END}, Poll: true})
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
	var authorizer = util.NullAuthorizer

	var defaultACL = []util.ACLEntry{util.LocalhostAllow()}
	var envACL = os.Getenv("WTAILF_ACL")
	if len(envACL) == 0 {
		authorizer = util.ComposeAuthorizers(authorizer,
			aclAuthorizer(util.NewACL(defaultACL...)),
		)
	} else {
		envACLParsed, err := util.ParseACL(envACL)
		if err != nil {
			panic(err)
		}
		authorizer = util.ComposeAuthorizers(authorizer,
			aclAuthorizer(util.NewACL(envACLParsed...)),
		)
	}

	var jwtCertPath = os.Getenv("WTAILF_JWT_CERT_PATH")
	if len(jwtCertPath) > 0 {
		var a, err = util.NewPubKeyJwtAuthorizer(jwtCertPath)
		if err != nil {
			panic(err)
		}
		var jwtSubMatch = regexp.MustCompile(os.Getenv("WTAILF_JWT_SUB_MATCH"))
		var jwtIssMatch = regexp.MustCompile(os.Getenv("WTAILF_JWT_ISS_MATCH"))

		authorizer = util.ComposeAuthorizers(authorizer,
			jwtAuthorizer(a, func(jwt *util.JwtData) bool {
				return jwtSubMatch.Match([]byte(jwt.Sub)) && jwtIssMatch.Match([]byte(jwt.Iss))
			}),
		)
	}

	var bindAddr, _ = net.ResolveTCPAddr("tcp", bindAddrStr)

	var announceCh = make(chan *Service)

	var myIfaces = util.GetNetInterfaceAddresses()
	hostname, _ := os.Hostname()

	var envSvcURL = os.Getenv("WTAILF_URL")

	if true {
		for _, i := range myIfaces {
			log.Printf("%s\n", i)
			first, last := cidr.AddressRange(i.Net)
			log.Printf("%s %s %s\n", i, first, last)
			var svcURL = fmt.Sprintf("http://%s:%d", i.IP, bindAddr.Port)
			if len(envSvcURL) > 0 {
				svcURL = envSvcURL
			}
			var svcID = strings.Replace(fmt.Sprintf("%s_%d", hostname, bindAddr.Port), ".", "_", -1)
			go serviceAnnouncer(svcID, svcURL, bindAddr.IP)
		}

		go serviceListener(announceCh)
	}

	var neighbours sync.Map

	go func(ch <-chan *Service) {

		for {
			select {
			case message := <-ch:
				{
					neighbours.Store(message.Endpoint, message)
				}
			}
		}

	}(announceCh)

	var authWrap = util.AuthorizedHandlerBuilder(authorizer)

	http.HandleFunc("/", authWrap(fixFs(http.FileServer(packr.New("dist", "./dist")))))
	http.HandleFunc("/api/sources", authWrap(getSourcesHandler(sources)))
	http.HandleFunc("/api/neighbours", authWrap(getNeighboursHandler(&neighbours)))
	http.HandleFunc("/api/events", authWrap(eventHandler(sources)))
	log.Print("Listening on " + bindAddrStr)
	log.Fatal(http.ListenAndServe(bindAddrStr, nil))
}

func serviceAnnouncer(serviceID string, serviceURL string, localip net.IP) {
	laddr := &net.UDPAddr{localip, 0, ""}
	hostname, _ := os.Hostname()
	osname := runtime.GOOS
	svc := Service{Endpoint: serviceURL, ID: serviceID, Hostname: hostname, When: time.Now(), OS: osname}
	for {
		delay := 10 * time.Second
		err := broadcastServiceAnnounce(laddr, &svc)
		if err != nil {
			log.Println(err)
			delay = 60 * time.Second
		}
		time.Sleep(delay)
	}
}

func broadcastServiceAnnounce(laddr *net.UDPAddr, s *Service) error {
	addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:18081")
	if err != nil {
		return err
	}

	c, err := net.DialUDP("udp", laddr, addr)
	if err != nil {
		log.Println(err)
		return err
	}
	defer c.Close()

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.Encode(s)
	_, err = c.Write(buffer.Bytes())
	if err != nil {
		log.Println(err)
	}
	return err
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
		length, a, err := l1.ReadFrom(inputBytes)
		if err != nil {
			log.Println(err)
			continue
		}

		var _ = a

		buffer := bytes.NewBuffer(inputBytes[:length])
		decoder := json.NewDecoder(buffer)
		err = decoder.Decode(&message)
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
