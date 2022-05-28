package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Count struct {
	Count int `json:"count"`
}

var count = Count{Count: 0}

func main() {
	addr, found := syscall.Getenv("ADDR")

	if !found {
		addr = ":8080"
	}

	file, err := ioutil.ReadFile("data/count.json")

	if err == nil {
		json.Unmarshal(file, &count)
	}

	hub := newHub()

	go hub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	fs := http.FileServer(http.Dir("static"))

	http.Handle("/", fs)

	metrics := promhttp.Handler()

	http.HandleFunc("/metrics", metrics.ServeHTTP)

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c

	res, err := json.Marshal(count)

	if err != nil {
		panic(err)
	}

	os.Mkdir("data", 0777)

	ioutil.WriteFile("./data/count.json", res, 0644)
}
