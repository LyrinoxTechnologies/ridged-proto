package rdg

import (
	"bytes"

	"github.com/LyrinoxTechnologies/ridged-proto/rdgproto"
)

type Metrics struct {
	A uint64
	B uint64
	C uint64
	D uint64
	E uint64
}

func (m *Metrics) Marshal() ([]byte, error) {
	buf := rdgproto.GetBuffer()
	defer rdgproto.PutBuffer(buf)

	rdgproto.WriteUint64(buf, m.A)
	rdgproto.WriteUint64(buf, m.B)
	rdgproto.WriteUint64(buf, m.C)
	rdgproto.WriteUint64(buf, m.D)
	rdgproto.WriteUint64(buf, m.E)

	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

func (m *Metrics) Unmarshal(data []byte) error {
	r := bytes.NewReader(data)

	var err error
	if m.A, err = rdgproto.ReadUint64(r); err != nil {
		return err
	}
	if m.B, err = rdgproto.ReadUint64(r); err != nil {
		return err
	}
	if m.C, err = rdgproto.ReadUint64(r); err != nil {
		return err
	}
	if m.D, err = rdgproto.ReadUint64(r); err != nil {
		return err
	}
	if m.E, err = rdgproto.ReadUint64(r); err != nil {
		return err
	}
	return nil
}