package main

import (
    "fmt"
    "log"
    "net/http"
)

func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s: OK", r.URL.Path[1:])
}

func Serve() {
    http.HandleFunc("/health", health)

    addr := fmt.Sprintf("%s:%s", Server["host"], Server["port"])

    log.Printf("Rhapsode serves on %s", addr)
    log.Fatal(http.ListenAndServe(addr, nil))
}
