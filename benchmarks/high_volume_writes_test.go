package benchmarks_test

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/cluster-logging-benchmarks/internal/config"
	"github.com/openshift/cluster-logging-benchmarks/internal/k8s"
	"github.com/openshift/cluster-logging-benchmarks/internal/logger"
)

var _ = Describe("Scenario: High Volume Writes", func() {
	var (
		beforeOnce  sync.Once
		afterOnce   sync.Once
		scenarioCfg config.HighVolumeWrites
	)

	BeforeEach(func() {
		scenarioCfg = benchCfg.Scenarios.HighVolumeWrites
		if !scenarioCfg.Enabled {
			Skip("High Volumes Writes Benchmark not enabled!")

			return
		}

		beforeOnce.Do(func() {
			writerCfg := scenarioCfg.Writers

			err := logger.Deploy(k8sClient, benchCfg.Logger, writerCfg, benchCfg.ClusterLogging.BulkIndexURL())
			Expect(err).Should(Succeed(), "Failed to deploy logger")

			err = k8s.WaitForReadyDeployment(k8sClient, benchCfg.Logger.Namespace, benchCfg.Logger.Name, writerCfg.Replicas, defaultRetry, defaulTimeout)
			Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")
		})

		time.Sleep(scenarioCfg.Samples.Interval)
	})

	AfterEach(func() {
		afterOnce.Do(func() {
			Expect(logger.Undeploy(k8sClient, benchCfg.Logger)).Should(Succeed(), "Failed to delete logger deployment")
		})
	})

	Measure("should result in measurements of p99, p50 and avg for all successful write requests to elasticsearch", func(b Benchmarker) {
		defaultRange := scenarioCfg.Samples.Range

		//
		// Collect measurements for the elasticsearch
		//
		job := benchCfg.Metrics.ElasticsearchJob()

		// Record Reads QPS
		qps, err := metricsClient.RequestWritesQPS(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read QPS for all elasticsearch push with status code 2xx")
		b.RecordValue("All elasticsearch 2xx push QPS", qps)

		// Record latency p99
		p99, err := metricsClient.RequestDurationOkBulkIndexP99(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read p99 for all elasticsearch push requests with status code 2xx")
		b.RecordValue("All elasticsearch 2xx push p99", p99)

		// Record latency p50
		p50, err := metricsClient.RequestDurationOkBulkIndexP50(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read p50 for all elasticsearch push requests with status code 2xx")
		b.RecordValue("All elasticsearch 2xx push p50", p50)

		// Record latency avg
		avg, err := metricsClient.RequestDurationOkBulkIndexAvg(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read average for all elasticsearch push requests with status code 2xx")
		b.RecordValue("All elasticsearch 2xx push avg", avg)
	}, benchCfg.Scenarios.HighVolumeWrites.Samples.Total)
})
