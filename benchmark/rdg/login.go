package rdg

import (
	"bytes"

	"github.com/LyrinoxTechnologies/ridged-proto/rdgproto"
)

type LoginRequest struct {
	Username string
	Password string
	ClientId string
}

func (l *LoginRequest) Marshal() ([]byte, error) {
	buf := rdgproto.GetBuffer()
	defer rdgproto.PutBuffer(buf)

	rdgproto.WriteString(buf, l.Username)
	rdgproto.WriteString(buf, l.Password)
	rdgproto.WriteString(buf, l.ClientId)

	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

func (l *LoginRequest) Unmarshal(data []byte) error {
	r := bytes.NewReader(data)

	var err error
	if l.Username, err = rdgproto.ReadString(r); err != nil {
		return err
	}
	if l.Password, err = rdgproto.ReadString(r); err != nil {
		return err
	}
	if l.ClientId, err = rdgproto.ReadString(r); err != nil {
		return err
	}
	return nil
}