package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	serverSleepTime  = 2 * time.Second
	clientCancelTime = time.Second
)

func echo(w http.ResponseWriter, r *http.Request) {
	time.Sleep(serverSleepTime)
	ctx := r.Context()
	if ctx.Err() != nil {
		fmt.Printf("CTX ERROR before read body: %s\n", ctx.Err())
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
	fmt.Printf("server got body: %s\n", string(body))
	if ctx.Err() != nil {
		fmt.Printf("CTX ERROR after reading body: %s\n", ctx.Err())
		return
	}
	w.Write(body)
	if ctx.Err() != nil {
		fmt.Printf("CTX ERROR after write body: %s\n", ctx.Err())
		return
	}
	fmt.Println("server finished handling, ctx.Err()=", ctx.Err())
}

func sendRequest(ctx context.Context, body io.Reader) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:80/", body)
	if err != nil {
		return err
	}

	go func() {
		time.Sleep(clientCancelTime)
		fmt.Println("call cancel()")
		cancel()
	}()
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	rbody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	fmt.Printf("client got response: %s\n", string(rbody))
	return nil
}

func main() {
	http.HandleFunc("/", echo)
	go http.ListenAndServe(":80", nil)
	time.Sleep(time.Millisecond)

	err := sendRequest(context.Background(), nil)
	if err != nil {
		fmt.Printf("ERROR client: %s\n", err)
	}
	time.Sleep(serverSleepTime)
}
