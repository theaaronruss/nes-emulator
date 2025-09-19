package nes

// status flag masks
const (
	carryFlagMask uint8 = 1 << iota
	zeroFlagMask
	intDisableFlagMask
	decimalFlagMask
	breakFlagMask
	unusedFlagMask
	overflowFlagMask
	negativeFlagMask
)

const (
	nmiVector           uint16 = 0xFFFA
	resetVector         uint16 = 0xFFFC
	irqVector           uint16 = 0xFFFE
	stackBase           uint16 = 0x0100
	initialStackPointer uint8  = 0xFD
	initialStatus       uint8  = 0x24
)

type Cpu struct {
	a      uint8
	x      uint8
	y      uint8
	sp     uint8
	pc     uint16
	status uint8

	sys         *System
	cycleDelay  int
	totalCycles int
	handleIrq   bool
	handleNmi   bool
}

func NewCpu(sys *System) *Cpu {
	pcLow := sys.Read(resetVector)
	pcHigh := sys.Read(resetVector + 1)

	return &Cpu{
		sp:         initialStackPointer,
		pc:         uint16(pcHigh)<<8 | uint16(pcLow),
		status:     initialStatus,
		sys:        sys,
		cycleDelay: 7,
	}
}

func (cpu *Cpu) Clock() {
	if cpu.cycleDelay <= 0 {
		if cpu.handleIrq && cpu.status&intDisableFlagMask == 0 {
			low := cpu.sys.Read(irqVector)
			high := cpu.sys.Read(irqVector + 1)
			address := uint16(high)<<8 | uint16(low)
			cpu.pc = address
			cpu.handleIrq = false
		} else if cpu.handleNmi {
			low := cpu.sys.Read(nmiVector)
			high := cpu.sys.Read(nmiVector + 1)
			address := uint16(high)<<8 | uint16(low)
			cpu.pc = address
			cpu.handleNmi = false
		}

		opcode := cpu.sys.Read(cpu.pc)
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

func (cpu *Cpu) updateFlag(flag uint8, value bool) {
	if value {
		cpu.status |= flag
	} else {
		cpu.status &= ^flag
	}
}

func (cpu *Cpu) testFlag(flag uint8) bool {
	return cpu.status&flag > 0
}

func (cpu *Cpu) stackPush(data uint8) {
	address := stackBase + uint16(cpu.sp)
	cpu.sys.Write(address, data)
	cpu.sp--
}

func (cpu *Cpu) stackPop() uint8 {
	cpu.sp++
	address := stackBase + uint16(cpu.sp)
	return cpu.sys.Read(address)
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
	address := cpu.sys.Read(pc + 1)
	return uint16(address)
}

func (cpu *Cpu) getZeroPageOffsetAddress(pc uint16, offset uint8) uint16 {
	zeroPageAddr := cpu.sys.Read(pc + 1)
	zeroPageAddr += offset
	return uint16(zeroPageAddr)
}

func (cpu *Cpu) getAbsoluteAddress(pc uint16) uint16 {
	low := cpu.sys.Read(pc + 1)
	high := cpu.sys.Read(pc + 2)
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
	offset := int8(cpu.sys.Read(pc + 1))
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
	low := cpu.sys.Read(jumpAddress)

	// the 6502 has a bug where it wraps incorrectly if the jump address is at
	// a page boundary
	var high uint8
	if jumpAddress&0x00FF == 0x00FF {
		high = cpu.sys.Read(jumpAddress & 0xFF00)
	} else {
		high = cpu.sys.Read(jumpAddress + 1)
	}

	address := uint16(high)<<8 | uint16(low)
	return address
}

func (cpu *Cpu) getIndexedIndirectAddress(pc uint16, offset uint8) uint16 {
	zeroPageAddr := cpu.sys.Read(pc + 1)
	zeroPageAddr += offset
	low := cpu.sys.Read(uint16(zeroPageAddr))
	high := cpu.sys.Read(uint16(zeroPageAddr + 1))
	return uint16(high)<<8 | uint16(low)
}

func (cpu *Cpu) getIndirectIndexedAddress(pc uint16, offset uint8) (uint16, bool) {
	zeroPageAddr := cpu.sys.Read(pc + 1)
	low := cpu.sys.Read(uint16(zeroPageAddr))
	high := cpu.sys.Read(uint16(zeroPageAddr + 1))
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
		value = cpu.sys.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	result := uint16(cpu.a) + uint16(value)
	if cpu.testFlag(carryFlagMask) {
		result++
	}

	cpu.updateFlag(carryFlagMask, result > 255)
	cpu.updateFlag(zeroFlagMask, result == 0)
	cpu.updateFlag(overflowFlagMask, (uint8(result)^cpu.a)&(uint8(result)^value)&0x80 > 0)
	cpu.updateFlag(negativeFlagMask, result&0x80 > 0)

	cpu.a = uint8(result)
}

// bitwise and
func (cpu *Cpu) and(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.a &= value

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// arithmetic shift left
func (cpu *Cpu) asl(addrMode addressMode, pc uint16) {
	var value uint8
	var address uint16
	if addrMode == addrModeAccumulator {
		value = cpu.a
	} else {
		address, _ = cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.Read(address)
	}

	cpu.updateFlag(carryFlagMask, value&0x80 > 0)
	value <<= 1

	cpu.updateFlag(zeroFlagMask, value == 0)
	cpu.updateFlag(negativeFlagMask, value&0x80 > 0)

	if addrMode == addrModeAccumulator {
		cpu.a = value
	} else {
		cpu.sys.Write(address, value)
	}
}

// branch if carry set
func (cpu *Cpu) bcs(addrMode addressMode, pc uint16) {
	if !cpu.testFlag(carryFlagMask) {
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
	if cpu.testFlag(carryFlagMask) {
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
	value := cpu.sys.Read(address)
	result := cpu.a & value

	cpu.updateFlag(zeroFlagMask, result == 0)
	cpu.updateFlag(overflowFlagMask, value&0x40 > 0)
	cpu.updateFlag(negativeFlagMask, value&0x80 > 0)
}

// branch if minus
func (cpu *Cpu) bmi(addrMode addressMode, pc uint16) {
	if !cpu.testFlag(negativeFlagMask) {
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
	if cpu.testFlag(zeroFlagMask) {
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
	if cpu.testFlag(negativeFlagMask) {
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
	cpu.stackPush(cpu.status | unusedFlagMask | breakFlagMask)
	newPcLow := cpu.sys.Read(irqVector)
	newPcHigh := cpu.sys.Read(irqVector + 1)
	cpu.pc = uint16(newPcHigh)<<8 | uint16(newPcLow)
}

// branch if overflow clear
func (cpu *Cpu) bvc(addrMode addressMode, pc uint16) {
	if cpu.testFlag(overflowFlagMask) {
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
	if !cpu.testFlag(overflowFlagMask) {
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
	if !cpu.testFlag(zeroFlagMask) {
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
	cpu.updateFlag(carryFlagMask, false)
}

// clear decimal
func (cpu *Cpu) cld(addrMode addressMode, pc uint16) {
	cpu.updateFlag(decimalFlagMask, false)
}

// clear interrupt disable
func (cpu *Cpu) cli(addrMode addressMode, pc uint16) {
	cpu.updateFlag(intDisableFlagMask, false)
}

// clear overflow
func (cpu *Cpu) clv(addrMode addressMode, pc uint16) {
	cpu.updateFlag(overflowFlagMask, false)
}

// compare a
func (cpu *Cpu) cmp(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	result := cpu.a - value

	cpu.updateFlag(carryFlagMask, cpu.a >= value)
	cpu.updateFlag(zeroFlagMask, cpu.a == value)
	cpu.updateFlag(negativeFlagMask, result&0x80 > 0)
}

// compare x
func (cpu *Cpu) cpx(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.Read(pc + 1)
	} else {
		address, _ := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.Read(address)
	}

	result := cpu.x - value

	cpu.updateFlag(carryFlagMask, cpu.x >= value)
	cpu.updateFlag(zeroFlagMask, cpu.x == value)
	cpu.updateFlag(negativeFlagMask, result&0x80 > 0)
}

// compare y
func (cpu *Cpu) cpy(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.Read(pc + 1)
	} else {
		address, _ := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.Read(address)
	}

	result := cpu.y - value

	cpu.updateFlag(carryFlagMask, cpu.y >= value)
	cpu.updateFlag(zeroFlagMask, cpu.y == value)
	cpu.updateFlag(negativeFlagMask, result&0x80 > 0)
}

// decrement memory and compare a
func (cpu *Cpu) dcp(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.Read(address)
	value--
	cpu.sys.Write(address, value)

	result := cpu.a - value

	cpu.updateFlag(carryFlagMask, cpu.a >= value)
	cpu.updateFlag(zeroFlagMask, cpu.a == value)
	cpu.updateFlag(negativeFlagMask, result&0x80 > 0)
}

// decrement memory
func (cpu *Cpu) dec(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.Read(address)
	value--
	cpu.sys.Write(address, value)

	cpu.updateFlag(zeroFlagMask, value == 0)
	cpu.updateFlag(negativeFlagMask, value&0x80 > 0)
}

// decrement x
func (cpu *Cpu) dex(addrMode addressMode, pc uint16) {
	cpu.x--

	cpu.updateFlag(zeroFlagMask, cpu.x == 0)
	cpu.updateFlag(negativeFlagMask, cpu.x&0x80 > 0)
}

// decrement y
func (cpu *Cpu) dey(addrMode addressMode, pc uint16) {
	cpu.y--

	cpu.updateFlag(zeroFlagMask, cpu.y == 0)
	cpu.updateFlag(negativeFlagMask, cpu.y&0x80 > 0)
}

// bitwise exclusive or
func (cpu *Cpu) eor(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.a ^= value

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// increment memory
func (cpu *Cpu) inc(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.Read(address)
	value++
	cpu.sys.Write(address, value)

	cpu.updateFlag(zeroFlagMask, value == 0)
	cpu.updateFlag(negativeFlagMask, value&0x80 > 0)
}

// increment x
func (cpu *Cpu) inx(addrMode addressMode, pc uint16) {
	cpu.x++

	cpu.updateFlag(zeroFlagMask, cpu.x == 0)
	cpu.updateFlag(negativeFlagMask, cpu.x&0x80 > 0)
}

// increment y
func (cpu *Cpu) iny(addrMode addressMode, pc uint16) {
	cpu.y++

	cpu.updateFlag(zeroFlagMask, cpu.y == 0)
	cpu.updateFlag(negativeFlagMask, cpu.y&0x80 > 0)
}

// increment memory and subtract with carry
func (cpu *Cpu) isb(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.Read(address)
	value++
	cpu.sys.Write(address, value)

	result := int16(cpu.a) - int16(value)
	if !cpu.testFlag(carryFlagMask) {
		result--
	}

	cpu.updateFlag(carryFlagMask, result >= 0)
	cpu.updateFlag(zeroFlagMask, result == 0)
	cpu.updateFlag(overflowFlagMask, (uint8(result)^cpu.a)&(uint8(result)^^value)&0x80 > 0)
	cpu.updateFlag(negativeFlagMask, result&0x80 > 0)

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
	value := cpu.sys.Read(address)

	cpu.a = value
	cpu.x = value

	cpu.updateFlag(zeroFlagMask, value == 0)
	cpu.updateFlag(negativeFlagMask, value&0x80 > 0)
}

// load a
func (cpu *Cpu) lda(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.a = value

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// load x
func (cpu *Cpu) ldx(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.x = value

	cpu.updateFlag(zeroFlagMask, cpu.x == 0)
	cpu.updateFlag(negativeFlagMask, cpu.x&0x80 > 0)
}

// load y
func (cpu *Cpu) ldy(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.y = value

	cpu.updateFlag(zeroFlagMask, cpu.y == 0)
	cpu.updateFlag(negativeFlagMask, cpu.y&0x80 > 0)
}

// logical shift right
func (cpu *Cpu) lsr(addrMode addressMode, pc uint16) {
	var value uint8
	var address uint16
	if addrMode == addrModeAccumulator {
		value = cpu.a
	} else {
		address, _ = cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.Read(address)
	}

	cpu.updateFlag(carryFlagMask, value&0x01 > 0)
	value >>= 1

	cpu.updateFlag(zeroFlagMask, value == 0)
	cpu.updateFlag(negativeFlagMask, false)
	if addrMode == addrModeAccumulator {
		cpu.a = value
	} else {
		cpu.sys.Write(address, value)
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
		value = cpu.sys.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.a |= value

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// push a
func (cpu *Cpu) pha(addrMode addressMode, pc uint16) {
	cpu.stackPush(cpu.a)
}

// push processor status
func (cpu *Cpu) php(addrMode addressMode, pc uint16) {
	cpu.stackPush(cpu.status | unusedFlagMask | breakFlagMask)
}

// pull a
func (cpu *Cpu) pla(addrMode addressMode, pc uint16) {
	cpu.a = cpu.stackPop()

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// pull processor status
func (cpu *Cpu) plp(addrMode addressMode, pc uint16) {
	flags := cpu.stackPop()
	cpu.status = flags & 0xCF
	cpu.status |= unusedFlagMask
}

// rotate left and bitwise and
func (cpu *Cpu) rla(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.Read(address)

	carry := cpu.testFlag(carryFlagMask)
	cpu.updateFlag(carryFlagMask, value&0x80 > 0)

	value <<= 1
	if carry {
		value |= 0x01
	}

	cpu.sys.Write(address, value)

	cpu.a &= value

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// rotate left
func (cpu *Cpu) rol(addrMode addressMode, pc uint16) {
	var value uint8
	var address uint16
	if addrMode == addrModeAccumulator {
		value = cpu.a
	} else {
		address, _ = cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.Read(address)
	}

	carry := cpu.testFlag(carryFlagMask)
	cpu.updateFlag(carryFlagMask, value&0x80 > 0)
	value <<= 1

	if carry {
		value |= 0x01
	}

	cpu.updateFlag(zeroFlagMask, value == 0)
	cpu.updateFlag(negativeFlagMask, value&0x80 > 0)

	if addrMode == addrModeAccumulator {
		cpu.a = value
	} else {
		cpu.sys.Write(address, value)
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
		value = cpu.sys.Read(address)
	}

	carry := cpu.testFlag(carryFlagMask)
	cpu.updateFlag(carryFlagMask, value&0x01 > 0)

	value >>= 1
	if carry {
		value |= 0x80
	}

	cpu.updateFlag(zeroFlagMask, value == 0)
	cpu.updateFlag(negativeFlagMask, value&0x80 > 0)

	if addrMode == addrModeAccumulator {
		cpu.a = value
	} else {
		cpu.sys.Write(address, value)
	}
}

// rotate right and add with carry
func (cpu *Cpu) rra(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.Read(address)

	carry := cpu.testFlag(carryFlagMask)
	cpu.updateFlag(carryFlagMask, value&0x01 > 0)

	value >>= 1
	if carry {
		value |= 0x80
	}

	cpu.sys.Write(address, value)

	result := uint16(cpu.a) + uint16(value)
	if cpu.testFlag(carryFlagMask) {
		result++
	}

	cpu.updateFlag(carryFlagMask, result > 255)
	cpu.updateFlag(zeroFlagMask, result == 0)
	cpu.updateFlag(overflowFlagMask, (uint8(result)^cpu.a)&(uint8(result)^value)&0x80 > 0)
	cpu.updateFlag(negativeFlagMask, result&0x80 > 0)

	cpu.a = uint8(result)
}

// return from interrupt
func (cpu *Cpu) rti(addrMode addressMode, pc uint16) {
	flags := cpu.stackPop()
	low := cpu.stackPop()
	high := cpu.stackPop()
	cpu.status = flags & 0xCF
	cpu.status |= unusedFlagMask
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
	cpu.sys.Write(address, value)
}

// subtract with carry
func (cpu *Cpu) sbc(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.Read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.Read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	result := int16(cpu.a) - int16(value)
	if !cpu.testFlag(carryFlagMask) {
		result--
	}

	cpu.updateFlag(carryFlagMask, result >= 0)
	cpu.updateFlag(zeroFlagMask, result == 0)
	cpu.updateFlag(overflowFlagMask, (uint8(result)^cpu.a)&(uint8(result)^^value)&0x80 > 0)
	cpu.updateFlag(negativeFlagMask, result&0x80 > 0)

	cpu.a = uint8(result)
}

// set carry
func (cpu *Cpu) sec(addrMode addressMode, pc uint16) {
	cpu.updateFlag(carryFlagMask, true)
}

// set decimal
func (cpu *Cpu) sed(addrMode addressMode, pc uint16) {
	cpu.updateFlag(decimalFlagMask, true)
}

// set interrupt disable
func (cpu *Cpu) sei(addrMode addressMode, pc uint16) {
	cpu.updateFlag(intDisableFlagMask, true)
}

// arithmetic shift left and bitwise or
func (cpu *Cpu) slo(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.Read(address)

	cpu.updateFlag(carryFlagMask, value&0x80 > 0)
	value <<= 1

	cpu.sys.Write(address, value)
	cpu.a |= value

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// logical shift right and bitwise exclusive or
func (cpu *Cpu) sre(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.Read(address)

	cpu.updateFlag(carryFlagMask, value&0x01 > 0)
	value >>= 1
	cpu.sys.Write(address, value)

	cpu.a ^= value

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// store a
func (cpu *Cpu) sta(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	cpu.sys.Write(address, cpu.a)
}

// store x
func (cpu *Cpu) stx(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	cpu.sys.Write(address, cpu.x)
}

// store y
func (cpu *Cpu) sty(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	cpu.sys.Write(address, cpu.y)
}

// transfer a to x
func (cpu *Cpu) tax(addrMode addressMode, pc uint16) {
	cpu.x = cpu.a

	cpu.updateFlag(zeroFlagMask, cpu.x == 0)
	cpu.updateFlag(negativeFlagMask, cpu.x&0x80 > 0)
}

// transfer a to y
func (cpu *Cpu) tay(addrMode addressMode, pc uint16) {
	cpu.y = cpu.a

	cpu.updateFlag(zeroFlagMask, cpu.y == 0)
	cpu.updateFlag(negativeFlagMask, cpu.y&0x80 > 0)
}

// transfer stack pointer to x
func (cpu *Cpu) tsx(addrMode addressMode, pc uint16) {
	cpu.x = cpu.sp

	cpu.updateFlag(zeroFlagMask, cpu.x == 0)
	cpu.updateFlag(negativeFlagMask, cpu.x&0x80 > 0)
}

// transfer x to a
func (cpu *Cpu) txa(addrMode addressMode, pc uint16) {
	cpu.a = cpu.x

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// transfer x to stack pointer
func (cpu *Cpu) txs(addrMode addressMode, pc uint16) {
	cpu.sp = cpu.x
}

// transfer y to a
func (cpu *Cpu) tya(addrMode addressMode, pc uint16) {
	cpu.a = cpu.y

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}
