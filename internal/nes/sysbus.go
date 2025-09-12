package nes

const (
	cpuRamSize int = 2048
)

type BusReadWriter interface {
	Read(uint16) uint8
	Write(uint16, uint8)
}

type SysBus struct {
	cpuRam    [cpuRamSize]uint8
	cartridge *Cartridge
}

func NewSysBus() *SysBus {
	return &SysBus{}
}

func (bus *SysBus) InsertCartridge(cartridge *Cartridge) {
	bus.cartridge = cartridge
}

func (bus *SysBus) Read(address uint16) uint8 {
	if address < 0x0800 {
		return bus.cpuRam[address]
	}
	return 0x00
}

func (bus *SysBus) Write(address uint16, data uint8) {
	if address < 0x0800 {
		bus.cpuRam[address] = data
	}
}
