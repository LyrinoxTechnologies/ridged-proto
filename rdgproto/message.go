package rdgproto

import (
"bytes"
"encoding/binary"
"errors"
"io"
"sync"
)

var (
ErrInvalidMessage    = errors.New("invalid message format")
ErrPayloadTooLarge   = errors.New("payload exceeds maximum size")
ErrSignatureRequired = errors.New("signature required but not present")
ErrInvalidSignature  = errors.New("invalid signature")
ErrStreamInterrupted = errors.New("stream was interrupted")
ErrStreamMismatch    = errors.New("stream chunk mismatch")
)

// MaxPayloadSize is the maximum allowed payload size (100MB)
const MaxPayloadSize = 100 * 1024 * 1024

// messageIDCounter is a global counter for generating unique message IDs
var messageIDCounter uint32 = 0
var messageIDMu sync.Mutex

// nextGlobalMessageID returns a unique message ID
func nextGlobalMessageID() uint32 {
messageIDMu.Lock()
defer messageIDMu.Unlock()
messageIDCounter++
return messageIDCounter
}

// Marshal serializes a payload with the given message type into rdgproto binary format.
// This is the simplest way to create a binary message for transmission.
// Returns the complete binary message ready to be sent over any transport.
//
// Example:
//
//data, err := rdgproto.Marshal(MyMessageType, &MyPayload{...})
func Marshal(messageType byte, payload interface{}) ([]byte, error) {
return MarshalMessage(messageType, nextGlobalMessageID(), payload, nil)
}

// MarshalWithID serializes a payload with a specific message ID.
// Use this when you need to correlate request/response pairs.
func MarshalWithID(messageType byte, messageID uint32, payload interface{}) ([]byte, error) {
return MarshalMessage(messageType, messageID, payload, nil)
}

// MarshalSecure serializes a payload with HMAC signing for message authentication.
// The same secret must be used to verify the message on the receiving end.
func MarshalSecure(messageType byte, payload interface{}, secret []byte) ([]byte, error) {
opts := SecureMessageOptions(secret)
return MarshalMessage(messageType, nextGlobalMessageID(), payload, opts)
}

// Unmarshal deserializes an rdgproto binary message into its components.
// Returns the message header and the deserialized payload.
//
// Example:
//
//msg, payload, err := rdgproto.Unmarshal(data)
//if err != nil {
//    // handle error
//}
//switch msg.Type {
//case MyMessageType:
//    myPayload := payload.(*MyPayload)
//    // use myPayload
//}
func Unmarshal(data []byte) (*Message, interface{}, error) {
return UnmarshalMessage(data, nil)
}

// UnmarshalSecure deserializes and verifies an rdgproto message with HMAC verification.
// Returns an error if the signature is invalid or missing.
func UnmarshalSecure(data []byte, secret []byte) (*Message, interface{}, error) {
opts := SecureMessageOptions(secret)
return UnmarshalMessage(data, opts)
}

// UnmarshalStrict deserializes with strict mode - rejects unknown message types
// This is recommended for production use to prevent processing malicious messages
func UnmarshalStrict(data []byte) (*Message, interface{}, error) {
return UnmarshalMessage(data, StrictMessageOptions())
}

// UnmarshalInto deserializes the payload from a binary message directly into a target struct.
// The target must be a pointer to a struct that implements PayloadUnmarshaler.
//
// Example:
//
//var myPayload MyPayload
//msg, err := rdgproto.UnmarshalInto(data, &myPayload)
func UnmarshalInto(data []byte, target PayloadUnmarshaler) (*Message, error) {
msg, _, err := UnmarshalMessage(data, nil)
if err != nil {
return nil, err
}
if err := target.Unmarshal(msg.Payload); err != nil {
return nil, err
}
return msg, nil
}

// MessageOptions contains optional settings for message handling
type MessageOptions struct {
Signer       Signer
Verifier     Verifier
Registry     *PayloadRegistry
StreamConfig *StreamConfig
// StrictMode when true, rejects messages with unknown message types
// This prevents processing of unregistered message types for security
StrictMode   bool
}

// StrictMessageOptions creates options with strict mode enabled
// In strict mode, only registered message types are accepted
func StrictMessageOptions() *MessageOptions {
return &MessageOptions{
StrictMode: true,
}
}

// SecureStrictMessageOptions creates options with both HMAC signing and strict mode
func SecureStrictMessageOptions(secret []byte) *MessageOptions {
opts := SecureMessageOptions(secret)
opts.StrictMode = true
return opts
}

// MarshalMessage serializes a message with header and payload into binary format
// Format: [Type(1)][ID(4)][PayloadLen(4)][Payload(N)][SignatureLen(4)][Signature(N)]
func MarshalMessage(messageType byte, messageID uint32, payload interface{}, opts *MessageOptions) ([]byte, error) {
// Serialize payload
payloadBytes, err := MarshalPayload(payload)
if err != nil {
return nil, err
}

if len(payloadBytes) > MaxPayloadSize {
return nil, ErrPayloadTooLarge
}

buf := new(bytes.Buffer)

// Write header
if err := buf.WriteByte(messageType); err != nil {
return nil, err
}
if err := binary.Write(buf, binary.BigEndian, messageID); err != nil {
return nil, err
}
if err := binary.Write(buf, binary.BigEndian, uint32(len(payloadBytes))); err != nil {
return nil, err
}

// Write payload
if _, err := buf.Write(payloadBytes); err != nil {
return nil, err
}

// Optional signature
if opts != nil && opts.Signer != nil {
// Sign the message (header + payload)
messageData := buf.Bytes()
signature, err := opts.Signer.Sign(messageData)
if err != nil {
return nil, err
}
// Append signature length and signature
if err := binary.Write(buf, binary.BigEndian, uint32(len(signature))); err != nil {
return nil, err
}
if _, err := buf.Write(signature); err != nil {
return nil, err
}
} else {
// No signature - write zero length
if err := binary.Write(buf, binary.BigEndian, uint32(0)); err != nil {
return nil, err
}
}

return buf.Bytes(), nil
}

// UnmarshalMessage deserializes a binary message into its components
func UnmarshalMessage(data []byte, opts *MessageOptions) (*Message, interface{}, error) {
if len(data) < HeaderSize+SignatureLengthSize {
return nil, nil, ErrInvalidMessage
}

r := bytes.NewReader(data)

// Read header
messageType, err := r.ReadByte()
if err != nil {
return nil, nil, err
}

var messageID uint32
if err := binary.Read(r, binary.BigEndian, &messageID); err != nil {
return nil, nil, err
}

var payloadLen uint32
if err := binary.Read(r, binary.BigEndian, &payloadLen); err != nil {
return nil, nil, err
}

if payloadLen > MaxPayloadSize {
return nil, nil, ErrPayloadTooLarge
}

// Read payload
payload := make([]byte, payloadLen)
if _, err := io.ReadFull(r, payload); err != nil {
return nil, nil, err
}

// Read signature length
var sigLen uint32
if err := binary.Read(r, binary.BigEndian, &sigLen); err != nil {
return nil, nil, err
}

var signature []byte
if sigLen > 0 {
signature = make([]byte, sigLen)
if _, err := io.ReadFull(r, signature); err != nil {
return nil, nil, err
}
}

// Verify signature if verifier is provided
if opts != nil && opts.Verifier != nil {
if sigLen == 0 {
return nil, nil, ErrSignatureRequired
}
// Message data for verification is header + payload
messageDataLen := HeaderSize + int(payloadLen)
if err := opts.Verifier.Verify(data[:messageDataLen], signature); err != nil {
return nil, nil, ErrInvalidSignature
}
}

msg := &Message{
Type:      messageType,
ID:        messageID,
Payload:   payload,
Signature: signature,
}

// Deserialize payload using registry if provided, respecting strict mode
var payloadObj interface{}
strictMode := opts != nil && opts.StrictMode
if opts != nil && opts.Registry != nil {
payloadObj, err = UnmarshalPayloadWithRegistryStrict(messageType, payload, opts.Registry, strictMode)
} else {
payloadObj, err = UnmarshalPayloadStrict(messageType, payload, strictMode)
}
if err != nil {
return msg, nil, err
}

return msg, payloadObj, nil
}

// Protocol handles message sending and receiving over a connection
type Protocol struct {
conn         Connection
opts         *MessageOptions
mu           sync.Mutex
readMu       sync.Mutex
nextID       uint32
idMu         sync.Mutex
streamConfig *StreamConfig

// Stream assembly state
streamMu        sync.Mutex
activeStreams   map[uint32]*streamAssembler
}

// streamAssembler collects chunks for a streamed message
type streamAssembler struct {
header   *StreamHeader
chunks   map[uint32][]byte
received uint32
}

// NewProtocol creates a new Protocol instance with the given connection
func NewProtocol(conn Connection, opts *MessageOptions) *Protocol {
streamCfg := DefaultStreamConfig()
if opts != nil && opts.StreamConfig != nil {
streamCfg = opts.StreamConfig
}

return &Protocol{
conn:          conn,
opts:          opts,
nextID:        1,
streamConfig:  streamCfg,
activeStreams: make(map[uint32]*streamAssembler),
}
}

// NextMessageID returns the next message ID and increments the counter
func (p *Protocol) NextMessageID() uint32 {
p.idMu.Lock()
defer p.idMu.Unlock()
id := p.nextID
p.nextID++
return id
}

// SendMessage serializes and sends a message over the connection
// Automatically uses streaming for large payloads
func (p *Protocol) SendMessage(messageType byte, messageID uint32, payload interface{}) error {
// Serialize payload first to check size
payloadBytes, err := MarshalPayload(payload)
if err != nil {
return err
}

// Check if streaming is needed
if p.streamConfig.Enabled && len(payloadBytes) >= p.streamConfig.Threshold {
return p.sendStreamed(messageType, messageID, payloadBytes)
}

return p.sendDirect(messageType, messageID, payload)
}

// sendDirect sends a message without streaming
func (p *Protocol) sendDirect(messageType byte, messageID uint32, payload interface{}) error {
data, err := MarshalMessage(messageType, messageID, payload, p.opts)
if err != nil {
return err
}

p.mu.Lock()
defer p.mu.Unlock()

// Write message length first (for framing)
lenBuf := make([]byte, 4)
binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
if _, err := p.conn.Write(lenBuf); err != nil {
return err
}

// Write message data
_, err = p.conn.Write(data)
return err
}

// sendStreamed sends a large payload as multiple chunks
func (p *Protocol) sendStreamed(messageType byte, messageID uint32, payloadBytes []byte) error {
chunkSize := p.streamConfig.ChunkSize
totalSize := uint64(len(payloadBytes))
totalChunks := uint32((len(payloadBytes) + chunkSize - 1) / chunkSize)

// Send stream start header
header := &StreamHeader{
OriginalType: messageType,
TotalSize:    totalSize,
TotalChunks:  totalChunks,
}
if err := p.sendDirect(MessageTypeStreamStart, messageID, header); err != nil {
return err
}

// Send chunks
for i := uint32(0); i < totalChunks; i++ {
start := int(i) * chunkSize
end := start + chunkSize
if end > len(payloadBytes) {
end = len(payloadBytes)
}

chunk := &StreamChunk{
ChunkIndex: i,
Data:       payloadBytes[start:end],
}
if err := p.sendDirect(MessageTypeStreamChunk, messageID, chunk); err != nil {
return err
}
}

// Send stream end marker
if err := p.sendDirect(MessageTypeStreamEnd, messageID, []byte{}); err != nil {
return err
}

return nil
}

// ReceiveMessage reads and deserializes a message from the connection
// Automatically reassembles streamed messages
func (p *Protocol) ReceiveMessage() (*Message, interface{}, error) {
for {
msg, payload, err := p.receiveRaw()
if err != nil {
return nil, nil, err
}

// Handle streaming messages internally
switch msg.Type {
case MessageTypeStreamStart:
header := payload.(*StreamHeader)
p.startStream(msg.ID, header)
continue

case MessageTypeStreamChunk:
chunk := payload.(*StreamChunk)
complete, assembledPayload := p.addChunk(msg.ID, chunk)
if complete {
// Stream complete, unmarshal the assembled payload
p.streamMu.Lock()
assembler := p.activeStreams[msg.ID]
originalType := assembler.header.OriginalType
delete(p.activeStreams, msg.ID)
p.streamMu.Unlock()

payloadObj, err := UnmarshalPayload(originalType, assembledPayload)
if err != nil {
return nil, nil, err
}

return &Message{
Type:    originalType,
ID:      msg.ID,
Payload: assembledPayload,
}, payloadObj, nil
}
continue

case MessageTypeStreamEnd:
// Stream end marker - check if we have all chunks
p.streamMu.Lock()
assembler, exists := p.activeStreams[msg.ID]
if exists && assembler.received == assembler.header.TotalChunks {
originalType := assembler.header.OriginalType
assembledPayload := p.assembleStream(assembler)
delete(p.activeStreams, msg.ID)
p.streamMu.Unlock()

payloadObj, err := UnmarshalPayload(originalType, assembledPayload)
if err != nil {
return nil, nil, err
}

return &Message{
Type:    originalType,
ID:      msg.ID,
Payload: assembledPayload,
}, payloadObj, nil
}
p.streamMu.Unlock()
continue

default:
return msg, payload, nil
}
}
}

// receiveRaw reads a single raw message from the connection
func (p *Protocol) receiveRaw() (*Message, interface{}, error) {
p.readMu.Lock()
defer p.readMu.Unlock()

// Read message length
lenBuf := make([]byte, 4)
if _, err := io.ReadFull(p.conn, lenBuf); err != nil {
return nil, nil, err
}
msgLen := binary.BigEndian.Uint32(lenBuf)

if msgLen > MaxPayloadSize+HeaderSize+SignatureLengthSize+1024 {
return nil, nil, ErrPayloadTooLarge
}

// Read message data
data := make([]byte, msgLen)
if _, err := io.ReadFull(p.conn, data); err != nil {
return nil, nil, err
}

return UnmarshalMessage(data, p.opts)
}

// startStream initializes a new stream assembler
func (p *Protocol) startStream(msgID uint32, header *StreamHeader) {
p.streamMu.Lock()
defer p.streamMu.Unlock()

p.activeStreams[msgID] = &streamAssembler{
header:   header,
chunks:   make(map[uint32][]byte),
received: 0,
}
}

// addChunk adds a chunk to a stream and returns true if complete
func (p *Protocol) addChunk(msgID uint32, chunk *StreamChunk) (bool, []byte) {
p.streamMu.Lock()
defer p.streamMu.Unlock()

assembler, exists := p.activeStreams[msgID]
if !exists {
return false, nil
}

// Store chunk
assembler.chunks[chunk.ChunkIndex] = chunk.Data
assembler.received++

// Check if complete
if assembler.received == assembler.header.TotalChunks {
return true, p.assembleStream(assembler)
}

return false, nil
}

// assembleStream combines all chunks into the final payload
func (p *Protocol) assembleStream(assembler *streamAssembler) []byte {
result := make([]byte, 0, assembler.header.TotalSize)
for i := uint32(0); i < assembler.header.TotalChunks; i++ {
result = append(result, assembler.chunks[i]...)
}
return result
}

// Send is a convenience method that sends a message with auto-generated ID
func (p *Protocol) Send(messageType byte, payload interface{}) (uint32, error) {
id := p.NextMessageID()
return id, p.SendMessage(messageType, id, payload)
}

// Close closes the underlying connection
func (p *Protocol) Close() error {
return p.conn.Close()
}

// SendRaw sends raw bytes with a custom message type
func (p *Protocol) SendRaw(messageType byte, data []byte) (uint32, error) {
return p.Send(messageType, data)
}

// SetStreamConfig updates the streaming configuration
func (p *Protocol) SetStreamConfig(config *StreamConfig) {
p.streamConfig = config
}

// GetStreamConfig returns the current streaming configuration
func (p *Protocol) GetStreamConfig() *StreamConfig {
return p.streamConfig
}
