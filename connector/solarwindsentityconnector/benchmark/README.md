# Benchmark

You can find full documentation about Go benchmarking in the [testing package](https://pkg.go.dev/testing#B).


## Benchmark Output
Each row in the benchmark output follows this general format:

`<BenchmarkName> <Operations> <Time/op> <B/op> <Allocs/op>`

- **BenchmarkName** – name of the benchmark case.

- **Operations** – number of operations that the loop body execute.

- **Time/op** – average time per operation (nanoseconds).

- **B/op** – average number of bytes allocated per operation.

- **Allocs/op**– average number of memory allocations per operation.


### Example
| Benchmark name                       | Operations | Time/op            | B/op             | Allocs/op*          |
|--------------------------------------|------------|--------------------|------------------|---------------------|
BenchmarkMetrics/100_metrics_0.0_false | 2360	    | 522211 ns/op	     | 410380 B/op	    | 10166 allocs/op
BenchmarkMetrics/10K_metrics_0.0_false | 22	        | 51158437 ns/op	 | 41474485 B/op	| 1015050 allocs/op
BenchmarkMetrics/1M_metrics_0.0_false  | 1	        | 5850317708 ns/op	 | 4190139808 B/op	| 101500820 allocs/op

## Format of Benchmark Test Cases
Each benchmark test case is described using the following format:

`<BenchmarkName> <TotalData> <InvalidRatio> <Multiple>`

- **BenchmarkName** – name identifier for the test case. It follows the format: `<TotalData>_<DataType>_<InvalidRatio>_<Multiple>`. For example, *"10K_metrics_0.5_true"* represents a case with 10 000 metrics, 50% of them invalid, passed as a batch (multiple).

- **TotalData** – total number of metrics/logs used in the benchmark (e.g. 100, 10 000, 1 000 000).

- **InvalidRatio** –  the ratio of invalid metrics/logs relative to the total amount of data.

- **Multiple** – defines how data is passed:
    - false – metrics/logs are passed separately.
    - true – metrics/logs are grouped and passed as a single batch.
    #### Examples 
    ##### multiple = true
    ```
    {
        "resource_logs": [
            {
                "resource": {
                    "attributes": [
                        {
                            "key": "k8s.cluster.name",
                            "value": {
                                "stringValue": "Cluster 1"
                            }
                        }
                    ]
                },
                "scope_logs": [
                    {
                        "log_records": [
                            {
                                "observedTimeUnixNano": 1719740310000000000,
                                "body": {
                                    "value": {
                                        "stringValue": "test"
                                    }
                                }
                            }
                        ]
                    }
                ]
            },
            {
                "resource": {
                    "attributes": [
                        {
                            "key": "k8s.cluster.name",
                            "value": {
                                "stringValue": "Cluster 2"
                            }
                        }
                    ]
                },
                "scope_logs": [
                    {
                        "log_records": [
                            {
                                "observedTimeUnixNano": 1719740315000000000,
                                "body": {
                                    "value": {
                                        "stringValue": "test"
                                    }
                                }
                            }
                        ]
                    }
                ]
            },
            ...
        ]
    }
    ```
    ##### multiple = false
    ```
    [
        {
            "resource_logs": [
                {
                    "resource": {
                        "attributes": [
                            {
                                "key": "k8s.cluster.name",
                                "value": {
                                    "stringValue": "Cluster 1"
                                }
                            }
                        ]
                    },
                    "scope_logs": [
                        {
                            "log_records": [
                                {
                                    "observedTimeUnixNano": 1719740310000000000,
                                    "body": {
                                        "value": {
                                            "stringValue": "test"
                                        }
                                    }
                                }
                            ]
                        }
                    ]
                }
            ]
        },
        {
            "resource_logs": [
                {
                    "resource": {
                        "attributes": [
                            {
                                "key": "k8s.cluster.name",
                                "value": {
                                    "stringValue": "Cluster 2"
                                }
                            }
                        ]
                    },
                    "scope_logs": [
                        {
                            "log_records": [
                                {
                                    "observedTimeUnixNano": 1719740310000000000,
                                    "body": {
                                        "value": {
                                            "stringValue": "test"
                                        }
                                    }
                                }
                            ]
                        }
                    ]
                }
            ]
        },
        ...
    ]
    ```
