package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func echo(w http.ResponseWriter, r *http.Request) {
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

	client := &http.Client{
		Transport: &http.Transport{
			Proxy:                  nil,
			OnProxyConnectResponse: nil,
			DialContext:            nil,
			Dial:                   nil,
			DialTLSContext:         nil,
			DialTLS:                nil,
			TLSClientConfig:        nil,
			TLSHandshakeTimeout:    0,
			DisableKeepAlives:      false,
			DisableCompression:     false,
			MaxIdleConns:           0,
			MaxIdleConnsPerHost:    0,
			MaxConnsPerHost:        0,
			IdleConnTimeout:        0,
			ResponseHeaderTimeout:  0,
			ExpectContinueTimeout:  0,
			TLSNextProto:           nil,
			ProxyConnectHeader:     nil,
			GetProxyConnectHeader:  nil,
			MaxResponseHeaderBytes: 0,
			WriteBufferSize:        0,
			ReadBufferSize:         0,
			ForceAttemptHTTP2:      false,
		},
		Timeout: 0,
	}

	request, err := http.NewRequestWithContext(context.Background(), "GET", "http://127.0.0.1:80/",
		strings.NewReader("Hello World!"))
	if err != nil {
		log.Fatal(err)
	}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("client got response: %s\n", string(body))
}
