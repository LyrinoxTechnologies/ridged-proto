package benchmark

import (
	"bytes"
	"testing"

	"google.golang.org/protobuf/proto"
	"github.com/LyrinoxTechnologies/ridged-proto/rdgproto"
)

// rdgproto implementation
type LoginRequestRdg struct {
	Username string
	Password string
	ClientId string
}

func (l *LoginRequestRdg) Marshal() ([]byte, error) {
	buf := rdgproto.GetBuffer()
	defer rdgproto.PutBuffer(buf)
	
	rdgproto.WriteString(buf, l.Username)
	rdgproto.WriteString(buf, l.Password)
	rdgproto.WriteString(buf, l.ClientId)
	
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

func (l *LoginRequestRdg) Unmarshal(data []byte) error {
	r := bytes.NewReader(data)
	var err error
	l.Username, err = rdgproto.ReadString(r)
	if err != nil {
		return err
	}
	l.Password, err = rdgproto.ReadString(r)
	if err != nil {
		return err
	}
	l.ClientId, err = rdgproto.ReadString(r)
	return err
}

// Test data
var testUsername = "john.doe@example.com"
var testPassword = "super_secret_password_123"
var testClientId = "client-abc-123-xyz"

// Benchmark: Protobuf Marshal
func BenchmarkProtobuf_Marshal(b *testing.B) {
	msg := &LoginRequest{
		Username: testUsername,
		Password: testPassword,
		ClientId: testClientId,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data, err := proto.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
		_ = data
	}
}

// Benchmark: rdgproto Marshal
func BenchmarkRdgproto_Marshal(b *testing.B) {
	msg := &LoginRequestRdg{
		Username: testUsername,
		Password: testPassword,
		ClientId: testClientId,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data, err := msg.Marshal()
		if err != nil {
			b.Fatal(err)
		}
		_ = data
	}
}

// Benchmark: Protobuf Unmarshal
func BenchmarkProtobuf_Unmarshal(b *testing.B) {
	msg := &LoginRequest{
		Username: testUsername,
		Password: testPassword,
		ClientId: testClientId,
	}
	data, _ := proto.Marshal(msg)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := &LoginRequest{}
		err := proto.Unmarshal(data, result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark: rdgproto Unmarshal
func BenchmarkRdgproto_Unmarshal(b *testing.B) {
	msg := &LoginRequestRdg{
		Username: testUsername,
		Password: testPassword,
		ClientId: testClientId,
	}
	data, _ := msg.Marshal()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := &LoginRequestRdg{}
		err := result.Unmarshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Size comparison test
func TestMessageSize(t *testing.T) {
	protoMsg := &LoginRequest{
		Username: testUsername,
		Password: testPassword,
		ClientId: testClientId,
	}
	protoData, _ := proto.Marshal(protoMsg)

	rdgMsg := &LoginRequestRdg{
		Username: testUsername,
		Password: testPassword,
		ClientId: testClientId,
	}
	rdgData, _ := rdgMsg.Marshal()

	t.Logf("Protobuf size: %d bytes", len(protoData))
	t.Logf("rdgproto size: %d bytes", len(rdgData))
	
	if len(rdgData) > len(protoData) {
		diff := len(rdgData) - len(protoData)
		pct := float64(diff) / float64(len(protoData)) * 100
		t.Logf("rdgproto is LARGER by %d bytes (+%.2f%%)", diff, pct)
	} else {
		diff := len(protoData) - len(rdgData)
		pct := float64(diff) / float64(len(protoData)) * 100
		t.Logf("rdgproto is SMALLER by %d bytes (-%.2f%%)", diff, pct)
	}
}