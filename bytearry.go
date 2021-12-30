package FabricEmu

import (
	"bytes"
	"encoding/binary"

	"github.com/tg123/phabrik/serialization"
)

type bytearray struct {
	data interface{}
}

var _ serialization.CustomMarshaler = (*bytearray)(nil)

func (b *bytearray) Marshal(s serialization.Encoder) error {
	if err := s.WriteTypeMeta(serialization.FabricSerializationTypeArray | serialization.FabricSerializationTypeUChar); err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, b.data); err != nil {
		return err
	}

	data := buf.Bytes()

	if err := s.WriteCompressedUInt32(uint32(len(data))); err != nil {
		return err
	}

	return s.WriteBinary(data)
}

func (b *bytearray) Unmarshal(_ serialization.FabricSerializationType, _ serialization.Decoder) error {
	panic("not implemented") // TODO: Implement
}
