package metricCollector

import (
	"log"
	"traffic-scheduler/prometheusClient"
)

// PodInfo는 파드의 이름, IP 주소, 호스트 이름 및 호스트 IP 주소를 나타냅니다.
type PodInfo struct {
	PodName  string `json:"pod_name"`
	PodIP    string `json:"pod_ip"`
	HostName string `json:"host_name"`
	HostIP   string `json:"host_ip"`
}

// ComponentPodMap은 컴포넌트 이름을 키로 하고, 해당 컴포넌트에 속한 파드들의 정보를 포함하는 맵입니다.
type ComponentPodMap struct {
	Components map[string][]PodInfo `json:"components"`
}

// MapComponentToPodInfo 함수는 주어진 네임스페이스에서 각 컴포넌트에 대응하는 파드의 상세 정보를 반환합니다.
func MapComponentToPodInfo(promClient *prometheusClient.PrometheusClient, namespace string) (*ComponentPodMap, error) {
	componentsPodMap := &ComponentPodMap{
		Components: make(map[string][]PodInfo),
	}

	// Prometheus 쿼리를 통해 네임스페이스 내의 모든 컴포넌트를 가져옴
	componentQuery := `kube_deployment_labels{namespace="` + namespace + `"}`
	componentResult, err := promClient.Query(componentQuery)
	if err != nil {
		log.Fatalf("Error querying Prometheus for components: %v", err)
		return nil, err
	}

	// 각 컴포넌트(label_app)를 순회하면서 해당 컴포넌트에 속한 파드 정보를 가져옴
	for _, sample := range componentResult {
		component := string(sample.Metric["label_app"])

		query := `kube_pod_labels{label_app="` + component + `", namespace="` + namespace + `"}`

		labelResult, err := promClient.Query(query)
		if err != nil {
			log.Fatalf("Error querying Prometheus for labels: %v", err)
			return nil, err
		}

		var podNames []string
		for _, sample := range labelResult {
			podName := string(sample.Metric["pod"])
			podNames = append(podNames, podName)
		}

		for _, podName := range podNames {
			query := `kube_pod_info{namespace="` + namespace + `", pod="` + podName + `"}`

			infoResult, err := promClient.Query(query)
			if err != nil {
				log.Fatalf("Error querying Prometheus for pod info: %v", err)
				return nil, err
			}

			for _, sample := range infoResult {
				podInfo := PodInfo{
					PodName:  string(sample.Metric["pod"]),
					PodIP:    string(sample.Metric["pod_ip"]),
					HostName: string(sample.Metric["node"]),
					HostIP:   string(sample.Metric["host_ip"]),
				}

				componentsPodMap.Components[component] = append(componentsPodMap.Components[component], podInfo)
			}
		}
	}

	return componentsPodMap, nil
}
