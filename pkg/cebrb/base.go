package cebrb

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	sourceHost = "http://localhost:8080"
	//byzantineFactor = 5
)

var (
	hosts []string
	n     int
	f     int
)

type Instance struct {
	SentInit    bool
	SentWitness bool
	Delivered   bool
	Witnesses   map[string]string
}

func (i *Instance) Clear(_ http.ResponseWriter, _ *http.Request) {
	i.SentInit = false
	i.SentWitness = false
	i.Delivered = false
	i.Witnesses = make(map[string]string)
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

	allData := strings.Split(string(data), "\n")
	hosts = allData[1:]
	f1, _ := strconv.ParseInt(allData[0], 10, 32)
	f = int(f1)
	//log.Println("hosts: ", hosts)

	n = len(hosts)
	//f = n / byzantineFactor
}

func makeRequest(rType, message, from, host string) {
	//log.Println(fmt.Sprintf("%s -> %s - %s %s", from, host, rType, message))
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/cebrb/deliver/%s", host, rType), nil)
	if err != nil {
		//log.Fatalln("unable to reach host: ", err)
	}

	q := req.URL.Query()
	q.Add("message", message)
	q.Add("from", from)
	req.URL.RawQuery = q.Encode()

	//log.Println(req.URL.String())

	_, err = http.Get(req.URL.String())
	if err != nil {
		//log.Fatalf("unable to reach host '%s'", host)
	}
}

func (i *Instance) Init(_ http.ResponseWriter, r *http.Request, currentHost string) {
	message := r.URL.Query().Get("message")
	from := r.URL.Query().Get("from")

	if from == sourceHost && !i.SentInit {
		i.SentInit = true
		for _, host := range hosts {
			makeRequest("witness", message, currentHost, host)
		}
	}
}

func (i *Instance) Witness(_ http.ResponseWriter, r *http.Request, currentHost string) {
	initMessage := r.URL.Query().Get("message")
	from := r.URL.Query().Get("from")

	if _, ok := i.Witnesses[from]; !ok {
		i.Witnesses[from] = initMessage
	}

	messagesCnt := make(map[string]int)

	for _, host := range hosts {
		messagesCnt[i.Witnesses[host]]++
	}

	for message, cnt := range messagesCnt {
		if message != "" && cnt >= n-2*f && !i.SentWitness {
			i.SentWitness = true

			for _, host := range hosts {
				makeRequest("witness", message, currentHost, host)
			}
		}

		if message != "" && cnt >= n-f && !i.Delivered {
			i.Delivered = true
			log.Printf("%s delivered message %s", currentHost, "0") // replace 0 with message
		}
	}
}
