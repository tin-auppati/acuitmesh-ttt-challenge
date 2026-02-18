package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Tic-Tac-Toe Backend is Running!")
	})
	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}
