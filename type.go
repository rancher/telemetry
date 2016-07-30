package main

import (
	"github.com/vincent99/telemetry/collector"
)

type Record struct {
	Version      int                    `json:"version"`
	Installation collector.Installation `json:"installation"`
	Os           collector.Os           `json:"os"`
}
