package handlers

import (
	"bytes"
	"context"
	"fmt"
	"go-chat/pkg/database"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
)

const (
	MESSAGE_NEW_USER  = "New User"
	MESSAGE_CHAT      = "Chat"
	MESSAGE_LEAVE     = "Leave"
	MESSAGE_EMPTY     = "Empty"
	MESSAGE_SYSTEM    = "System"
	MESSAGE_IMAGE     = "Image"
	MESSAGE_USER_LIST = "UserList"
)

type SocketPayload struct {
	Message    string     `json:"message"`
	TargetUser string     `json:"targetUser"`
	Image      *ImageFile `json:"image,omitempty"`
}

type SocketResponse struct {
	From    string      `json:"from"`
	Type    string      `json:"type"`
	Message interface{} `json:"message"`
}

type ImageFile struct {
	Data        []byte `json:"data"`
	ContentType string `json:"contentType"`
}

var connections = make(map[*websocket.Conn]string)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.String(http.StatusBadRequest, "Could not open websocket connection")
	}

	username := c.QueryParam("username")

	if username == "" {
		return c.String(http.StatusBadRequest, "Username cannot be empty")
	}

	username = generateUniqueUsername(username)

	connections[ws] = username

	broadcastUserList()

	go handleIO(ws)

	return nil
}
func generateUniqueUsername(username string) string {
	// Check if the username is already in use
	for i := 1; isUsernameTaken(username); i++ {
		username = fmt.Sprintf("%s%d", username, i)
	}
	return username
}

func isUsernameTaken(username string) bool {
	for _, existingUsername := range connections {
		if existingUsername == username {
			return true
		}
	}
	return false
}

func broadcastUserList() {
	var onlineUsers []string
	for _, username := range connections {
		onlineUsers = append(onlineUsers, username)
	}

	for conn := range connections {
		err := conn.WriteJSON(SocketResponse{
			From:    "Server",
			Type:    MESSAGE_USER_LIST,
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

		if payload.Message == "" && payload.Image == nil {
			broadcastMessage(ws, MESSAGE_EMPTY, "")
			continue
		}

		if payload.TargetUser != "" {
			if payload.Image != nil {
				uploadImage(ws, payload.TargetUser, payload.Image)
			} else {
				sendPrivateMessage(ws, payload.TargetUser, MESSAGE_CHAT, payload.Message)
			}
		} else {
			if payload.Image != nil {
				uploadImage(ws, "", payload.Image)
			} else {
				broadcastMessage(ws, MESSAGE_CHAT, payload.Message)
			}
		}
	}
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

	if kind == MESSAGE_CHAT {
		log.Printf("Broadcasted message from %s: %s\n", connections[sender], message)
	} else if kind == MESSAGE_IMAGE {
		log.Printf("Broadcasted image link from %s: %s\n", connections[sender], message)
	}
}

func uploadImage(sender *websocket.Conn, targetUser string, image *ImageFile) {
	minioClient, bucketName := database.MinioClient()

	fileName := generateUniqueFileName()

	_, err := minioClient.PutObject(context.Background(), bucketName, fileName, bytes.NewReader(image.Data), int64(len(image.Data)), minio.PutObjectOptions{
		ContentType: image.ContentType,
	})
	if err != nil {
		log.Println("ERROR", err.Error())
		return
	}

	reqParams := make(url.Values)
	imageURL, _ := minioClient.PresignedGetObject(context.Background(), bucketName, fileName, time.Hour*168, reqParams)

	if targetUser != "" {
		sendPrivateMessage(sender, targetUser, MESSAGE_IMAGE, imageURL.String())
	} else {
		broadcastMessage(sender, MESSAGE_IMAGE, imageURL.String())
	}
}

func generateUniqueFileName() string {
	return time.Now().Format("20060102150405")
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

func closeWebSocket(ws *websocket.Conn) {
	delete(connections, ws)
	ws.Close()
}
