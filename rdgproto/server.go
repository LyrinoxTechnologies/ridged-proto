package rdgproto

import (
"net"
"sync"
)

// Listener interface for transport-agnostic server listening
type Listener interface {
Accept() (Connection, error)
Close() error
}

// netListenerAdapter wraps net.Listener to satisfy our Listener interface
type netListenerAdapter struct {
listener net.Listener
}

func (n *netListenerAdapter) Accept() (Connection, error) {
conn, err := n.listener.Accept()
if err != nil {
return nil, err
}
return conn, nil
}

func (n *netListenerAdapter) Close() error {
return n.listener.Close()
}

// ConnectionHandler is called for each new client connection
type ConnectionHandler func(client *Client)

// Server manages client connections and message routing
type Server struct {
listener Listener
opts     *MessageOptions
handler  ConnectionHandler

mu       sync.RWMutex
clients  map[*Client]struct{}
running  bool
done     chan struct{}
}

// NewServer creates a new server with the given listener
// The listener can be a net.Listener (TCP, Unix socket, etc.) or any custom Listener
func NewServer(listener interface{}, opts *MessageOptions) *Server {
var l Listener
switch v := listener.(type) {
case Listener:
l = v
case net.Listener:
l = &netListenerAdapter{listener: v}
default:
panic("listener must implement Listener interface or be a net.Listener")
}

return &Server{
listener: l,
opts:     opts,
clients:  make(map[*Client]struct{}),
done:     make(chan struct{}),
}
}

// SetConnectionHandler sets the handler called for each new connection
func (s *Server) SetConnectionHandler(handler ConnectionHandler) {
s.mu.Lock()
defer s.mu.Unlock()
s.handler = handler
}

// Start begins accepting connections (blocking)
func (s *Server) Start() error {
s.mu.Lock()
if s.running {
s.mu.Unlock()
return nil
}
s.running = true
s.mu.Unlock()

for {
conn, err := s.listener.Accept()
if err != nil {
s.mu.RLock()
running := s.running
s.mu.RUnlock()
if !running {
return nil
}
continue
}

client := NewClient(conn, s.opts)
s.addClient(client)

s.mu.RLock()
handler := s.handler
s.mu.RUnlock()

go func(c *Client) {
defer func() {
s.removeClient(c)
c.Close()
}()
if handler != nil {
handler(c)
}
}(client)
}
}

// StartAsync begins accepting connections (non-blocking)
func (s *Server) StartAsync() {
go s.Start()
}

// Stop stops the server
func (s *Server) Stop() error {
s.mu.Lock()
s.running = false
s.mu.Unlock()

err := s.listener.Close()

// Close all client connections
s.mu.RLock()
clients := make([]*Client, 0, len(s.clients))
for c := range s.clients {
clients = append(clients, c)
}
s.mu.RUnlock()

for _, c := range clients {
c.Close()
}

close(s.done)
return err
}

// addClient adds a client to the server's client list
func (s *Server) addClient(client *Client) {
s.mu.Lock()
defer s.mu.Unlock()
s.clients[client] = struct{}{}
}

// removeClient removes a client from the server's client list
func (s *Server) removeClient(client *Client) {
s.mu.Lock()
defer s.mu.Unlock()
delete(s.clients, client)
}

// ClientCount returns the number of connected clients
func (s *Server) ClientCount() int {
s.mu.RLock()
defer s.mu.RUnlock()
return len(s.clients)
}

// Broadcast sends a message to all connected clients
func (s *Server) Broadcast(messageType byte, payload interface{}) error {
s.mu.RLock()
clients := make([]*Client, 0, len(s.clients))
for c := range s.clients {
clients = append(clients, c)
}
s.mu.RUnlock()

for _, c := range clients {
if _, err := c.Send(messageType, payload); err != nil {
// Log or handle error, but continue broadcasting
continue
}
}
return nil
}

// Done returns a channel that is closed when the server stops
func (s *Server) Done() <-chan struct{} {
return s.done
}
