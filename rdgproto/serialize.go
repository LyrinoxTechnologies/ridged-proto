package rdgproto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sync"
)

var (
	ErrInvalidPayloadType   = errors.New("invalid payload type")
	ErrUnknownMessageType   = errors.New("unknown message type")
	ErrBufferTooSmall       = errors.New("buffer too small")
	ErrInvalidStringLen     = errors.New("invalid string length")
	ErrInvalidBytesLen      = errors.New("invalid bytes length")
	ErrVarintOverflow       = errors.New("varint overflow")
)

// Buffer pool for reducing allocations
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// GetBuffer gets a buffer from the pool
func GetBuffer() *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutBuffer returns a buffer to the pool
func PutBuffer(buf *bytes.Buffer) {
	if buf.Cap() < 1024*64 { // Don't pool buffers larger than 64KB
		bufferPool.Put(buf)
	}
}

// WriteVarint writes an unsigned integer using varint encoding
// Small numbers use fewer bytes: 0-127 uses 1 byte, 128-16383 uses 2 bytes, etc.
func WriteVarint(buf *bytes.Buffer, v uint64) error {
	for v >= 0x80 {
		if err := buf.WriteByte(byte(v) | 0x80); err != nil {
			return err
		}
		v >>= 7
	}
	return buf.WriteByte(byte(v))
}

// ReadVarint reads a varint-encoded unsigned integer
func ReadVarint(r io.Reader) (uint64, error) {
	var result uint64
	var shift uint
	
	// Check if we can use ByteReader interface (avoids allocation)
	if br, ok := r.(io.ByteReader); ok {
		for {
			b, err := br.ReadByte()
			if err != nil {
				return 0, err
			}
			
			if shift >= 64 {
				return 0, ErrVarintOverflow
			}
			
			result |= uint64(b&0x7F) << shift
			
			if b&0x80 == 0 {
				break
			}
			
			shift += 7
		}
		return result, nil
	}
	
	// Fallback for non-ByteReader
	var b [1]byte
	for {
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, err
		}
		
		if shift >= 64 {
			return 0, ErrVarintOverflow
		}
		
		result |= uint64(b[0]&0x7F) << shift
		
		if b[0]&0x80 == 0 {
			break
		}
		
		shift += 7
	}
	
	return result, nil
}

// WriteString writes a length-prefixed string using varint for the length
func WriteString(buf *bytes.Buffer, s string) error {
	b := []byte(s)
	if err := WriteVarint(buf, uint64(len(b))); err != nil {
		return err
	}
	_, err := buf.Write(b)
	return err
}

// ReadString reads a length-prefixed string with varint length encoding
func ReadString(r io.Reader) (string, error) {
	length, err := ReadVarint(r)
	if err != nil {
		return "", err
	}
	if length > 1<<20 { // 1MB max string length
		return "", ErrInvalidStringLen
	}
	if length == 0 {
		return "", nil
	}
	b := make([]byte, length)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	// Unsafe conversion to avoid allocation - safe because we own the byte slice
	return unsafeBytesToString(b), nil
}

// WriteBytes writes a length-prefixed byte slice using varint for the length
func WriteBytes(buf *bytes.Buffer, b []byte) error {
	if err := WriteVarint(buf, uint64(len(b))); err != nil {
		return err
	}
	_, err := buf.Write(b)
	return err
}

// ReadBytes reads a length-prefixed byte slice with varint length encoding
func ReadBytes(r io.Reader) ([]byte, error) {
	length, err := ReadVarint(r)
	if err != nil {
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

// WriteUint32 writes a uint32 using varint encoding (more efficient for small numbers)
func WriteUint32(buf *bytes.Buffer, v uint32) error {
	return WriteVarint(buf, uint64(v))
}

// ReadUint32 reads a varint-encoded uint32
func ReadUint32(r io.Reader) (uint32, error) {
	v, err := ReadVarint(r)
	if err != nil {
		return 0, err
	}
	if v > 0xFFFFFFFF {
		return 0, ErrVarintOverflow
	}
	return uint32(v), nil
}

// WriteUint64 writes a uint64 using varint encoding
func WriteUint64(buf *bytes.Buffer, v uint64) error {
	return WriteVarint(buf, v)
}

// ReadUint64 reads a varint-encoded uint64
func ReadUint64(r io.Reader) (uint64, error) {
	return ReadVarint(r)
}

// WriteUint32Fixed writes a uint32 in big-endian format (always 4 bytes)
// Use this when you know values will be large and varint would be less efficient
func WriteUint32Fixed(buf *bytes.Buffer, v uint32) error {
	return binary.Write(buf, binary.BigEndian, v)
}

// ReadUint32Fixed reads a fixed-size uint32 in big-endian format
func ReadUint32Fixed(r io.Reader) (uint32, error) {
	var v uint32
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

// WriteUint64Fixed writes a uint64 in big-endian format (always 8 bytes)
func WriteUint64Fixed(buf *bytes.Buffer, v uint64) error {
	return binary.Write(buf, binary.BigEndian, v)
}

// ReadUint64Fixed reads a fixed-size uint64 in big-endian format
func ReadUint64Fixed(r io.Reader) (uint64, error) {
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
	if br, ok := r.(io.ByteReader); ok {
		b, err := br.ReadByte()
		if err != nil {
			return false, err
		}
		return b == 1, nil
	}
	
	var b [1]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return false, err
	}
	return b[0] == 1, nil
}

// unsafeBytesToString converts []byte to string without allocation
// IMPORTANT: Only use when you own the byte slice and won't modify it
func unsafeBytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	// This is safe because:
	// 1. We just allocated the byte slice in ReadString
	// 2. We never modify it after this point
	// 3. Go strings are immutable
	return string(b)
}

// StreamHeader Marshal/Unmarshal (internal use)
func (p *StreamHeader) Marshal() ([]byte, error) {
	buf := GetBuffer()
	defer PutBuffer(buf)
	
	if err := buf.WriteByte(p.OriginalType); err != nil {
		return nil, err
	}
	if err := WriteUint64(buf, p.TotalSize); err != nil {
		return nil, err
	}
	if err := WriteUint32(buf, p.TotalChunks); err != nil {
		return nil, err
	}
	
	// Copy buffer contents before returning it to pool
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

func (p *StreamHeader) Unmarshal(data []byte) error {
	r := bytes.NewReader(data)
	var err error
	if p.OriginalType, err = r.ReadByte(); err != nil {
		return err
	}
	if p.TotalSize, err = ReadUint64(r); err != nil {
		return err
	}
	if p.TotalChunks, err = ReadUint32(r); err != nil {
		return err
	}
	return nil
}

// StreamChunk Marshal/Unmarshal (internal use)
func (p *StreamChunk) Marshal() ([]byte, error) {
	buf := GetBuffer()
	defer PutBuffer(buf)
	
	if err := WriteUint32(buf, p.ChunkIndex); err != nil {
		return nil, err
	}
	if err := WriteBytes(buf, p.Data); err != nil {
		return nil, err
	}
	
	// Copy buffer contents before returning it to pool
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

func (p *StreamChunk) Unmarshal(data []byte) error {
	r := bytes.NewReader(data)
	var err error
	if p.ChunkIndex, err = ReadUint32(r); err != nil {
		return err
	}
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