package handlers

import (
	"fmt"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"log"
	"sort"
)

var upgradeConnection = websocket.FastHTTPUpgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}
type WebSocketConnection struct {
	*websocket.Conn
}

type WsJSONResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

type WsPayload struct {
	Action   string              `json:"action"`
	Username string              `json:"username"`
	Message  string              `json:"message"`
	Conn     WebSocketConnection `json:"-"`
}

// Store which client with client name
var clients = make(map[WebSocketConnection]string)
var wsChan = make(chan WsPayload)



func WsHandler(c *fiber.Ctx) error {
	err := upgradeConnection.Upgrade(c.Context(), wsConnectController)
	if err != nil {
		log.Println("Can not connect to the server")
		return  err
	}
	return nil
}


func wsConnectController(connection *websocket.Conn) {
	log.Println("Connect success!!!")
	var response WsJSONResponse
	response.Message = `<em><small>Connected to server</small></em>`

	conn := WebSocketConnection {connection}
	clients[conn] = ""
	err := conn.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}
	go ListenWs(&conn)
}

func ListenWs(conn *WebSocketConnection) {
	defer func(){
		if r := recover(); r != nil {
			log.Println("error ", fmt.Sprintf("#{r}"))
		}
	}()
	var payload WsPayload
	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// do nothing
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

func ListenToWsChannel() {
	var response WsJSONResponse
	for {
		e := <-wsChan
		switch e.Action {
		case"username":
				users := getUserList()
				response.Action = "list_users"
				response.ConnectedUsers = users
				broadcastToAll(response)

		case "left":
			delete(clients, e.Conn)
			users := getUserList()
			response.Action = "list_users"
			response.ConnectedUsers = users
			broadcastToAll(response)
		case "broadcast":
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s</strong>: %s", e.Username, e.Message)
			broadcastToAll(response)
		}
	}
}
func getUserList() []string {
	var userList []string
	for _, x := range clients {
		if x != "" {
			userList = append(userList, x)
		}
	}
	sort.Strings(userList)

	return userList
}

func broadcastToAll(response WsJSONResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("web socket err")
			_ = client.Close()
			delete(clients, client)
		}
	}
}