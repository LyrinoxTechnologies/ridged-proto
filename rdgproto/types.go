// Package rdgproto provides a binary-only protocol framework for communication
// between different client platforms and a central server.
//
// rdgproto is a framework - it provides the protocol infrastructure (send/receive,
// marshal/unmarshal, streaming) and developers define their own message types.
package rdgproto

import (
"io"
"sync"
)

// Reserved message types for internal streaming protocol
// Developers should NOT use these types for their own messages
const (
MessageTypeStreamStart byte = 250
MessageTypeStreamChunk byte = 251
MessageTypeStreamEnd   byte = 252
)

// IsReservedType returns true if the message type is reserved for internal use
func IsReservedType(messageType byte) bool {
return messageType >= 250
}

// Header sizes
const (
MessageTypeSize     = 1
MessageIDSize       = 4
PayloadLengthSize   = 4
HeaderSize          = MessageTypeSize + MessageIDSize + PayloadLengthSize
SignatureLengthSize = 4
)

// Streaming constants
const (
// StreamingThreshold is the payload size at which streaming is automatically used
StreamingThreshold = 1024 * 1024 // 1MB

// DefaultChunkSize is the default size for stream chunks
DefaultChunkSize = 64 * 1024 // 64KB
)

// Message represents a protocol message with header and payload
type Message struct {
Type      byte
ID        uint32
Payload   []byte
Signature []byte
}

// StreamHeader represents metadata for a streamed message (internal use)
type StreamHeader struct {
OriginalType byte
TotalSize    uint64
TotalChunks  uint32
}

// StreamChunk represents a chunk in a streamed message (internal use)
type StreamChunk struct {
ChunkIndex uint32
Data       []byte
}

// Connection interface for transport-agnostic communication
// Developers implement this interface to provide their transport layer
type Connection interface {
io.Reader
io.Writer
io.Closer
}

// Signer interface for optional message signing
type Signer interface {
Sign(data []byte) ([]byte, error)
}

// Verifier interface for optional signature verification
type Verifier interface {
Verify(data []byte, signature []byte) error
}

// PayloadMarshaler interface for custom payload types
// Implement this on your payload structs to enable serialization
type PayloadMarshaler interface {
Marshal() ([]byte, error)
}

// PayloadUnmarshaler interface for custom payload types
// Implement this on your payload structs to enable deserialization
type PayloadUnmarshaler interface {
Unmarshal(data []byte) error
}

// Payload combines both marshal and unmarshal capabilities
type Payload interface {
PayloadMarshaler
PayloadUnmarshaler
}

// PayloadFactory creates a new instance of a payload type
type PayloadFactory func() PayloadUnmarshaler

// PayloadRegistry manages custom payload type registrations
type PayloadRegistry struct {
mu       sync.RWMutex
handlers map[byte]PayloadFactory
}

// globalRegistry is the default registry for payload types
var globalRegistry = NewPayloadRegistry()

// NewPayloadRegistry creates a new payload registry
func NewPayloadRegistry() *PayloadRegistry {
r := &PayloadRegistry{
handlers: make(map[byte]PayloadFactory),
}
// Register internal streaming types only
r.registerStreamingTypes()
return r
}

// registerStreamingTypes registers the internal streaming types
func (r *PayloadRegistry) registerStreamingTypes() {
r.handlers[MessageTypeStreamStart] = func() PayloadUnmarshaler { return &StreamHeader{} }
r.handlers[MessageTypeStreamChunk] = func() PayloadUnmarshaler { return &StreamChunk{} }
}

// Register adds or replaces a payload handler for a message type
// Use this to register your custom message types
func (r *PayloadRegistry) Register(messageType byte, factory PayloadFactory) {
if IsReservedType(messageType) {
return // Silently ignore attempts to register reserved types
}
r.mu.Lock()
defer r.mu.Unlock()
r.handlers[messageType] = factory
}

// Unregister removes a payload handler for a message type
func (r *PayloadRegistry) Unregister(messageType byte) {
if IsReservedType(messageType) {
return // Cannot unregister reserved types
}
r.mu.Lock()
defer r.mu.Unlock()
delete(r.handlers, messageType)
}

// Get returns the factory for a message type, or nil if not registered
func (r *PayloadRegistry) Get(messageType byte) PayloadFactory {
r.mu.RLock()
defer r.mu.RUnlock()
return r.handlers[messageType]
}

// Has checks if a message type is registered
func (r *PayloadRegistry) Has(messageType byte) bool {
r.mu.RLock()
defer r.mu.RUnlock()
_, ok := r.handlers[messageType]
return ok
}

// RegisterPayloadType registers a custom payload type with the global registry
// This is the main way developers register their message types
//
// Example:
//
//const MyMessageType byte = 1
//rdgproto.RegisterPayloadType(MyMessageType, func() rdgproto.PayloadUnmarshaler {
//    return &MyPayload{}
//})
func RegisterPayloadType(messageType byte, factory PayloadFactory) {
globalRegistry.Register(messageType, factory)
}

// UnregisterPayloadType removes a payload type from the global registry
func UnregisterPayloadType(messageType byte) {
globalRegistry.Unregister(messageType)
}

// GetPayloadFactory returns the factory for a message type from the global registry
func GetPayloadFactory(messageType byte) PayloadFactory {
return globalRegistry.Get(messageType)
}

// HasPayloadType checks if a message type is registered in the global registry
func HasPayloadType(messageType byte) bool {
return globalRegistry.Has(messageType)
}

// StreamConfig configures streaming behavior
type StreamConfig struct {
// Threshold is the payload size at which streaming is used (default: 1MB)
Threshold int

// ChunkSize is the size of each stream chunk (default: 64KB)
ChunkSize int

// Enabled controls whether streaming is active
Enabled bool
}

// DefaultStreamConfig returns the default streaming configuration
func DefaultStreamConfig() *StreamConfig {
return &StreamConfig{
Threshold: StreamingThreshold,
ChunkSize: DefaultChunkSize,
Enabled:   true,
}
}
