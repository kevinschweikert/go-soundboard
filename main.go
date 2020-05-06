package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/faiface/beep"
	"github.com/kevinschweikert/go-soundboard/audio"
	"github.com/kevinschweikert/go-soundboard/web"
)

//-------------------------------------------------------------------------------------------------//

func main() {

	folderPath := flag.String("path", "./sounds", "path to sound files")
	speakerSampleRate := flag.Int("samplerate", 48000, "Output Samplerate in Hz")
	buffSize := flag.Int("buff", 256, "Output buffer size in bytes")
	flag.Parse()

	ap := audio.NewAudioPanel(beep.SampleRate(*speakerSampleRate))
	ap.Init(*buffSize)

	dir := audio.GetFilesInFolder(*folderPath)
	//fmt.Println(dir)

	WebSocketConfig := web.NewWebSocketData()
	fs := http.FileServer(http.Dir("./webinterface/public"))

	http.HandleFunc("/websocket", web.HandleWebSocket(WebSocketConfig, ap, dir))
	http.HandleFunc("/test", web.TestHandler(dir))
	http.Handle("/", fs)
	log.Println("Server listening on 0.0.0.0:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))

}
