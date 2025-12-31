// marshal_bench_test.go
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
	loginTestUsername = "john.doe@example.com"
	loginTestPassword = "super_secret_password_123"
	loginTestClientID = "client-abc-123-xyz"

	blobTestData = []byte("this is some test blob data")
	bulkTestData = []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	metricsTestData = rdg.Metrics{
		A: 100,
		B: 200,
		C: 300,
		D: 400,
		E: 500,
	}
)

// --------------------
// Benchmarks: Login
// --------------------

// Protobuf Marshal
func BenchmarkProtobuf_Login_Marshal(b *testing.B) {
	msg := &protobuff.LoginRequest{
		Username: loginTestUsername,
		Password: loginTestPassword,
		ClientId: loginTestClientID,
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

// rdgproto Marshal
func BenchmarkRdgproto_Login_Marshal(b *testing.B) {
	msg := &rdg.LoginRequest{
		Username: loginTestUsername,
		Password: loginTestPassword,
		ClientId: loginTestClientID,
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
// Benchmarks: Blob
// --------------------

func BenchmarkProtobuf_Blob_Marshal(b *testing.B) {
	msg := &protobuff.Blob{
		Data: blobTestData,
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

func BenchmarkRdgproto_Blob_Marshal(b *testing.B) {
	msg := &rdg.Blob{
		Data: blobTestData,
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
// Benchmarks: Bulk
// --------------------

func BenchmarkProtobuf_Bulk_Marshal(b *testing.B) {
	msg := &protobuff.BulkData{
		Values: bulkTestData,
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

func BenchmarkRdgproto_Bulk_Marshal(b *testing.B) {
	msg := &rdg.BulkData{
		Values: bulkTestData,
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
// Benchmarks: Metrics
// --------------------

// Protobuf Marshal
func BenchmarkProtobuf_Metrics_Marshal(b *testing.B) {
	msg := &protobuff.Metrics{
		A: 1,
		B: 2,
		C: 3,
		D: 4,
		E: 5,
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

// rdgproto Marshal
func BenchmarkRdgproto_Metrics_Marshal(b *testing.B) {
	msg := &rdg.Metrics{
		A: 1,
		B: 2,
		C: 3,
		D: 4,
		E: 5,
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
