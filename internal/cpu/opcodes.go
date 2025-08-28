package cpu

type addressMode int

const (
	addrModeZeroPageX addressMode = iota
	addrModeZeroPageY
	addrModeAbsoluteX
	addrModeAbsoluteY
	addrModeIndexIndirX
	addrModeIndirIndexY
	addrModeImplied
	addrModeAccumulator
	addrModeImmediate
	addrModeZeroPage
	addrModeAbsolute
	addrModeRelative
	addrModeIndirect
)

type instruction struct {
	mnemonic string
	addrMode addressMode
	bytes    int
	cycles   int
	fn       func(*Cpu, *instruction)
}

var opcodes = [256]instruction{
	0x00: {"BRK", addrModeImplied, 2, 7, (*Cpu).forceBreak},
	0x01: {"ORA", addrModeIndexIndirX, 2, 6, (*Cpu).bitwiseOr},
	0x05: {"ORA", addrModeZeroPage, 2, 3, (*Cpu).bitwiseOr},
	0x06: {"ASL", addrModeZeroPage, 2, 5, (*Cpu).arithmeticShiftLeft},
	0x08: {"PHP", addrModeImplied, 1, 3, (*Cpu).pushProcessorStatus},
	0x09: {"ORA", addrModeImmediate, 2, 2, (*Cpu).bitwiseOr},
	0x0A: {"ASL", addrModeAccumulator, 1, 2, (*Cpu).arithmeticShiftLeft},
}
