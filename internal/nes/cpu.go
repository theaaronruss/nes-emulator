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

func (cpu *Cpu) mustGetAddress(addrMode addressMode) uint16 {
	switch addrMode {
	case addrModeZeroPage:
		return cpu.getZeroPageAddress()
	}
	panic("invalid address mode")
}

func (cpu *Cpu) getZeroPageAddress() uint16 {
	address := cpu.bus.Read(cpu.pc + 1)
	return uint16(address)
}

func (cpu *Cpu) getAbsoluteAddress() uint16 {
	low := cpu.bus.Read(cpu.pc + 1)
	high := cpu.bus.Read(cpu.pc + 2)
	return uint16(high)<<8 | uint16(low)
}
