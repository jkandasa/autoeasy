package utils

import (
	"errors"
	"fmt"
	"strings"

	mcUtils "github.com/mycontroller-org/server/v2/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func GetObjectMeta(cfg map[string]interface{}) (*metav1.ObjectMeta, error) {
	metaCfgRaw := mcUtils.GetMapValue(cfg, "metadata", nil)
	if metaCfgRaw == nil {
		return nil, errors.New("metadata can not be empty")
	}

	metaCfg, ok := metaCfgRaw.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid metadata format")
	}

	metadata := &metav1.ObjectMeta{}
	err := mcUtils.MapToStruct(mcUtils.TagNameJSON, metaCfg, metadata)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func IgnoreNotFoundError(err error) error {
	if err == nil {
		return nil
	}
	if strings.HasSuffix(err.Error(), "not found") {
		return nil
	}
	return err
}

// FindItem returns the availability status and location
func FindItem(slice []string, value string) (int, bool) {
	for i, item := range slice {
		if item == value {
			return i, true
		}
	}
	return -1, false
}

// ContainsString returns the available status
func ContainsString(slice []string, value string) bool {
	_, available := FindItem(slice, value)
	return available
}

// ContainsNamespacedName returns the available status
func ContainsNamespacedName(slice []types.NamespacedName, target metav1.ObjectMeta) bool {
	for _, n := range slice {
		if target.Name == n.Name && target.Namespace == n.Namespace {
			return true
		}
	}
	return false
}

func ToStringSlice(items []interface{}) []string {
	names := []string{}
	for _, item := range items {
		strItem, ok := item.(string)
		if !ok {
			strItem = fmt.Sprintf("%v", item)
		}
		names = append(names, strItem)
	}
	return names
}

func ToNamespacedNameSlice(rawItems []interface{}) []types.NamespacedName {
	items := make([]types.NamespacedName, 0)
	for _, rawItem := range rawItems {
		item, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}
		name, nameFound := item["name"]
		namepsace, nsFound := item["namespace"]
		if nameFound && nsFound {
			items = append(items, types.NamespacedName{
				Name:      fmt.Sprintf("%v", name),
				Namespace: fmt.Sprintf("%v", namepsace),
			})
		}
	}
	return items
}
