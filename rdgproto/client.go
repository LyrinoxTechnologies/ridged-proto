package rdgproto

import (
"errors"
"sync"
)

var (
ErrNotConnected = errors.New("not connected")
ErrClosed       = errors.New("connection closed")
)

// MessageHandler is called when a message is received
type MessageHandler func(msg *Message, payload interface{}) error

// Client provides a high-level API for sending and receiving messages
type Client struct {
proto   *Protocol
opts    *MessageOptions
handler MessageHandler

mu       sync.RWMutex
running  bool
done     chan struct{}
errChan  chan error
}

// NewClient creates a new client with the given connection
func NewClient(conn Connection, opts *MessageOptions) *Client {
return &Client{
proto:   NewProtocol(conn, opts),
opts:    opts,
done:    make(chan struct{}),
errChan: make(chan error, 1),
}
}

// SetHandler sets the message handler for incoming messages
func (c *Client) SetHandler(handler MessageHandler) {
c.mu.Lock()
defer c.mu.Unlock()
c.handler = handler
}

// Start begins listening for incoming messages in a goroutine
// Messages are dispatched to the registered handler
func (c *Client) Start() error {
c.mu.Lock()
if c.running {
c.mu.Unlock()
return nil
}
c.running = true
c.mu.Unlock()

go c.listen()
return nil
}

// listen processes incoming messages
func (c *Client) listen() {
defer func() {
c.mu.Lock()
c.running = false
c.mu.Unlock()
close(c.done)
}()

for {
msg, payload, err := c.proto.ReceiveMessage()
if err != nil {
select {
case c.errChan <- err:
default:
}
return
}

c.mu.RLock()
handler := c.handler
c.mu.RUnlock()

if handler != nil {
if err := handler(msg, payload); err != nil {
select {
case c.errChan <- err:
default:
}
}
}
}
}

// Send sends a message with the specified type and payload
// Returns the message ID assigned to this message
func (c *Client) Send(messageType byte, payload interface{}) (uint32, error) {
return c.proto.Send(messageType, payload)
}

// SendWithID sends a message with a specific message ID
func (c *Client) SendWithID(messageType byte, messageID uint32, payload interface{}) error {
return c.proto.SendMessage(messageType, messageID, payload)
}

// SendRaw sends raw bytes with the specified message type
func (c *Client) SendRaw(messageType byte, data []byte) (uint32, error) {
return c.proto.SendRaw(messageType, data)
}

// Wait blocks until the client is closed or an error occurs
// Returns the error that caused the client to stop, or nil if closed cleanly
func (c *Client) Wait() error {
select {
case err := <-c.errChan:
return err
case <-c.done:
return nil
}
}

// Close closes the connection and stops the client
func (c *Client) Close() error {
return c.proto.Close()
}

// Errors returns a channel that receives errors from the listener
func (c *Client) Errors() <-chan error {
return c.errChan
}

// Done returns a channel that is closed when the client stops
func (c *Client) Done() <-chan struct{} {
return c.done
}

// Protocol returns the underlying Protocol for advanced usage
func (c *Client) Protocol() *Protocol {
return c.proto
}
