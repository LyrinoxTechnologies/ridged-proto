package rdg

import (
	"bytes"

	"github.com/LyrinoxTechnologies/ridged-proto/rdgproto"
)

type Blob struct {
	Data []byte
}

func (b *Blob) Marshal() ([]byte, error) {
	buf := rdgproto.GetBuffer()
	defer rdgproto.PutBuffer(buf)

	rdgproto.WriteBytes(buf, b.Data)

	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

func (b *Blob) Unmarshal(data []byte) error {
	r := bytes.NewReader(data)

	var err error
	b.Data, err = rdgproto.ReadBytes(r)
	return err
}