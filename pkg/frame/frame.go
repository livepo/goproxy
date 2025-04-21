package frame

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type Frame struct {
	Type   uint8
	ConnID uint32
	Length uint32
	Data   []byte
}

func DecodeFrame(data []byte) (*Frame, error) {
	if len(data) < 9 {
		return nil, errors.New("data too short")
	}

	buf := bytes.NewReader(data)

	var f Frame
	binary.Read(buf, binary.BigEndian, &f.Type)
	binary.Read(buf, binary.BigEndian, &f.ConnID)
	binary.Read(buf, binary.BigEndian, &f.Length)

	if int(f.Length) != len(data)-9 {
		return nil, errors.New("invalid length")
	}

	f.Data = data[9:]
	return &f, nil
}

func EncodeFrame(f *Frame) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, f.Type)
	binary.Write(buf, binary.BigEndian, f.ConnID)
	binary.Write(buf, binary.BigEndian, f.Length)
	buf.Write(f.Data)
	return buf.Bytes()
}
