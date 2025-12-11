package rdgproto

import (
"bytes"
"encoding/binary"
"errors"
"io"
)

var (
ErrInvalidPayloadType   = errors.New("invalid payload type")
ErrUnknownMessageType   = errors.New("unknown message type")
ErrBufferTooSmall       = errors.New("buffer too small")
ErrInvalidStringLen     = errors.New("invalid string length")
ErrInvalidBytesLen      = errors.New("invalid bytes length")
)

// WriteString writes a length-prefixed string to a buffer
// This is a helper for implementing custom payload marshaling
func WriteString(buf *bytes.Buffer, s string) error {
b := []byte(s)
if err := binary.Write(buf, binary.BigEndian, uint32(len(b))); err != nil {
return err
}
_, err := buf.Write(b)
return err
}

// ReadString reads a length-prefixed string from a reader
// This is a helper for implementing custom payload unmarshaling
func ReadString(r io.Reader) (string, error) {
var length uint32
if err := binary.Read(r, binary.BigEndian, &length); err != nil {
return "", err
}
if length > 1<<20 { // 1MB max string length
return "", ErrInvalidStringLen
}
b := make([]byte, length)
if _, err := io.ReadFull(r, b); err != nil {
return "", err
}
return string(b), nil
}

// WriteBytes writes a length-prefixed byte slice to a buffer
// This is a helper for implementing custom payload marshaling
func WriteBytes(buf *bytes.Buffer, b []byte) error {
if err := binary.Write(buf, binary.BigEndian, uint32(len(b))); err != nil {
return err
}
_, err := buf.Write(b)
return err
}

// ReadBytes reads a length-prefixed byte slice from a reader
// This is a helper for implementing custom payload unmarshaling
func ReadBytes(r io.Reader) ([]byte, error) {
var length uint32
if err := binary.Read(r, binary.BigEndian, &length); err != nil {
return nil, err
}
if length > 1<<30 { // 1GB max bytes length
return nil, ErrInvalidBytesLen
}
b := make([]byte, length)
if _, err := io.ReadFull(r, b); err != nil {
return nil, err
}
return b, nil
}

// WriteUint32 writes a uint32 to a buffer in big-endian format
func WriteUint32(buf *bytes.Buffer, v uint32) error {
return binary.Write(buf, binary.BigEndian, v)
}

// ReadUint32 reads a uint32 from a reader in big-endian format
func ReadUint32(r io.Reader) (uint32, error) {
var v uint32
err := binary.Read(r, binary.BigEndian, &v)
return v, err
}

// WriteUint64 writes a uint64 to a buffer in big-endian format
func WriteUint64(buf *bytes.Buffer, v uint64) error {
return binary.Write(buf, binary.BigEndian, v)
}

// ReadUint64 reads a uint64 from a reader in big-endian format
func ReadUint64(r io.Reader) (uint64, error) {
var v uint64
err := binary.Read(r, binary.BigEndian, &v)
return v, err
}

// WriteBool writes a boolean as a single byte (0 or 1)
func WriteBool(buf *bytes.Buffer, v bool) error {
var b byte
if v {
b = 1
}
return buf.WriteByte(b)
}

// ReadBool reads a boolean from a reader
func ReadBool(r io.Reader) (bool, error) {
b := make([]byte, 1)
if _, err := io.ReadFull(r, b); err != nil {
return false, err
}
return b[0] == 1, nil
}

// StreamHeader Marshal/Unmarshal (internal use)
func (p *StreamHeader) Marshal() ([]byte, error) {
buf := new(bytes.Buffer)
if err := buf.WriteByte(p.OriginalType); err != nil {
return nil, err
}
if err := binary.Write(buf, binary.BigEndian, p.TotalSize); err != nil {
return nil, err
}
if err := binary.Write(buf, binary.BigEndian, p.TotalChunks); err != nil {
return nil, err
}
return buf.Bytes(), nil
}

func (p *StreamHeader) Unmarshal(data []byte) error {
r := bytes.NewReader(data)
var err error
if p.OriginalType, err = r.ReadByte(); err != nil {
return err
}
if err = binary.Read(r, binary.BigEndian, &p.TotalSize); err != nil {
return err
}
if err = binary.Read(r, binary.BigEndian, &p.TotalChunks); err != nil {
return err
}
return nil
}

// StreamChunk Marshal/Unmarshal (internal use)
func (p *StreamChunk) Marshal() ([]byte, error) {
buf := new(bytes.Buffer)
if err := binary.Write(buf, binary.BigEndian, p.ChunkIndex); err != nil {
return nil, err
}
if err := WriteBytes(buf, p.Data); err != nil {
return nil, err
}
return buf.Bytes(), nil
}

func (p *StreamChunk) Unmarshal(data []byte) error {
r := bytes.NewReader(data)
if err := binary.Read(r, binary.BigEndian, &p.ChunkIndex); err != nil {
return err
}
var err error
if p.Data, err = ReadBytes(r); err != nil {
return err
}
return nil
}

// MarshalPayload serializes any supported payload type to binary
func MarshalPayload(payload interface{}) ([]byte, error) {
// Handle nil case first
if payload == nil {
return []byte{}, nil
}

// Handle raw bytes passthrough
if b, ok := payload.([]byte); ok {
return b, nil
}

// Try interface-based marshaling (covers all types with Marshal method)
if m, ok := payload.(PayloadMarshaler); ok {
return m.Marshal()
}

return nil, ErrInvalidPayloadType
}

// UnmarshalPayload deserializes binary data into the appropriate payload type
// Uses the global registry to look up payload types
// Returns raw bytes for unknown message types (non-strict mode)
func UnmarshalPayload(messageType byte, data []byte) (interface{}, error) {
return UnmarshalPayloadStrict(messageType, data, false)
}

// UnmarshalPayloadStrict deserializes binary data with optional strict mode
// In strict mode, returns ErrUnknownMessageType for unregistered message types
func UnmarshalPayloadStrict(messageType byte, data []byte, strict bool) (interface{}, error) {
// Try to get factory from registry first
factory := globalRegistry.Get(messageType)
if factory != nil {
p := factory()
if err := p.Unmarshal(data); err != nil {
return nil, err
}
return p, nil
}

// In strict mode, reject unknown message types
if strict {
return nil, ErrUnknownMessageType
}

// Return raw bytes for unknown message types (non-strict mode)
return data, nil
}

// UnmarshalPayloadWithRegistry deserializes binary data using a specific registry
func UnmarshalPayloadWithRegistry(messageType byte, data []byte, registry *PayloadRegistry) (interface{}, error) {
return UnmarshalPayloadWithRegistryStrict(messageType, data, registry, false)
}

// UnmarshalPayloadWithRegistryStrict deserializes binary data with optional strict mode
func UnmarshalPayloadWithRegistryStrict(messageType byte, data []byte, registry *PayloadRegistry, strict bool) (interface{}, error) {
factory := registry.Get(messageType)
if factory != nil {
p := factory()
if err := p.Unmarshal(data); err != nil {
return nil, err
}
return p, nil
}

// In strict mode, reject unknown message types
if strict {
return nil, ErrUnknownMessageType
}

// Return raw bytes for unknown message types (non-strict mode)
return data, nil
}
