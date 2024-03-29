package types

import (
	iostreamTY "github.com/jkandasa/autoeasy/pkg/types/iostream"
)

type PortForwardRequest struct {
	Namespace  string                `json:"namespace"`
	Pod        string                `json:"pod"`
	Deployment string                `json:"deployment"`
	Addresses  []string              `json:"addresses"`
	Ports      []string              `json:"ports"`
	Streams    *iostreamTY.IOStreams `json:"-"`
}
