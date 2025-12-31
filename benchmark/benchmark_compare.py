import re
import matplotlib.pyplot as plt
from glob import glob
from collections import defaultdict

def parse_bench_files(pattern):
    """Parse multiple benchmark files and return average ns/op per benchmark."""
    benchmarks = defaultdict(list)
    for filepath in glob(pattern):
        with open(filepath) as f:
            for line in f:
                m = re.match(r"Benchmark(\w+)-\d+\s+\d+\s+([\d\.]+) ns/op", line)
                if m:
                    name, ns_per_op = m.group(1), float(m.group(2))
                    benchmarks[name].append(ns_per_op)
    # Average over multiple runs/files
    return {k: sum(v)/len(v) for k, v in benchmarks.items()}

# Parse multiple rounds
protobuf = parse_bench_files("protobuf_bench*.txt")
rdgproto = parse_bench_files("rdgproto_bench*.txt")

# Find benchmarks that exist in both
common_labels = sorted(
    k.replace("Protobuf_", "")
    for k in protobuf.keys()
    if f"Rdgproto_{k.replace('Protobuf_', '')}" in rdgproto
)

x = range(len(common_labels))

# Plot
plt.figure(figsize=(14,7))
plt.bar([i-0.2 for i in x], [protobuf[f"Protobuf_{k}"] for k in common_labels],
        width=0.4, label="Protobuf", color="blue")
plt.bar([i+0.2 for i in x], [rdgproto[f"Rdgproto_{k}"] for k in common_labels],
        width=0.4, label="Rdgproto", color="green")

plt.xticks(x, common_labels, rotation=45, ha="right")
plt.yscale('log')
plt.ylabel("Average ns/op (log scale)")
plt.title("Benchmark Comparison: Rdgproto vs Protobuf (averaged over multiple runs)")
plt.grid(True, which="both", linestyle="--", linewidth=0.5)
plt.legend()
plt.tight_layout()
plt.show()
