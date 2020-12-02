package metrics

import "github.com/prometheus/common/model"

func (c *client) BytesReceivedTotal() (float64, error) {
	// TODO(@periklis) Provide implementation
	return 0.0, nil
}

func (c *client) RequestDurationOkSearchAvg(job string, duration model.Duration) (float64, error) {
	// TODO(@periklis): Declare route for search calls
	return c.requestDurationAvg(job, "GET", "", "2.*", duration)
}

func (c *client) RequestDurationOkSearchP50(job string, duration model.Duration) (float64, error) {
	// TODO(@periklis): Declare route for search calls
	return c.requestDurationQuantile(job, "GET", ".*", "2.*", duration, 50)
}

func (c *client) RequestDurationOkSearchP99(job string, duration model.Duration) (float64, error) {
	// TODO(@periklis): Declare route for search calls
	return c.requestDurationQuantile(job, "GET", "search", "2.*", duration, 99)
}

func (c *client) RequestDurationOkBulkIndexAvg(job string, duration model.Duration) (float64, error) {
	// TODO(@periklis): Declare route for bulk index calls
	return c.requestDurationAvg(job, "POST", "", "2.*", duration)
}

func (c *client) RequestDurationOkBulkIndexP50(job string, duration model.Duration) (float64, error) {
	// TODO(@periklis): Declare route for bulk index calls
	return c.requestDurationQuantile(job, "POST", ".*", "2.*", duration, 50)
}

func (c *client) RequestDurationOkBulkIndexP99(job string, duration model.Duration) (float64, error) {
	// TODO(@periklis): Declare route for bulk index calls
	return c.requestDurationQuantile(job, "POST", ".*", "2.*", duration, 99)
}

func (c *client) RequestReadsQPS(job string, duration model.Duration) (float64, error) {
	// TODO(@periklis): Declare route for search calls
	route := ".*"
	return c.requestQPS(job, route, "2.*", duration)
}

func (c *client) RequestWritesQPS(job string, duration model.Duration) (float64, error) {
	// TODO(@periklis): Declare route for bulk index calls
	route := ".*"
	return c.requestQPS(job, route, ".*", duration)
}
