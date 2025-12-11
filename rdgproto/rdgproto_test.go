package rdgproto

import (
"bytes"
"encoding/binary"
"net"
"testing"
)

// =========================================================================
// Example custom payload types (as a developer would define them)
// =========================================================================

const (
MsgTypeLogin    byte = 1
MsgTypeResponse byte = 2
MsgTypeData     byte = 3
)

// LoginPayload is an example custom payload
type LoginPayload struct {
Username string
Password string
ClientID string
}

func (p *LoginPayload) Marshal() ([]byte, error) {
buf := new(bytes.Buffer)
if err := WriteString(buf, p.Username); err != nil {
return nil, err
}
if err := WriteString(buf, p.Password); err != nil {
return nil, err
}
if err := WriteString(buf, p.ClientID); err != nil {
return nil, err
}
return buf.Bytes(), nil
}

func (p *LoginPayload) Unmarshal(data []byte) error {
r := bytes.NewReader(data)
var err error
if p.Username, err = ReadString(r); err != nil {
return err
}
if p.Password, err = ReadString(r); err != nil {
return err
}
if p.ClientID, err = ReadString(r); err != nil {
return err
}
return nil
}

// ResponsePayload is an example response payload
type ResponsePayload struct {
Success bool
Message string
}

func (p *ResponsePayload) Marshal() ([]byte, error) {
buf := new(bytes.Buffer)
if err := WriteBool(buf, p.Success); err != nil {
return nil, err
}
if err := WriteString(buf, p.Message); err != nil {
return nil, err
}
return buf.Bytes(), nil
}

func (p *ResponsePayload) Unmarshal(data []byte) error {
r := bytes.NewReader(data)
var err error
if p.Success, err = ReadBool(r); err != nil {
return err
}
if p.Message, err = ReadString(r); err != nil {
return err
}
return nil
}

// DataPayload is an example data transfer payload
type DataPayload struct {
ID         string
ChunkIndex uint32
TotalChunks uint32
Data       []byte
}

func (p *DataPayload) Marshal() ([]byte, error) {
buf := new(bytes.Buffer)
if err := WriteString(buf, p.ID); err != nil {
return nil, err
}
if err := binary.Write(buf, binary.BigEndian, p.ChunkIndex); err != nil {
return nil, err
}
if err := binary.Write(buf, binary.BigEndian, p.TotalChunks); err != nil {
return nil, err
}
if err := WriteBytes(buf, p.Data); err != nil {
return nil, err
}
return buf.Bytes(), nil
}

func (p *DataPayload) Unmarshal(data []byte) error {
r := bytes.NewReader(data)
var err error
if p.ID, err = ReadString(r); err != nil {
return err
}
if err = binary.Read(r, binary.BigEndian, &p.ChunkIndex); err != nil {
return err
}
if err = binary.Read(r, binary.BigEndian, &p.TotalChunks); err != nil {
return err
}
if p.Data, err = ReadBytes(r); err != nil {
return err
}
return nil
}

// Register our test payloads
func init() {
RegisterPayloadType(MsgTypeLogin, func() PayloadUnmarshaler { return &LoginPayload{} })
RegisterPayloadType(MsgTypeResponse, func() PayloadUnmarshaler { return &ResponsePayload{} })
RegisterPayloadType(MsgTypeData, func() PayloadUnmarshaler { return &DataPayload{} })
}

// =========================================================================
// Tests
// =========================================================================

func TestCustomPayloadMarshalUnmarshal(t *testing.T) {
original := &LoginPayload{
Username: "testuser",
Password: "testpass123",
ClientID: "client-001",
}

data, err := original.Marshal()
if err != nil {
t.Fatalf("Marshal failed: %v", err)
}

decoded := &LoginPayload{}
if err := decoded.Unmarshal(data); err != nil {
t.Fatalf("Unmarshal failed: %v", err)
}

if decoded.Username != original.Username {
t.Errorf("Username mismatch: got %s, want %s", decoded.Username, original.Username)
}
if decoded.Password != original.Password {
t.Errorf("Password mismatch: got %s, want %s", decoded.Password, original.Password)
}
if decoded.ClientID != original.ClientID {
t.Errorf("ClientID mismatch: got %s, want %s", decoded.ClientID, original.ClientID)
}
}

func TestMarshalUnmarshalMessage(t *testing.T) {
payload := &LoginPayload{
Username: "testuser",
Password: "secret",
ClientID: "client-123",
}

data, err := MarshalMessage(MsgTypeLogin, 42, payload, nil)
if err != nil {
t.Fatalf("MarshalMessage failed: %v", err)
}

msg, payloadObj, err := UnmarshalMessage(data, nil)
if err != nil {
t.Fatalf("UnmarshalMessage failed: %v", err)
}

if msg.Type != MsgTypeLogin {
t.Errorf("Type mismatch: got %d, want %d", msg.Type, MsgTypeLogin)
}
if msg.ID != 42 {
t.Errorf("ID mismatch: got %d, want %d", msg.ID, 42)
}

decoded, ok := payloadObj.(*LoginPayload)
if !ok {
t.Fatalf("Payload type assertion failed, got %T", payloadObj)
}
if decoded.Username != payload.Username {
t.Errorf("Username mismatch: got %s, want %s", decoded.Username, payload.Username)
}
}

func TestMarshalUnmarshalMessageWithHMAC(t *testing.T) {
secret := []byte("super-secret-key")
opts := SecureMessageOptions(secret)

payload := &ResponsePayload{
Success: true,
Message: "Operation completed",
}

data, err := MarshalMessage(MsgTypeResponse, 100, payload, opts)
if err != nil {
t.Fatalf("MarshalMessage failed: %v", err)
}

msg, payloadObj, err := UnmarshalMessage(data, opts)
if err != nil {
t.Fatalf("UnmarshalMessage failed: %v", err)
}

if msg.Type != MsgTypeResponse {
t.Errorf("Type mismatch: got %d, want %d", msg.Type, MsgTypeResponse)
}
if len(msg.Signature) == 0 {
t.Error("Expected signature to be present")
}

decoded, ok := payloadObj.(*ResponsePayload)
if !ok {
t.Fatalf("Payload type assertion failed, got %T", payloadObj)
}
if decoded.Success != payload.Success {
t.Errorf("Success mismatch: got %v, want %v", decoded.Success, payload.Success)
}
}

func TestMarshalUnmarshalMessageWithRSA(t *testing.T) {
privateKey, publicKey, err := GenerateRSAKeyPair(2048)
if err != nil {
t.Fatalf("Failed to generate RSA key pair: %v", err)
}

signerOpts := RSAMessageOptions(privateKey, nil)
verifierOpts := RSAMessageOptions(nil, publicKey)

payload := &ResponsePayload{
Success: false,
Message: "Internal error",
}

data, err := MarshalMessage(MsgTypeResponse, 200, payload, signerOpts)
if err != nil {
t.Fatalf("MarshalMessage failed: %v", err)
}

msg, payloadObj, err := UnmarshalMessage(data, verifierOpts)
if err != nil {
t.Fatalf("UnmarshalMessage failed: %v", err)
}

if msg.Type != MsgTypeResponse {
t.Errorf("Type mismatch: got %d, want %d", msg.Type, MsgTypeResponse)
}
if len(msg.Signature) == 0 {
t.Error("Expected signature to be present")
}

decoded, ok := payloadObj.(*ResponsePayload)
if !ok {
t.Fatalf("Payload type assertion failed, got %T", payloadObj)
}
if decoded.Message != payload.Message {
t.Errorf("Message mismatch: got %s, want %s", decoded.Message, payload.Message)
}
}

func TestInvalidSignature(t *testing.T) {
secret1 := []byte("secret-1")
secret2 := []byte("secret-2")
opts1 := SecureMessageOptions(secret1)
opts2 := SecureMessageOptions(secret2)

payload := &LoginPayload{
Username: "user",
Password: "pass",
ClientID: "client",
}

data, err := MarshalMessage(MsgTypeLogin, 1, payload, opts1)
if err != nil {
t.Fatalf("MarshalMessage failed: %v", err)
}

_, _, err = UnmarshalMessage(data, opts2)
if err != ErrInvalidSignature {
t.Errorf("Expected ErrInvalidSignature, got: %v", err)
}
}

func TestProtocolSendReceive(t *testing.T) {
listener, err := net.Listen("tcp", "127.0.0.1:0")
if err != nil {
t.Fatalf("Failed to create listener: %v", err)
}
defer listener.Close()

done := make(chan error, 1)
var receivedMsg *Message
var receivedPayload interface{}

go func() {
conn, err := listener.Accept()
if err != nil {
done <- err
return
}
defer conn.Close()

proto := NewProtocol(conn, nil)
receivedMsg, receivedPayload, err = proto.ReceiveMessage()
done <- err
}()

clientConn, err := net.Dial("tcp", listener.Addr().String())
if err != nil {
t.Fatalf("Failed to connect: %v", err)
}
defer clientConn.Close()

clientProto := NewProtocol(clientConn, nil)
_, err = clientProto.Send(MsgTypeLogin, &LoginPayload{
Username: "testuser",
Password: "testpass",
ClientID: "client-001",
})
if err != nil {
t.Fatalf("Send failed: %v", err)
}

if err := <-done; err != nil {
t.Fatalf("Server receive failed: %v", err)
}

if receivedMsg.Type != MsgTypeLogin {
t.Errorf("Type mismatch: got %d, want %d", receivedMsg.Type, MsgTypeLogin)
}

payload, ok := receivedPayload.(*LoginPayload)
if !ok {
t.Fatalf("Payload type assertion failed, got %T", receivedPayload)
}
if payload.Username != "testuser" {
t.Errorf("Username mismatch: got %s, want testuser", payload.Username)
}
}

func TestProtocolSendReceiveWithHMAC(t *testing.T) {
secret := []byte("shared-secret")
opts := SecureMessageOptions(secret)

listener, err := net.Listen("tcp", "127.0.0.1:0")
if err != nil {
t.Fatalf("Failed to create listener: %v", err)
}
defer listener.Close()

done := make(chan error, 1)
var receivedMsg *Message
var receivedPayload interface{}

go func() {
conn, err := listener.Accept()
if err != nil {
done <- err
return
}
defer conn.Close()

proto := NewProtocol(conn, opts)
receivedMsg, receivedPayload, err = proto.ReceiveMessage()
done <- err
}()

clientConn, err := net.Dial("tcp", listener.Addr().String())
if err != nil {
t.Fatalf("Failed to connect: %v", err)
}
defer clientConn.Close()

clientProto := NewProtocol(clientConn, opts)
_, err = clientProto.Send(MsgTypeResponse, &ResponsePayload{
Success: true,
Message: "Hello",
})
if err != nil {
t.Fatalf("Send failed: %v", err)
}

if err := <-done; err != nil {
t.Fatalf("Server receive failed: %v", err)
}

if receivedMsg.Type != MsgTypeResponse {
t.Errorf("Type mismatch: got %d, want %d", receivedMsg.Type, MsgTypeResponse)
}
if len(receivedMsg.Signature) == 0 {
t.Error("Expected signature to be present")
}

payload, ok := receivedPayload.(*ResponsePayload)
if !ok {
t.Fatalf("Payload type assertion failed, got %T", receivedPayload)
}
if payload.Message != "Hello" {
t.Errorf("Message mismatch: got %s, want Hello", payload.Message)
}
}

func TestGenericMarshalUnmarshal(t *testing.T) {
payload := &LoginPayload{
Username: "testuser",
Password: "testpass",
ClientID: "client-123",
}

// Marshal
data, err := Marshal(MsgTypeLogin, payload)
if err != nil {
t.Fatalf("Marshal failed: %v", err)
}

// Unmarshal
msg, payloadObj, err := Unmarshal(data)
if err != nil {
t.Fatalf("Unmarshal failed: %v", err)
}

if msg.Type != MsgTypeLogin {
t.Errorf("Type mismatch: got %d, want %d", msg.Type, MsgTypeLogin)
}

decoded, ok := payloadObj.(*LoginPayload)
if !ok {
t.Fatalf("Payload type assertion failed, got %T", payloadObj)
}
if decoded.Username != payload.Username {
t.Errorf("Username mismatch: got %s, want %s", decoded.Username, payload.Username)
}
}

func TestMarshalWithID(t *testing.T) {
payload := &ResponsePayload{
Success: true,
Message: "OK",
}

data, err := MarshalWithID(MsgTypeResponse, 42, payload)
if err != nil {
t.Fatalf("MarshalWithID failed: %v", err)
}

msg, _, err := Unmarshal(data)
if err != nil {
t.Fatalf("Unmarshal failed: %v", err)
}

if msg.ID != 42 {
t.Errorf("ID mismatch: got %d, want 42", msg.ID)
}
}

func TestMarshalSecure(t *testing.T) {
secret := []byte("my-secret-key")
payload := &ResponsePayload{
Success: true,
Message: "Secured",
}

data, err := MarshalSecure(MsgTypeResponse, payload, secret)
if err != nil {
t.Fatalf("MarshalSecure failed: %v", err)
}

msg, payloadObj, err := UnmarshalSecure(data, secret)
if err != nil {
t.Fatalf("UnmarshalSecure failed: %v", err)
}

if msg.Type != MsgTypeResponse {
t.Errorf("Type mismatch: got %d, want %d", msg.Type, MsgTypeResponse)
}
if len(msg.Signature) == 0 {
t.Error("Expected signature to be present")
}

decoded, ok := payloadObj.(*ResponsePayload)
if !ok {
t.Fatalf("Payload type assertion failed, got %T", payloadObj)
}
if decoded.Message != payload.Message {
t.Errorf("Message mismatch: got %s, want %s", decoded.Message, payload.Message)
}
}

func TestUnmarshalInto(t *testing.T) {
original := &LoginPayload{
Username: "user",
Password: "pass",
ClientID: "client",
}

data, err := Marshal(MsgTypeLogin, original)
if err != nil {
t.Fatalf("Marshal failed: %v", err)
}

var target LoginPayload
msg, err := UnmarshalInto(data, &target)
if err != nil {
t.Fatalf("UnmarshalInto failed: %v", err)
}

if msg.Type != MsgTypeLogin {
t.Errorf("Type mismatch: got %d, want %d", msg.Type, MsgTypeLogin)
}
if target.Username != original.Username {
t.Errorf("Username mismatch: got %s, want %s", target.Username, original.Username)
}
}

func TestRawBytesPayload(t *testing.T) {
rawPayload := []byte{0xDE, 0xAD, 0xBE, 0xEF}

// Use an unregistered message type
const UnknownType byte = 200
data, err := MarshalMessage(UnknownType, 1, rawPayload, nil)
if err != nil {
t.Fatalf("MarshalMessage with raw bytes failed: %v", err)
}

msg, payloadObj, err := UnmarshalMessage(data, nil)
if err != nil {
t.Fatalf("UnmarshalMessage failed: %v", err)
}

if !bytes.Equal(msg.Payload, rawPayload) {
t.Errorf("Payload mismatch: got %v, want %v", msg.Payload, rawPayload)
}

// For unknown message types, payloadObj should be the raw bytes
rawBytes, ok := payloadObj.([]byte)
if !ok {
t.Fatalf("Expected raw bytes payload, got %T", payloadObj)
}
if !bytes.Equal(rawBytes, rawPayload) {
t.Errorf("Payload object mismatch: got %v, want %v", rawBytes, rawPayload)
}
}

func TestStrictModeRejectsUnknownTypes(t *testing.T) {
const UnknownType byte = 200
rawData := []byte{0xDE, 0xAD, 0xBE, 0xEF}
data, err := Marshal(UnknownType, rawData)
if err != nil {
t.Fatalf("Marshal failed: %v", err)
}

// Non-strict mode should accept it
msg, payload, err := Unmarshal(data)
if err != nil {
t.Fatalf("Non-strict Unmarshal should succeed: %v", err)
}
if msg.Type != UnknownType {
t.Errorf("Type mismatch: got %d, want %d", msg.Type, UnknownType)
}
_, ok := payload.([]byte)
if !ok {
t.Error("Non-strict mode should return raw bytes for unknown types")
}

// Strict mode should reject it
_, _, err = UnmarshalStrict(data)
if err != ErrUnknownMessageType {
t.Errorf("Strict mode should reject unknown types, got: %v", err)
}
}

func TestStrictModeAcceptsKnownTypes(t *testing.T) {
payload := &LoginPayload{
Username: "user",
Password: "pass",
ClientID: "client",
}
data, err := Marshal(MsgTypeLogin, payload)
if err != nil {
t.Fatalf("Marshal failed: %v", err)
}

msg, payloadObj, err := UnmarshalStrict(data)
if err != nil {
t.Fatalf("Strict Unmarshal failed for known type: %v", err)
}
if msg.Type != MsgTypeLogin {
t.Errorf("Type mismatch: got %d, want %d", msg.Type, MsgTypeLogin)
}

decoded, ok := payloadObj.(*LoginPayload)
if !ok {
t.Fatalf("Payload type assertion failed, got %T", payloadObj)
}
if decoded.Username != "user" {
t.Errorf("Username mismatch: got %s, want user", decoded.Username)
}
}

func TestStreamingLargePayload(t *testing.T) {
// Create a 2MB payload
largeData := make([]byte, 2*1024*1024)
for i := range largeData {
largeData[i] = byte(i % 256)
}

listener, err := net.Listen("tcp", "127.0.0.1:0")
if err != nil {
t.Fatalf("Failed to create listener: %v", err)
}
defer listener.Close()

done := make(chan error, 1)
var receivedMsg *Message
var receivedPayload interface{}

go func() {
conn, err := listener.Accept()
if err != nil {
done <- err
return
}
defer conn.Close()

proto := NewProtocol(conn, nil)
receivedMsg, receivedPayload, err = proto.ReceiveMessage()
done <- err
}()

clientConn, err := net.Dial("tcp", listener.Addr().String())
if err != nil {
t.Fatalf("Failed to connect: %v", err)
}
defer clientConn.Close()

streamCfg := &StreamConfig{
Threshold: 1024 * 1024, // 1MB threshold
ChunkSize: 64 * 1024,   // 64KB chunks
Enabled:   true,
}
opts := &MessageOptions{StreamConfig: streamCfg}
clientProto := NewProtocol(clientConn, opts)

// Send large payload
payload := &DataPayload{
ID:          "large-data",
ChunkIndex:  0,
TotalChunks: 1,
Data:        largeData,
}
_, err = clientProto.Send(MsgTypeData, payload)
if err != nil {
t.Fatalf("Send failed: %v", err)
}

if err := <-done; err != nil {
t.Fatalf("Server receive failed: %v", err)
}

if receivedMsg.Type != MsgTypeData {
t.Errorf("Type mismatch: got %d, want %d", receivedMsg.Type, MsgTypeData)
}

recvPayload, ok := receivedPayload.(*DataPayload)
if !ok {
t.Fatalf("Payload type assertion failed, got %T", receivedPayload)
}
if len(recvPayload.Data) != len(largeData) {
t.Errorf("Data length mismatch: got %d, want %d", len(recvPayload.Data), len(largeData))
}
if !bytes.Equal(recvPayload.Data, largeData) {
t.Error("Data content mismatch")
}
}

func TestStreamHeader(t *testing.T) {
original := &StreamHeader{
OriginalType: MsgTypeData,
TotalSize:    1024 * 1024 * 10,
TotalChunks:  160,
}

data, err := original.Marshal()
if err != nil {
t.Fatalf("Marshal failed: %v", err)
}

decoded := &StreamHeader{}
if err := decoded.Unmarshal(data); err != nil {
t.Fatalf("Unmarshal failed: %v", err)
}

if decoded.OriginalType != original.OriginalType {
t.Errorf("OriginalType mismatch: got %d, want %d", decoded.OriginalType, original.OriginalType)
}
if decoded.TotalSize != original.TotalSize {
t.Errorf("TotalSize mismatch: got %d, want %d", decoded.TotalSize, original.TotalSize)
}
if decoded.TotalChunks != original.TotalChunks {
t.Errorf("TotalChunks mismatch: got %d, want %d", decoded.TotalChunks, original.TotalChunks)
}
}

func TestStreamChunk(t *testing.T) {
original := &StreamChunk{
ChunkIndex: 42,
Data:       []byte{0xDE, 0xAD, 0xBE, 0xEF},
}

data, err := original.Marshal()
if err != nil {
t.Fatalf("Marshal failed: %v", err)
}

decoded := &StreamChunk{}
if err := decoded.Unmarshal(data); err != nil {
t.Fatalf("Unmarshal failed: %v", err)
}

if decoded.ChunkIndex != original.ChunkIndex {
t.Errorf("ChunkIndex mismatch: got %d, want %d", decoded.ChunkIndex, original.ChunkIndex)
}
if !bytes.Equal(decoded.Data, original.Data) {
t.Errorf("Data mismatch: got %v, want %v", decoded.Data, original.Data)
}
}

func TestIsReservedType(t *testing.T) {
if IsReservedType(1) {
t.Error("Type 1 should not be reserved")
}
if IsReservedType(249) {
t.Error("Type 249 should not be reserved")
}
if !IsReservedType(250) {
t.Error("Type 250 should be reserved")
}
if !IsReservedType(255) {
t.Error("Type 255 should be reserved")
}
}

// Benchmarks

func BenchmarkMarshalMessage(b *testing.B) {
payload := &LoginPayload{
Username: "testuser",
Password: "testpassword123",
ClientID: "client-abc-123",
}

b.ResetTimer()
for i := 0; i < b.N; i++ {
_, err := MarshalMessage(MsgTypeLogin, uint32(i), payload, nil)
if err != nil {
b.Fatal(err)
}
}
}

func BenchmarkUnmarshalMessage(b *testing.B) {
payload := &LoginPayload{
Username: "testuser",
Password: "testpassword123",
ClientID: "client-abc-123",
}
data, _ := MarshalMessage(MsgTypeLogin, 1, payload, nil)

b.ResetTimer()
for i := 0; i < b.N; i++ {
_, _, err := UnmarshalMessage(data, nil)
if err != nil {
b.Fatal(err)
}
}
}

func BenchmarkMarshalMessageWithHMAC(b *testing.B) {
secret := []byte("benchmark-secret-key")
opts := SecureMessageOptions(secret)
payload := &LoginPayload{
Username: "testuser",
Password: "testpassword123",
ClientID: "client-abc-123",
}

b.ResetTimer()
for i := 0; i < b.N; i++ {
_, err := MarshalMessage(MsgTypeLogin, uint32(i), payload, opts)
if err != nil {
b.Fatal(err)
}
}
}

func BenchmarkRoundTrip(b *testing.B) {
payload := &LoginPayload{
Username: "testuser",
Password: "testpassword123",
ClientID: "client-abc-123",
}

b.ResetTimer()
b.ReportAllocs()
for i := 0; i < b.N; i++ {
data, err := MarshalMessage(MsgTypeLogin, uint32(i), payload, nil)
if err != nil {
b.Fatal(err)
}
_, _, err = UnmarshalMessage(data, nil)
if err != nil {
b.Fatal(err)
}
}
}

func BenchmarkRoundTripWithHMAC(b *testing.B) {
secret := []byte("benchmark-secret")
opts := SecureMessageOptions(secret)
payload := &LoginPayload{
Username: "testuser",
Password: "testpassword123",
ClientID: "client-abc-123",
}

b.ResetTimer()
b.ReportAllocs()
for i := 0; i < b.N; i++ {
data, err := MarshalMessage(MsgTypeLogin, uint32(i), payload, opts)
if err != nil {
b.Fatal(err)
}
_, _, err = UnmarshalMessage(data, opts)
if err != nil {
b.Fatal(err)
}
}
}

func BenchmarkParallel(b *testing.B) {
payload := &LoginPayload{
Username: "testuser",
Password: "testpassword123",
ClientID: "client-abc-123",
}

b.RunParallel(func(pb *testing.PB) {
i := uint32(0)
for pb.Next() {
data, err := MarshalMessage(MsgTypeLogin, i, payload, nil)
if err != nil {
b.Fatal(err)
}
_, _, err = UnmarshalMessage(data, nil)
if err != nil {
b.Fatal(err)
}
i++
}
})
}
