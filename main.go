package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/kevinschweikert/go-soundboard/audio"
)

//-------------------------------------------------------------------------------------------------//

func main() {

	folderPath := flag.String("path", "./sounds", "path to sound files")
	speakerSampleRate := flag.Int("samplerate", 48000, "Output Samplerate in Hz")
	buffSize := flag.Int("buff", 256, "Output buffer size in bytes")
	port := flag.Int("port", 8000, "Port to listen for the webinterface")
	flag.Parse()

	dir, err := audio.GetFilesInFolder(*folderPath)
	if err != nil {
		log.Println(err)
	}

	ap := audio.NewPanel(*speakerSampleRate, dir)
	err = ap.Init(*buffSize)
	if err != nil {
		log.Println(err)
	}

	WebSocketConfig := NewWebSocketData()
	fs := http.FileServer(http.Dir("./webinterface/public"))

	http.HandleFunc("/websocket", handleWebSocket(WebSocketConfig, ap))
	http.Handle("/", fs)
	log.Printf("Server listening on 0.0.0.0:%d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))

}
