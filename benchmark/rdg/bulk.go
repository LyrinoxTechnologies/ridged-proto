package rdg

import (
	"bytes"

	"github.com/LyrinoxTechnologies/ridged-proto/rdgproto"
)

type BulkData struct {
	Values []uint32
}

func (b *BulkData) Marshal() ([]byte, error) {
	buf := rdgproto.GetBuffer()
	defer rdgproto.PutBuffer(buf)

	// Length prefix (number of elements)
	rdgproto.WriteUint32(buf, uint32(len(b.Values)))

	for _, v := range b.Values {
		rdgproto.WriteUint32(buf, v)
	}

	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

func (b *BulkData) Unmarshal(data []byte) error {
	r := bytes.NewReader(data)

	count, err := rdgproto.ReadUint32(r)
	if err != nil {
		return err
	}

	b.Values = make([]uint32, count)
	for i := uint32(0); i < count; i++ {
		v, err := rdgproto.ReadUint32(r)
		if err != nil {
			return err
		}
		b.Values[i] = v
	}
	return nil
}