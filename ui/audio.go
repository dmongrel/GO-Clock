package ui

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

var soundEffects = make(map[string]*beep.Buffer)

func LoadSound(name string, data []byte) error {
	clear(soundEffects)
	streamer, format, err := mp3.Decode(io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return fmt.Errorf("mp3 decoding error: %w", err)
	}
	defer streamer.Close()

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)

	soundEffects[name] = buffer
	return nil
}

func PlaySound(name string, loop bool) {
	buffer, exists := soundEffects[name]
	if !exists {
		return
	}

	speaker.Clear()
	freshStreamer := buffer.Streamer(0, buffer.Len())
	var s beep.Streamer = freshStreamer
	if loop {
		s = beep.Loop(-1, freshStreamer)
	}

	// Apply volume control (0 means 100% volume)
	volume := &effects.Volume{
		Streamer: s,
		Base:     2,
		Volume:   0,
	}

	speaker.Play(volume)
}

const targetRate = beep.SampleRate(44100)

var speakerInitialized bool

func InitAudio() {
	if !speakerInitialized {
		// Initialize speaker with a default sample rate
		speaker.Init(targetRate, targetRate.N(time.Second/10))
		speakerInitialized = true
	}
}

func StopSound() {
	speaker.Clear()
}
