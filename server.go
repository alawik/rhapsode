package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/gorilla/mux"
)

func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s: OK", r.URL.Path[1:])
}

func fxIdentifier(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Function ID: %v\n", vars["fxid"])
}

func Serve() {
    r := mux.NewRouter()
    r.HandleFunc("/health", health)
    r.HandleFunc("/f/{fxid}", fxIdentifier)
    http.Handle("/", r)

    addr := fmt.Sprintf("%s:%s", Server["host"], Server["port"])

    log.Printf("Rhapsode serves on %s", addr)
    log.Fatal(http.ListenAndServe(addr, nil))
}
