package metricCollector

import (
	"fmt"
	"math"
	"traffic-scheduler/graphGenerator"
	"traffic-scheduler/prometheusClient"
)

// PodCpuUtilizationData는 특정 파드의 CPU 활용도를 나타냅니다.
type PodCpuUtilizationData struct {
	PodName        string `json:"pod_name"`
	CpuUtilization int    `json:"cpu_utilization"` // 퍼센트 단위
}

// ComponentCpuUtilizationData는 특정 컴포넌트와 그 컴포넌트의 파드들이 수신한 CPU 활용도를 나타냅니다.
type ComponentCpuUtilizationData struct {
	ComponentName string                  `json:"component_name"`
	PodCpuUsage   []PodCpuUtilizationData `json:"pod_cpu_usage"`
}

// CpuUtilizationData는 전체 CPU 활용도 데이터를 포함합니다.
type CpuUtilizationData struct {
	Components []ComponentCpuUtilizationData `json:"components"`
}

// CollectCpuUtilizationPerPod는 컴포넌트 그래프와 컴포넌트-파드 매핑을 기반으로 각 파드의 CPU 활용도를 수집합니다.
func CollectCpuUtilizationPerPod(promClient *prometheusClient.PrometheusClient, namespace string, componentGraph *graphGenerator.ComponentGraph, componentPodMap map[string][]string, duration string) (*CpuUtilizationData, error) {
	var utilizationData CpuUtilizationData

	// 모든 컴포넌트에 대해 CPU 활용도를 수집
	for component := range componentGraph.Components {
		pods, exists := componentPodMap[component]
		if !exists {
			continue
		}

		var podCpuUsageList []PodCpuUtilizationData

		for _, pod := range pods {
			// `container_cpu_usage_seconds_total` 쿼리를 실행하여 CPU 사용량을 가져옴
			query := fmt.Sprintf(`  sum(rate(container_cpu_usage_seconds_total{cpu="total",pod="%s",container="%s"}[%s])) * 1000 `,
				pod, component, duration)

			result, err := promClient.Query(query)
			if err != nil {
				return nil, err
			}

			var totalCpuUsage float64
			for _, sample := range result {
				totalCpuUsage += float64(sample.Value)
			}

			roundedCpuUsage := int(math.Round(totalCpuUsage))

			podCpuUsageList = append(podCpuUsageList, PodCpuUtilizationData{
				PodName:        pod,
				CpuUtilization: roundedCpuUsage,
			})
		}

		utilizationData.Components = append(utilizationData.Components, ComponentCpuUtilizationData{
			ComponentName: component,
			PodCpuUsage:   podCpuUsageList,
		})
	}

	return &utilizationData, nil
}
