package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kevinschweikert/go-soundboard/audio"
)

// Switch constants
const (
	Load   = "load"
	Play   = "play"
	Error  = "error"
	Volume = "volume"
	Stop   = "stop"
)

// Msg struct to marshal and unmarshal the websocket json data
type Msg struct {
	Type       string            `json:"type"`
	Msg        string            `json:"msg"`
	SoundFiles []audio.SoundFile `json:"soundfiles"`
	Volume     float64           `json:"volume"`
}

// WebSocketData hold all necessary connection data
type WebSocketData struct {
	Clients  map[*websocket.Conn]bool
	Upgrader websocket.Upgrader
}

// NewWebSocketData returns a pointer to a new WebSocketData struct
func NewWebSocketData() *WebSocketData {
	c := make(map[*websocket.Conn]bool)
	u := websocket.Upgrader{}
	return &WebSocketData{c, u}
}

func handleWebSocket(wd *WebSocketData, ap *audio.Panel) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		c, err := wd.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		wd.Clients[c] = true

		defer c.Close()
		//defer delete(wd.Clients, c)

		handleConnection(c, wd, ap)

	}
}

func (wd *WebSocketData) broadcast(m Msg) {
	for client := range wd.Clients {
		client.WriteJSON(m)
	}
}

func (wd *WebSocketData) nMinusOne(c *websocket.Conn, m Msg) {
	for client := range wd.Clients {
		if c == client {
			return
		}
		client.WriteJSON(m)
	}
}

func initialData(ap *audio.Panel) (Msg, Msg) {
	load := Msg{
		Type:       Load,
		SoundFiles: ap.SoundDir.SoundFiles,
	}
	volume := Msg{
		Type:   Volume,
		Volume: ap.Volume.Volume,
	}
	return load, volume
}

func handleConnection(c *websocket.Conn, wd *WebSocketData, ap *audio.Panel) {
	m1, m2 := initialData(ap)
	c.WriteJSON(m1)
	c.WriteJSON(m2)

	for {
		payload := new(Msg)
		err := c.ReadJSON(payload)
		if err != nil {
			//log.Println("read:", err)
			break
		}

		switch payload.Type {
		case Play:
			err := ap.PlaySound(payload.SoundFiles[0])
			if err != nil {
				c.WriteJSON(Msg{
					Type: Error,
					Msg:  err.Error(),
				})
			}
		case Error:
			log.Println("error received: ", payload.Msg)
		case Load:
			ap.Reload()
			newPayload := Msg{
				Type:       "load",
				SoundFiles: ap.SoundDir.SoundFiles,
			}
			wd.broadcast(newPayload)
		case Stop:
			ap.Stop()
		case Volume:
			v := ap.ChangeVolume(payload.Volume)
			m := Msg{
				Type:   Volume,
				Volume: v,
			}
			wd.nMinusOne(c, m)
		}

	}
}
