package store

import (
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	K8SClient    client.Client
	K8SClientSet *kubernetes.Clientset
)
