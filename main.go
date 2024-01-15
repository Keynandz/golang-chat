package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const MESSAGE_NEW_USER = "New User"
const MESSAGE_CHAT = "Chat"
const MESSAGE_LEAVE = "Leave"
const MESSAGE_EMPTY = "Empty"
const MESSAGE_SYSTEM = "System"


type SocketPayload struct {
    Message     string `json:"message"`
    TargetUser  string `json:"targetUser"`
}

type SocketResponse struct {
	From    string `json:"from"`
	Type    string `json:"type"`
	Message interface{} `json:"message"`
}

var connections = make(map[*websocket.Conn]string)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", home)
	e.GET("/ws", handleWebSocket)

	e.Start("0.0.0.0:8080")
}

func home(c echo.Context) error {
	content, err := os.ReadFile("index.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not open requested file")
	}

	return c.HTML(http.StatusOK, string(content))
}

func handleWebSocket(c echo.Context) error {
    ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
    if err != nil {
        return c.String(http.StatusBadRequest, "Could not open websocket connection")
    }

    username := c.QueryParam("username")
    connections[ws] = username

    broadcastUserList()

    go handleIO(ws)

    return nil
}

func broadcastUserList() {
    var onlineUsers []string
    for _, username := range connections {
        onlineUsers = append(onlineUsers, username)
    }

    for conn := range connections {
        err := conn.WriteJSON(SocketResponse{
            From:    "Server",
            Type:    "UserList",
            Message: onlineUsers,
        })
        if err != nil {
            log.Println("ERROR", err.Error())
            closeWebSocket(conn)
        }
    }
}

func handleIO(ws *websocket.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR", fmt.Sprintf("%v", r))
		}
	}()

	broadcastMessage(ws, MESSAGE_NEW_USER, "")

	for {
		var payload SocketPayload
		err := ws.ReadJSON(&payload)
		if err != nil {
			if strings.Contains(err.Error(), "websocket: close") {
				broadcastMessage(ws, MESSAGE_LEAVE, "")
				closeWebSocket(ws)
				return
			}

			log.Println("ERROR", err.Error())
			continue
		}

		if payload.Message == "" {
			broadcastMessage(ws, MESSAGE_EMPTY, "")
			continue
		}

		if payload.TargetUser != "" {
            sendPrivateMessage(ws, payload.TargetUser, MESSAGE_CHAT, payload.Message)
        } else {
            broadcastMessage(ws, MESSAGE_CHAT, payload.Message)
        }
	}
}


func closeWebSocket(ws *websocket.Conn) {
	delete(connections, ws)
	ws.Close()
}

func broadcastMessage(sender *websocket.Conn, kind, message string) {
	for conn := range connections {
		if conn == sender {
			continue
		}

		err := conn.WriteJSON(SocketResponse{
			From:    connections[sender],
			Type:    kind,
			Message: message,
		})
		if err != nil {
			log.Println("ERROR", err.Error())
			closeWebSocket(conn)
		}
	}

	if kind != MESSAGE_EMPTY {
		log.Printf("Broadcasted message from %s: %s\n", connections[sender], message)
	}
}

func sendPrivateMessage(sender *websocket.Conn, targetUser, kind, message string) {
    for conn, username := range connections {
        if username == targetUser && conn != sender {
            err := conn.WriteJSON(SocketResponse{
                From:    connections[sender],
                Type:    kind,
                Message: message,
            })
            if err != nil {
                log.Println("ERROR", err.Error())
                closeWebSocket(conn)
            }
            return
        }
    }
}