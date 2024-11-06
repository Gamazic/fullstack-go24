package main

import (
	"fmt"
	"net/http"
)

func main() {
	admin := http.NewServeMux()
	admin.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "users: 1,2,3")
	})
	admin.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "posts: 1,2,3,4,5,6")
	})
	adminHandler := http.StripPrefix("/admin", admin)

	api := http.NewServeMux()
	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "api: ", r.URL.Path)
	})
	apiHandler := http.StripPrefix("/api", api)

	mux := http.NewServeMux()
	mux.Handle("/admin/", adminHandler)
	mux.Handle("/api/", apiHandler)

	server := http.Server{
		Addr:    ":80",
		Handler: mux,
	}
	server.ListenAndServe()
}
