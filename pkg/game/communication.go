package game

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Peterculazh/block_fighter/pkg/tmp_storage"
	"github.com/lxzan/gws"
)

const (
	PingInterval         = 5 * time.Second
	HeartbeatWaitTimeout = 10 * time.Second
)

type ClientMessageEvent string

const (
	KEY_PRESS ClientMessageEvent = "KEY_PRESS"
)

type KeyData struct {
	Key string `json:"key"`
}

type BaseMessage struct {
	Event ClientMessageEvent `json:"event"`
}

type KeyPressed struct {
	BaseMessage
	Data KeyData `json:"data"`
}

var communication = CreateWebSocket()

var Upgrader = gws.NewUpgrader(communication, &gws.ServerOption{
	PermessageDeflate: gws.PermessageDeflate{
		Enabled:               true,
		ServerContextTakeover: true,
		ClientContextTakeover: true,
	},
	Authorize: func(r *http.Request, session gws.SessionStorage) bool {
		token := r.URL.Query().Get("token")
		nickname := r.URL.Query().Get("nickname")
		if token == "" || nickname == "" {
			return false
		}
		rdb := tmp_storage.GetRedis()
		storedToken, err := rdb.Get(r.Context(), nickname).Result()
		if err != nil || storedToken != token {
			log.Printf("Authorization failed for %s with token %s\n", nickname, token)
			return false
		}
		rdb.Del(r.Context(), nickname).Result()
		// Authorization successful
		id, err := room.JoinPlayer(nickname)
		if err != nil {
			log.Println("HERE")
			return false
		}
		session.Store("id", id)
		log.Printf("Id of joined player: %s", id)
		return true
	},
})

func CreateWebSocket() *WebSocket {
	return &WebSocket{
		sessions: gws.NewConcurrentMap[string, *gws.Conn](),
	}
}

type WebSocket struct {
	gws.BuiltinEventHandler
	sessions *gws.ConcurrentMap[string, *gws.Conn]
}

func (w *WebSocket) OnOpen(socket *gws.Conn) {
	idAny, _ := socket.Session().Load("id")

	id, ok := idAny.(string)
	if !ok {
		log.Println("Error: id isn't string")
		return
	}
	log.Println("opened connection", id)
	w.sessions.Store(id, socket)
}

func (w *WebSocket) getSession(socket *gws.Conn, key string) any {
	s, _ := socket.Session().Load(key)
	return s
}

func (w *WebSocket) Send(socket *gws.Conn, payload []byte) {
	var channel = w.getSession(socket, "channel").(chan []byte)
	select {
	case channel <- payload:
	default:
		return
	}
}

func (w *WebSocket) OnMessage(socket *gws.Conn, message *gws.Message) {

	log.Println("got message")
	// w.sessions.Range(func(key string, value *gws.Conn) bool {
	// 	message := fmt.Sprintf("Message from server for id - %s", key)
	// 	value.WriteMessage(gws.OpcodeText, []byte(message))
	// 	return true
	// })

	id, _ := socket.Session().Load("id")

	log.Printf("Received message from id: %s", id)

	defer message.Close()
	content := message.Bytes()

	var jsonMap BaseMessage
	err := json.Unmarshal(content, &jsonMap)
	if err != nil {
		log.Println("error", err)
		return
	}

	switch jsonMap.Event {
	case KEY_PRESS:
		var parsedMessage KeyPressed
		if err := json.Unmarshal(content, &parsedMessage); err != nil {
			log.Printf("Error unmarshaling KeyPressed: %v", err)
			return
		}
		game.HandleKeyPress(id.(string), PlayerMovingDirection(parsedMessage.Data.Key))
		log.Printf("Key %s", parsedMessage.Data.Key)
	}

}

func (w *WebSocket) OnClose(socket *gws.Conn, err error) {
	id, ok := socket.Session().Load("id")
	if !ok || id == nil {
		return
	}

	if strID, ok := id.(string); ok {
		room.RemovePlayer(strID)
	}
}

func (w *WebSocket) SendMessage(id string, message []byte) {
	socket, ok := communication.sessions.Load(id)
	if !ok {
		// TODO: Remove player from room if he don't exist in sockets
		return
	}
	socket.WriteMessage(gws.OpcodeText, message)
}
