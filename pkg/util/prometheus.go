package util

import (
	"github.com/prometheus/client_golang/prometheus"
)

// RegisterLenient has the same effect as calling premetheus.Register(), but treats
// the special "AlreadyRegisteredError" as a success.
func RegisterLenient(c prometheus.Collector) error {
	err := prometheus.Register(c)
	if err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return err
		}
	}
	return nil
}
