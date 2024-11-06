package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	serverSleepTime  = 2 * time.Second
	clientCancelTime = 1 * time.Second
)

func echo(w http.ResponseWriter, r *http.Request) {
	time.Sleep(serverSleepTime)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
	fmt.Printf("server got body: %s\n", string(body))
	w.Write(body)
}

func main() {
	http.HandleFunc("/", echo)
	go http.ListenAndServe(":80", nil)
	time.Sleep(time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	request, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:80/",
		strings.NewReader("Hello World!"))
	if err != nil {
		panic(err)
	}

	go func() {
		time.Sleep(clientCancelTime)
		fmt.Println("call cancel()")
		cancel()
	}()

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("client got response: %s\n", string(body))
}
