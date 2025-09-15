package cpu

import "github.com/azraelsec/chip-8/internal/ram"

type CPU struct {
	// 15 general purpose 8bit registers (V0...VE) + carry flag (VF)
	V [16]byte
	// index register and program counter (0x000 -> 0xFFF)
	I, PC uint16

	DelayTimer byte
	SoundTimer byte

	// stack and stack pointer for stack frames (call and jump...)
	Stack [16]uint16
	SP    uint16
}

func New() *CPU {
	return &CPU{
		PC: ram.ProgramStartAddr,
	}
}
