# Benchmark

You can find full documentation about Go benchmarking in the [testing package](https://pkg.go.dev/testing#B).


## Benchmark Output
Each row in the benchmark output follows this general format:

`<BenchmarkName> <N> <time/op> <B/op> <allocs/op>`

- **BenchmarkName** – name of the benchmark case.

- **N** – number of iterations the benchmark was run.

- **time/op** – average time per operation (nanoseconds per op).

- **B/op** – average number of bytes allocated per operation.

- **allocs/op**– average number of memory allocations per operation.



### Example
| Benchmark name                          | Iterations | Time per op       | Bytes per op     | Allocs per op       |
|----------------------------------------|------------|--------------------|------------------|----------------------|
| BenchmarkMetrics/100_metrics_0.0_false-10 | 1000000000 | 0.0002275 ns/op     | 0 B/op           | 0 allocs/op          |
| BenchmarkMetrics/10K_metrics_0.0_false-10 | 1000000000 | 0.03321 ns/op       | 0 B/op           | 0 allocs/op          |
| BenchmarkMetrics/1M_metrics_0.0_false-10  | 1          | 3017625833 ns/op    | 3534081728 B/op  | 80000377 allocs/op   |

## Benchmark Test Case Format
The benchmark cases are defined using the following structure:

`<BenchmarkName> <TotalData> <InvalidRatio> <Multiple>`

- **BenchmarkName** – a label used to identify the case in the benchmark. It follows the format: `<TotalData>_<DataType>_<InvalidRatio>_<Multiple>`. For example, *"10K_metrics_0.5_true"* represents a case with 10 000 metrics, 50% of them invalid, and using multiple log.

- **TotalData** – the total number of metrics/logs passed into the benchmark (e.g. 100, 10 000, 1 000 000).

- **InvalidRatio** –  the ratio of metrics/logs that are invalid.

- **Multiple** – defines how the data is passed:
    - false – metrics/logs are passed separately.
    - true – metrics/logs are grouped and passed as a single batch.