# ridged-proto

A binary-only protocol **framework** in Go for building custom communication protocols between different client platforms (CLI, Web apps, etc.) and a central server. rdgproto provides the protocol infrastructure - you define your own message types and payloads.

## Features

- **Pure Binary Protocol**: Exclusively uses binary serialization for compact, efficient data transfer
- **Framework Architecture**: You define your own message types - rdgproto handles the rest
- **Transport-Agnostic**: Works with any transport layer (WebSocket, TCP, HTTP, etc.)
- **Automatic Streaming**: Large payloads (≥1MB) are automatically chunked and streamed
- **Serialization Helpers**: Built-in helpers for strings, bytes, integers, booleans
- **Optional Security**: HMAC-SHA256 and RSA-SHA256 cryptographic signing
- **Strict Mode**: Reject unknown message types for security
- **High-Level API**: Simple Client/Server abstractions

## Installation

```bash
go get github.com/LyrinoxTechnologies/ridged-proto/rdgproto
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Your Application                     │
│         (Your message types & payload structs)          │
├─────────────────────────────────────────────────────────┤
│                   rdgproto Framework                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │
│  │   Client    │  │   Server    │  │    Protocol     │  │
│  │   (send/    │  │  (listener/ │  │ (marshal/parse) │  │
│  │   receive)  │  │   handler)  │  │   + streaming   │  │
│  └─────────────┘  └─────────────┘  └─────────────────┘  │
├─────────────────────────────────────────────────────────┤
│           Your Transport (TCP, WebSocket, etc.)         │
│       Implements rdgproto.Connection interface          │
└─────────────────────────────────────────────────────────┘
```

## Quick Start

### 1. Define Your Message Types

```go
package myprotocol

import (
    "bytes"
    "github.com/LyrinoxTechnologies/ridged-proto/rdgproto"
)

// Define your message type constants
const (
    MsgTypeLogin    byte = 1
    MsgTypeResponse byte = 2
    MsgTypeData     byte = 3
)

// Define your payload structs
type LoginPayload struct {
    Username string
    Password string
    ClientID string
}

// Implement PayloadMarshaler - serialize to binary
func (p *LoginPayload) Marshal() ([]byte, error) {
    buf := new(bytes.Buffer)
    rdgproto.WriteString(buf, p.Username)
    rdgproto.WriteString(buf, p.Password)
    rdgproto.WriteString(buf, p.ClientID)
    return buf.Bytes(), nil
}

// Implement PayloadUnmarshaler - deserialize from binary
func (p *LoginPayload) Unmarshal(data []byte) error {
    r := bytes.NewReader(data)
    var err error
    p.Username, err = rdgproto.ReadString(r)
    if err != nil { return err }
    p.Password, err = rdgproto.ReadString(r)
    if err != nil { return err }
    p.ClientID, err = rdgproto.ReadString(r)
    return err
}
```

### 2. Register Your Payload Types

```go
func init() {
    rdgproto.RegisterPayloadType(MsgTypeLogin, func() rdgproto.PayloadUnmarshaler {
        return &LoginPayload{}
    })
    rdgproto.RegisterPayloadType(MsgTypeResponse, func() rdgproto.PayloadUnmarshaler {
        return &ResponsePayload{}
    })
}
```

### 3. Marshal and Unmarshal Messages

```go
// Marshal a message to binary
data, err := rdgproto.Marshal(MsgTypeLogin, &LoginPayload{
    Username: "user",
    Password: "pass",
    ClientID: "client-1",
})

// Unmarshal binary data back to a message
msg, payload, err := rdgproto.Unmarshal(data)
if err != nil {
    // handle error
}
login := payload.(*LoginPayload)
fmt.Printf("Username: %s\n", login.Username)
```

### 4. Use With a Transport

```go
// Server side
listener, _ := net.Listen("tcp", ":8080")
server := rdgproto.NewServer(listener, nil)

server.SetConnectionHandler(func(client *rdgproto.Client) {
    client.SetHandler(func(msg *rdgproto.Message, payload interface{}) error {
        switch msg.Type {
        case MsgTypeLogin:
            login := payload.(*LoginPayload)
            // Handle login...
            client.Send(MsgTypeResponse, &ResponsePayload{Success: true})
        }
        return nil
    })
    client.Start()
    client.Wait()
})

server.Start()

// Client side
conn, _ := net.Dial("tcp", "localhost:8080")
client := rdgproto.NewClient(conn, nil)

client.SetHandler(func(msg *rdgproto.Message, payload interface{}) error {
    switch msg.Type {
    case MsgTypeResponse:
        resp := payload.(*ResponsePayload)
        fmt.Printf("Response: %v\n", resp.Success)
    }
    return nil
})

client.Start()
client.Send(MsgTypeLogin, &LoginPayload{Username: "user", Password: "pass"})
client.Wait()
```

## Serialization Helpers

rdgproto provides helper functions for common binary serialization tasks:

```go
buf := new(bytes.Buffer)

// Strings (length-prefixed)
rdgproto.WriteString(buf, "hello")
str, _ := rdgproto.ReadString(reader)

// Byte slices (length-prefixed)
rdgproto.WriteBytes(buf, []byte{0x01, 0x02})
data, _ := rdgproto.ReadBytes(reader)

// Integers (big-endian)
rdgproto.WriteUint32(buf, 42)
n, _ := rdgproto.ReadUint32(reader)

rdgproto.WriteUint64(buf, 123456789)
n64, _ := rdgproto.ReadUint64(reader)

// Booleans
rdgproto.WriteBool(buf, true)
b, _ := rdgproto.ReadBool(reader)
```

## Security Features

### HMAC Signing

```go
secret := []byte("shared-secret-key")

// Marshal with signature
data, err := rdgproto.MarshalSecure(MsgTypeLogin, payload, secret)

// Unmarshal and verify signature
msg, payload, err := rdgproto.UnmarshalSecure(data, secret)
if err == rdgproto.ErrInvalidSignature {
    // signature verification failed
}
```

### Strict Mode (Reject Unknown Message Types)

For production use, enable strict mode to reject unregistered message types:

```go
// Strict mode rejects messages with unregistered types
msg, payload, err := rdgproto.UnmarshalStrict(data)
if err == rdgproto.ErrUnknownMessageType {
    // Unknown message type - rejected for security
}

// Combine strict mode with HMAC signing
opts := rdgproto.SecureStrictMessageOptions(secret)
server := rdgproto.NewServer(listener, opts)
```

### RSA Signing

```go
privateKey, publicKey, _ := rdgproto.GenerateRSAKeyPair(2048)

// Sign with private key
signerOpts := rdgproto.RSAMessageOptions(privateKey, nil)
data, _ := rdgproto.MarshalMessage(MsgType, 1, payload, signerOpts)

// Verify with public key
verifierOpts := rdgproto.RSAMessageOptions(nil, publicKey)
msg, payload, _ := rdgproto.UnmarshalMessage(data, verifierOpts)
```

## Message Format

Messages use the following binary structure:

```
+---------------------------+
| Message Type (1 byte)     |
+---------------------------+
| Message ID (4 bytes)      |
+---------------------------+
| Payload Length (4 bytes)  |
+---------------------------+
| Payload (variable)        |
+---------------------------+
| Signature Length (4 bytes)|
+---------------------------+
| Signature (variable)      |
+---------------------------+
```

## Reserved Message Types

Types 250-255 are reserved for internal streaming protocol. Do not use these for your message types:

| Type | Usage |
|------|-------|
| 1-249 | Available for your message types |
| 250 | Reserved: Stream Start |
| 251 | Reserved: Stream Chunk |
| 252 | Reserved: Stream End |
| 253-255 | Reserved for future use |

Use `rdgproto.IsReservedType(t)` to check if a type is reserved.

## Automatic Streaming

Large payloads (≥1MB by default) are automatically chunked and streamed:

```go
streamCfg := &rdgproto.StreamConfig{
    Threshold: 1024 * 1024,  // 1MB threshold
    ChunkSize: 64 * 1024,    // 64KB chunks
    Enabled:   true,
}
opts := &rdgproto.MessageOptions{StreamConfig: streamCfg}
client := rdgproto.NewClient(conn, opts)

// Large payloads are automatically chunked on send
// and reassembled on receive - transparent to your code
client.Send(MsgTypeData, &DataPayload{Data: largeData})
```

## Interfaces

### Connection Interface

Your transport must implement this:

```go
type Connection interface {
    io.Reader  // Read([]byte) (int, error)
    io.Writer  // Write([]byte) (int, error)
    io.Closer  // Close() error
}
```

### Payload Interfaces

Your payload structs must implement:

```go
type PayloadMarshaler interface {
    Marshal() ([]byte, error)
}

type PayloadUnmarshaler interface {
    Unmarshal(data []byte) error
}
```

## API Reference

### Functions

```go
// Simple marshaling
rdgproto.Marshal(messageType, payload) ([]byte, error)
rdgproto.MarshalWithID(messageType, msgID, payload) ([]byte, error)
rdgproto.MarshalSecure(messageType, payload, secret) ([]byte, error)

// Simple unmarshaling
rdgproto.Unmarshal(data) (*Message, interface{}, error)
rdgproto.UnmarshalSecure(data, secret) (*Message, interface{}, error)
rdgproto.UnmarshalStrict(data) (*Message, interface{}, error)
rdgproto.UnmarshalInto(data, &target) (*Message, error)

// Registration
rdgproto.RegisterPayloadType(msgType, factory)
rdgproto.UnregisterPayloadType(msgType)
rdgproto.HasPayloadType(msgType) bool
rdgproto.IsReservedType(msgType) bool
```

### Client

```go
client := rdgproto.NewClient(conn, opts)
client.SetHandler(func(msg, payload) error { ... })
client.Start()
client.Send(messageType, payload) (msgID, error)
client.SendWithID(messageType, msgID, payload) error
client.SendRaw(messageType, data) (msgID, error)
client.Wait() error
client.Close() error
client.Done() <-chan struct{}
client.Errors() <-chan error
```

### Server

```go
server := rdgproto.NewServer(listener, opts)
server.SetConnectionHandler(func(client) { ... })
server.Start()             // blocking
server.StartAsync()        // non-blocking
server.Stop()
server.ClientCount() int
server.Broadcast(messageType, payload)
server.Done() <-chan struct{}
```

## License

See [LICENSE](LICENSE) file.
>>>>>>> 92e41a8 (rdgproto v 0.1)
