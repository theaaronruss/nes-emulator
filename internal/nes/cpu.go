package nes

import (
	"fmt"
	"strings"
)

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

type cpu struct {
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

func NewCpu(sys *System) *cpu {
	pcLow := sys.read(resetVector)
	pcHigh := sys.read(resetVector + 1)

	return &cpu{
		sp:         initialStackPointer,
		pc:         uint16(pcHigh)<<8 | uint16(pcLow),
		status:     initialStatus,
		sys:        sys,
		cycleDelay: 7,
	}
}

func (cpu *cpu) Clock() {
	if cpu.cycleDelay <= 0 {
		if cpu.handleNmi {
			oldAddrLow := uint8(cpu.pc)
			oldAddrHigh := uint8(cpu.pc & 0xFF00 >> 8)
			cpu.stackPush(oldAddrHigh)
			cpu.stackPush(oldAddrLow)
			cpu.stackPush(cpu.status & ^breakFlagMask | unusedFlagMask)
			cpu.updateFlag(intDisableFlagMask, true)
			low := cpu.sys.read(nmiVector)
			high := cpu.sys.read(nmiVector + 1)
			address := uint16(high)<<8 | uint16(low)
			cpu.pc = address
			cpu.handleNmi = false
			cpu.cycleDelay += 7
		} else if cpu.handleIrq && cpu.status&intDisableFlagMask == 0 {
			oldAddrLow := uint8(cpu.pc)
			oldAddrHigh := uint8(cpu.pc & 0xFF00 >> 8)
			cpu.stackPush(oldAddrHigh)
			cpu.stackPush(oldAddrLow)
			cpu.stackPush(cpu.status & ^breakFlagMask | unusedFlagMask)
			cpu.updateFlag(intDisableFlagMask, true)
			newAddrLow := cpu.sys.read(irqVector)
			newAddrHigh := cpu.sys.read(irqVector + 1)
			address := uint16(newAddrHigh)<<8 | uint16(newAddrLow)
			cpu.pc = address
			cpu.handleIrq = false
			cpu.cycleDelay += 7
		}

		opcode := cpu.sys.read(cpu.pc)
		instruction := opcodes[opcode]
		currPc := cpu.pc
		cpu.pc += uint16(instruction.bytes)
		instruction.fn(cpu, instruction.addrMode, currPc)
		cpu.cycleDelay += instruction.cycles
	}
	cpu.cycleDelay--
	cpu.totalCycles++
}

func (cpu *cpu) logInstruction(pc uint16, instr *instruction) {
	var byteStr strings.Builder
	for i := range instr.bytes {
		nextByte := fmt.Sprintf("%02X ", cpu.sys.read(cpu.pc+uint16(i)))
		byteStr.WriteString(nextByte)
	}
	fmt.Printf("%04X  %-9s %-4s A:%02X X:%02X Y:%02X P:%02X SP:%02X\n", pc,
		byteStr.String(), instr.mnemonic, cpu.a, cpu.x, cpu.y, cpu.status, cpu.sp)
}

func (cpu *cpu) Irq() {
	cpu.handleIrq = true
}

func (cpu *cpu) Nmi() {
	cpu.handleNmi = true
}

func (cpu *cpu) updateFlag(flag uint8, value bool) {
	if value {
		cpu.status |= flag
	} else {
		cpu.status &= ^flag
	}
}

func (cpu *cpu) testFlag(flag uint8) bool {
	return cpu.status&flag > 0
}

func (cpu *cpu) stackPush(data uint8) {
	address := stackBase + uint16(cpu.sp)
	cpu.sys.write(address, data)
	cpu.sp--
}

func (cpu *cpu) stackPop() uint8 {
	cpu.sp++
	address := stackBase + uint16(cpu.sp)
	return cpu.sys.read(address)
}

func (cpu *cpu) mustGetAddress(addrMode addressMode, pc uint16) (uint16, bool) {
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

func (cpu *cpu) getZeroPageAddress(pc uint16) uint16 {
	address := cpu.sys.read(pc + 1)
	return uint16(address)
}

func (cpu *cpu) getZeroPageOffsetAddress(pc uint16, offset uint8) uint16 {
	zeroPageAddr := cpu.sys.read(pc + 1)
	zeroPageAddr += offset
	return uint16(zeroPageAddr)
}

func (cpu *cpu) getAbsoluteAddress(pc uint16) uint16 {
	low := cpu.sys.read(pc + 1)
	high := cpu.sys.read(pc + 2)
	return uint16(high)<<8 | uint16(low)
}

func (cpu *cpu) getAbsoluteOffsetAddress(pc uint16, offset uint8) (uint16, bool) {
	address := cpu.getAbsoluteAddress(pc)
	offsetAddress := address + uint16(offset)
	if address&0xFF00 != offsetAddress&0xFF00 {
		return offsetAddress, true
	} else {
		return offsetAddress, false
	}
}

func (cpu *cpu) getRelativeAddress(pc uint16) (uint16, bool) {
	offset := int8(cpu.sys.read(pc + 1))
	offsetAddress := int16(pc+2) + int16(offset)
	address := uint16(offsetAddress)
	if (pc+2)&0xFF00 != address&0xFF00 {
		return address, true
	} else {
		return address, false
	}
}

func (cpu *cpu) getIndirectAddress(pc uint16) uint16 {
	jumpAddress := cpu.getAbsoluteAddress(pc)
	low := cpu.sys.read(jumpAddress)

	// the 6502 has a bug where it wraps incorrectly if the jump address is at
	// a page boundary
	var high uint8
	if jumpAddress&0x00FF == 0x00FF {
		high = cpu.sys.read(jumpAddress & 0xFF00)
	} else {
		high = cpu.sys.read(jumpAddress + 1)
	}

	address := uint16(high)<<8 | uint16(low)
	return address
}

func (cpu *cpu) getIndexedIndirectAddress(pc uint16, offset uint8) uint16 {
	zeroPageAddr := cpu.sys.read(pc + 1)
	zeroPageAddr += offset
	low := cpu.sys.read(uint16(zeroPageAddr))
	high := cpu.sys.read(uint16(zeroPageAddr + 1))
	return uint16(high)<<8 | uint16(low)
}

func (cpu *cpu) getIndirectIndexedAddress(pc uint16, offset uint8) (uint16, bool) {
	zeroPageAddr := cpu.sys.read(pc + 1)
	low := cpu.sys.read(uint16(zeroPageAddr))
	high := cpu.sys.read(uint16(zeroPageAddr + 1))
	address := uint16(high)<<8 | uint16(low)
	offsetAddress := address + uint16(offset)
	if address&0xFF00 != offsetAddress&0xFF00 {
		return offsetAddress, true
	} else {
		return offsetAddress, false
	}
}

// add with carry
func (cpu *cpu) adc(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	result := uint16(cpu.a) + uint16(value)
	if cpu.testFlag(carryFlagMask) {
		result++
	}

	cpu.updateFlag(carryFlagMask, result > 255)
	cpu.updateFlag(zeroFlagMask, uint8(result) == 0)
	cpu.updateFlag(overflowFlagMask, (uint8(result)^cpu.a)&(uint8(result)^value)&0x80 > 0)
	cpu.updateFlag(negativeFlagMask, result&0x80 > 0)

	cpu.a = uint8(result)
}

// bitwise and
func (cpu *cpu) and(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.a &= value

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// arithmetic shift left
func (cpu *cpu) asl(addrMode addressMode, pc uint16) {
	var value uint8
	var address uint16
	if addrMode == addrModeAccumulator {
		value = cpu.a
	} else {
		address, _ = cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
	}

	cpu.updateFlag(carryFlagMask, value&0x80 > 0)
	value <<= 1

	cpu.updateFlag(zeroFlagMask, value == 0)
	cpu.updateFlag(negativeFlagMask, value&0x80 > 0)

	if addrMode == addrModeAccumulator {
		cpu.a = value
	} else {
		cpu.sys.write(address, value)
	}
}

// branch if carry set
func (cpu *cpu) bcs(addrMode addressMode, pc uint16) {
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
func (cpu *cpu) bcc(addrMode addressMode, pc uint16) {
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
func (cpu *cpu) bit(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.read(address)
	result := cpu.a & value

	cpu.updateFlag(zeroFlagMask, result == 0)
	cpu.updateFlag(overflowFlagMask, value&0x40 > 0)
	cpu.updateFlag(negativeFlagMask, value&0x80 > 0)
}

// branch if minus
func (cpu *cpu) bmi(addrMode addressMode, pc uint16) {
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
func (cpu *cpu) bne(addrMode addressMode, pc uint16) {
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
func (cpu *cpu) bpl(addrMode addressMode, pc uint16) {
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
func (cpu *cpu) brk(addrMode addressMode, pc uint16) {
	pc += 2
	oldPcLow := uint8(pc & 0x00FF)
	oldPcHigh := uint8(pc & 0xFF00 >> 8)
	cpu.stackPush(oldPcHigh)
	cpu.stackPush(oldPcLow)
	cpu.stackPush(cpu.status | unusedFlagMask | breakFlagMask)
	newPcLow := cpu.sys.read(irqVector)
	newPcHigh := cpu.sys.read(irqVector + 1)
	cpu.pc = uint16(newPcHigh)<<8 | uint16(newPcLow)
}

// branch if overflow clear
func (cpu *cpu) bvc(addrMode addressMode, pc uint16) {
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
func (cpu *cpu) bvs(addrMode addressMode, pc uint16) {
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
func (cpu *cpu) beq(addrMode addressMode, pc uint16) {
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
func (cpu *cpu) clc(addrMode addressMode, pc uint16) {
	cpu.updateFlag(carryFlagMask, false)
}

// clear decimal
func (cpu *cpu) cld(addrMode addressMode, pc uint16) {
	cpu.updateFlag(decimalFlagMask, false)
}

// clear interrupt disable
func (cpu *cpu) cli(addrMode addressMode, pc uint16) {
	cpu.updateFlag(intDisableFlagMask, false)
}

// clear overflow
func (cpu *cpu) clv(addrMode addressMode, pc uint16) {
	cpu.updateFlag(overflowFlagMask, false)
}

// compare a
func (cpu *cpu) cmp(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
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
func (cpu *cpu) cpx(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.read(pc + 1)
	} else {
		address, _ := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
	}

	result := cpu.x - value

	cpu.updateFlag(carryFlagMask, cpu.x >= value)
	cpu.updateFlag(zeroFlagMask, cpu.x == value)
	cpu.updateFlag(negativeFlagMask, result&0x80 > 0)
}

// compare y
func (cpu *cpu) cpy(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.read(pc + 1)
	} else {
		address, _ := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
	}

	result := cpu.y - value

	cpu.updateFlag(carryFlagMask, cpu.y >= value)
	cpu.updateFlag(zeroFlagMask, cpu.y == value)
	cpu.updateFlag(negativeFlagMask, result&0x80 > 0)
}

// decrement memory and compare a
func (cpu *cpu) dcp(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.read(address)
	value--
	cpu.sys.write(address, value)

	result := cpu.a - value

	cpu.updateFlag(carryFlagMask, cpu.a >= value)
	cpu.updateFlag(zeroFlagMask, cpu.a == value)
	cpu.updateFlag(negativeFlagMask, result&0x80 > 0)
}

// decrement memory
func (cpu *cpu) dec(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.read(address)
	value--
	cpu.sys.write(address, value)

	cpu.updateFlag(zeroFlagMask, value == 0)
	cpu.updateFlag(negativeFlagMask, value&0x80 > 0)
}

// decrement x
func (cpu *cpu) dex(addrMode addressMode, pc uint16) {
	cpu.x--

	cpu.updateFlag(zeroFlagMask, cpu.x == 0)
	cpu.updateFlag(negativeFlagMask, cpu.x&0x80 > 0)
}

// decrement y
func (cpu *cpu) dey(addrMode addressMode, pc uint16) {
	cpu.y--

	cpu.updateFlag(zeroFlagMask, cpu.y == 0)
	cpu.updateFlag(negativeFlagMask, cpu.y&0x80 > 0)
}

// bitwise exclusive or
func (cpu *cpu) eor(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.a ^= value

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// increment memory
func (cpu *cpu) inc(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.read(address)
	value++
	cpu.sys.write(address, value)

	cpu.updateFlag(zeroFlagMask, value == 0)
	cpu.updateFlag(negativeFlagMask, value&0x80 > 0)
}

// increment x
func (cpu *cpu) inx(addrMode addressMode, pc uint16) {
	cpu.x++

	cpu.updateFlag(zeroFlagMask, cpu.x == 0)
	cpu.updateFlag(negativeFlagMask, cpu.x&0x80 > 0)
}

// increment y
func (cpu *cpu) iny(addrMode addressMode, pc uint16) {
	cpu.y++

	cpu.updateFlag(zeroFlagMask, cpu.y == 0)
	cpu.updateFlag(negativeFlagMask, cpu.y&0x80 > 0)
}

// increment memory and subtract with carry
func (cpu *cpu) isb(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.read(address)
	value++
	cpu.sys.write(address, value)

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
func (cpu *cpu) jmp(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	cpu.pc = address
}

// jump to subroutine
func (cpu *cpu) jsr(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	pc += 2
	low := uint8(pc & 0x00FF)
	high := uint8((pc & 0xFF00) >> 8)
	cpu.stackPush(high)
	cpu.stackPush(low)
	cpu.pc = address
}

// load a and load x
func (cpu *cpu) lax(addrMode addressMode, pc uint16) {
	address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
	if pageCrossed {
		cpu.cycleDelay++
	}
	value := cpu.sys.read(address)

	cpu.a = value
	cpu.x = value

	cpu.updateFlag(zeroFlagMask, value == 0)
	cpu.updateFlag(negativeFlagMask, value&0x80 > 0)
}

// load a
func (cpu *cpu) lda(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.a = value

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// load x
func (cpu *cpu) ldx(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.x = value

	cpu.updateFlag(zeroFlagMask, cpu.x == 0)
	cpu.updateFlag(negativeFlagMask, cpu.x&0x80 > 0)
}

// load y
func (cpu *cpu) ldy(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.y = value

	cpu.updateFlag(zeroFlagMask, cpu.y == 0)
	cpu.updateFlag(negativeFlagMask, cpu.y&0x80 > 0)
}

// logical shift right
func (cpu *cpu) lsr(addrMode addressMode, pc uint16) {
	var value uint8
	var address uint16
	if addrMode == addrModeAccumulator {
		value = cpu.a
	} else {
		address, _ = cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
	}

	cpu.updateFlag(carryFlagMask, value&0x01 > 0)
	value >>= 1

	cpu.updateFlag(zeroFlagMask, value == 0)
	cpu.updateFlag(negativeFlagMask, false)
	if addrMode == addrModeAccumulator {
		cpu.a = value
	} else {
		cpu.sys.write(address, value)
	}
}

// no operation
func (cpu *cpu) nop(addrMode addressMode, pc uint16) {
	_, pageCrossed := cpu.mustGetAddress(addrMode, pc)
	if pageCrossed {
		cpu.cycleDelay++
	}
	// do nothing
}

// bitwise or
func (cpu *cpu) ora(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
		if pageCrossed {
			cpu.cycleDelay++
		}
	}

	cpu.a |= value

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// push a
func (cpu *cpu) pha(addrMode addressMode, pc uint16) {
	cpu.stackPush(cpu.a)
}

// push processor status
func (cpu *cpu) php(addrMode addressMode, pc uint16) {
	cpu.stackPush(cpu.status | unusedFlagMask | breakFlagMask)
}

// pull a
func (cpu *cpu) pla(addrMode addressMode, pc uint16) {
	cpu.a = cpu.stackPop()

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// pull processor status
func (cpu *cpu) plp(addrMode addressMode, pc uint16) {
	flags := cpu.stackPop()
	cpu.status = flags & 0xCF
	cpu.status |= unusedFlagMask
}

// rotate left and bitwise and
func (cpu *cpu) rla(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.read(address)

	carry := cpu.testFlag(carryFlagMask)
	cpu.updateFlag(carryFlagMask, value&0x80 > 0)

	value <<= 1
	if carry {
		value |= 0x01
	}

	cpu.sys.write(address, value)

	cpu.a &= value

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// rotate left
func (cpu *cpu) rol(addrMode addressMode, pc uint16) {
	var value uint8
	var address uint16
	if addrMode == addrModeAccumulator {
		value = cpu.a
	} else {
		address, _ = cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
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
		cpu.sys.write(address, value)
	}
}

// rotate right
func (cpu *cpu) ror(addrMode addressMode, pc uint16) {
	var value uint8
	var address uint16
	if addrMode == addrModeAccumulator {
		value = cpu.a
	} else {
		address, _ = cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
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
		cpu.sys.write(address, value)
	}
}

// rotate right and add with carry
func (cpu *cpu) rra(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.read(address)

	carry := cpu.testFlag(carryFlagMask)
	cpu.updateFlag(carryFlagMask, value&0x01 > 0)

	value >>= 1
	if carry {
		value |= 0x80
	}

	cpu.sys.write(address, value)

	result := uint16(cpu.a) + uint16(value)
	if cpu.testFlag(carryFlagMask) {
		result++
	}

	cpu.updateFlag(carryFlagMask, result > 255)
	cpu.updateFlag(zeroFlagMask, uint8(result) == 0)
	cpu.updateFlag(overflowFlagMask, (uint8(result)^cpu.a)&(uint8(result)^value)&0x80 > 0)
	cpu.updateFlag(negativeFlagMask, result&0x80 > 0)

	cpu.a = uint8(result)
}

// return from interrupt
func (cpu *cpu) rti(addrMode addressMode, pc uint16) {
	flags := cpu.stackPop()
	low := cpu.stackPop()
	high := cpu.stackPop()
	cpu.status = flags & 0xCF
	cpu.status |= unusedFlagMask
	cpu.pc = uint16(high)<<8 | uint16(low)
}

// return from subroutine
func (cpu *cpu) rts(addrMode addressMode, pc uint16) {
	low := cpu.stackPop()
	high := cpu.stackPop()
	address := uint16(high)<<8 | uint16(low)
	cpu.pc = address + 1
}

// store a and x
func (cpu *cpu) sax(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.a & cpu.x
	cpu.sys.write(address, value)
}

// subtract with carry
func (cpu *cpu) sbc(addrMode addressMode, pc uint16) {
	var value uint8
	if addrMode == addrModeImmediate {
		value = cpu.sys.read(pc + 1)
	} else {
		address, pageCrossed := cpu.mustGetAddress(addrMode, pc)
		value = cpu.sys.read(address)
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
func (cpu *cpu) sec(addrMode addressMode, pc uint16) {
	cpu.updateFlag(carryFlagMask, true)
}

// set decimal
func (cpu *cpu) sed(addrMode addressMode, pc uint16) {
	cpu.updateFlag(decimalFlagMask, true)
}

// set interrupt disable
func (cpu *cpu) sei(addrMode addressMode, pc uint16) {
	cpu.updateFlag(intDisableFlagMask, true)
}

// arithmetic shift left and bitwise or
func (cpu *cpu) slo(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.read(address)

	cpu.updateFlag(carryFlagMask, value&0x80 > 0)
	value <<= 1

	cpu.sys.write(address, value)
	cpu.a |= value

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// logical shift right and bitwise exclusive or
func (cpu *cpu) sre(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	value := cpu.sys.read(address)

	cpu.updateFlag(carryFlagMask, value&0x01 > 0)
	value >>= 1
	cpu.sys.write(address, value)

	cpu.a ^= value

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// store a
func (cpu *cpu) sta(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	cpu.sys.write(address, cpu.a)
}

// store x
func (cpu *cpu) stx(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	cpu.sys.write(address, cpu.x)
}

// store y
func (cpu *cpu) sty(addrMode addressMode, pc uint16) {
	address, _ := cpu.mustGetAddress(addrMode, pc)
	cpu.sys.write(address, cpu.y)
}

// transfer a to x
func (cpu *cpu) tax(addrMode addressMode, pc uint16) {
	cpu.x = cpu.a

	cpu.updateFlag(zeroFlagMask, cpu.x == 0)
	cpu.updateFlag(negativeFlagMask, cpu.x&0x80 > 0)
}

// transfer a to y
func (cpu *cpu) tay(addrMode addressMode, pc uint16) {
	cpu.y = cpu.a

	cpu.updateFlag(zeroFlagMask, cpu.y == 0)
	cpu.updateFlag(negativeFlagMask, cpu.y&0x80 > 0)
}

// transfer stack pointer to x
func (cpu *cpu) tsx(addrMode addressMode, pc uint16) {
	cpu.x = cpu.sp

	cpu.updateFlag(zeroFlagMask, cpu.x == 0)
	cpu.updateFlag(negativeFlagMask, cpu.x&0x80 > 0)
}

// transfer x to a
func (cpu *cpu) txa(addrMode addressMode, pc uint16) {
	cpu.a = cpu.x

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}

// transfer x to stack pointer
func (cpu *cpu) txs(addrMode addressMode, pc uint16) {
	cpu.sp = cpu.x
}

// transfer y to a
func (cpu *cpu) tya(addrMode addressMode, pc uint16) {
	cpu.a = cpu.y

	cpu.updateFlag(zeroFlagMask, cpu.a == 0)
	cpu.updateFlag(negativeFlagMask, cpu.a&0x80 > 0)
}
