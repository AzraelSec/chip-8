package chip8

import (
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/azraelsec/chip-8/internal/cpu"
	"github.com/azraelsec/chip-8/internal/display"
	"github.com/azraelsec/chip-8/internal/keys"
	"github.com/azraelsec/chip-8/internal/ram"
)

type AudioController interface {
	StartBeep()
	StopBeep()
}

type Chip8 struct {
	cpu     *cpu.CPU
	ram     ram.Ram
	display display.GFX
	keys    keys.Keys

	opcode uint16

	shouldDraw bool

	ophandlers map[uint16]func() error
	audio      AudioController
}

func New(program []byte) *Chip8 {
	c := &Chip8{
		cpu:     cpu.New(),
		ram:     ram.New(program),
		display: display.GFX{},
		keys:    keys.Keys{},

		ophandlers: make(map[uint16]func() error),
		audio:      nil,
	}

	c.registerOpHandler(0x0000, c.zxSwitch)
	c.registerOpHandler(0x1000, c.jump)
	c.registerOpHandler(0x2000, c.callFunc)
	c.registerOpHandler(0x3000, c.skipIf)
	c.registerOpHandler(0x4000, c.skipIfNot)
	c.registerOpHandler(0x5000, c.skipIfReg)
	c.registerOpHandler(0x6000, c.setRegister)
	c.registerOpHandler(0x7000, c.addRegister)
	c.registerOpHandler(0x8000, c.eightxSwitch)
	c.registerOpHandler(0x9000, c.skipIfNotReg)
	c.registerOpHandler(0xa000, c.setIndex)
	c.registerOpHandler(0xb000, c.jumpToVZPlus)
	c.registerOpHandler(0xc000, c.rand)
	c.registerOpHandler(0xd000, c.draw)
	c.registerOpHandler(0xe000, c.exSwitch)
	c.registerOpHandler(0xf000, c.fxSwitch)

	return c
}

func (c *Chip8) SetKey(k uint8, s bool) {
	if s {
		c.keys[k] = 1
	} else {
		c.keys[k] = 0
	}
}

func (c *Chip8) SetAudio(audio AudioController) {
	c.audio = audio
}

func (c *Chip8) registerOpHandler(opcode uint16, h func() error) {
	c.ophandlers[opcode] = h
}

func (c *Chip8) Tick() error {
	c.fetch()
	if err := c.decodeExecute(); err != nil {
		return err
	}

	if c.cpu.DelayTimer > 0 {
		c.cpu.DelayTimer -= 1
	}

	if c.cpu.SoundTimer > 0 {
		c.cpu.SoundTimer -= 1
		if c.audio != nil {
			c.audio.StartBeep()
		}
	} else {
		if c.audio != nil {
			c.audio.StopBeep()
		}
	}

	return nil
}

func (c *Chip8) Buffer() []byte {
	return c.display[:]
}

func (c *Chip8) CheckDraw() bool {
	v := c.shouldDraw
	c.shouldDraw = false
	return v
}

func (c *Chip8) fetch() {
	c.opcode = uint16(c.ram[c.cpu.PC])<<8 | uint16(c.ram[c.cpu.PC+1])
}

func (c *Chip8) decodeExecute() error {
	slog.Info("decoding", slog.String("opcode", fmt.Sprintf("0x%x", uint16(c.opcode))), slog.String("opcode&0xF000", fmt.Sprintf("0x%x", uint16(c.opcode&0xF000))))
	return c.ophandlers[uint16(c.opcode&0xF000)]()
}

func (c *Chip8) zxSwitch() error {
	switch c.opcode & 0xF00F {
	case 0x0000:
		c.clearScreen()
	case 0x000E:
		c.returnFromSubroutine()
	default:
		if (c.opcode & 0xF000) == 0x0000 {
			c.callRCA1802()
		} else {
			return c.unknowOpCodeErr()
		}
	}
	return nil
}

func (c *Chip8) eightxSwitch() error {
	switch c.opcode & 0x000F {
	case 0x0000:
		c.assign()
	case 0x0001:
		c.or()
	case 0x0002:
		c.and()
	case 0x0003:
		c.xor()
	case 0x0004:
		c.add()
	case 0x0005:
		c.sub()
	case 0x0006:
		c.shiftr()
	case 0x0007:
		c.vxSubVy()
	case 0x000E:
		c.shiftl()
	default:
		return c.unknowOpCodeErr()
	}
	return nil
}

func (c *Chip8) exSwitch() error {
	switch c.opcode & 0x000F {
	case 0x000E:
		c.skipIfKeyPressed()
	case 0x0001:
		c.skipIfKeyNotPressed()
	default:
		return c.unknowOpCodeErr()
	}
	return nil
}

func (c *Chip8) fxSwitch() error {
	switch c.opcode & 0x00FF {
	case 0x0007:
		c.delayToVx()
	case 0x000A:
		c.waitForInput()
	case 0x0015:
		c.vxToDelay()
	case 0x0018:
		c.vxToSound()
	case 0x001E:
		c.addVxToI()
	case 0x0029:
		c.setCharToI()
	case 0x0033:
		c.vxToRAM()
	case 0x0055:
		c.regDump()
	case 0x0065:
		c.regLoad()
	default:
		return c.unknowOpCodeErr()
	}
	return nil
}

// 0x0NNN
func (c *Chip8) callRCA1802() {
	c.cpu.PC += 2
}

// 0x00E0
func (c *Chip8) clearScreen() {
	c.display.Clear()
	c.cpu.PC += 2
}

// 0x00EE
func (c *Chip8) returnFromSubroutine() {
	c.cpu.PC = c.cpu.Stack[c.cpu.SP]
	c.cpu.SP -= 1
}

// 0x1NNN
func (c *Chip8) jump() error {
	c.cpu.PC = c.opcode & 0x0FFF
	return nil
}

// 0x2NNN
func (c *Chip8) callFunc() error {
	c.cpu.SP += 1
	c.cpu.Stack[c.cpu.SP] = c.cpu.PC + 2
	c.cpu.PC = c.opcode & 0x0FFF
	return nil
}

// 0x3XNN
func (c *Chip8) skipIf() error {
	vi := (c.opcode & 0x0F00) >> 8
	d := uint8(c.opcode & 0x00FF)
	if c.cpu.V[vi] == d {
		c.cpu.PC += 4
	} else {
		c.cpu.PC += 2
	}
	return nil
}

// 0x4XNN
func (c *Chip8) skipIfNot() error {
	vi := (c.opcode & 0x0F00) >> 8
	d := uint8(c.opcode & 0x00FF)
	if c.cpu.V[vi] != d {
		c.cpu.PC += 4
	} else {
		c.cpu.PC += 2
	}
	return nil
}

// 0x5XY0
func (c *Chip8) skipIfReg() error {
	vx, vy := (c.opcode&0x0F00)>>8, (c.opcode&0x00F0)>>4
	if c.cpu.V[vx] == c.cpu.V[vy] {
		c.cpu.PC += 4
	} else {
		c.cpu.PC += 2
	}
	return nil
}

// 0x6XNN
func (c *Chip8) setRegister() error {
	vi := (c.opcode & 0x0F00) >> 8
	d := c.opcode & 0x00FF
	c.cpu.V[vi] = uint8(d)
	c.cpu.PC += 2
	return nil
}

// 0x7XNN
func (c *Chip8) addRegister() error {
	vi := (c.opcode & 0x0F00) >> 8
	d := c.opcode & 0x00FF
	c.cpu.V[vi] += uint8(d)
	c.cpu.PC += 2
	return nil
}

// 0x8XY0
func (c *Chip8) assign() {
	c.cpu.V[(c.opcode&0x0F00)>>8] = c.cpu.V[(c.opcode&0x00F0)>>4]
	c.cpu.PC += 2
}

// 0x8XY1
func (c *Chip8) or() {
	c.cpu.V[(c.opcode&0x0F00)>>8] |= c.cpu.V[(c.opcode&0x00F0)>>4]
	c.cpu.PC += 2
}

// 0x8XY2
func (c *Chip8) and() {
	c.cpu.V[(c.opcode&0x0F00)>>8] &= c.cpu.V[(c.opcode&0x00F0)>>4]
	c.cpu.PC += 2
}

// 0x8XY3
func (c *Chip8) xor() {
	c.cpu.V[(c.opcode&0x0F00)>>8] ^= c.cpu.V[(c.opcode&0x00F0)>>4]
	c.cpu.PC += 2
}

// 0x8XY4
func (c *Chip8) add() {
	if uint16(c.cpu.V[(c.opcode&0x0F00)>>8]+c.cpu.V[(c.opcode&0x00F0)>>4]) > 0xFF {
		c.cpu.V[0xF] = 1
	} else {
		c.cpu.V[0xF] = 0
	}

	c.cpu.V[(c.opcode&0x0F00)>>8] += c.cpu.V[(c.opcode&0x00F0)>>4]
	c.cpu.PC += 2
}

// 0x8XY5
func (c *Chip8) sub() {
	if c.cpu.V[(c.opcode&0x0F00)>>8] >= c.cpu.V[(c.opcode&0x00F0)>>4] {
		c.cpu.V[0xF] = 1
	} else {
		c.cpu.V[0xF] = 0
	}

	c.cpu.V[(c.opcode&0x0F00)>>8] -= c.cpu.V[(c.opcode&0x00F0)>>4]
	c.cpu.PC += 2
}

// 0x8XY6
func (c *Chip8) shiftr() {
	c.cpu.V[0xF] = c.cpu.V[(c.opcode&0x0F00)>>8] & 0x1
	c.cpu.V[(c.opcode&0x0F00)>>8] >>= 1
	c.cpu.PC += 2
}

// 0x8XY7
func (c *Chip8) vxSubVy() {
	if c.cpu.V[(c.opcode&0x00F0)>>4] >= c.cpu.V[(c.opcode&0x0F00)>>8] {
		c.cpu.V[0xF] = 1
	} else {
		c.cpu.V[0xF] = 0
	}
	c.cpu.V[(c.opcode&0x0F00)>>8] = c.cpu.V[(c.opcode&0x00F0)>>4] - c.cpu.V[(c.opcode&0x0F00)>>8]
	c.cpu.PC += 2
}

// 0x8XYE
func (c *Chip8) shiftl() {
	c.cpu.V[0xF] = (c.cpu.V[(c.opcode&0x0F00)>>8] & 0x80) >> 7
	c.cpu.V[(c.opcode&0x0F00)>>8] <<= 1
	c.cpu.PC += 2
}

// 0x9XY0
func (c *Chip8) skipIfNotReg() error {
	vx, vy := (c.opcode&0x0F00)>>8, (c.opcode&0x00F0)>>4
	if c.cpu.V[vx] != c.cpu.V[vy] {
		c.cpu.PC += 4
	} else {
		c.cpu.PC += 2
	}
	return nil
}

// 0xANNN
func (c *Chip8) setIndex() error {
	c.cpu.I = c.opcode & 0x0FFF
	c.cpu.PC += 2
	return nil
}

// 0xBNNN
func (c *Chip8) jumpToVZPlus() error {
	nnn := c.opcode & 0x0FFF
	c.cpu.PC = nnn + uint16(c.cpu.V[0])
	return nil
}

// 0xCXNN
func (c *Chip8) rand() error {
	vx, nn := (c.opcode&0x0F00)>>8, uint8(c.opcode&0x00FF)
	c.cpu.V[vx] = uint8(rand.Intn(256)) & nn
	c.cpu.PC += 2
	return nil
}

// 0xDXYN
func (c *Chip8) draw() error {
	x, y := uint16(c.cpu.V[(c.opcode&0x0F00)>>8]), uint16(c.cpu.V[(c.opcode&0x00F0)>>4])
	h := uint16(c.opcode & 0x000F)

	c.cpu.V[0xF] = 0

	for yline := range h {
		pixel := c.ram[c.cpu.I+yline]
		for xline := range 8 {
			if (pixel & (0b10000000 >> xline)) != 0 {
				px := (x + uint16(xline)) % display.DisplayWidth
				py := (y + yline) % display.DisplayHeight
				cfgPixel := px + (py * display.DisplayWidth)
				if c.display[cfgPixel] == 1 {
					c.cpu.V[0xF] = 1
				}
				c.display[cfgPixel] ^= 1
			}
		}
	}

	c.shouldDraw = true
	c.cpu.PC += 2
	return nil
}

// 0xEX9E
func (c *Chip8) skipIfKeyPressed() {
	if c.keys[c.cpu.V[(c.opcode&0x0F00)>>8]] == 1 {
		c.cpu.PC += 4
	} else {
		c.cpu.PC += 2
	}
}

// 0xEXA1
func (c *Chip8) skipIfKeyNotPressed() {
	if c.keys[c.cpu.V[(c.opcode&0x0F00)>>8]] == 1 {
		c.cpu.PC += 2
	} else {
		c.cpu.PC += 4
	}
}

// 0xFX07
func (c *Chip8) delayToVx() {
	c.cpu.V[(c.opcode&0x0F00)>>8] = c.cpu.DelayTimer
	c.cpu.PC += 2
}

// 0xFX0A
func (c *Chip8) waitForInput() {
	for i, k := range c.keys {
		if k != 0 {
			c.cpu.V[(c.opcode&0x0F00)>>8] = uint8(i)
			c.cpu.PC += 2
		}
	}
}

// 0xFX15
func (c *Chip8) vxToDelay() {
	c.cpu.DelayTimer = c.cpu.V[(c.opcode&0x0F00)>>8]
	c.cpu.PC += 2
}

// 0xFX18
func (c *Chip8) vxToSound() {
	c.cpu.SoundTimer = c.cpu.V[(c.opcode&0x0F00)>>8]
	c.cpu.PC += 2
}

// 0xFX1E
func (c *Chip8) addVxToI() {
	c.cpu.I += uint16(c.cpu.V[(c.opcode&0x0F00)>>8])
	c.cpu.PC += 2
}

// 0xFX29
func (c *Chip8) setCharToI() {
	c.cpu.I = uint16(c.cpu.V[(c.opcode&0x0F00)>>8]) * 5
	c.cpu.PC += 2
}

// 0xFX33
func (c *Chip8) vxToRAM() {
	c.ram[c.cpu.I] = c.cpu.V[(c.opcode&0x0F00)>>8] / 100
	c.ram[c.cpu.I+1] = (c.cpu.V[(c.opcode&0x0F00)>>8] / 10) % 10
	c.ram[c.cpu.I+2] = (c.cpu.V[(c.opcode&0x0F00)>>8] % 100) % 10
	c.cpu.PC += 2
}

// 0xFX55
func (c *Chip8) regDump() {
	x := (c.opcode & 0x0F00) >> 8
	for acc := uint16(0); acc <= x; acc++ {
		c.ram[c.cpu.I+acc] = c.cpu.V[acc]
	}
	c.cpu.PC += 2
}

// 0xFX65
func (c *Chip8) regLoad() {
	x := (c.opcode & 0x0F00) >> 8
	for acc := uint16(0); acc <= x; acc++ {
		c.cpu.V[acc] = c.ram[c.cpu.I+acc]
	}
	c.cpu.PC += 2
}

func (c *Chip8) unknowOpCodeErr() error {
	return fmt.Errorf("unknown opcode: 0x%x", c.opcode)
}
