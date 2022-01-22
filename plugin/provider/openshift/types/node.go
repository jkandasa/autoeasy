package types

import (
	corev1 "k8s.io/api/core/v1"
)

type Node struct {
	Name     string                `json:"name"`
	Labels   map[string]string     `json:"labels"`
	IsMaster bool                  `json:"isMaster"`
	Capacity corev1.ResourceList   `json:"capacity"`
	NodeInfo corev1.NodeSystemInfo `json:"nodeInfo"`
}
