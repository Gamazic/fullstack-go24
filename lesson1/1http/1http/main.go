package main

import (
	"net/http"
)

func path(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(r.URL.Path))
}

func main() {
	http.HandleFunc("/", path)
	http.ListenAndServe("0.0.0.0:80", nil)
}
