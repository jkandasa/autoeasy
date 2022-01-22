package types

import (
	"github.com/operator-framework/api/pkg/lib/version"
	corsosv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

type Info struct {
	Name        string
	DisplayName string
	Version     version.OperatorVersion
	Images      []string
	Phase       corsosv1alpha1.ClusterServiceVersionPhase
}
