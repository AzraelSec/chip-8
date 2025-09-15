package display

const (
	DisplayWidth  = 64
	DisplayHeight = 32
)

type GFX [DisplayWidth * DisplayHeight]byte

func (gfx *GFX) Clear() {
	*gfx = GFX{}
}
