package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	sourceHost      = "http://localhost:8080"
	byzantineFactor = 3
)

var (
	hosts []string
	n     int
	f     int
)

type Instance struct {
	SentEcho  bool
	SentReady bool
	Delivered bool
	Echos     map[string]string
	Readys    map[string]string
}

func (i *Instance) clear() {
	i.SentEcho = false
	i.SentReady = false
	i.Delivered = false
	i.Echos = make(map[string]string)
	i.Readys = make(map[string]string)
}

func init() {
	file, err := os.OpenFile("/Users/mukkhakimov/Documents/itmo/thesis/logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetFlags(log.Lmicroseconds | log.Ldate)
	log.SetOutput(file)

	data, err := ioutil.ReadFile("/Users/mukkhakimov/Documents/itmo/thesis/hosts.txt")
	if err != nil {
		log.Fatalln("unable to read hosts: ", err)
	}

	hosts = strings.Split(string(data), "\n")
	log.Println("hosts: ", hosts)

	n = len(hosts)
	f = n / byzantineFactor
}

func makeRequest(rType, message, from, host string) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/deliver/%s", host, rType), nil)
	if err != nil {
		log.Fatalln("unable to reach host: ", err)
	}

	q := req.URL.Query()
	q.Add("message", message)
	q.Add("from", from)
	req.URL.RawQuery = q.Encode()

	log.Println(req.URL.String())

	_, err = http.Get(req.URL.String())
	if err != nil {
		log.Fatalf("unable to reach host '%s'", host)
	}
}

func (i *Instance) doChecks(currentHost string) {
	i.checkReady1(currentHost)
	i.checkReady2(currentHost)
	i.checkDeliver(currentHost)
}

func (i *Instance) Send(_ http.ResponseWriter, r *http.Request, currentHost string) {
	message := r.URL.Query().Get("message")
	from := r.URL.Query().Get("from")

	if from == sourceHost && !i.SentEcho {
		i.SentEcho = true
		for _, host := range hosts {
			makeRequest("echo", message, currentHost, host)
		}
	}

	i.doChecks(currentHost)
}

func (i *Instance) Echo(_ http.ResponseWriter, r *http.Request, currentHost string) {
	message := r.URL.Query().Get("message")
	from := r.URL.Query().Get("from")

	if _, ok := i.Echos[from]; !ok {
		i.Echos[from] = message
	}

	i.doChecks(currentHost)
}

func (i *Instance) Ready(_ http.ResponseWriter, r *http.Request, currentHost string) {
	message := r.URL.Query().Get("message")
	from := r.URL.Query().Get("from")

	if _, ok := i.Readys[from]; !ok {
		i.Readys[from] = message
	}

	i.doChecks(currentHost)
}

func (i *Instance) checkReady1(currentHost string) {
	messagesCnt := make(map[string]int)

	for _, host := range hosts {
		messagesCnt[i.Echos[host]]++
	}

	for message, cnt := range messagesCnt {
		if message != "" && cnt > (n+f)/2 && !i.SentReady {
			i.SentReady = true
			for _, host := range hosts {
				makeRequest("ready", message, currentHost, host)
			}
		}
	}
}

func (i *Instance) checkReady2(currentHost string) {
	messagesCnt := make(map[string]int)

	for _, host := range hosts {
		messagesCnt[i.Readys[host]]++
	}

	for message, cnt := range messagesCnt {
		if message != "" && cnt > f && !i.SentReady {
			i.SentReady = true
			for _, host := range hosts {
				makeRequest("ready", message, currentHost, host)
			}
		}
	}
}

func (i *Instance) checkDeliver(currentHost string) {
	messagesCnt := make(map[string]int)

	for _, host := range hosts {
		messagesCnt[i.Readys[host]]++
	}

	for message, cnt := range messagesCnt {
		if message != "" && cnt > 2*f && !i.Delivered {
			i.Delivered = true
			log.Printf("%s delivered message %s", currentHost, message)

			i.clear()
		}
	}
}
