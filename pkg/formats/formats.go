package formats

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

func Memory(q resource.Quantity) string {
	return fmt.Sprintf("%vMi", q.Value()/(1024*1024))
}

func MemoryUsage(q resource.Quantity, percentage int32) string {
	return fmt.Sprintf("%s (%d%%)", Memory(q), percentage)
}

func CPU(q resource.Quantity) string {
	return fmt.Sprintf("%vm", q.MilliValue())
}

func CPUUsage(q resource.Quantity, percentage int32) string {
	return fmt.Sprintf("%s (%d%%)", CPU(q), percentage)
}
