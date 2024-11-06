package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/page",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Page path:", r.URL.String())
		})

	http.HandleFunc("/collection/test",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Collection subpath:", r.URL.String())
		})

	http.HandleFunc("/collection/",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Collection path:", r.URL.String())
		})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Main: ", r.URL.String())
	})

	http.ListenAndServe(":80", nil)
}
