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

	bus         BusReadWriter
	cycleDelay  int
	totalCycles int
	handleIrq   bool
	handleNmi   bool
}

func NewCpu(bus BusReadWriter) *Cpu {
	pcLow := bus.Read(resetVector)
	pcHigh := bus.Read(resetVector + 1)
	cpu := &Cpu{
		a: 0, x: 0, y: 0,
		sp:     initialStackPointer,
		pc:     uint16(pcHigh)<<8 | uint16(pcLow),
		status: initialStatus, bus: bus,
		cycleDelay: 7, totalCycles: 0,
		handleIrq: false, handleNmi: false,
	}
	return cpu
}

func (cpu *Cpu) Clock() {
	if cpu.cycleDelay <= 0 {
		if cpu.handleIrq && cpu.status&flagIntDisable == 0 {
			low := cpu.bus.Read(irqVector)
			high := cpu.bus.Read(irqVector + 1)
			address := uint16(high)<<8 | uint16(low)
			cpu.pc = address
			cpu.handleIrq = false
		} else if cpu.handleNmi {
			low := cpu.bus.Read(nmiVector)
			high := cpu.bus.Read(nmiVector + 1)
			address := uint16(high)<<8 | uint16(low)
			cpu.pc = address
			cpu.handleNmi = false
		}

		opcode := cpu.bus.Read(cpu.pc)
		instruction := opcodes[opcode]
		currPc := cpu.pc
		cpu.pc += uint16(instruction.bytes)
		instruction.fn(cpu, instruction.addrMode, currPc)
		cpu.cycleDelay += instruction.cycles
	}
	cpu.cycleDelay--
	cpu.totalCycles++
}

func (cpu *Cpu) Irq() {
	cpu.handleIrq = true
}

func (cpu *Cpu) Nmi() {
	cpu.handleNmi = true
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

func (cpu *Cpu) mustGetAddress(addrMode addressMode, pc uint16) (uint16, bool) {
	switch addrMode {
	case addrModeImplied:
		return pc, false
	case addrModeImmediate:
		return pc, false
	case addrModeZeroPage:
		return cpu.getZeroPageAddress(pc), false
	case addrModeZeroPageX:
		return cpu.getZeroPageOffsetAddress(pc, cpu.x), false
	case addrModeZeroPageY:
		return cpu.getZeroPageOffsetAddress(pc, cpu.y), false
	case addrModeAbsolute:
		return cpu.getAbsoluteAddress(pc), false
	case addrModeAbsoluteX:
		return cpu.getAbsoluteOffsetAddress(pc, cpu.x)
	case addrModeAbsoluteY:
		return cpu.getAbsoluteOffsetAddress(pc, cpu.y)
	case addrModeRelative:
		return cpu.getRelativeAddress(pc)
	case addrModeIndirect:
		return cpu.getIndirectAddress(pc), false
	case addrModeIndexedIndir:
		return cpu.getIndexedIndirectAddress(pc, cpu.x), false
	case addrModeIndirIndexed:
		return cpu.getIndirectIndexedAddress(pc, cpu.y)
	}
	panic("invalid address mode")
}

func (cpu *Cpu) getZeroPageAddress(pc uint16) uint16 {
	address := cpu.bus.Read(pc + 1)
	return uint16(address)
}

func (cpu *Cpu) getZeroPageOffsetAddress(pc uint16, offset uint8) uint16 {
	zeroPageAddr := cpu.bus.Read(pc + 1)
	zeroPageAddr += offset
	return uint16(zeroPageAddr)
}

func (cpu *Cpu) getAbsoluteAddress(pc uint16) uint16 {
	low := cpu.bus.Read(pc + 1)
	high := cpu.bus.Read(pc + 2)
	return uint16(high)<<8 | uint16(low)
}

func (cpu *Cpu) getAbsoluteOffsetAddress(pc uint16, offset uint8) (uint16, bool) {
	address := cpu.getAbsoluteAddress(pc)
	offsetAddress := address + uint16(offset)
	if address&0xFF00 != offsetAddress&0xFF00 {
		return offsetAddress, true
	} else {
		return offsetAddress, false
	}
}

func (cpu *Cpu) getRelativeAddress(pc uint16) (uint16, bool) {
	offset := int8(cpu.bus.Read(pc + 1))
	offsetAddress := int16(pc+2) + int16(offset)
	address := uint16(offsetAddress)
	if (pc+2)&0xFF00 != address&0xFF00 {
		return address, true
	} else {
		return address, false
	}
}

func (cpu *Cpu) getIndirectAddress(pc uint16) uint16 {
	jumpAddress := cpu.getAbsoluteAddress(pc)
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

func (cpu *Cpu) getIndexedIndirectAddress(pc uint16, offset uint8) uint16 {
	zeroPageAddr := cpu.bus.Read(pc + 1)
	zeroPageAddr += offset
	low := cpu.bus.Read(uint16(zeroPageAddr))
	high := cpu.bus.Read(uint16(zeroPageAddr + 1))
	return uint16(high)<<8 | uint16(low)
}

func (cpu *Cpu) getIndirectIndexedAddress(pc uint16, offset uint8) (uint16, bool) {
	zeroPageAddr := cpu.bus.Read(pc + 1)
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
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
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
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
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
		address, _ = cpu.mustGetAddress(addrMode, pc)
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

// branch if carry set
func (cpu *Cpu) bcs(addrMode addressMode, pc uint16) {
	if !cpu.testFlag(flagCarry) {
		return
	}
	address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
	cpu.pc = address

	cpu.cycleDelay++
	if pageCrossed {
		cpu.cycleDelay++
	}
}

// branch if carry clear
func (cpu *Cpu) bcc(addrMode addressMode, pc uint16) {
	if cpu.testFlag(flagCarry) {
		return
	}
	address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
	cpu.pc = address

	cpu.cycleDelay++
	if pageCrossed {
		cpu.cycleDelay++
	}
}

// bit test
func (cpu *Cpu) bit(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
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
	address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
	cpu.pc = address

	cpu.cycleDelay++
	if pageCrossed {
		cpu.cycleDelay++
	}
}

// branch if not equal
func (cpu *Cpu) bne(addrMode addressMode, pc uint16) {
	if cpu.testFlag(flagZero) {
		return
	}
	address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
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
	address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
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
	address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
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
	address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
	cpu.pc = address

	cpu.cycleDelay++
	if pageCrossed {
		cpu.cycleDelay++
	}
}

// branch if equal
func (cpu *Cpu) beq(addrMode addressMode, pc uint16) {
	if !cpu.testFlag(flagZero) {
		return
	}
	address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
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

// clear decimal
func (cpu *Cpu) cld(addrMode addressMode, pc uint16) {
	cpu.clearFlag(flagDecimal)
}

// clear interrupt disable
func (cpu *Cpu) cli(addrMode addressMode, pc uint16) {
	cpu.clearFlag(flagIntDisable)
}

// clear overflow
func (cpu *Cpu) clv(addrMode addressMode, pc uint16) {
	cpu.clearFlag(flagOverflow)
}

// compare a
func (cpu *Cpu) cmp(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.bus.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.bus.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	result := cpu.a - value

	if cpu.a >= value {
		cpu.setFlag(flagCarry)
	} else {
		cpu.clearFlag(flagCarry)
	}

	if cpu.a == value {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if result&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// compare x
func (cpu *Cpu) cpx(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.bus.Read(pc + 1)
	} else {
		address, _ := cpu.mustGetAddress(addrMode, pc)
		value = cpu.bus.Read(address)
	}

	result := cpu.x - value

	if cpu.x >= value {
		cpu.setFlag(flagCarry)
	} else {
		cpu.clearFlag(flagCarry)
	}

	if cpu.x == value {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if result&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// compare y
func (cpu *Cpu) cpy(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.bus.Read(pc + 1)
	} else {
		address, _ := cpu.mustGetAddress(addrMode, pc)
		value = cpu.bus.Read(address)
	}

	result := cpu.y - value

	if cpu.y >= value {
		cpu.setFlag(flagCarry)
	} else {
		cpu.clearFlag(flagCarry)
	}

	if cpu.y == value {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if result&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// decrement memory and compare a
func (cpu *Cpu) dcp(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.bus.Read(address)
	value--
	cpu.bus.Write(address, value)

	result := cpu.a - value

	if cpu.a >= value {
		cpu.setFlag(flagCarry)
	} else {
		cpu.clearFlag(flagCarry)
	}

	if cpu.a == value {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if result&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// decrement memory
func (cpu *Cpu) dec(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.bus.Read(address)
	value--
	cpu.bus.Write(address, value)

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
}

// decrement x
func (cpu *Cpu) dex(addrMode addressMode, pc uint16) {
	cpu.x--

	if cpu.x == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if cpu.x&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// decrement y
func (cpu *Cpu) dey(addrMode addressMode, pc uint16) {
	cpu.y--

	if cpu.y == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if cpu.y&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// bitwise exclusive or
func (cpu *Cpu) eor(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.bus.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
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

// increment memory
func (cpu *Cpu) inc(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.bus.Read(address)
	value++
	cpu.bus.Write(address, value)

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
}

// increment x
func (cpu *Cpu) inx(addrMode addressMode, pc uint16) {
	cpu.x++

	if cpu.x == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if cpu.x&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// increment y
func (cpu *Cpu) iny(addrMode addressMode, pc uint16) {
	cpu.y++

	if cpu.y == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if cpu.y&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// increment memory and subtract with carry
func (cpu *Cpu) isb(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.bus.Read(address)
	value++
	cpu.bus.Write(address, value)

	result := int16(cpu.a) - int16(value)
	if !cpu.testFlag(flagCarry) {
		result--
	}

	if result >= 0 {
		cpu.setFlag(flagCarry)
	} else {
		cpu.clearFlag(flagCarry)
	}

	if result == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if (uint8(result)^cpu.a)&(uint8(result)^^value)&0x80 > 0 {
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

// jump
func (cpu *Cpu) jmp(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	cpu.pc = address
}

// jump to subroutine
func (cpu *Cpu) jsr(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	pc += 2
	low := uint8(pc & 0x00FF)
	high := uint8((pc & 0xFF00) >> 8)
	cpu.stackPush(high)
	cpu.stackPush(low)
	cpu.pc = address
}

// load a and load x
func (cpu *Cpu) lax(addrMode addressMode, pc uint16) {
	address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
	if pageCrossed {
		cpu.cycleDelay++
	}
	value := cpu.bus.Read(address)

	cpu.a = value
	cpu.x = value

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
}

// load a
func (cpu *Cpu) lda(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.bus.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.bus.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.a = value

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

// load x
func (cpu *Cpu) ldx(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.bus.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.bus.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.x = value

	if cpu.x == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if cpu.x&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// load y
func (cpu *Cpu) ldy(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.bus.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.bus.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.y = value

	if cpu.y == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if cpu.y&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// logical shift right
func (cpu *Cpu) lsr(addrMode addressMode, pc uint16) {
	var value uint8
	var address uint16
	if addrMode == addrModeAccumulator {
		value = cpu.a
	} else {
		address, _ = cpu.mustGetAddress(addrMode, pc)
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
	_, pageCrossed := cpu.mustGetAddress(addrMode, pc)
	if pageCrossed {
		cpu.cycleDelay++
	}
	// do nothing
}

// bitwise or
func (cpu *Cpu) ora(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.bus.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
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
	cpu.status |= flagUnused
}

// rotate left and bitwise and
func (cpu *Cpu) rla(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.bus.Read(address)

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

	cpu.bus.Write(address, value)

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

// rotate left
func (cpu *Cpu) rol(addrMode addressMode, pc uint16) {
	var value uint8
	var address uint16
	if addrMode == addrModeAccumulator {
		value = cpu.a
	} else {
		address, _ = cpu.mustGetAddress(addrMode, pc)
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
		address, _ = cpu.mustGetAddress(addrMode, pc)
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
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.bus.Read(address)

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

	cpu.bus.Write(address, value)

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

// return from interrupt
func (cpu *Cpu) rti(addrMode addressMode, pc uint16) {
	flags := cpu.stackPop()
	low := cpu.stackPop()
	high := cpu.stackPop()
	cpu.status = flags & 0xCF
	cpu.status |= flagUnused
	cpu.pc = uint16(high)<<8 | uint16(low)
}

// return from subroutine
func (cpu *Cpu) rts(addrMode addressMode, pc uint16) {
	low := cpu.stackPop()
	high := cpu.stackPop()
	address := uint16(high)<<8 | uint16(low)
	cpu.pc = address + 1
}

// store a and x
func (cpu *Cpu) sax(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.a & cpu.x
	cpu.bus.Write(address, value)
}

// subtract with carry
func (cpu *Cpu) sbc(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.bus.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.bus.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	result := int16(cpu.a) - int16(value)
	if !cpu.testFlag(flagCarry) {
		result--
	}

	if result >= 0 {
		cpu.setFlag(flagCarry)
	} else {
		cpu.clearFlag(flagCarry)
	}

	if result == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if (uint8(result)^cpu.a)&(uint8(result)^^value)&0x80 > 0 {
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

// set carry
func (cpu *Cpu) sec(addrMode addressMode, pc uint16) {
	cpu.setFlag(flagCarry)
}

// set decimal
func (cpu *Cpu) sed(addrMode addressMode, pc uint16) {
	cpu.setFlag(flagDecimal)
}

// set interrupt disable
func (cpu *Cpu) sei(addrMode addressMode, pc uint16) {
	cpu.setFlag(flagIntDisable)
}

// arithmetic shift left and bitwise or
func (cpu *Cpu) slo(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.bus.Read(address)

	if value&0x80 > 0 {
		cpu.setFlag(flagCarry)
	} else {
		cpu.clearFlag(flagCarry)
	}

	value <<= 1

	cpu.bus.Write(address, value)

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

// logical shift right and bitwise exclusive or
func (cpu *Cpu) sre(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.bus.Read(address)

	if value&0x01 > 0 {
		cpu.setFlag(flagCarry)
	} else {
		cpu.clearFlag(flagCarry)
	}

	value >>= 1
	cpu.bus.Write(address, value)

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

// store a
func (cpu *Cpu) sta(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	cpu.bus.Write(address, cpu.a)
}

// store x
func (cpu *Cpu) stx(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	cpu.bus.Write(address, cpu.x)
}

// store y
func (cpu *Cpu) sty(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	cpu.bus.Write(address, cpu.y)
}

// transfer a to x
func (cpu *Cpu) tax(addrMode addressMode, pc uint16) {
	cpu.x = cpu.a

	if cpu.x == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if cpu.x&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// transfer a to y
func (cpu *Cpu) tay(addrMode addressMode, pc uint16) {
	cpu.y = cpu.a

	if cpu.y == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}

	if cpu.y&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// transfer stack pointer to x
func (cpu *Cpu) tsx(addrMode addressMode, pc uint16) {
	cpu.x = cpu.sp

	if cpu.x == 0 {
		cpu.setFlag(flagZero)
	} else {
		cpu.clearFlag(flagZero)
	}
	if cpu.x&0x80 > 0 {
		cpu.setFlag(flagNegative)
	} else {
		cpu.clearFlag(flagNegative)
	}
}

// transfer x to a
func (cpu *Cpu) txa(addrMode addressMode, pc uint16) {
	cpu.a = cpu.x

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

// transfer x to stack pointer
func (cpu *Cpu) txs(addrMode addressMode, pc uint16) {
	cpu.sp = cpu.x
}

// transfer y to a
func (cpu *Cpu) tya(addrMode addressMode, pc uint16) {
	cpu.a = cpu.y

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
