# cluster-logging benchmarks

This suite consists of benchmarks tests for the following cluster-logging ES scenarios. Each scenario asserts recorded measurements against a selected profile from the [config](./config) directory:

1. **High Volume Writes**: Measure `QPS`, `p99`, `p50` and `avg` request duration for all 2xx write requests to ES cluster.
2. **High Volume Reads**: Measure `QPS`, `p99`, `p50` and `avg` request duration for all 2xx read requests to ES cluster
3. **High Volume Aggregate**: Measure `QPS`, `p99`, `p50` and `avg` request duration for all 2xx read requests to ES cluster.
