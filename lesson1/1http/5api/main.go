package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			myParam := r.URL.Query().Get("param")
			fmt.Fprintf(w, "`param` is `%s`\n", myParam)
			for k, v := range r.Header {
				fmt.Fprintln(w, k, ": ", v)
			}
		}
		if r.Method == "POST" {
			body, _ := io.ReadAll(r.Body)
			fmt.Fprintln(w, string(body))
		}
	})

	http.ListenAndServe(":80", nil)
}
