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
	0x0D: {"ORA", addrModeAbsolute, 3, 4, (*Cpu).bitwiseOr},
	0x0E: {"ASL", addrModeAbsolute, 3, 6, (*Cpu).arithmeticShiftLeft},
	0x10: {"BPL", addrModeRelative, 2, 0, (*Cpu).branchIfPlus},
	0x11: {"ORA", addrModeIndirIndexY, 2, 0, (*Cpu).bitwiseOr},
	0x15: {"ORA", addrModeZeroPageX, 2, 4, (*Cpu).bitwiseOr},
	0x16: {"ASL", addrModeZeroPageX, 2, 6, (*Cpu).arithmeticShiftLeft},
	0x18: {"CLC", addrModeImplied, 1, 2, (*Cpu).clearCarry},
	0x19: {"ORA", addrModeAbsoluteY, 3, 0, (*Cpu).bitwiseOr},
	0x1D: {"ORA", addrModeAbsoluteX, 3, 0, (*Cpu).bitwiseOr},
	0x1E: {"ASL", addrModeAbsoluteX, 3, 7, (*Cpu).arithmeticShiftLeft},
}
