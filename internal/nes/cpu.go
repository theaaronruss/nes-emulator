package nes

const (
	nmiVector           uint16 = 0xFFFA
	resetVector         uint16 = 0xFFFC
	irqVector           uint16 = 0xFFFE
	stackBase           uint16 = 0x0100
	initialStackPointer uint8  = 0xFD
	initialStatus       uint8  = 0x24
)

// status flag masks
const (
	flagCarry uint8 = 1 << iota
	flagZero
	flagIntDisable
	flagDecimal
	flagBreak
	flagUnused
	flagOverflow
	flagNegative
)

type Cpu struct {
	a      uint8
	x      uint8
	y      uint8
	sp     uint8
	pc     uint16
	status uint8

	bus        BusReadWriter
	cycleDelay int
}

func NewCpu(bus BusReadWriter) *Cpu {
	pcLow := bus.Read(resetVector)
	pcHigh := bus.Read(resetVector + 1)
	cpu := &Cpu{
		a: 0, x: 0, y: 0,
		sp:     initialStackPointer,
		pc:     uint16(pcHigh)<<8 | uint16(pcLow),
		status: initialStatus, bus: bus,
		cycleDelay: 0,
	}
	return cpu
}

func (cpu *Cpu) setFlag(flag uint8) {
	cpu.status |= flag
}

func (cpu *Cpu) clearFlag(flag uint8) {
	cpu.status &= ^flag
}

func (cpu *Cpu) testFlag(flag uint8) bool {
	return cpu.status&flag > 0
}

func (cpu *Cpu) stackPush(data uint8) {
	address := stackBase + uint16(cpu.sp)
	cpu.bus.Write(address, data)
	cpu.sp--
}

func (cpu *Cpu) stackPop() uint8 {
	cpu.sp++
	address := stackBase + uint16(cpu.sp)
	return cpu.bus.Read(address)
}

func (cpu *Cpu) mustGetAddress(addrMode addressMode) (uint16, bool) {
	switch addrMode {
	case addrModeZeroPage:
		return cpu.getZeroPageAddress(), false
	case addrModeZeroPageX:
		return cpu.getZeroPageOffsetAddress(cpu.x), false
	case addrModeZeroPageY:
		return cpu.getZeroPageOffsetAddress(cpu.y), false
	case addrModeAbsolute:
		return cpu.getAbsoluteAddress(), false
	case addrModeAbsoluteX:
		return cpu.getAbsoluteOffsetAddress(cpu.x)
	case addrModeAbsoluteY:
		return cpu.getAbsoluteOffsetAddress(cpu.y)
	case addrModeRelative:
		return cpu.getRelativeAddress()
	case addrModeIndirect:
		return cpu.getIndirectAddress(), false
	case addrModeIndexedIndir:
		return cpu.getIndexedIndirectAddress(cpu.x), false
	case addrModeIndirIndexed:
		return cpu.getIndirectIndexedAddress(cpu.y)
	}
	panic("invalid address mode")
}

func (cpu *Cpu) getZeroPageAddress() uint16 {
	address := cpu.bus.Read(cpu.pc + 1)
	return uint16(address)
}

func (cpu *Cpu) getZeroPageOffsetAddress(offset uint8) uint16 {
	zeroPageAddr := cpu.bus.Read(cpu.pc + 1)
	zeroPageAddr += offset
	return uint16(zeroPageAddr)
}

func (cpu *Cpu) getAbsoluteAddress() uint16 {
	low := cpu.bus.Read(cpu.pc + 1)
	high := cpu.bus.Read(cpu.pc + 2)
	return uint16(high)<<8 | uint16(low)
}

func (cpu *Cpu) getAbsoluteOffsetAddress(offset uint8) (uint16, bool) {
	address := cpu.getAbsoluteAddress()
	offsetAddress := address + uint16(offset)
	if address&0xFF00 != offsetAddress&0xFF00 {
		return offsetAddress, true
	} else {
		return offsetAddress, false
	}
}

func (cpu *Cpu) getRelativeAddress() (uint16, bool) {
	offset := int8(cpu.bus.Read(cpu.pc + 1))
	offsetAddress := int16(cpu.pc+2) + int16(offset)
	address := uint16(offsetAddress)
	if (cpu.pc+2)&0xFF00 != address&0xFF00 {
		return address, true
	} else {
		return address, false
	}
}

func (cpu *Cpu) getIndirectAddress() uint16 {
	jumpAddress := cpu.getAbsoluteAddress()
	low := cpu.bus.Read(jumpAddress)

	// the 6502 has a bug where it wraps incorrectly if the jump address is at
	// a page boundary
	var high uint8
	if jumpAddress&0x00FF == 0x00FF {
		high = cpu.bus.Read(jumpAddress & 0xFF00)
	} else {
		high = cpu.bus.Read(jumpAddress + 1)
	}

	address := uint16(high)<<8 | uint16(low)
	return address
}

func (cpu *Cpu) getIndexedIndirectAddress(offset uint8) uint16 {
	zeroPageAddr := cpu.bus.Read(cpu.pc + 1)
	zeroPageAddr += offset
	low := cpu.bus.Read(uint16(zeroPageAddr))
	high := cpu.bus.Read(uint16(zeroPageAddr + 1))
	return uint16(high)<<8 | uint16(low)
}

func (cpu *Cpu) getIndirectIndexedAddress(offset uint8) (uint16, bool) {
	zeroPageAddr := cpu.bus.Read(cpu.pc + 1)
	low := cpu.bus.Read(uint16(zeroPageAddr))
	high := cpu.bus.Read(uint16(zeroPageAddr + 1))
	address := uint16(high)<<8 | uint16(low)
	offsetAddress := address + uint16(offset)
	if address&0xFF00 != offsetAddress&0xFF00 {
		return offsetAddress, true
	} else {
		return offsetAddress, false
	}
}
