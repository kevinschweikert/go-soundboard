package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

var soundBtn []fyne.CanvasObject
var volumeLabel *widget.Label
var files []soundfile
var mainVol float64

type soundfile struct {
	Path string
	Name string
}

func main() {
	mainVol = 0

	folderPath := flag.String("path", "./sounds", "Path to sound files")
	sampleRate := flag.Int("samplerate", 48000, "Output Samplerate in Hz")
	buffSize := flag.Int("buff", 256, "Output buffer size in bytes")
	flag.Parse()

	speaker.Init(beep.SampleRate(*sampleRate), *buffSize)

	err := filepath.Walk(*folderPath, onFileFound(&files))
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		name := file.Name
		path := file.Path
		newSoundBtn := widget.NewButton(name, func() { buttonCallback(path) })
		soundBtn = append(soundBtn, newSoundBtn)
	}

	a := app.New()
	w := a.NewWindow("Soundboard")

	volumeSlider := widget.NewSlider(-5, 0)
	volumeSlider.Value = 0
	volumeSlider.Step = 0.1
	//volumeSlider.Value = 0
	volumeSlider.OnChanged = volumeChanged

	volumeLabel = widget.NewLabel(fmt.Sprintf("%.1f", mainVol))

	closeBtn := widget.NewButton("Close", func() { w.Close() })

	buttonContainer := fyne.NewContainerWithLayout(
		layout.NewGridLayoutWithColumns(5),
		soundBtn...,
	)

	closeContainer := fyne.NewContainerWithLayout(
		layout.NewCenterLayout(),
		closeBtn,
	)

	appContainer := fyne.NewContainerWithLayout(
		layout.NewGridLayout(1),
		buttonContainer,
		closeContainer,
		volumeLabel,
		volumeSlider,
	)

	w.SetContent(appContainer)

	w.Resize(fyne.NewSize(800, 800))
	w.ShowAndRun()
}

func onFileFound(files *[]soundfile) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			filename := strings.Split(info.Name(), ".")[0]
			*files = append(*files, soundfile{path, filename})
		}
		return nil
	}
}

func buttonCallback(path string) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	resampled := beep.Resample(4, format.SampleRate, 48000, streamer)
	volume := &effects.Volume{
		Streamer: resampled,
		Base:     2,
		Volume:   mainVol,
		Silent:   false,
	}
	speaker.Play(beep.Seq(volume, beep.Callback(func() { streamer.Close() })))
}

func volumeChanged(volume float64) {
	mainVol = volume
	volumeLabel.Text = fmt.Sprintf("%.1f", mainVol)
	volumeLabel.Refresh()
}
