# 🎮 CHIP-8 Emulator

[![Go Version](https://img.shields.io/badge/Go-1.24.5-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)

A high-performance CHIP-8 emulator written in Go, featuring pixel-perfect graphics rendering and authentic sound emulation. Experience classic 8-bit games exactly as they were meant to be played!

## ✨ Features

- 🎯 **Complete CHIP-8 instruction set** - All 35 opcodes implemented
- 🖥️ **Pixel-perfect display** - Authentic 64x32 monochrome graphics
- ⌨️ **Intuitive keyboard mapping** - Modern QWERTY layout support
- 🎵 **Configurable audio** - Built-in beep sound system with volume control
- 🚀 **High performance** - Configurable CPU cycles at 60 FPS
- 🎮 **Game compatibility** - Supports classic CHIP-8 ROMs
- 📦 **Cross-platform** - Runs on Windows, macOS, and Linux

## 🛠️ Installation

### Prerequisites

- Go 1.24.5 or later
- Git

### Quick Start

```bash
# Clone the repository
git clone https://github.com/azraelsec/chip-8.git
cd chip-8

# Build the emulator
make build

# Run a game
./chip-8 -rom-path roms/pong.c8
```

## 🎮 Usage

### Command Line Options

```bash
./chip-8 [OPTIONS]

Options:
  -rom-path string
        path to the rom you want to emulate (required)
  -cycles-per-update int
        number of cycles to run per update (default 10)
  -volume float
        audio volume (0.0 to 1.0) (default 0.7)
```

### Basic Examples

```bash
# Play Pong with default settings
./chip-8 -rom-path roms/pong.c8

# Play Space Invaders
./chip-8 -rom-path roms/invaders.c8

# Play Breakout
./chip-8 -rom-path roms/pong.ch8
```

### Advanced Examples

```bash
# Run with custom volume (quiet)
./chip-8 -rom-path roms/pong.c8 -volume 0.3

# Run with faster CPU cycles
./chip-8 -rom-path roms/pong.c8 -cycles-per-update 15

# Run with custom volume and CPU cycles
./chip-8 -rom-path roms/invaders.c8 -volume 0.5 -cycles-per-update 12

# Run silently (no audio)
./chip-8 -rom-path roms/pong.c8 -volume 0.0
```

### Keyboard Controls

The CHIP-8 keypad is mapped to your keyboard as follows:

```
CHIP-8 Keypad    Your Keyboard
┌─┬─┬─┬─┐        ┌─┬─┬─┬─┐
│1│2│3│C│   →    │1│2│3│4│
├─┼─┼─┼─┤        ├─┼─┼─┼─┤
│4│5│6│D│   →    │Q│W│E│R│
├─┼─┼─┼─┤        ├─┼─┼─┼─┤
│7│8│9│E│   →    │A│S│D│F│
├─┼─┼─┼─┤        ├─┼─┼─┼─┤
│A│0│B│F│   →    │Z│X│C│V│
└─┴─┴─┴─┘        └─┴─┴─┴─┘
```

## 🏗️ Architecture

The emulator follows a clean, modular architecture:

```
pkg/
├── chip8/          # Core CHIP-8 system implementation
└── emulator/       # Ebiten game engine integration

internal/
├── audio/          # Audio system and beep generation
├── cpu/            # CPU registers and stack management
├── display/        # Graphics buffer and rendering
├── keys/           # Input handling
└── ram/            # Memory management and font data
```

## 🎯 Technical Specifications

- **Memory**: 4KB RAM (0x000-0xFFF)
- **Display**: 64x32 pixels, monochrome
- **CPU**: 16 8-bit registers (V0-VF)
- **Stack**: 16 levels for subroutines
- **Timers**: 60Hz delay and sound timers
- **Clock Speed**: ~540 Hz (configurable via -cycles-per-update)
- **Audio**: 440Hz sine wave beep with configurable volume

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

_Built with ❤️ in Go_
