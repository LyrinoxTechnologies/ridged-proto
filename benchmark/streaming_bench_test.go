// streaming_bench_test.go
package benchmark

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/LyrinoxTechnologies/ridged-proto/benchmark/protobuff"
	"github.com/LyrinoxTechnologies/ridged-proto/benchmark/rdg"
)

var (
	// Large payloads to simulate streaming
	largeBlobData = make([]byte, 1024*1024) // 1 MB
	hugeBlobData  = make([]byte, 10*1024*1024) // 10 MB
)

func init() {
	// Fill payloads with pseudo-random data
	for i := range largeBlobData {
		largeBlobData[i] = byte(i % 256)
	}
	for i := range hugeBlobData {
		hugeBlobData[i] = byte(i % 256)
	}
}

// --------------------
// Protobuf Streaming Benchmarks
// --------------------
func BenchmarkProtobuf_LargeBlob_Marshal(b *testing.B) {
	msg := &protobuff.Blob{
		Data: largeBlobData,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := proto.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
		_ = data
	}
}

func BenchmarkProtobuf_HugeBlob_Marshal(b *testing.B) {
	msg := &protobuff.Blob{
		Data: hugeBlobData,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := proto.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
		_ = data
	}
}

// --------------------
// rdgproto Streaming Benchmarks
// --------------------
func BenchmarkRdgproto_LargeBlob_Marshal(b *testing.B) {
	msg := &rdg.Blob{
		Data: largeBlobData,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := msg.Marshal()
		if err != nil {
			b.Fatal(err)
		}
		_ = data
	}
}

func BenchmarkRdgproto_HugeBlob_Marshal(b *testing.B) {
	msg := &rdg.Blob{
		Data: hugeBlobData,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := msg.Marshal()
		if err != nil {
			b.Fatal(err)
		}
		_ = data
	}
}

// --------------------
// Optional: Chunked Streaming Simulation
// --------------------
// Simulate sending the payload in 64 KB chunks
func BenchmarkRdgproto_LargeBlob_Chunked(b *testing.B) {
	const chunkSize = 64 * 1024 // 64 KB
	msg := &rdg.Blob{
		Data: largeBlobData,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := msg.Marshal()
		if err != nil {
			b.Fatal(err)
		}

		// Simulate sending in chunks
		for offset := 0; offset < len(data); offset += chunkSize {
			end := offset + chunkSize
			if end > len(data) {
				end = len(data)
			}
			chunk := data[offset:end]
			_ = chunk // would be sent over network
		}
	}
}
