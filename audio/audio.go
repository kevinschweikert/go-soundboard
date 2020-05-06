package audio

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

type SoundFile struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Id        int    `json:"id"`
}

type SoundDirectory struct {
	SoundFiles []SoundFile
	Path       string
}

type AudioPanel struct {
	speakerSampleRate beep.SampleRate
	mixer             *beep.Mixer
	ctrl              *beep.Ctrl
	volume            *effects.Volume
	Err               error
	OnErr             func(string)
}

func NewAudioPanel(speakerSampleRate beep.SampleRate) *AudioPanel {
	Err := errors.New("")
	mixer := &beep.Mixer{}
	ctrl := &beep.Ctrl{Streamer: mixer}
	volume := &effects.Volume{Streamer: mixer, Base: 2}
	return &AudioPanel{speakerSampleRate, mixer, ctrl, volume, Err, func(string) {}}
}

func (ap *AudioPanel) Play(streamer beep.StreamSeekCloser, audioSampleRate beep.SampleRate) {
	ap.Stop()
	resampler := beep.Resample(4, audioSampleRate, ap.speakerSampleRate, streamer)
	ap.mixer.Add(resampler)
}

func (ap *AudioPanel) sendError() {
	ap.OnErr(ap.Err.Error())
}

func (ap *AudioPanel) Stop() {
	ap.mixer.Clear()
}

func (ap *AudioPanel) ChangeVolume(newVolume float64) {
	speaker.Lock()
	ap.volume.Volume = newVolume
	speaker.Unlock()
}

func (ap *AudioPanel) Init(buffSize int) {
	Err := speaker.Init(ap.speakerSampleRate, buffSize)
	if Err != nil {
		log.Fatal("Could not initialize speaker")
	}
	speaker.Play(ap.volume)
}

func GetFilesInFolder(folderPath string) SoundDirectory {
	var files []SoundFile
	id := 0
	Err := filepath.Walk(folderPath, OnFileFound(&files, &id))
	if Err != nil {
		log.Fatal(Err)
	}

	return SoundDirectory{SoundFiles: files, Path: folderPath}
}

func OnFileFound(files *[]SoundFile, counter *int) filepath.WalkFunc {
	return func(path string, info os.FileInfo, Err error) error {
		if Err != nil {
			return Err
		}
		if !info.IsDir() {
			filenameParts := strings.Split(info.Name(), ".")
			filename := filenameParts[0]
			extension := strings.ToLower(filenameParts[len(filenameParts)-1])
			*files = append(*files, SoundFile{Path: path, Name: filename, Extension: extension, Id: *counter})
			*counter++
		}
		return nil
	}
}

func (d *SoundDirectory) Reload() {
	d.SoundFiles = GetFilesInFolder(d.Path).SoundFiles
}

func (d *SoundDirectory) PlaySound(ap *AudioPanel, soundId int) {
	var streamer beep.StreamSeekCloser
	var format beep.Format
	var Err error

	sound := d.SoundFiles[soundId]

	f, Err := os.Open(sound.Path)
	if Err != nil {
		ap.Err = Err
		return
	}

	switch sound.Extension {
	case "mp3":
		streamer, format, Err = mp3.Decode(f)
		if Err != nil {
			fmt.Println(Err)
			ap.Err = Err
			ap.sendError()
			return
		}
	case "wav":
		streamer, format, Err = wav.Decode(f)
		if Err != nil {
			fmt.Println(Err)
			ap.Err = Err
			return
		}
	default:
		fmt.Println("File format not supported")
		ap.Err = errors.New("File format not supported")
		ap.sendError()
		return
	}

	ap.Play(streamer, format.SampleRate)
}
