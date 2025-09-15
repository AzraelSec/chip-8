package audio

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	SampleRate     = 44100
	BeepFreq       = 440
	BufferDuration = 0.1
	FadeDuration   = 0.01
)

type Audio struct {
	context      *audio.Context
	player       *audio.Player
	infiniteLoop *audio.InfiniteLoop
	isBeeping    bool
}

func generateBeepData() []byte {
	const amplitude = 0.3
	numSamples := int(SampleRate * BufferDuration)
	fadeSamples := int(SampleRate * FadeDuration)

	buf := make([]byte, numSamples*4)

	for i := range numSamples {
		sample := amplitude * math.Sin(2*math.Pi*BeepFreq*float64(i)/SampleRate)

		fadeMultiplier := 1.0
		if i < fadeSamples {
			fadeMultiplier = float64(i) / float64(fadeSamples)
		} else if i >= numSamples-fadeSamples {
			fadeMultiplier = float64(numSamples-i) / float64(fadeSamples)
		}

		sample *= fadeMultiplier
		sampleInt := int16(sample * 32767)

		binary.LittleEndian.PutUint16(buf[i*4:i*4+2], uint16(sampleInt))
		binary.LittleEndian.PutUint16(buf[i*4+2:i*4+4], uint16(sampleInt))
	}

	return buf
}

func New(volume float64) (*Audio, error) {
	ctx := audio.NewContext(SampleRate)
	beepData := generateBeepData()
	reader := bytes.NewReader(beepData)
	infiniteLoop := audio.NewInfiniteLoop(reader, int64(len(beepData)))

	player, err := ctx.NewPlayer(infiniteLoop)
	if err != nil {
		return nil, err
	}

	if volume < 0.0 {
		volume = 0.0
	} else if volume > 1.0 {
		volume = 1.0
	}

	player.SetVolume(volume)

	return &Audio{
		context:      ctx,
		player:       player,
		infiniteLoop: infiniteLoop,
		isBeeping:    false,
	}, nil
}

func (a *Audio) StartBeep() {
	if a.isBeeping {
		return
	}

	a.player.Rewind()
	a.player.Play()
	a.isBeeping = true
}

func (a *Audio) StopBeep() {
	if !a.isBeeping {
		return
	}

	a.player.Pause()
	a.isBeeping = false
}

func (a *Audio) Close() error {
	if a.player != nil {
		return a.player.Close()
	}
	return nil
}
