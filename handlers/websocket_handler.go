package handlers

// import (
// 	"log"
// 	"net/http"

// 	"github.com/gorilla/websocket"
// 	"github.com/issy20/go-websocket/domain"
// )

// type WebsocketHandler struct {
// 	hub *domain.Hub
// }

// func NewWebsocketHandler(hub *domain.Hub) *WebsocketHandler {
// 	return &WebsocketHandler{hub: hub}
// }

// func (h *WebsocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
// 	upgrader := &websocket.Upgrader{
// 		CheckOrigin: func(r *http.Request) bool {
// 			return true
// 		},
// 	}
// 	ws, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	client := domain.NewClient(ws)
// 	go client.ReadLoop(h.hub.BroadcastCh, h.hub.UnRegisterCh)
// 	go client.WriteLoop()
// 	h.hub.RegisterCh <- client
// }
