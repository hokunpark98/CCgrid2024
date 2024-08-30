package metricCollector

import (
	"traffic-scheduler/prometheusClient"
)

// NodeCPUFrequencyData는 노드의 이름과 최소 CPU 주파수를 나타냅니다.
type NodeCPUFrequencyData struct {
	NodeName string `json:"node_name"`
	Hertz    int64  `json:"hertz"`
}

// CollectNodeCPUFrequency는 Kubernetes 클러스터의 모든 워커 노드에 대해
// node_cpu_frequency_min_hertz 값을 수집하고 노드 이름과 함께 반환합니다.
func CollectNodeCPUFrequency(promClient *prometheusClient.PrometheusClient) ([]NodeCPUFrequencyData, error) {
	var nodeFrequencyData []NodeCPUFrequencyData

	// Prometheus 쿼리 작성
	query := `node_cpu_frequency_max_hertz{cpu="0"}`

	// 쿼리 실행
	result, err := promClient.Query(query)
	if err != nil {
		return nil, err
	}

	// 쿼리 결과를 반복하여 노드 이름과 주파수 값을 수집
	for _, sample := range result {
		nodeName := string(sample.Metric["instance"])
		hertz := int64(sample.Value) / 100000000

		nodeFrequencyData = append(nodeFrequencyData, NodeCPUFrequencyData{
			NodeName: nodeName,
			Hertz:    hertz,
		})
	}

	return nodeFrequencyData, nil
}
