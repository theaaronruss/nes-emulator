package nes

type BusReadWriter interface {
	Read(uint16) uint8
	Write(uint16, uint8)
}

type SysBus struct {
}

func (bus *SysBus) Read(address uint16) uint8 {
	return 0x00
}

func (bus *SysBus) Write(address uint16, data uint8) {
}
