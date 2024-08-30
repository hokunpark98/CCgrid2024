package main

import (
	"fmt"
	"log"
	"os"
	"traffic-scheduler/graphGenerator"
	"traffic-scheduler/metricCollector"
	"traffic-scheduler/prometheusClient"
)

// getRequestCountPerPod는 DC의 각 파드들이 UC로부터 수신한 요청 수를 가져옴
func getRequestCountPerPod(promClient *prometheusClient.PrometheusClient, namespace string, componentGraph *graphGenerator.ComponentGraph, componentPodMap map[string][]string, duration string, logFile *os.File) (*metricCollector.RequestCountData, error) {
	requestCountData, err := metricCollector.CollectRequestCountPerPod(promClient, namespace, componentGraph, componentPodMap, duration)
	if err != nil {
		return nil, err
	}
	logFile.WriteString(fmt.Sprintf("Request Count Per Pod:\n"))
	log.Print("Request Count Per Pod:")
	for _, component := range requestCountData.Components {
		logFile.WriteString(fmt.Sprintf("  Component: %s\n", component.ComponentName))
		log.Print(fmt.Sprintf("  Component: %s", component.ComponentName))
		for _, podRequest := range component.PodRequests {
			logFile.WriteString(fmt.Sprintf("    Pod: %s, Request Count: %d\n", podRequest.PodName, podRequest.RequestCount))
			log.Print(fmt.Sprintf("    Pod: %s, Request Count: %d", podRequest.PodName, podRequest.RequestCount))
		}
	}
	logFile.WriteString(fmt.Sprintf("\n"))
	log.Print("")
	return requestCountData, nil
}

// getRequestDurationData는 요청 수 데이터를 기반으로 평균 요청 지속 시간을 수집하고 반환
func getRequestDurationData(promClient *prometheusClient.PrometheusClient, namespace string, componentGraph *graphGenerator.ComponentGraph, componentPodMap map[string][]string, duration string, requestCountData *metricCollector.RequestCountData, logFile *os.File) (*metricCollector.RequestDurationData, error) {
	requestDurationData, err := metricCollector.CollectRequestDurationPerPod(promClient, namespace, componentGraph, componentPodMap, duration, requestCountData)
	if err != nil {
		return nil, err
	}

	logFile.WriteString(fmt.Sprintf("Request Duration Per Pod:\n"))
	log.Print("Request Duration Per Pod:")
	for _, component := range requestDurationData.Components {
		logFile.WriteString(fmt.Sprintf("  Component: %s\n", component.ComponentName))
		log.Print(fmt.Sprintf("  Component: %s", component.ComponentName))
		for _, podDuration := range component.PodDurations {
			logFile.WriteString(fmt.Sprintf("    Pod: %s, Request Duration: %d ms\n", podDuration.PodName, podDuration.RequestDuration))
			log.Print(fmt.Sprintf("    Pod: %s, Request Duration: %d ms", podDuration.PodName, podDuration.RequestDuration))
		}
	}
	logFile.WriteString(fmt.Sprintf("\n"))
	log.Print("")
	return requestDurationData, nil
}

// getCpuUtilizationPerPod는 각 파드별 CPU 활용도를 수집하고 반환
func getCpuUtilizationPerPod(promClient *prometheusClient.PrometheusClient, namespace string, componentGraph *graphGenerator.ComponentGraph, componentPodMap map[string][]string, duration string, logFile *os.File) (*metricCollector.CpuUtilizationData, error) {
	cpuUtilizationData, err := metricCollector.CollectCpuUtilizationPerPod(promClient, namespace, componentGraph, componentPodMap, duration)
	if err != nil {
		return nil, err
	}

	logFile.WriteString(fmt.Sprintf("CPU Utilization Per Pod:\n"))
	log.Print("CPU Utilization Per Pod:")
	for _, component := range cpuUtilizationData.Components {
		logFile.WriteString(fmt.Sprintf("  Component: %s\n", component.ComponentName))
		log.Print(fmt.Sprintf("  Component: %s", component.ComponentName))
		for _, podCpu := range component.PodCpuUsage {
			logFile.WriteString(fmt.Sprintf("    Pod: %s, CPU Utilization: %d%%\n", podCpu.PodName, podCpu.CpuUtilization))
			log.Print(fmt.Sprintf("    Pod: %s, CPU Utilization: %d%%", podCpu.PodName, podCpu.CpuUtilization))
		}
	}
	logFile.WriteString(fmt.Sprintf("\n"))
	log.Print("")
	return cpuUtilizationData, nil
}

// getNodeCpuFrequencies는 노드별 CPU Hz를 수집
func getNodeCpuFrequencies(promClient *prometheusClient.PrometheusClient, logFile *os.File) ([]metricCollector.NodeCPUFrequencyData, error) {
	nodeFrequencies, err := metricCollector.CollectNodeCPUFrequency(promClient)
	if err != nil {
		return nil, err
	}

	// 수집된 데이터를 로그 파일에 기록
	logFile.WriteString("Node CPU Frequencies:\n")
	log.Print("Node CPU Frequencies:")
	for _, nodeFreq := range nodeFrequencies {
		logFile.WriteString(fmt.Sprintf("  NodeName: %s, Hertz: %d\n", nodeFreq.NodeName, nodeFreq.Hertz))
		log.Print(fmt.Sprintf("  NodeName: %s, Hertz: %d", nodeFreq.NodeName, nodeFreq.Hertz))
	}
	logFile.WriteString(fmt.Sprintf("\n"))
	log.Print("")

	return nodeFrequencies, nil
}

// getMapPodsToIP는 파드 이름을 키로, IP 주소와 관련 정보를 값으로 하는 맵을 반환
func getMapPodsToIP(promClient *prometheusClient.PrometheusClient, namespace string, logFile *os.File) (map[string]graphGenerator.PodInfo, error) {
	podInfoMap, err := graphGenerator.MapPodsToIP(promClient, namespace)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error mapping pods to IPs: %v\n", err))
		log.Print(fmt.Sprintf("Error mapping pods to IPs: %v", err))
		return nil, err
	}

	// 로그 파일에 기록
	logFile.WriteString(fmt.Sprintf("Pod to IP Mapping:\n"))
	log.Print("Pod to IP Mapping:")
	for podName, podInfo := range podInfoMap {
		logFile.WriteString(fmt.Sprintf("  PodName: %s, PodIP: %s, HostName: %s, HostIP: %s\n", podName, podInfo.PodIP, podInfo.HostName, podInfo.HostIP))
		log.Print(fmt.Sprintf("  PodName: %s, PodIP: %s, HostName: %s, HostIP: %s", podName, podInfo.PodIP, podInfo.HostName, podInfo.HostIP))
	}
	logFile.WriteString(fmt.Sprintf("\n"))
	log.Print("")

	return podInfoMap, nil
}
