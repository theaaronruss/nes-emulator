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

func (cpu *Cpu) Clock() {
	if cpu.cycleDelay <= 0 {
		opcode := cpu.bus.Read(cpu.pc)
		instruction := opcodes[opcode]
		currPc := cpu.pc
		cpu.pc += uint16(instruction.bytes)
		instruction.fn(cpu, instruction.addrMode, currPc)
		cpu.cycleDelay += instruction.cycles
	}
	cpu.cycleDelay--
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

// add with carry
func (cpu *Cpu) adc(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.bus.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode)
		value = cpu.bus.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	result := uint16(cpu.a) + uint16(value)
	if cpu.testFlag(flagCarry) {
		result++
	}

	if result > 255 {
		cpu.setFlag(flagCarry)
	} else {
		cpu.clearFlag(flagCarry)
	}

	if uint8(result) == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if (uint8(result)^cpu.a)&(uint8(result)^value)&0x80 > 0 {
		cpu.setFlag(flagOverflow)
	} else {
		cpu.clearFlag(flagOverflow)
	}

	if result&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}

	cpu.a = uint8(result)
}

// bitwise and
func (cpu *Cpu) and(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.bus.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode)
		value = cpu.bus.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.a &= value

	if cpu.a == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if cpu.a&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// arithmetic shift left
func (cpu *Cpu) asl(addrMode addressMode, pc uint16) {
	var value uint8
	var address uint16
	if addrMode == addrModeAccumulator {
		value = cpu.a
	} else {
		address, _ = cpu.mustGetAddress(addrMode)
		value = cpu.bus.Read(address)
	}

	if value&0x80 > 0 {
		cpu.setFlag(flagCarry)
	} else {
		cpu.clearFlag(flagCarry)
	}

	value <<= 1

	if value == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if value&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}

	if addrMode == addrModeAccumulator {
		cpu.a = value
	} else {
		cpu.bus.Write(address, value)
	}
}

// bit test
func (cpu *Cpu) bit(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode)
	value := cpu.bus.Read(address)
	result := cpu.a & value

	if result == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if value&0x40 > 0 {
		cpu.setFlag(flagOverflow)
	} else {
		cpu.clearFlag(flagOverflow)
	}

	if value&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// branch if minus
func (cpu *Cpu) bmi(addrMode addressMode, pc uint16) {
	if !cpu.testFlag(flagNegative) {
		return
	}
	address, pageCrossed := cpu.mustGetAddress(addrMode)
	cpu.pc = address

	cpu.cycleDelay++
	if pageCrossed {
		cpu.cycleDelay++
	}
}

// branch if plus
func (cpu *Cpu) bpl(addrMode addressMode, pc uint16) {
	if cpu.testFlag(flagNegative) {
		return
	}
	address, pageCrossed := cpu.mustGetAddress(addrMode)
	cpu.pc = address

	cpu.cycleDelay++
	if pageCrossed {
		cpu.cycleDelay++
	}
}

// force break
func (cpu *Cpu) brk(addrMode addressMode, pc uint16) {
	pc += 2
	oldPcLow := uint8(pc & 0x00FF)
	oldPcHigh := uint8(pc & 0xFF00 >> 8)
	cpu.stackPush(oldPcHigh)
	cpu.stackPush(oldPcLow)
	cpu.stackPush(cpu.status | flagUnused | flagBreak)
	newPcLow := cpu.bus.Read(irqVector)
	newPcHigh := cpu.bus.Read(irqVector + 1)
	cpu.pc = uint16(newPcHigh)<<8 | uint16(newPcLow)
}

// branch if overflow clear
func (cpu *Cpu) bvc(addrMode addressMode, pc uint16) {
	if cpu.testFlag(flagOverflow) {
		return
	}
	address, pageCrossed := cpu.mustGetAddress(addrMode)
	cpu.pc = address

	cpu.cycleDelay++
	if pageCrossed {
		cpu.cycleDelay++
	}
}

// branch if overflow set
func (cpu *Cpu) bvs(addrMode addressMode, pc uint16) {
	if !cpu.testFlag(flagOverflow) {
		return
	}
	address, pageCrossed := cpu.mustGetAddress(addrMode)
	cpu.pc = address

	cpu.cycleDelay++
	if pageCrossed {
		cpu.cycleDelay++
	}
}

// clear carry
func (cpu *Cpu) clc(addrMode addressMode, pc uint16) {
	cpu.clearFlag(flagCarry)
}

// clear interrupt disable
func (cpu *Cpu) cli(addrMode addressMode, pc uint16) {
	cpu.clearFlag(flagIntDisable)
}

// bitwise exclusive or
func (cpu *Cpu) eor(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.bus.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode)
		value = cpu.bus.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.a ^= value

	if cpu.a == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}
	if cpu.a&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// jump
func (cpu *Cpu) jmp(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode)
	cpu.pc = address
}

// jump to subroutine
func (cpu *Cpu) jsr(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode)
	cpu.pc += 2
	low := uint8(pc & 0x00FF)
	high := uint8((pc & 0xFF00) >> 8)
	cpu.stackPush(high)
	cpu.stackPush(low)
	cpu.pc = address
}

// logical shift right
func (cpu *Cpu) lsr(addrMode addressMode, pc uint16) {
	var value uint8
	var address uint16
	if addrMode == addrModeAccumulator {
		value = cpu.a
	} else {
		address, _ = cpu.mustGetAddress(addrMode)
		value = cpu.bus.Read(address)
	}

	if value&0x01 > 0 {
		cpu.setFlag(flagCarry)
	} else {
		cpu.clearFlag(flagCarry)
	}

	value >>= 1

	if value == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	cpu.clearFlag(flagNegative)
	if addrMode == addrModeAccumulator {
		cpu.a = value
	} else {
		cpu.bus.Write(address, value)
	}
}

// no operation
func (cpu *Cpu) nop(addrMode addressMode, pc uint16) {
	// do nothing
}

// bitwise or
func (cpu *Cpu) ora(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.bus.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode)
		value = cpu.bus.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.a |= value

	if cpu.a == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if cpu.a&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// push a
func (cpu *Cpu) pha(addrMode addressMode, pc uint16) {
	cpu.stackPush(cpu.a)
}

// push processor status
func (cpu *Cpu) php(addrMode addressMode, pc uint16) {
	cpu.stackPush(cpu.status | flagUnused | flagBreak)
}

// pull a
func (cpu *Cpu) pla(addrMode addressMode, pc uint16) {
	cpu.a = cpu.stackPop()

	if cpu.a == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if cpu.a&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// pull processor status
func (cpu *Cpu) plp(addrMode addressMode, pc uint16) {
	flags := cpu.stackPop()
	cpu.status = flags & 0xCF
}

// rotate left and bitwise and
func (cpu *Cpu) rla(addrMode addressMode, pc uint16) {
	cpu.rol(addrMode, cpu.pc)
	cpu.and(addrMode, cpu.pc)
}

// rotate left
func (cpu *Cpu) rol(addrMode addressMode, pc uint16) {
	var value uint8
	var address uint16
	if addrMode == addrModeAccumulator {
		value = cpu.a
	} else {
		address, _ = cpu.mustGetAddress(addrMode)
		value = cpu.bus.Read(address)
	}

	carry := cpu.testFlag(flagCarry)
	if value&0x80 > 0 {
		cpu.setFlag(flagCarry)
	} else {
		cpu.clearFlag(flagCarry)
	}

	value <<= 1

	if carry {
		value |= 0x01
	}

	if value == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if value&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}

	if addrMode == addrModeAccumulator {
		cpu.a = value
	} else {
		cpu.bus.Write(address, value)
	}
}

// rotate right
func (cpu *Cpu) ror(addrMode addressMode, pc uint16) {
	var value uint8
	var address uint16
	if addrMode == addrModeAccumulator {
		value = cpu.a
	} else {
		address, _ = cpu.mustGetAddress(addrMode)
		value = cpu.bus.Read(address)
	}

	carry := cpu.testFlag(flagCarry)
	if value&0x01 > 0 {
		cpu.setFlag(flagCarry)
	} else {
		cpu.clearFlag(flagCarry)
	}

	value >>= 1
	if carry {
		value |= 0x80
	}

	if value == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if value&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}

	if addrMode == addrModeAccumulator {
		cpu.a = value
	} else {
		cpu.bus.Write(address, value)
	}
}

// rotate right and add with carry
func (cpu *Cpu) rra(addrMode addressMode, pc uint16) {
	cpu.ror(addrMode, pc)
	cpu.adc(addrMode, cpu.pc)
}

// return from interrupt
func (cpu *Cpu) rti(addrMode addressMode, pc uint16) {
	flags := cpu.stackPop()
	low := cpu.stackPop()
	high := cpu.stackPop()
	cpu.status = flags & 0xCF
	cpu.pc = uint16(high)<<8 | uint16(low)
}

// return from subroutine
func (cpu *Cpu) rts(addrMode addressMode, pc uint16) {
	low := cpu.stackPop()
	high := cpu.stackPop()
	address := uint16(high)<<8 | uint16(low)
	cpu.pc = address + 1
}

// set carry
func (cpu *Cpu) sec(addrMode addressMode, pc uint16) {
	cpu.setFlag(flagCarry)
}

// arithmetic shift left and bitwise or
func (cpu *Cpu) slo(addrMode addressMode, pc uint16) {
	cpu.asl(addrMode, cpu.pc)
	cpu.ora(addrMode, cpu.pc)
}

// logical shift right and bitwise exclusive or
func (cpu *Cpu) sre(addrMode addressMode, pc uint16) {
	cpu.lsr(addrMode, cpu.pc)
	cpu.eor(addrMode, cpu.pc)
}
