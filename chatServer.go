package main

type Broadcast struct {
	sender *Client
	data []byte
}

type WsServer struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Broadcast
}

func NewWebsocketServer() *WsServer {
	return &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast: make(chan *Broadcast),
	}
}

func (server *WsServer) Run() {
	for {
		select {

		case client := <-server.register:
			server.registerClient(client)

		case client := <-server.unregister:
			server.unregisterClient(client)

		case message := <-server.broadcast:
			server.broadcastToClients(message)
		}
	}
}

func (server *WsServer) registerClient(client *Client) {
	server.clients[client] = true
}

func (server *WsServer) unregisterClient(client *Client) {
	if server.clients[client] {
		delete(server.clients, client)
		close(client.send)
	}
}

func (server *WsServer) broadcastToClients(message *Broadcast){
	for client := range server.clients{
		if client == message.sender{
			continue
		}
		select {
		case client.send <- message.data:
			default:
				if server.clients[client] {
					close(client.send)
					delete(server.clients, client)
				}
		}
	}
}
