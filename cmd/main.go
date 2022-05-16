package main

import (
	"flag"
	"fmt"
	"github.com/mrhakimov/broadcasts-single-instance/pkg/brb"
	"github.com/mrhakimov/broadcasts-single-instance/pkg/cebrb"
	"net/http"
	"sync"
)

func createServer(hostAndPort string, port int) *http.Server {
	mux := http.NewServeMux()

	brbI := &brb.Instance{
		SentEcho:  false,
		SentReady: false,
		Delivered: false,
		Echos:     make(map[string]string),
		Readys:    make(map[string]string),
	}

	cebrbI := &cebrb.Instance{
		SentInit:    false,
		SentWitness: false,
		Delivered:   false,
		Witnesses:   make(map[string]string),
	}

	mux.HandleFunc("/brb/clear", func(w http.ResponseWriter, r *http.Request) {
		brbI.Clear(w, r)
	})
	mux.HandleFunc("/brb/deliver/send", func(w http.ResponseWriter, r *http.Request) {
		brbI.Send(w, r, hostAndPort)
	})
	mux.HandleFunc("/brb/deliver/echo", func(w http.ResponseWriter, r *http.Request) {
		brbI.Echo(w, r, hostAndPort)
	})
	mux.HandleFunc("/brb/deliver/ready", func(w http.ResponseWriter, r *http.Request) {
		brbI.Ready(w, r, hostAndPort)
	})

	mux.HandleFunc("/cebrb/clear", func(w http.ResponseWriter, r *http.Request) {
		cebrbI.Clear(w, r)
	})
	mux.HandleFunc("/cebrb/deliver/init", func(w http.ResponseWriter, r *http.Request) {
		cebrbI.Init(w, r, hostAndPort)
	})
	mux.HandleFunc("/cebrb/deliver/witness", func(w http.ResponseWriter, r *http.Request) {
		cebrbI.Witness(w, r, hostAndPort)
	})

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: mux,
	}

	return &server
}

func main() {
	n := flag.Int("n", 4, "the number of processes")
	f := flag.Int("f", 0, "the number of byzantine processes")
	flag.Parse()

	wg := new(sync.WaitGroup)

	for port := 9000; port < 9000+(*n-*f-1); port++ {
		wg.Add(1)
		go func(port int) {
			server := createServer(fmt.Sprintf("http://localhost:%d", port), port)
			fmt.Println(server.ListenAndServe())
			wg.Done()
		}(port)
	}

	wg.Wait()
}
