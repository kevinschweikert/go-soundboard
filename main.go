package main

import (
	"errors"
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
	"github.com/faiface/beep/wav"
)

type soundFile struct {
	path      string
	name      string
	extension string
}

type soundDirectory struct {
	soundFiles []soundFile
	path       string
}

type audioPanel struct {
	speakerSampleRate beep.SampleRate
	mixer             *beep.Mixer
	ctrl              *beep.Ctrl
	volume            *effects.Volume
	err               error
}

func newAudioPanel(speakerSampleRate beep.SampleRate) *audioPanel {
	err := errors.New("")
	mixer := &beep.Mixer{}
	ctrl := &beep.Ctrl{Streamer: mixer}
	volume := &effects.Volume{Streamer: mixer, Base: 2}
	return &audioPanel{speakerSampleRate, mixer, ctrl, volume, err}
}

func (ap *audioPanel) play(streamer beep.StreamSeekCloser, audioSampleRate beep.SampleRate) {
	ap.stop()
	resampler := beep.Resample(4, audioSampleRate, ap.speakerSampleRate, streamer)
	ap.mixer.Add(resampler)

}

func (ap *audioPanel) stop() {
	ap.mixer.Clear()
}

func (ap *audioPanel) changeVolume(newVolume float64) {
	speaker.Lock()
	ap.volume.Volume = newVolume
	speaker.Unlock()
}

func (ap *audioPanel) init(buffSize int) {
	err := speaker.Init(ap.speakerSampleRate, buffSize)
	if err != nil {
		log.Fatal("Could not initialize speaker")
	}
	speaker.Play(ap.volume)
}

func getFilesInFolder(folderPath string) soundDirectory {
	var files []soundFile
	err := filepath.Walk(folderPath, onFileFound(&files))
	if err != nil {
		log.Fatal(err)
	}

	return soundDirectory{files, folderPath}
}

func onFileFound(files *[]soundFile) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			filenameParts := strings.Split(info.Name(), ".")
			filename := filenameParts[0]
			extension := strings.ToLower(filenameParts[len(filenameParts)-1])
			*files = append(*files, soundFile{path, filename, extension})
		}
		return nil
	}
}

func (s *soundDirectory) getSoundButtons(ap *audioPanel) []fyne.CanvasObject {
	var buttons []fyne.CanvasObject
	for _, file := range s.soundFiles {
		name := file.name
		soundfile := file
		button := widget.NewButton(name, func() { playSound(ap, soundfile) })
		buttons = append(buttons, button)
	}

	return buttons
}

func main() {

	folderPath := flag.String("path", "./sounds", "path to sound files")
	speakerSampleRate := flag.Int("samplerate", 48000, "Output Samplerate in Hz")
	buffSize := flag.Int("buff", 256, "Output buffer size in bytes")
	flag.Parse()

	ap := newAudioPanel(beep.SampleRate(*speakerSampleRate))
	ap.init(*buffSize)

	dir := getFilesInFolder(*folderPath)
	buttons := dir.getSoundButtons(ap)

	a := app.New()
	w := a.NewWindow("Soundboard")

	volumeLabel := widget.NewLabel(fmt.Sprintf("%.1f", ap.volume.Volume))
	volumeLabel.TextStyle.Bold = true

	volumeSlider := widget.NewSlider(-5, 0)
	volumeSlider.Value = 0
	volumeSlider.Step = 0.05
	//volumeSlider.Value = 0
	volumeSlider.OnChanged = onVolumeChanged(ap, volumeLabel)

	errorLabel := widget.NewLabel(ap.err.Error())

	stopButton := widget.NewButton("Stop All", func() { ap.stop() })
	stopButton.Style = widget.PrimaryButton

	buttonContainer := fyne.NewContainerWithLayout(
		layout.NewGridLayoutWithColumns(5),
		buttons...,
	)

	volumeContainer := widget.NewVBox(
		volumeLabel,
		volumeSlider,
		stopButton,
		errorLabel,
	)

	appContainer := fyne.NewContainerWithLayout(
		layout.NewGridLayoutWithColumns(1),
		buttonContainer,
		volumeContainer,
	)

	w.SetContent(appContainer)

	w.Resize(fyne.NewSize(800, 800))
	w.ShowAndRun()
}

func playSound(ap *audioPanel, file soundFile) {
	var streamer beep.StreamSeekCloser
	var format beep.Format
	var err error

	f, err := os.Open(file.path)
	if err != nil {
		log.Fatal(err)
		return
	}

	switch file.extension {
	case "mp3":
		streamer, format, err = mp3.Decode(f)
		if err != nil {
			fmt.Println(err)
			ap.err = err
			return
		}
	case "wav":
		streamer, format, err = wav.Decode(f)
		if err != nil {
			fmt.Println(err)
			ap.err = err
			return
		}
	default:
		fmt.Println("File format not supported")
		ap.err = errors.New("File format not supported")
		return
	}

	ap.play(streamer, format.SampleRate)
}

func onVolumeChanged(ap *audioPanel, volumeLabel *widget.Label) func(float64) {
	return func(volume float64) {
		ap.changeVolume(volume)
		volumeLabel.Text = fmt.Sprintf("%.1f", ap.volume.Volume)
		volumeLabel.Refresh()
	}
}
