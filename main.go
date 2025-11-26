package main

import (
    "fmt"
    "net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
    // Write to the ResponseWriter (w) to send data back to the user
    fmt.Fprintf(w, "System is online. Status: 200 OK")
}

func main() {
    port := ":8080"
    http.HandleFunc("/", helloHandler)
    fmt.Printf("Starting server on port %s...\n", port)
    if err := http.ListenAndServe(port, nil); err != nil {
        panic(err)
    }
}
