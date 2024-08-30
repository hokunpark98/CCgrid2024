package trafficMonitor

import (
	"fmt"
	"math"
	"traffic-scheduler/graphGenerator"
	"traffic-scheduler/prometheusClient"
)

// PodRequestCount는 특정 파드가 수신한 요청 수를 나타냅니다.
type PodRequestCountData struct {
	PodName      string `json:"pod_name"`
	RequestCount int    `json:"request_count"`
}

// ComponentRequestCount는 특정 컴포넌트와 그 컴포넌트의 파드들이 수신한 요청 수를 나타냅니다.
type ComponentRequestCountData struct {
	ComponentName string                `json:"component_name"`
	PodRequests   []PodRequestCountData `json:"pod_requests"`
}

// RequestData는 전체 요청 수 데이터를 포함합니다.
type RequestCountData struct {
	Components []ComponentRequestCountData `json:"components"`
}

// CollectRequestCountPerPod는 컴포넌트 그래프와 컴포넌트-파드 매핑을 기반으로 각 파드의 수신 요청 수를 수집합니다.
func CollectRequestCountPerPod(promClient *prometheusClient.PrometheusClient, namespace string, componentGraph *graphGenerator.ComponentGraph, componentPodMap map[string][]string, duration string) (*RequestCountData, error) {
	var requestData RequestCountData

	for _, link := range componentGraph.Links {
		dc := link.DC
		uc := link.UC

		dcPods, exists := componentPodMap[dc]
		if !exists {
			continue
		}

		var podRequestList []PodRequestCountData

		for _, pod := range dcPods {
			query := fmt.Sprintf(`increase(istio_requests_total{kubernetes_namespace="%s", kubernetes_pod_name="%s", source_app="%s"}[%s])`,
				namespace, pod, uc, duration)

			result, err := promClient.Query(query)
			if err != nil {
				return nil, err
			}

			var totalRequests float64
			for _, sample := range result {
				totalRequests += float64(sample.Value)
			}

			// 소수점을 첫째 자리에서 반올림하고 정수로 변환
			roundedRequests := int(math.Round(totalRequests))

			podRequestList = append(podRequestList, PodRequestCountData{
				PodName:      pod,
				RequestCount: roundedRequests,
			})
		}

		requestData.Components = append(requestData.Components, ComponentRequestCountData{
			ComponentName: dc,
			PodRequests:   podRequestList,
		})
	}

	return &requestData, nil
}
