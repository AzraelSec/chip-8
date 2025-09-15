package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/azraelsec/chip-8/pkg/emulator"
	ebiten "github.com/hajimehoshi/ebiten/v2"
)

func main() {
	rom := flag.String("rom-path", "", "path to the rom you want to emulate")
	cyclesPerUpdate := flag.Int("cycles-per-update", 10, "number of cycles to run per update")
	volume := flag.Float64("volume", 0.7, "audio volume (0.0 to 1.0)")
	flag.Parse()

	if *rom == "" {
		slog.Error("-rom-path is required")
		os.Exit(1)
	}

	if *cyclesPerUpdate < 1 {
		slog.Error("-cycles-per-update must be greater than 0")
		os.Exit(1)
	}

	if *volume < 0.0 || *volume > 1.0 {
		slog.Error("-volume must be between 0.0 and 1.0")
		os.Exit(1)
	}

	romData, err := os.ReadFile(*rom)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	romName := strings.TrimSuffix(filepath.Base(*rom), filepath.Ext(*rom))

	ebiten.SetWindowSize(640, 320)
	ebiten.SetWindowTitle(fmt.Sprintf("Chip8 - %s", romName))

	e := emulator.New(romData, *cyclesPerUpdate, *volume)
	if err := ebiten.RunGame(e); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
