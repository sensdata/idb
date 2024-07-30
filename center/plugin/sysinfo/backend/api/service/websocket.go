package service

// import (
// 	"encoding/json"
// 	"log"
// 	"net/http"

// 	"github.com/gorilla/websocket"
// 	"github.com/sensdata/idb/core/plugin"
// )

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// 	CheckOrigin: func(r *http.Request) bool {
// 		return true
// 	},
// }

// func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Println("Failed to set websocket upgrade: ", err)
// 		return
// 	}
// 	defer conn.Close()

// 	// 根据请求路径决定处理逻辑
// 	path := r.URL.Path
// 	if path != config.Plugin.Entry {
// 		log.Println("Unknown WebSocket path: ", path)
// 		return
// 	}

// 	for {
// 		_, message, err := conn.ReadMessage()
// 		if err != nil {
// 			log.Println("Error reading message: ", err)
// 			break
// 		}

// 		var req plugin.PluginRequest
// 		if err := json.Unmarshal(message, &req); err != nil {
// 			log.Println("Error unmarshalling request: ", err)
// 			sendError(conn, "Invalid request format")
// 			continue
// 		}

// 		switch req.Type {
// 		case "plugin":
// 			handlePluginInfo(conn)
// 		case "menu":
// 			handleMenu(conn)
// 		default:
// 			sendError(conn, "Unknown request type")
// 		}
// 	}
// }

// func handleMenu(conn *websocket.Conn) {
// 	menuItems, err := getMenus()
// 	if err != nil {
// 		sendError(conn, err.Error())
// 		return
// 	}
// 	sendResponse(conn, plugin.PluginResponse{Type: "menuItems", Payload: menuItems})
// }

// func handlePluginInfo(conn *websocket.Conn) {
// 	pluginInfo, err := getPluginInfo()
// 	if err != nil {
// 		sendError(conn, err.Error())
// 		return
// 	}
// 	sendResponse(conn, plugin.PluginResponse{Type: "pluginInfo", Payload: pluginInfo})
// }

// func sendResponse(conn *websocket.Conn, res plugin.PluginResponse) {
// 	message, err := json.Marshal(res)
// 	if err != nil {
// 		log.Println("Error marshalling response: ", err)
// 		return
// 	}
// 	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
// 		log.Println("Error writing message: ", err)
// 	}
// }

// func sendError(conn *websocket.Conn, errorMsg string) {
// 	sendResponse(conn, plugin.PluginResponse{Type: "error", Error: errorMsg})
// }
