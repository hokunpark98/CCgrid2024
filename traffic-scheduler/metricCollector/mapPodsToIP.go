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

// MapPodsToIP 함수는 파드의 이름, IP 주소, 호스트 이름 및 호스트 IP 주소를 포함하는 맵을 반환합니다.
func MapPodsToIP(promClient *prometheusClient.PrometheusClient, namespace string) (map[string]PodInfo, error) {
	podInfoMap := make(map[string]PodInfo)

	query := `kube_pod_info{namespace="` + namespace + `"}`

	result, err := promClient.Query(query)
	if err != nil {
		log.Fatalf("Error querying Prometheus: %v", err)
		return nil, err
	}

	for _, sample := range result {
		podName := string(sample.Metric["pod"])
		podIP := string(sample.Metric["pod_ip"])
		hostName := string(sample.Metric["node"])
		hostIP := string(sample.Metric["host_ip"])

		podInfoMap[podName] = PodInfo{
			PodName:  podName,
			PodIP:    podIP,
			HostName: hostName,
			HostIP:   hostIP,
		}
	}

	return podInfoMap, nil
}
