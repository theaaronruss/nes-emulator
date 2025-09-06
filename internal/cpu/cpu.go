package cpu

import "github.com/theaaronruss/nes-emulator/internal/sysbus"

// interrupt vectors
const (
	nmiVector   uint16 = 0xFFFA
	resetVector uint16 = 0xFFFC
	irqVector   uint16 = 0xFFFE
)

// status flag masks
const (
	flagCarry      uint8 = 1 << iota
	flagZero       uint8 = 1 << iota
	flagIntDisable uint8 = 1 << iota
	flagDecimal    uint8 = 1 << iota
	flagBreak      uint8 = 1 << iota
	flagUnused     uint8 = 1 << iota
	flagOverflow   uint8 = 1 << iota
	flagNegative   uint8 = 1 << iota
)

const (
	stackBase           uint16 = 0x0100
	initialStackPointer uint8  = 0xFD
)

// registers
var (
	a      uint8
	x      uint8
	y      uint8
	sp     uint8
	pc     uint16
	status uint8
)

var cycleDelay int

func Reset() {
	sp = initialStackPointer
	pcLow := sysbus.Read(resetVector)
	pcHigh := sysbus.Read(resetVector + 1)
	pc = uint16(pcHigh)<<8 | uint16(pcLow)
	setFlag(flagIntDisable)
	cycleDelay = 0
}

func Clock() {
	if cycleDelay <= 0 {
		opcode := sysbus.Read(pc)
		instruction := opcodes[opcode]
		instruction.fn(&instruction)
		cycleDelay += instruction.cycles
		return
	}
	cycleDelay--
}

func setFlag(flag uint8) {
	status |= flag
}

func clearFlag(flag uint8) {
	status &= ^flag
}

func testFlag(flag uint8) bool {
	return status&flag > 0
}

func stackPush(data uint8) {
	address := stackBase + uint16(sp)
	sysbus.Write(address, data)
	sp--
}

func stackPop() uint8 {
	sp++
	address := stackBase + uint16(sp)
	return sysbus.Read(address)
}

func getAddress(addrMode addressMode) (uint16, bool) {
	switch addrMode {
	case addrModeZeroPageX:
		address := sysbus.Read(pc + 1)
		return uint16(address + x), false
	case addrModeZeroPageY:
		address := sysbus.Read(pc + 1)
		return uint16(address + y), false
	case addrModeAbsoluteX:
		baseAddress, _ := getAddress(addrModeAbsolute)
		address := baseAddress + uint16(x)
		if baseAddress&0xFF00 == address&0xFF00 {
			return address, false
		} else {
			return address, true
		}
	case addrModeAbsoluteY:
		baseAddress, _ := getAddress(addrModeAbsolute)
		address := baseAddress + uint16(y)
		if baseAddress&0xFF00 == address&0xFF00 {
			return address, false
		} else {
			return address, true
		}
	case addrModeIndexIndirX:
		zeroPageAddr := sysbus.Read(pc + 1)
		zeroPageAddr += x
		low := sysbus.Read(uint16(zeroPageAddr))
		high := sysbus.Read(uint16(zeroPageAddr) + 1)
		return uint16(high)<<8 | uint16(low), false
	case addrModeIndirIndexY:
		zeroPageAddr := sysbus.Read(pc + 1)
		low := sysbus.Read(uint16(zeroPageAddr))
		high := sysbus.Read(uint16(zeroPageAddr) + 1)
		baseAddress := uint16(high)<<8 | uint16(low)
		address := baseAddress + uint16(y)
		if baseAddress&0xFF00 == address&0xFF00 {
			return address, false
		} else {
			return address, true
		}
	case addrModeZeroPage:
		return uint16(sysbus.Read(pc + 1)), false
	case addrModeAbsolute:
		low := sysbus.Read(pc + 1)
		high := sysbus.Read(pc + 2)
		return uint16(high)<<8 | uint16(low), false
	case addrModeRelative:
		offset := int8(sysbus.Read(pc))
		return pc + uint16(offset), false
	case addrModeIndirect:
		address, _ := getAddress(addrModeAbsolute)
		low := sysbus.Read(address)
		high := sysbus.Read(address + 1)
		return uint16(high)<<8 | uint16(low), false
	}
	return 0x0000, false
}

func forceBreak(instr *instruction) {
	pc += 2
	oldPcLow := uint8(pc & 0x00FF)
	oldPcHigh := uint8(pc & 0xFF00 >> 8)
	stackPush(oldPcHigh)
	stackPush(oldPcLow)
	stackPush(status | flagUnused | flagBreak)
	newPcLow := sysbus.Read(irqVector)
	newPcHigh := sysbus.Read(irqVector + 1)
	newPc := uint16(newPcHigh)<<8 | uint16(newPcLow)
	pc = newPc
}

func returnFromInterrupt(instr *instruction) {
	flags := stackPop()
	low := stackPop()
	high := stackPop()
	status = flags & 0b11001111
	pc = uint16(high)<<8 | uint16(low)
}

func bitTest(instr *instruction) {
	address, _ := getAddress(instr.addrMode)
	value := sysbus.Read(address)
	result := a & value
	if result == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}
	if value&0x40 > 0 {
		setFlag(flagOverflow)
	} else {
		clearFlag(flagOverflow)
	}
	if value&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}
}

func bitwiseOr(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(pc + 1)
	} else {
		address, pageCrossed := getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			cycleDelay++
		}
	}
	a |= value
	if a == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}
	if a&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}
	pc += uint16(instr.bytes)
}

func bitwiseXor(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(pc + 1)
	} else {
		address, pageCrossed := getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			cycleDelay++
		}
	}
	a ^= value
	if a == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}
	if a&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}
	pc += uint16(instr.bytes)
}

func bitwiseAnd(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(pc + 1)
	} else {
		address, pageCrossed := getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			cycleDelay++
		}
	}
	a &= value
	if a == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}
	if a&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}
	pc += uint16(instr.bytes)
}

func arithmeticShiftLeft(instr *instruction) {
	var value uint8
	var address uint16
	if instr.addrMode == addrModeAccumulator {
		value = a
	} else {
		address, _ = getAddress(instr.addrMode)
		value = sysbus.Read(address)
	}
	if value&0x80 > 0 {
		setFlag(flagCarry)
	} else {
		clearFlag(flagCarry)
	}
	value <<= 1
	if value == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}
	if value&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}
	if instr.addrMode == addrModeAccumulator {
		a = value
	} else {
		sysbus.Write(address, value)
	}
	pc += uint16(instr.bytes)
}

func logicalShiftRight(instr *instruction) {
	var value uint8
	var address uint16
	if instr.addrMode == addrModeAccumulator {
		value = a
	} else {
		address, _ = getAddress(instr.addrMode)
		value = sysbus.Read(address)
	}
	if value&0x01 > 0 {
		setFlag(flagCarry)
	} else {
		clearFlag(flagCarry)
	}
	value >>= 1
	if value == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}
	clearFlag(flagNegative)
	if instr.addrMode == addrModeAccumulator {
		a = value
	} else {
		sysbus.Write(address, value)
	}
	pc += uint16(instr.bytes)
}

func rotateLeft(instr *instruction) {
	var value uint8
	var address uint16
	if instr.addrMode == addrModeAccumulator {
		value = a
	} else {
		address, _ = getAddress(instr.addrMode)
		value = sysbus.Read(address)
	}

	if value&0x80 > 0 {
		setFlag(flagCarry)
	} else {
		clearFlag(flagCarry)
	}

	value <<= 1

	if testFlag(flagCarry) {
		value |= 0x01
	}

	if value == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if value&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	if instr.addrMode == addrModeAccumulator {
		a = value
	} else {
		sysbus.Write(address, value)
	}

	pc += uint16(instr.bytes)
}

func rotateRight(instr *instruction) {
	var value uint8
	var address uint16
	if instr.addrMode == addrModeAccumulator {
		value = a
	} else {
		address, _ = getAddress(instr.addrMode)
		value = sysbus.Read(address)
	}

	if value&0x01 > 0 {
		setFlag(flagCarry)
	} else {
		clearFlag(flagCarry)
	}

	value >>= 1

	if testFlag(flagCarry) {
		value |= 0x80
	}

	if value == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if value&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	if instr.addrMode == addrModeAccumulator {
		a = value
	} else {
		sysbus.Write(address, value)
	}

	pc += uint16(instr.bytes)
}

func pushProcessorStatus(instr *instruction) {
	stackPush(status | flagUnused | flagBreak)
	pc += uint16(instr.bytes)
}

func pullProcessorStatus(instr *instruction) {
	flags := stackPop()
	status = flags & 0b11001111
	pc += uint16(instr.bytes)
}

func pushA(instr *instruction) {
	stackPush(a)
	pc += uint16(instr.bytes)
}

func pullA(instr *instruction) {
	a = stackPop()

	if a == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if a&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func storeA(instr *instruction) {
	address, _ := getAddress(instr.addrMode)
	sysbus.Write(address, a)
	pc += uint16(instr.bytes)
}

func loadA(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(pc + 1)
	} else {
		address, pageCrossed := getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			cycleDelay++
		}
	}
	a = value

	if a == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if a&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func storeX(instr *instruction) {
	address, _ := getAddress(instr.addrMode)
	sysbus.Write(address, x)
	pc += uint16(instr.bytes)
}

func loadX(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(pc + 1)
	} else {
		address, pageCrossed := getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			cycleDelay++
		}
	}
	x = value

	if x == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if x&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func storeY(instr *instruction) {
	address, _ := getAddress(instr.addrMode)
	sysbus.Write(address, y)
	pc += uint16(instr.bytes)
}

func loadY(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(pc + 1)
	} else {
		address, pageCrossed := getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			cycleDelay++
		}
	}
	y = value

	if y == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if y&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func transferAToX(instr *instruction) {
	x = a

	if x == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if x&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func transferAToY(instr *instruction) {
	y = a

	if y == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if y&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func transferXToA(instr *instruction) {
	a = x

	if a == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if a&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func transferYToA(instr *instruction) {
	a = x

	if a == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if a&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func transferXToStackPointer(instr *instruction) {
	sp = x
	pc += uint16(instr.bytes)
}

func transferStackPointerToX(instr *instruction) {
	x = sp
	pc += uint16(instr.bytes)
}

func compareA(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(pc + 1)
	} else {
		address, pageCrossed := getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			cycleDelay++
		}
	}

	if a >= value {
		setFlag(flagCarry)
	} else {
		clearFlag(flagCarry)
	}

	if a == value {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	result := a - value

	if result&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func compareX(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(pc + 1)
	} else {
		address, _ := getAddress(instr.addrMode)
		value = sysbus.Read(address)
	}

	if x >= value {
		setFlag(flagCarry)
	} else {
		clearFlag(flagCarry)
	}

	if x == value {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	result := x - value

	if result&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func compareY(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(pc + 1)
	} else {
		address, _ := getAddress(instr.addrMode)
		value = sysbus.Read(address)
	}

	if y >= value {
		setFlag(flagCarry)
	} else {
		clearFlag(flagCarry)
	}

	if y == value {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	result := y - value

	if result&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func incrementY(instr *instruction) {
	y++

	if y == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if y&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func decrementY(instr *instruction) {
	y--

	if y == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if y&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func incrementX(instr *instruction) {
	x++

	if x == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if x&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func decrementX(instr *instruction) {
	x--

	if x == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if x&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func incrementMemory(instr *instruction) {
	address, _ := getAddress(instr.addrMode)
	value := sysbus.Read(address)
	value++
	sysbus.Write(address, value)

	if value == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if value&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func decrementMemory(instr *instruction) {
	address, _ := getAddress(instr.addrMode)
	value := sysbus.Read(address)
	value--
	sysbus.Write(address, value)

	if value == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if value&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	pc += uint16(instr.bytes)
}

func jump(instr *instruction) {
	address, _ := getAddress(instr.addrMode)
	pc = address
}

func jumpToSubroutine(instr *instruction) {
	address, _ := getAddress(instr.addrMode)
	pc += 2
	low := uint8(pc & 0x00FF)
	high := uint8((pc & 0xFF00) >> 8)
	stackPush(high)
	stackPush(low)
	pc = address
}

func returnFromSubroutine(instr *instruction) {
	low := stackPop()
	high := stackPop()
	address := uint16(high)<<8 | uint16(low)
	pc = address + 1
}

func branchIfPlus(instr *instruction) {
	if testFlag(flagNegative) {
		pc += uint16(instr.bytes)
		return
	}
	pc++
	address, pageCrossed := getAddress(instr.addrMode)
	pc = address + 1

	cycleDelay++
	if pageCrossed {
		cycleDelay++
	}
}

func branchIfMinus(instr *instruction) {
	if !testFlag(flagNegative) {
		pc += uint16(instr.bytes)
		return
	}
	pc++
	address, pageCrossed := getAddress(instr.addrMode)
	pc = address + 1

	cycleDelay++
	if pageCrossed {
		cycleDelay++
	}
}

func branchIfEqual(instr *instruction) {
	if !testFlag(flagZero) {
		pc += uint16(instr.bytes)
		return
	}
	pc++
	address, pageCrossed := getAddress(instr.addrMode)
	pc = address + 1

	cycleDelay++
	if pageCrossed {
		cycleDelay++
	}
}

func branchIfNotEqual(instr *instruction) {
	if testFlag(flagZero) {
		pc += uint16(instr.bytes)
		return
	}
	pc++
	address, pageCrossed := getAddress(instr.addrMode)
	pc = address + 1

	cycleDelay++
	if pageCrossed {
		cycleDelay++
	}
}

func branchIfCarrySet(instr *instruction) {
	if !testFlag(flagCarry) {
		pc += uint16(instr.bytes)
		return
	}
	pc++
	address, pageCrossed := getAddress(instr.addrMode)
	pc = address + 1

	cycleDelay++
	if pageCrossed {
		cycleDelay++
	}
}

func branchIfCarryClear(instr *instruction) {
	if testFlag(flagCarry) {
		pc += uint16(instr.bytes)
		return
	}
	pc++
	address, pageCrossed := getAddress(instr.addrMode)
	pc = address + 1

	cycleDelay++
	if pageCrossed {
		cycleDelay++
	}
}

func branchIfOverflowSet(instr *instruction) {
	if !testFlag(flagOverflow) {
		pc += uint16(instr.bytes)
		return
	}
	pc++
	address, pageCrossed := getAddress(instr.addrMode)
	pc = address + 1

	cycleDelay++
	if pageCrossed {
		cycleDelay++
	}
}

func branchIfOverflowClear(instr *instruction) {
	if testFlag(flagOverflow) {
		pc += uint16(instr.bytes)
		return
	}
	pc++
	address, pageCrossed := getAddress(instr.addrMode)
	pc = address + 1

	cycleDelay++
	if pageCrossed {
		cycleDelay++
	}
}

func setCarry(instr *instruction) {
	setFlag(flagCarry)
	pc += uint16(instr.bytes)
}

func clearCarry(instr *instruction) {
	clearFlag(flagCarry)
	pc += uint16(instr.bytes)
}

func clearOverflow(instr *instruction) {
	clearFlag(flagOverflow)
	pc += uint16(instr.bytes)
}

func setDecimal(instr *instruction) {
	setFlag(flagDecimal)
	pc += uint16(instr.bytes)
}

func clearDecimal(instr *instruction) {
	clearFlag(flagDecimal)
	pc += uint16(instr.bytes)
}

func setInterruptDisable(instr *instruction) {
	setFlag(flagIntDisable)
	pc += uint16(instr.bytes)
}

func clearInterruptDisable(instr *instruction) {
	clearFlag(flagIntDisable)
	pc += uint16(instr.bytes)
}

func addWithCarry(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(pc + 1)
	} else {
		address, pageCrossed := getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			cycleDelay++
		}
	}

	result := uint16(a) + uint16(value)
	if testFlag(flagCarry) {
		result++
	}

	if result > 255 {
		setFlag(flagCarry)
	} else {
		clearFlag(flagCarry)
	}

	if result == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if (uint8(result)^a)&(uint8(result)^value)&0x80 > 0 {
		setFlag(flagOverflow)
	} else {
		clearFlag(flagOverflow)
	}

	if result&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	a = uint8(result)
	pc += uint16(instr.bytes)
}

func subtractWithCarry(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(pc + 1)
	} else {
		address, pageCrossed := getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			cycleDelay++
		}
	}

	result := int16(a) - int16(value)
	if !testFlag(flagCarry) {
		result--
	}

	if result >= 0 {
		setFlag(flagCarry)
	} else {
		clearFlag(flagCarry)
	}

	if result == 0 {
		setFlag(flagZero)
	} else {
		clearFlag(flagZero)
	}

	if (uint8(result)^a)&(uint8(result)^^value)&0x80 > 0 {
		setFlag(flagOverflow)
	} else {
		clearFlag(flagOverflow)
	}

	if result&0x80 > 0 {
		setFlag(flagNegative)
	} else {
		clearFlag(flagNegative)
	}

	a = uint8(result)
	pc += uint16(instr.bytes)
}

func noOperation(instr *instruction) {
	pc += uint16(instr.bytes)
}
