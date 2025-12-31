// unmarshal_bench_test.go
package benchmark

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/LyrinoxTechnologies/ridged-proto/benchmark/protobuff"
	"github.com/LyrinoxTechnologies/ridged-proto/benchmark/rdg"
)

// --------------------
// Pre-marshaled test data
// --------------------
var (
	protoLoginData []byte
	rdgLoginData   []byte

	protoBlobData []byte
	rdgBlobData   []byte

	protoBulkData []byte
	rdgBulkData   []byte

	protoMetricsData []byte
	rdgMetricsData   []byte
)

func init() {
	// --------------------
	// Login
	// --------------------
	protoLoginMsg := &protobuff.LoginRequest{
		Username: loginTestUsername,
		Password: loginTestPassword,
		ClientId: loginTestClientID,
	}
	protoLoginData, _ = proto.Marshal(protoLoginMsg)

	rdgLoginMsg := &rdg.LoginRequest{
		Username: loginTestUsername,
		Password: loginTestPassword,
		ClientId: loginTestClientID,
	}
	rdgLoginData, _ = rdgLoginMsg.Marshal()

	// --------------------
	// Blob
	// --------------------
	protoBlobMsg := &protobuff.Blob{Data: blobTestData}
	protoBlobData, _ = proto.Marshal(protoBlobMsg)

	rdgBlobMsg := &rdg.Blob{Data: blobTestData}
	rdgBlobData, _ = rdgBlobMsg.Marshal()

	// --------------------
	// Bulk
	// --------------------
	protoBulkMsg := &protobuff.BulkData{Values: bulkTestData}
	protoBulkData, _ = proto.Marshal(protoBulkMsg)

	rdgBulkMsg := &rdg.BulkData{Values: bulkTestData}
	rdgBulkData, _ = rdgBulkMsg.Marshal()

	// --------------------
	// Metrics
	// --------------------
	protoMetricsMsg := &protobuff.Metrics{
		A: 1, B: 2, C: 3, D: 4, E: 5,
	}
	protoMetricsData, _ = proto.Marshal(protoMetricsMsg)

	rdgMetricsMsg := &rdg.Metrics{
		A: 1, B: 2, C: 3, D: 4, E: 5,
	}
	rdgMetricsData, _ = rdgMetricsMsg.Marshal()
}

// --------------------
// Benchmarks: Login
// --------------------
func BenchmarkProtobuf_Login_Unmarshal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg protobuff.LoginRequest
		if err := proto.Unmarshal(protoLoginData, &msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRdgproto_Login_Unmarshal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg rdg.LoginRequest
		if err := msg.Unmarshal(rdgLoginData); err != nil {
			b.Fatal(err)
		}
	}
}

// --------------------
// Benchmarks: Blob
// --------------------
func BenchmarkProtobuf_Blob_Unmarshal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg protobuff.Blob
		if err := proto.Unmarshal(protoBlobData, &msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRdgproto_Blob_Unmarshal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg rdg.Blob
		if err := msg.Unmarshal(rdgBlobData); err != nil {
			b.Fatal(err)
		}
	}
}

// --------------------
// Benchmarks: Bulk
// --------------------
func BenchmarkProtobuf_Bulk_Unmarshal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg protobuff.BulkData
		if err := proto.Unmarshal(protoBulkData, &msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRdgproto_Bulk_Unmarshal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg rdg.BulkData
		if err := msg.Unmarshal(rdgBulkData); err != nil {
			b.Fatal(err)
		}
	}
}

// --------------------
// Benchmarks: Metrics
// --------------------
func BenchmarkProtobuf_Metrics_Unmarshal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg protobuff.Metrics
		if err := proto.Unmarshal(protoMetricsData, &msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRdgproto_Metrics_Unmarshal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg rdg.Metrics
		if err := msg.Unmarshal(rdgMetricsData); err != nil {
			b.Fatal(err)
		}
	}
}
