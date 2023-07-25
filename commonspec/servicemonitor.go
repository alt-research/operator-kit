package commonspec

type ServiceMonitorSpec struct {
	//+optional
	Enabled bool `json:"enabled,omitempty"`
	//+optional
	Labels map[string]string `json:"labels,omitempty"`
	//+optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Interval at which metrics should be scraped
	// If not specified Prometheus' global scrape interval is used.
	Interval GoDuration `json:"interval,omitempty"`
	// Timeout after which the scrape is ended
	// If not specified, the Prometheus global scrape timeout is used unless it is less than `Interval` in which the latter is used.
	ScrapeTimeout GoDuration `json:"scrapeTimeout,omitempty"`
}
