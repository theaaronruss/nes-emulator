package cartridge

import (
	"errors"
	"io"
	"os"
)

var headerStart = [4]byte{0x4E, 0x45, 0x53, 0x1A}

const (
	headerSize     int = 16
	trainerSize    int = 512
	PrgRomBankSize int = 16384
)

type Cartridge struct {
	programDataBanks [][]byte
	characterData    []byte

	prgSize    int
	chrSize    int
	hasTrainer bool
}

func LoadCartridge(filePath string) (*Cartridge, error) {
	cartridge := &Cartridge{}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	err = cartridge.parseHeader(file)
	if err != nil {
		return nil, err
	}

	if cartridge.hasTrainer {
		file.Seek(int64(trainerSize), io.SeekCurrent)
	}

	err = cartridge.readProgramData(file)
	if err != nil {
		return nil, err
	}

	err = cartridge.readCharacterData(file)
	if err != nil {
		return nil, err
	}

	return cartridge, nil
}

func (cart *Cartridge) Read(bank int, address uint16) uint8 {
	if bank < 0 || bank >= len(cart.programDataBanks) || address > uint16(PrgRomBankSize) {
		return 0x00
	}
	return cart.programDataBanks[bank][address]
}

func (cart *Cartridge) parseHeader(file *os.File) error {
	buffer := make([]byte, headerSize)
	n, err := file.Read(buffer)
	if err != nil {
		return err
	}
	if n < headerSize {
		return errors.New("unexpected end of rom file")
	}

	for i, b := range headerStart {
		if buffer[i] != b {
			return errors.New("invalid rom file")
		}
	}

	cart.prgSize = int(buffer[4])
	cart.chrSize = int(buffer[5])

	flags := buffer[6]
	cart.hasTrainer = flags&0x04 > 0

	return nil
}

func (cart *Cartridge) readProgramData(file *os.File) error {
	cart.programDataBanks = make([][]byte, cart.prgSize)
	for i := range cart.prgSize {
		cart.programDataBanks[i] = make([]byte, PrgRomBankSize)
		n, err := file.Read(cart.programDataBanks[i])
		if err != nil {
			return err
		}
		if n < PrgRomBankSize {
			return errors.New("unexpected end of rom file")
		}
	}
	return nil
}

func (cart *Cartridge) readCharacterData(file *os.File) error {
	cart.characterData = make([]byte, cart.chrSize)
	n, err := file.Read(cart.characterData)
	if err != nil {
		return err
	}
	if n < cart.chrSize {
		return errors.New("unexpected end of rom file")
	}
	return nil
}
