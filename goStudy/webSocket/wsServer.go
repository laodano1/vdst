package main

import (
	"github.com/gorilla/sessions"
	"net/http"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
)


var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	store = sessions.NewCookieStore([]byte("my super duper secret"))
	sid = 0
	sns = make([]*sessions.Session, 0)
)

type (
	WebSocketTransport struct {
		Socket *websocket.Conn
		ChatID string
	}
)

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

// define a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		fmt.Println(string(msg))

		resp := fmt.Sprintf("%s", "Server Response!")
		if err := conn.WriteMessage(messageType, []byte(resp)); err != nil {
			log.Println(err)
			return
		}

	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	sid++
	id := strconv.Itoa(sid)
	sname := fmt.Sprintf("%s", "session-" + id)
	session, err := store.Get(r, sname)
	if err == nil {
		session.Values["user_id"] = "user-" + id
		session.Save(r, w)
	}
	sns = append(sns, session)

	fmt.Printf("session amount: %d\n", len(sns))

	//fmt.Fprintf(w, "Hello World")
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	//ws.CloseHandler()(1005, "disconnect!")
	// handle client close request
	go func() {
		select {
		case <-r.Context().Done() :
			fmt.Printf("%s Disconnect\n", r.Header.Get("user-agent"))
			return
		}
	}()
	// helpful log statement to show connections
	log.Println("Client Connected")

	reader(ws)

}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("Hello World")
	setupRoutes()
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}



