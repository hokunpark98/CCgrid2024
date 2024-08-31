package metricCollector

import (
	"log"
	"traffic-scheduler/graphGenerator"
	"traffic-scheduler/prometheusClient"
)

// MapComponentsToPods 함수는 컴포넌트를 받아 각 컴포넌트에 대응하는 파드 이름을 반환합니다.
func MapComponentsToPods(promClient *prometheusClient.PrometheusClient, componentGraph *graphGenerator.ComponentGraph) map[string][]string {
	componentPodMap := make(map[string][]string)

	for component := range componentGraph.Components {
		query := `kube_pod_labels{label_app="` + component + `"}`

		result, err := promClient.Query(query)
		if err != nil {
			log.Fatalf("Error querying Prometheus: %v", err)
		}

		for _, sample := range result {
			podName := string(sample.Metric["pod"])
			componentPodMap[component] = append(componentPodMap[component], podName)
		}
	}
	return componentPodMap
}
