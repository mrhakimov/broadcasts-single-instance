package main

import (
	"fmt"
	"github.com/mrhakimov/broadcasts-single-instance/pkg/handlers"
	"net/http"
	"sync"
)

const (
	n = 3
)

func createServer(hostAndPort string, port int) *http.Server {
	mux := http.NewServeMux()

	i := &handlers.Instance{
		SentEcho:  false,
		SentReady: false,
		Delivered: false,
		Echos:     make(map[string]string),
		Readys:    make(map[string]string),
	}

	mux.HandleFunc("/deliver/send", func(w http.ResponseWriter, r *http.Request) {
		i.Send(w, r, hostAndPort)
	})
	mux.HandleFunc("/deliver/echo", func(w http.ResponseWriter, r *http.Request) {
		i.Echo(w, r, hostAndPort)
	})
	mux.HandleFunc("/deliver/ready", func(w http.ResponseWriter, r *http.Request) {
		i.Ready(w, r, hostAndPort)
	})

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", port), // :{port}
		Handler: mux,
	}

	return &server
}

func main() {
	wg := new(sync.WaitGroup)

	for port := 9000; port < 9000+n; port++ {
		wg.Add(1)
		go func(port int) {
			server := createServer(fmt.Sprintf("http://localhost:%d", port), port)
			fmt.Println(server.ListenAndServe())
			wg.Done()
		}(port)
	}

	wg.Wait()
}
