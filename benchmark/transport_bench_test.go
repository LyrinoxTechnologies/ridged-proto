// transport_bench_test.go
package benchmark

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/LyrinoxTechnologies/ridged-proto/benchmark/protobuff"
	"github.com/LyrinoxTechnologies/ridged-proto/benchmark/rdg"
)

// --------------------
// Test data
// --------------------
var (
	smallPayload  = []byte("small message payload")
	mediumPayload = make([]byte, 512*1024)    // 512 KB
	largePayload  = make([]byte, 5*1024*1024) // 5 MB
)

func init() {
	// Fill medium and large payloads
	for i := range mediumPayload {
		mediumPayload[i] = byte(i % 256)
	}
	for i := range largePayload {
		largePayload[i] = byte(i % 256)
	}
}

// --------------------
// Helper: Simulated Transport
// --------------------

// simulateTransportSend simulates sending data over a transport without blocking
func simulateTransportSend(b *testing.B, data []byte) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// "Send" to transport (no actual network)
		buf := make([]byte, len(data))
		copy(buf, data)

		// "Receive" from transport
		received := make([]byte, len(data))
		copy(received, buf)
	}
}


// --------------------
// Benchmarks: Protobuf
// --------------------
func BenchmarkProtobuf_SmallTransport(b *testing.B) {
	msg := &protobuff.Blob{Data: smallPayload}
	data, _ := proto.Marshal(msg)
	simulateTransportSend(b, data)
}

func BenchmarkProtobuf_MediumTransport(b *testing.B) {
	msg := &protobuff.Blob{Data: mediumPayload}
	data, _ := proto.Marshal(msg)
	simulateTransportSend(b, data)
}

func BenchmarkProtobuf_LargeTransport(b *testing.B) {
	msg := &protobuff.Blob{Data: largePayload}
	data, _ := proto.Marshal(msg)
	simulateTransportSend(b, data)
}

// --------------------
// Benchmarks: rdgproto
// --------------------
func BenchmarkRdgproto_SmallTransport(b *testing.B) {
	msg := &rdg.Blob{Data: smallPayload}
	data, _ := msg.Marshal()
	simulateTransportSend(b, data)
}

func BenchmarkRdgproto_MediumTransport(b *testing.B) {
	msg := &rdg.Blob{Data: mediumPayload}
	data, _ := msg.Marshal()
	simulateTransportSend(b, data)
}

func BenchmarkRdgproto_LargeTransport(b *testing.B) {
	msg := &rdg.Blob{Data: largePayload}
	data, _ := msg.Marshal()
	simulateTransportSend(b, data)
}

// --------------------
// Optional: Chunked rdgproto transport
// --------------------
func BenchmarkRdgproto_LargeTransport_Chunked(b *testing.B) {
	const chunkSize = 64 * 1024
	msg := &rdg.Blob{Data: largePayload}
	data, _ := msg.Marshal()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for offset := 0; offset < len(data); offset += chunkSize {
			end := offset + chunkSize
			if end > len(data) {
				end = len(data)
			}
			chunk := data[offset:end]
			_ = chunk // would be sent over transport
		}
	}
}
