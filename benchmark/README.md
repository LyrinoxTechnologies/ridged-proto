# Benchmark Suite for rdgproto

This directory contains benchmarks comparing **rdgproto** against **Google Protocol Buffers** for various message types and payload sizes.

## Requirements

- Go 1.18 or later
- Protocol Buffers compiler (`protoc`) and Go plugin:

```bash
sudo apt install -y protobuf-compiler
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
````

* Python 3 with **matplotlib** for generating charts:

```bash
pip3 install matplotlib
```

## Running the Benchmarks

### 1. Clone the repository

```bash
git clone https://github.com/LyrinoxTechnologies/ridged-proto.git
cd ridged-proto/benchmark
```

### 2. Generate Protobuf Go code

The generated files will go into `./benchmark/protobuff`:

```bash 
cd protobuff && protoc --go_out=. --go_opt=paths=source_relative *.proto && cd ..
```
> **NOTE:** You must run the above command for Protocol Buffers benchmarks to work

### 3. Run Go benchmarks

```bash
go test -bench=. -benchmem -count=5 -v
```

* `-bench=.` runs all benchmarks.
* `-benchmem` reports memory allocations.
* `-count=5` repeats each benchmark 5 times to smooth out noise.
* `-v` prints detailed output.

> **NOTE:** If you wish to view a chart, there is an included python script for doing so. To do that, continue following the below directions

### 3.2. Save benchmark outputs separately

```bash
go test -bench='^BenchmarkProtobuf_' -benchmem -count=5 > protobuf_bench.txt
go test -bench='^BenchmarkRdgproto_' -benchmem -count=5 > rdgproto_bench.txt
```

These files can be used for plotting and comparison.

### 4. Generate Performance Charts

To visualize the results:

```bash
python3 benchmark_compare.py
```

* Parses Benchmark Outputs
* Computes averages
* Produces comparison charts for rdgproto vs Protocol Buffers

## Notes

* Benchmarks are CPU-bound; results may vary slightly depending on your hardware.
* For fair comparisons, run multiple rounds to avoid cache or scheduling effects.
* Large payloads may show differences in memory usage and allocations.

## License

See [LICENSE](../LICENSE) for licensing information.
