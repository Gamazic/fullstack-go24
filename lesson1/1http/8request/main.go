package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func path(w http.ResponseWriter, r *http.Request) {
	fmt.Println("server got request: ", r.URL.Path)
	w.Write([]byte(r.URL.Path))
}

func main() {
	http.HandleFunc("/", path)
	go http.ListenAndServe(":80", nil)
	time.Sleep(time.Millisecond)

	r, err := http.Get("http://127.0.0.1:80/hello/world")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("response: ", r.StatusCode, " ", string(body))
}
