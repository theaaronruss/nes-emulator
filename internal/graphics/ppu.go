package graphics

import "math/rand"

type Ppu struct {
	FrameBuffer   []uint8
	FrameComplete bool
	pixelOffset   int
}

func NewPpu() *Ppu {
	return &Ppu{
		FrameBuffer: make([]uint8, 4*256*240),
	}
}

func (p *Ppu) ClockCycle() {
	var color uint8
	if rand.Intn(2) == 1 {
		color = 0xFF
	} else {
		color = 0x00
	}
	p.FrameBuffer[p.pixelOffset*4] = color
	p.FrameBuffer[p.pixelOffset*4+1] = color
	p.FrameBuffer[p.pixelOffset*4+2] = color
	p.FrameBuffer[p.pixelOffset*4+3] = 0xFF
	p.pixelOffset++
	if p.pixelOffset >= 256*240 {
		p.pixelOffset = 0
		p.FrameComplete = true
	}
}
