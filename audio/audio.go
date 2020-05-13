package audio

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

// SoundFile holds a sound struct
type SoundFile struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Extension string `json:"extension"`
	ID        int    `json:"id"`
}

// SoundDirectory collects all SoundFiles from a specific path
type SoundDirectory struct {
	SoundFiles []SoundFile `json:"soundfiles"`
	Path       string      `json:"path"`
}

// Panel holds all Player structs like mixer, ctrl and Volume
type Panel struct {
	speakerSampleRate beep.SampleRate
	mixer             *beep.Mixer
	ctrl              *beep.Ctrl
	Volume            *effects.Volume
	SoundDir          SoundDirectory
}

// NewAudioPanel returns a pointer to a Panel struct
func NewPanel(speakerSampleRate int, dir SoundDirectory) *Panel {
	mixer := &beep.Mixer{}
	ctrl := &beep.Ctrl{Streamer: mixer}
	volume := &effects.Volume{Streamer: mixer, Base: 2}
	return &Panel{beep.SampleRate(speakerSampleRate), mixer, ctrl, volume, dir}
}

// play plays the streamer in the argument
func (ap *Panel) play(streamer beep.StreamSeekCloser, audioSampleRate beep.SampleRate) {
	ap.Stop()
	resampler := beep.Resample(4, audioSampleRate, ap.speakerSampleRate, streamer)
	ap.mixer.Add(resampler)
}

//Stop stops all playing streams
func (ap *Panel) Stop() {
	ap.mixer.Clear()
}

// ChangeVolume changes the Volume of the mixer
func (ap *Panel) ChangeVolume(newVolume float64) float64 {
	speaker.Lock()
	ap.Volume.Volume = newVolume
	speaker.Unlock()
	return ap.Volume.Volume
}

//Init initializes the speaker and plays the empty mixer
func (ap *Panel) Init(buffSize int) error {
	err := speaker.Init(ap.speakerSampleRate, buffSize)
	if err != nil {
		return err
	}
	speaker.Play(ap.Volume)
	return nil
}

// GetFilesInFolder returns a new SoundDorectory in a specified path
func GetFilesInFolder(folderPath string) (SoundDirectory, error) {
	var files []SoundFile
	id := 0
	err := filepath.Walk(folderPath, onFileFound(&files, &id))
	if err != nil {
		return SoundDirectory{}, err
	}

	return SoundDirectory{SoundFiles: files, Path: folderPath}, nil
}

func onFileFound(files *[]SoundFile, counter *int) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			filenameParts := strings.Split(info.Name(), ".")
			filename := filenameParts[0]
			extension := strings.ToLower(filenameParts[len(filenameParts)-1])
			*files = append(*files, SoundFile{Path: path, Name: filename, Extension: extension, ID: *counter})
			*counter++
		}
		return nil
	}
}

// Reload gets all the SoundFiles in the directory
func (ap *Panel) Reload() error {
	sd, err := GetFilesInFolder(ap.SoundDir.Path)
	if err != nil {
		return err
	}
	ap.SoundDir.SoundFiles = sd.SoundFiles
	return nil
}

// PlaySound plays a specified SoundFile
func (ap *Panel) PlaySound(s SoundFile) error {
	var streamer beep.StreamSeekCloser
	var format beep.Format

	sound := ap.SoundDir.SoundFiles[s.ID]

	f, err := os.Open(sound.Path)
	if err != nil {
		return errors.New("Can't open file: " + sound.Name)
	}

	switch sound.Extension {
	case "mp3":
		streamer, format, err = mp3.Decode(f)
		if err != nil {
			return errors.New("Can't decode file: " + sound.Name + "." + s.Extension)
		}
	case "wav":
		streamer, format, err = wav.Decode(f)
		if err != nil {
			return errors.New("Can't decode file: " + sound.Name + "." + s.Extension)
		}
	default:
		return errors.New("Unknown file format")
	}

	ap.play(streamer, format.SampleRate)

	return nil
}
