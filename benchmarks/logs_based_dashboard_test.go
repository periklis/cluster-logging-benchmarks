package benchmarks_test

import (
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/cluster-logging-benchmarks/internal/config"
	"github.com/openshift/cluster-logging-benchmarks/internal/k8s"
	"github.com/openshift/cluster-logging-benchmarks/internal/latch"
	"github.com/openshift/cluster-logging-benchmarks/internal/logger"
	"github.com/openshift/cluster-logging-benchmarks/internal/metrics"
	"github.com/openshift/cluster-logging-benchmarks/internal/querier"
)

var _ = Describe("Scenario: Logs-Based Dashboard", func() {
	var (
		beforeOnce  sync.Once
		afterOnce   sync.Once
		scenarioCfg config.LogsBasedDashboard
	)

	BeforeEach(func() {
		scenarioCfg = benchCfg.Scenarios.LogsBasedDashboard
		if !scenarioCfg.Enabled {
			Skip("Log Based Dashboard Benchmark not enabled!")

			return
		}

		beforeOnce.Do(func() {
			writerCfg := scenarioCfg.Writers
			readerCfg := scenarioCfg.Readers

			// Deploy the logger to ingest logs
			err := logger.Deploy(k8sClient, benchCfg.Logger, writerCfg, benchCfg.ClusterLogging.BulkIndexURL())
			Expect(err).Should(Succeed(), "Failed to deploy logger")

			err = k8s.WaitForReadyDeployment(k8sClient, benchCfg.Logger.Namespace, benchCfg.Logger.Name, writerCfg.Replicas, defaultRetry, defaulTimeout)
			Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")

			// Wait until we ingested enough logs based on startThreshold
			err = latch.WaitUntilGreaterOrEqual(metricsClient, metrics.BytesReceivedTotal, readerCfg.StartThreshold, defaultLatchTimeout)
			Expect(err).Should(Succeed(), "Failed to wait until latch activated")

			// Undeploy logger to assert only read traffic
			err = logger.Undeploy(k8sClient, benchCfg.Logger)
			Expect(err).Should(Succeed(), "Failed to delete logger deployment")

			// Deploy the query clients
			for id, query := range readerCfg.Queries {
				err = querier.Deploy(k8sClient, benchCfg.Querier, readerCfg, benchCfg.ClusterLogging.SearchURL(), id, query)
				Expect(err).Should(Succeed(), "Failed to deploy querier")
			}

			for id := range readerCfg.Queries {
				name := querier.DeploymentName(benchCfg.Querier, id)

				err = k8s.WaitForReadyDeployment(k8sClient, benchCfg.Querier.Namespace, name, readerCfg.Replicas, defaultRetry, defaulTimeout)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed to wait for ready querier deployment: %s", name))
			}
		})

		time.Sleep(scenarioCfg.Samples.Interval)
	})

	AfterEach(func() {
		afterOnce.Do(func() {
			readerCfg := scenarioCfg.Readers
			for id := range readerCfg.Queries {
				Expect(querier.Undeploy(k8sClient, benchCfg.Querier, id)).Should(Succeed(), "Failed to delete querier deployment")
			}
		})
	})

	Measure("should result in measurements of p99, p50 and avg for all successful dashboard read requests", func(b Benchmarker) {
		defaultRange := scenarioCfg.Samples.Range

		//
		// Collect measurements for the elasticsearch
		//
		job := benchCfg.Metrics.ElasticsearchJob()

		// Record Reads QPS
		qps, err := metricsClient.RequestReadsQPS(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read QPS for all elasticsearch dashboard reads with status code 2xx")
		b.RecordValue("All elasticsearch 2xx dashboard reads QPS", qps)

		// Record latency p99
		p99, err := metricsClient.RequestDurationOkSearchP99(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read p50 for all elasticsearch dashboard reads with status code 2xx")
		b.RecordValue("All elasticsearch 2xx dashboard reads p99", p99)

		// Record latency p50
		p50, err := metricsClient.RequestDurationOkSearchP50(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read p50 for all elasticsearch dashboard reads with status code 2xx")
		b.RecordValue("All elasticsearch 2xx dashboard reads p50", p50)

		// Record latency avg
		avg, err := metricsClient.RequestDurationOkSearchAvg(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read average for all elasticsearch dashboard reads with status code 2xx")
		b.RecordValue("All elasticsearch 2xx dashboard reads avg", avg)
	}, benchCfg.Scenarios.LogsBasedDashboard.Samples.Total)
})
