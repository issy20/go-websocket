package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/issy20/go-websocket/domain"
)

func main() {
	hub := domain.NewHub()
	go hub.RunLoop()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		domain.ServeWs(hub, w, r)
	})

	port := "80"
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		log.Panicln("Serve Error:", err)
	}
}
