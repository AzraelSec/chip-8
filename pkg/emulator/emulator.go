package emulator

import (
	"image/color"

	"github.com/azraelsec/chip-8/internal/audio"
	"github.com/azraelsec/chip-8/internal/display"
	"github.com/azraelsec/chip-8/pkg/chip8"
	"github.com/hajimehoshi/ebiten/v2"
)

var keyboardMap = map[ebiten.Key]byte{
	ebiten.Key1: 0x0,
	ebiten.Key2: 0x1,
	ebiten.Key3: 0x2,
	ebiten.Key4: 0x3,
	ebiten.KeyQ: 0x4,
	ebiten.KeyW: 0x5,
	ebiten.KeyE: 0x6,
	ebiten.KeyR: 0x7,
	ebiten.KeyA: 0x8,
	ebiten.KeyS: 0x9,
	ebiten.KeyD: 0xA,
	ebiten.KeyF: 0xB,
	ebiten.KeyZ: 0xC,
	ebiten.KeyX: 0xD,
	ebiten.KeyC: 0xE,
	ebiten.KeyV: 0xF,
}

type Emulator struct {
	chip            *chip8.Chip8
	buffer          *ebiten.Image
	cyclesPerUpdate int
	audio           *audio.Audio
}

var _ ebiten.Game = (*Emulator)(nil)

func New(program []byte, cyclesPerUpdate int, volume float64) *Emulator {
	audioSystem, err := audio.New(volume)
	if err != nil {
		audioSystem = nil
	}

	chip := chip8.New(program)
	if audioSystem != nil {
		chip.SetAudio(audioSystem)
	}

	return &Emulator{
		chip:            chip,
		cyclesPerUpdate: cyclesPerUpdate,
		audio:           audioSystem,
	}
}

// by default ebitengine runs this 60/1 second (which is in line with chip-8 spec)
func (e *Emulator) Update() error {
	for i := 0; i < e.cyclesPerUpdate; i++ {
		if err := e.chip.Tick(); err != nil {
			return err
		}

		if e.chip.CheckDraw() {
			e.buffer = ebiten.NewImage(display.DisplayWidth, display.DisplayHeight)
			buf := e.chip.Buffer()
			for j := range display.DisplayHeight {
				for i := range display.DisplayWidth {
					if buf[i+j*display.DisplayWidth] != 0 {
						e.buffer.Set(i, j, color.White)
					} else {
						e.buffer.Set(i, j, color.Black)
					}
				}
			}
		}

		for pk, ck := range keyboardMap {
			if ebiten.IsKeyPressed(pk) {
				e.chip.SetKey(ck, true)
			} else {
				e.chip.SetKey(ck, false)
			}
		}
	}
	return nil
}

func (e *Emulator) Draw(screen *ebiten.Image) {
	if e.buffer != nil {
		screen.DrawImage(e.buffer, nil)
	}
}

func (e *Emulator) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return display.DisplayWidth, display.DisplayHeight
}

func (e *Emulator) Close() error {
	if e.audio != nil {
		return e.audio.Close()
	}
	return nil
}
