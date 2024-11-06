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

	serverAdmin := http.Server{
		Addr:    ":8081",
		Handler: adminHandler,
	}
	go serverAdmin.ListenAndServe()

	api := http.NewServeMux()
	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "api: ", r.URL.Path)
	})
	serverApi := http.Server{
		Addr:    ":8080",
		Handler: api,
	}
	serverApi.ListenAndServe()
}
