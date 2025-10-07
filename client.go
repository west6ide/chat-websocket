package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 10000
)

var (
	newline = []byte{'\n'}
	space = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:8080" || origin == "http://127.0.0.1:8080"
	},
}

type Client struct {
	conn *websocket.Conn
	wsServer *WsServer
	send chan []byte
}

func newClient(conn *websocket.Conn, wsServer *WsServer) *Client {
	return &Client{
		conn: conn,
		wsServer: wsServer,
		send: make(chan []byte, 256),
	}
}

func ServeWs(wsServer *WsServer,w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := newClient(conn, wsServer)

	wsServer.register <- client
	go client.writePump()
	go client.readPump()

	fmt.Println("New Client joined the hub!")

	client.send <-[]byte(`{"message":"hello from server"}`)
	fmt.Println(client)
}

func (client *Client) disconnect(){
	client.wsServer.unregister <- client
	_ = client.conn.Close()
}

func (client *Client) readPump(){
	defer func() {
		client.disconnect()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, joinMessage, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure){
				log.Printf("unexpected close error: %v", err)
			}
			break
		}

		joinMessage = bytes.TrimSpace(bytes.Replace(joinMessage, newline, space, -1))
		client.wsServer.broadcast <- &Broadcast{sender: client, data: joinMessage}
	}
}

func (client *Client) writePump(){
	ticker := time.NewTicker(pingPeriod)
	defer func(){
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(client.send)
			for i := 0; i < n; i++ {
				if _, err := w.Write(newline); err != nil {
					break
				}
				if _, err := w.Write(<-client.send); err != nil {
					break
				}
			}

			if err := w.Close(); err != nil{
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil{
				return
			}
		}
	}
}