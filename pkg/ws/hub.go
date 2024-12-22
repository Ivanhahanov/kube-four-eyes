package ws

import (
	"log"

	"github.com/gofiber/contrib/websocket"
)

type client struct{} // Add more data to this type if needed

type Channels struct {
	Clients    map[*websocket.Conn]client
	Register   chan *websocket.Conn
	Broadcast  chan string
	Unregister chan *websocket.Conn
}

var Ch = NewChannels()

func NewChannels() *Channels {
	return &Channels{
		Clients:    make(map[*websocket.Conn]client),
		Register:   make(chan *websocket.Conn),
		Broadcast:  make(chan string),
		Unregister: make(chan *websocket.Conn),
	}
}

func (ch *Channels) RunHub() {
	for {
		select {
		case connection := <-ch.Register:
			ch.Clients[connection] = client{}
			log.Println("connection registered")

		case message := <-ch.Broadcast:
			log.Println("message received:", message)

			// Send the message to all clients
			for connection := range ch.Clients {
				if err := connection.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					log.Println("write error:", err)

					ch.Unregister <- connection
					connection.WriteMessage(websocket.CloseMessage, []byte{})
					connection.Close()
				}
			}

		case connection := <-ch.Unregister:
			// Remove the client from the hub
			delete(ch.Clients, connection)

			log.Println("connection unregistered")
		}
	}
}
