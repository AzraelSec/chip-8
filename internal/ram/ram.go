package ram

// 0x000-0x1FF - Chip 8 interpreter (contains font set in emu)
// 0x050-0x0A0 - Used for the built in 4x5 pixel font set (0-F)
// 0x200-0xFFF - Program ROM and work RAM

const (
	fontStartAddr    = 0x50
	ProgramStartAddr = 0x200
)

// 4KB in total
type Ram [4000]byte

func New(program []byte) Ram {
	mem := Ram{}
	// 0x050-0x0A0 - Used for the built in 4x5 pixel font set (0-F)
	copy(mem[fontStartAddr:], fontMap[:])
	// 0x200-0xFFF - Program ROM and work RAM
	copy(mem[ProgramStartAddr:], program)
	return mem
}
