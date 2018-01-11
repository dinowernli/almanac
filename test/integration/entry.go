package integration

type entry struct {
	Message     string `json:"message"`
	Logger      string `json:"logger"`
	TimestampMs int64  `json:"timestamp_ms"`
}
