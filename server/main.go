package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/issy20/go-websocket/auth"
	"github.com/issy20/go-websocket/config"
	"github.com/issy20/go-websocket/repository"
)

var addr = flag.String("addr", ":8080", "http server address")
var ctx = context.Background()

func main() {
	flag.Parse()
	config.CreateRedisClient()
	db, err := config.NewDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()
	defer func() {
		err := db.DB.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	userRepository := &repository.UserRepository{Db: db.DB}

	hub := NewHub(&repository.RoomRepository{Db: db.DB}, userRepository)
	go hub.RunLoop()

	api := &API{UserRepository: userRepository}

	http.HandleFunc("/ws", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	}))

	http.HandleFunc("/api/login", api.HandleLogin)
	http.HandleFunc("/api/create", api.HandleAddUser)

	port := "80"
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		log.Panicln("Serve Error:", err)
	}
	log.Fatal(http.ListenAndServe(*addr, nil))
}
