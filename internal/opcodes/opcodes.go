package opcodes

type AddressMode int

const (
	AddrModeZeroPageX AddressMode = iota
	AddrModeZeroPageY
	AddrModeAbsoluteX
	AddrModeAbsoluteY
	AddrModeIndirectX
	AddrModeIndirectY
	AddrModeImplicit
	AddrModeAccumulator
	AddrModeImmediate
	AddrModeZeroPage
	AddrModeAbsolute
	AddrModeRelative
	AddrModeIndirect
)

type operation int

const (
	OpADC operation = iota
	OpAND
	OpASL
	OpBCC
	OpBCS
	OpBEQ
	OpBIT
	OpBMI
	OpBNE
	OpBPL
	OpBRK
	OpBVC
	OpBVS
	OpCLC
	OpCLD
	OpCLI
	OpCLV
	OpCMP
	OpCPX
	OpCPY
	OpDEC
	OpDEX
	OpDEY
	OpEOR
	OpINC
	OpINX
	OpINY
	OpJMP
	OpJSR
	OpLDA
	OpLDX
	OpLDY
	OpLSR
	OpNOP
	OpORA
	OpPHA
	OpPHP
	OpPLA
	OpPLP
	OpROL
	OpROR
	OpRTI
	OpRTS
	OpSBC
	OpSEC
	OpSED
	OpSEI
	OpSTA
	OpSTX
	OpSTY
	OpTAX
	OpTAY
	OpTSX
	OpTXA
	OpTXY
)

type OpcodeInfo struct {
	Op       operation
	AddrMode AddressMode
}

var Ops [0xFF]*OpcodeInfo

func init() {
	Ops[0x00] = &OpcodeInfo{Op: OpBRK, AddrMode: AddrModeImplicit}
}
