package cpu

type Cpu struct {
	a			uint8	// accumulator
	x			uint8	// x index
	y			uint8	// y index
	pc			uint16	// program counter
	s			uint8	// stack pointer
	carry		bool	// carry flag
	zero		bool	// zero flag
	intDisable	bool	// interrupt disable flag
	decimal		bool	// decimal flag (decimal mode disabled on NES)
	overflow	bool	// overflow flag
	negative	bool	// negative flag
	cycleDelay	int
}

func NewCpu() *Cpu {
	return &Cpu{
		a: 0,
		x: 0,
		y: 0,
		pc: 0xFFFC,
		s: 0xFD,
		carry: false,
		zero: false,
		intDisable: true,
		decimal: false,
		overflow: false,
		negative: false,
		cycleDelay: 0,
	}
}

func (cpu *Cpu) Execute(instruction int32) {
	if cpu.cycleDelay > 0 {
		cpu.cycleDelay--
		return
	}
	byte1 := instruction & 0xFF0000 >> 16
	switch byte1 {
	case 0x61, 0x72, 0x65, 0x75, 0x69, 0x79, 0x6D, 0x7D:
		cpu.adc()
	case 0x21, 0x31, 0x25, 0x35, 0x29, 0x39, 0x2D, 0x3D:
		cpu.and()
	case 0x06, 0x16, 0x0A, 0x0E, 0x1E:
		cpu.asl()
	// TODO: implement other instructions
	}
}

// add with carry
func (cpu *Cpu) adc() {
}

// bitwise and
func (cpu *Cpu) and() {
}

// arithmetic shift left
func (cpu *Cpu) asl() {
}

// branch if carry clear
func (cpu *Cpu) bcc() {
}

// branch if carry set
func (cpu *Cpu) bcs() {
}

// branch if equal
func (cpu *Cpu) beq() {
}

// bit test
func (cpu *Cpu) bit() {
}

// branch if minus
func (cpu *Cpu) bmi() {
}

// branch if not equal
func (cpu *Cpu) bne() {
}

// branch if plus
func (cpu *Cpu) bpl() {
}

// break (software IRQ)
func (cpu *Cpu) brk() {
}

// branch if overflow clear
func (cpu *Cpu) bvc() {
}

// branch if overflow set
func (cpu *Cpu) bvs() {
}

// clear carry
func (cpu *Cpu) clc() {
}

// clear decimal
func (cpu *Cpu) cld() {
}

// clear interrupt disable
func (cpu *Cpu) cli() {
}

// clear overflow
func (cpu *Cpu) clv() {
}

// compare a
func (cpu *Cpu) cmp() {
}

// compare x
func (cpu *Cpu) cpx() {
}

// compare y
func (cpu *Cpu) cpy() {
}

// decrement memory
func (cpu *Cpu) dec() {
}

// decrement x
func (cpu *Cpu) dex() {
}

// decrement y
func (cpu *Cpu) dey() {
}

// bitwise exclusive or
func (cpu *Cpu) eor() {
}

// increment memory
func (cpu *Cpu) inc() {
}

// increment x
func (cpu *Cpu) inx() {
}

// increment y
func (cpu *Cpu) iny() {
}

// jump
func (cpu *Cpu) jmp() {
}

// jump to subroutine
func (cpu *Cpu) jsr() {
}

// load a
func (cpu *Cpu) lda() {
}

// load x
func (cpu *Cpu) ldx() {
}

// load y
func (cpu *Cpu) ldy() {
}

// logical shift right
func (cpu *Cpu) lsr() {
}

// no operation
func (cpu *Cpu) nop() {
}

// bitwise or
func (cpu *Cpu) ora() {
}

// push a
func (cpu *Cpu) pha() {
}

// push processor status
func (cpu *Cpu) php() {
}

// pull a
func (cpu *Cpu) pla() {
}

// pull processor status
func (cpu *Cpu) plp() {
}

// rotate left
func (cpu *Cpu) rol() {
}

// rotate right
func (cpu *Cpu) ror() {
}

// return from interrupt
func (cpu *Cpu) rti() {
}

// return from subroutine
func (cpu *Cpu) rts() {
}

// subtract with carry
func (cpu *Cpu) sbc() {
}

// set carry
func (cpu *Cpu) sec() {
}

// set decimal
func (cpu *Cpu) sed() {
}

// set interrupt disable
func (cpu *Cpu) sei() {
}

// store a
func (cpu *Cpu) sta() {
}

// store x
func (cpu *Cpu) stx() {
}

// store y
func (cpu *Cpu) sty() {
}

// transfer a to x
func (cpu *Cpu) tax() {
}

// transfer a to y
func (cpu *Cpu) tay() {
}

// transfer stack pointer to x
func (cpu *Cpu) tsx() {
}

// transfer x to a
func (cpu *Cpu) txa() {
}

// transfer y to stack pointer
func (cpu *Cpu) txs() {
}

// transfer y to a
func (cpu *Cpu) tya() {
}
