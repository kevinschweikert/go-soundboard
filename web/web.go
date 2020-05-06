package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kevinschweikert/go-soundboard/audio"
)

// Msg struct to marshal and unmarshal the websocket json data
type Msg struct {
	Type       string            `json:"type"`
	Msg        string            `json:"msg"`
	SoundID    int               `json:"soundID"`
	SoundFiles []audio.SoundFile `json:"soundFiles"`
	Volume     float64           `json:"volume"`
}

type WebSocketData struct {
	Active   []*websocket.Conn
	Upgrader websocket.Upgrader
}

func ErrorFunc(c *websocket.Conn) func(string) {

	return func(err string) {
		data := Msg{
			Type: "error",
			Msg:  err,
		}
		c.WriteJSON(data)
	}
}

func NewWebSocketData() *WebSocketData {
	var c []*websocket.Conn
	u := websocket.Upgrader{}
	return &WebSocketData{c, u}
}

// HandleWebSocket handles the /websocket route
func HandleWebSocket(wd *WebSocketData, ap *audio.AudioPanel, d audio.SoundDirectory) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := wd.Upgrader.Upgrade(w, r, nil)
		defer c.Close()
		if err != nil {
			log.Println("Upgrade:", err)
		}
		wd.Active = append(wd.Active, c)
		ap.OnErr = ErrorFunc(c)
		handleConnection(c, ap, d)

	}
}

// TestHandler handles test
func TestHandler(dir audio.SoundDirectory) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(dir)
		if err != nil {
			fmt.Println(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func handleConnection(c *websocket.Conn, ap *audio.AudioPanel, d audio.SoundDirectory) {
	initMsg := new(Msg)
	initMsg.Type = "load"
	initMsg.SoundFiles = d.SoundFiles
	c.WriteJSON(initMsg)

	for {
		payload := new(Msg)
		err := c.ReadJSON(payload)
		if err != nil {
			//log.Println("read:", err)
			break
		}

		switch payload.Type {
		case "play":
			fmt.Println("Playing: ", payload.Msg)
			d.PlaySound(ap, payload.SoundID)
		case "error":
			fmt.Println("error received", payload.Msg)
		case "load":
			fmt.Println("load")
			d.Reload()
			newPayload := Msg{
				Type:       "load",
				SoundFiles: d.SoundFiles,
			}
			c.WriteJSON(newPayload)
		case "echo":
			err := c.WriteJSON(payload)
			if err != nil {
				fmt.Println("Could not send echo")
			}
		case "stop":
			ap.Stop()
		case "volume":
			ap.ChangeVolume(payload.Volume)

		}

	}
}
