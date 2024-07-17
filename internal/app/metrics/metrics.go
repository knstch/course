package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	StatusCodesCounter *prometheus.CounterVec
}

func InitMetrics() *Metrics {
	statusCodes := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "returned_status_codes",
		Help: "HTTP статус коды, которые возвращает приложение",
	}, []string{"code", "method", "function"})

	prometheus.MustRegister(statusCodes)

	return &Metrics{
		statusCodes,
	}
}

func (m Metrics) RecordResponse(statusCode int, method, function string) {
	m.StatusCodesCounter.WithLabelValues(fmt.Sprint(statusCode), method, function).Inc()
}
