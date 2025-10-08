package main

type Broadcast struct {
	sender *Client
	data   []byte
}

type WsServer struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Broadcast
	rooms       map[*Room]bool
}

func NewWebsocketServer() *WsServer {
	return &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Broadcast),
		rooms:       make(map[*Room]bool),
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
	server.notifyClientJoined(client)
	server.listOnlineClients(client)
	server.clients[client] = true
}

func (server *WsServer) unregisterClient(client *Client) {
	if _, ok := server.clients[client]; ok {
		delete(server.clients, client)
		server.notifyClientLeft(client)
	}
}

func (server *WsServer) broadcastToClients(message *Broadcast) {
	for client := range server.clients {
		if client == message.sender {
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

func (server *WsServer) findRoomByName(name string) *Room {
	var foundRoom *Room
	for room := range server.rooms {
		if room.GetName() == name {
			foundRoom = room
			break
		}
	}

	return foundRoom
}

func (server *WsServer) createRoom(name string, private bool) *Room {
	room := NewRoom(name, private)
	go room.RunRoom()
	server.rooms[room] = true

	return room
}

func (server *WsServer) notifyClientJoined(client *Client) {
    msg := &Message{
        Action: UserJoinedAction,
        Sender: client,
    }
    server.broadcastToClients(&Broadcast{
        sender: client,
        data:   msg.encode(),
    })
}

func (server *WsServer) notifyClientLeft(client *Client) {
    msg := &Message{
        Action: UserLeftAction,
        Sender: client,
    }
    server.broadcastToClients(&Broadcast{
        sender: client,
        data:   msg.encode(),
    })
}

func (server *WsServer) listOnlineClients(client *Client) {
	for existingClient := range server.clients {
		message := &Message{
			Action: UserJoinedAction,
			Sender: existingClient,
		}
		client.send <- message.encode()
	}
}

func (server *WsServer) findRoomByID(ID string) *Room {
	var foundRoom *Room
	for room := range server.rooms {
		if room.GetID() == ID {
			foundRoom = room
			break
		}
	}

	return foundRoom
}

func (server *WsServer) findClientByID(ID string) *Client {
	var foundClient *Client
	for client := range server.clients {
		if client.ID.String() == ID {
			foundClient = client
			break
		}
	}

	return foundClient
}