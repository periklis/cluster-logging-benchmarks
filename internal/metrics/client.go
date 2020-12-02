package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type MetricType string

const (
	// TODO(@periklis) Replace with correct metric name
	BytesReceivedTotal MetricType = "bytes_received_total"
)

type Client interface {
	BytesReceivedTotal() (float64, error)

	// HTTP API
	RequestDurationOkSearchAvg(job string, duration model.Duration) (float64, error)
	RequestDurationOkSearchP50(job string, duration model.Duration) (float64, error)
	RequestDurationOkSearchP99(job string, duration model.Duration) (float64, error)

	RequestDurationOkBulkIndexAvg(job string, duration model.Duration) (float64, error)
	RequestDurationOkBulkIndexP50(job string, duration model.Duration) (float64, error)
	RequestDurationOkBulkIndexP99(job string, duration model.Duration) (float64, error)

	RequestReadsQPS(job string, duration model.Duration) (float64, error)
	RequestWritesQPS(job string, duration model.Duration) (float64, error)
}

type client struct {
	api     v1.API
	timeout time.Duration
}

func NewClient(url string, timeout time.Duration) (Client, error) {
	pc, err := api.NewClient(api.Config{Address: url})
	if err != nil {
		return nil, fmt.Errorf("failed creating prometheus client: %w", err)
	}

	return &client{
		api:     v1.NewAPI(pc),
		timeout: timeout,
	}, nil
}

func (c *client) requestDurationAvg(job, method, route, code string, duration model.Duration) (float64, error) {
	// TODO(@periklis): Adapt query to ES request duration metrics
	query := fmt.Sprintf(
		`100 * (sum by (job) (rate(request_duration_seconds_sum{job="%s", method="%s", route="%s", status_code=~"%s"}[%s])) / sum by (job) (rate(request_duration_seconds_count{job="%s", method="%s", route="%s", status_code=~"%s"}[%s])))`,
		job, method, route, code, duration,
		job, method, route, code, duration,
	)

	return c.executeScalarQuery(query)
}

func (c *client) requestDurationQuantile(job, method, route, code string, duration model.Duration, percentile int) (float64, error) {
	// TODO(@periklis): Adapt query to ES request duration metrics
	query := fmt.Sprintf(
		`histogram_quantile(0.%d, sum by (job, le) (rate(request_duration_seconds_bucket{job="%s", method="%s", route="%s", status_code=~"%s"}[%s])))`,
		percentile, job, method, route, code, duration,
	)

	return c.executeScalarQuery(query)
}

func (c *client) requestQPS(job, route, code string, duration model.Duration) (float64, error) {
	// TODO(@periklis): Adapt query to ES request duration metrics
	query := fmt.Sprintf(
		`sum(rate(request_duration_seconds_count{job="%s", route=~"%s", status_code=~"%s"}[%s]))`,
		job, route, code, duration,
	)

	return c.executeScalarQuery(query)
}

func (c *client) executeScalarQuery(query string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	res, _, err := c.api.Query(ctx, query, time.Now())
	if err != nil {
		return 0.0, fmt.Errorf("failed executing query %q: %w", query, err)
	}

	if res.Type() == model.ValScalar {
		value := res.(*model.Scalar)
		return float64(value.Value), nil
	}

	if res.Type() == model.ValVector {
		vec := res.(model.Vector)
		if vec.Len() == 0 {
			return 0.0, fmt.Errorf("empty result set for query: %s", query)
		}

		return float64(vec[0].Value), nil
	}

	return 0.0, fmt.Errorf("failed to parse result for query: %s", query)
}
