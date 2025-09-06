package cartridge

import (
	"errors"
	"io"
	"os"
)

var headerSignature = [...]byte{0x4E, 0x45, 0x53, 0x1A}

const (
	headerSize     int = 16
	trainerSize    int = 512
	prgRomBaseSize int = 16384
	chrRomBaseSize int = 8192
)

var (
	programData   []byte
	characterData []byte
	hasTrainer    bool
)

func LoadCartridge(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	err = parseHeader(file)
	if err != nil {
		return err
	}

	if hasTrainer {
		file.Seek(int64(trainerSize), io.SeekCurrent)
	}

	err = parseProgramData(file)
	if err != nil {
		return err
	}

	err = parseCharacterData(file)
	if err != nil {
		return err
	}

	return nil
}

func ReadProgramData(address uint16) uint8 {
	if programData == nil {
		return 0x00
	}
	return programData[address%uint16(prgRomBaseSize)]
}

func ReadCharacterData(address uint16) uint8 {
	if characterData == nil {
		return 0x00
	}
	return characterData[address%uint16(chrRomBaseSize)]
}

func parseHeader(file *os.File) error {
	buffer := make([]byte, headerSize)

	n, err := file.Read(buffer)
	if err != nil {
		return err
	}
	if n < headerSize {
		return errors.New("unexpected end of rom file")
	}

	for i, b := range headerSignature {
		if buffer[i] != b {
			return errors.New("invalid rom file")
		}
	}

	flags := buffer[6]
	hasTrainer = flags&0x04 > 0

	return nil
}

func parseProgramData(file *os.File) error {
	programData = make([]byte, prgRomBaseSize)
	n, err := file.Read(programData)
	if err != nil {
		return err
	}
	if n < prgRomBaseSize {
		return errors.New("unexpected end of rom file")
	}
	return nil
}

func parseCharacterData(file *os.File) error {
	characterData = make([]byte, chrRomBaseSize)
	n, err := file.Read(characterData)
	if err != nil {
		return err
	}
	if n < chrRomBaseSize {
		return errors.New("unexpected end of rom file")
	}
	return nil
}
