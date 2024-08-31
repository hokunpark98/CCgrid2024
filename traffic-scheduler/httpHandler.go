package main

import (
	"fmt"
	"log"
	"net/http"
	"traffic-scheduler/graphGenerator"
	"traffic-scheduler/logging"
	"traffic-scheduler/metricCollector"
	"traffic-scheduler/prometheusClient"
)

// handleGetGraph는 /get-graph 요청을 처리하고 컴포넌트 그래프를 생성하여 응답
func handleGetGraph(w http.ResponseWriter, r *http.Request, promClient *prometheusClient.PrometheusClient) {
	namespace := r.URL.Query().Get("namespace")

	if namespace == "" {
		logging.Alert(w, "fail\n")
		return
	}

	logFile := logging.GenerateLogFile()
	defer logFile.Close()

	componentGraph, err := graphGenerator.GenerateGraph(promClient, namespace)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Failed to generate graph: %v\n", err))
		log.Print(fmt.Sprintf("Failed to generate graph: %v", err))
		return
	}
	logging.LogComponentGraph(componentGraph, logFile)

	fmt.Printf("Component Graph: %+v\n", componentGraph)
}

// handleGetMonitoringInfo는 /get-monitoring-info 요청을 처리하여 컴포넌트 그래프를 생성하고 모니터링 정보를 반환
func handleGetMonitoringInfo(w http.ResponseWriter, r *http.Request, promClient *prometheusClient.PrometheusClient) {
	namespace := r.URL.Query().Get("namespace")
	duration := r.URL.Query().Get("duration")

	if namespace == "" || duration == "" {
		logging.Alert(w, "fail\n")
		return
	}

	logFile := logging.GenerateLogFile()
	defer logFile.Close()

	componentGraph, err := graphGenerator.GenerateGraph(promClient, namespace)
	if err != nil {
		logMessage := fmt.Sprintf("Failed to generate graph: %v\n", err)
		logging.LogMessage(logFile, logMessage)
		logging.Alert(w, "fail\n")
		return
	}
	logging.LogComponentGraph(componentGraph, logFile)

	componentPodMap, err := metricCollector.MapComponentToPodInfo(promClient, namespace)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Failed to generate component pod map: %v\n", err))
		log.Print(fmt.Sprintf("Failed to generate component pod map: %v", err))
		return
	}
	logging.LogComponentPodMap(componentPodMap, logFile)

	requestCountMap, err := metricCollector.CollectRequestCountPerPod(promClient, namespace, componentGraph, componentPodMap, duration)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting request count data: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting request count data: %v\n", err))
		return
	}
	logging.LogRequestCountPerPod(requestCountMap, logFile)

	requestDurationMap, err := metricCollector.CollectRequestDurationPerPod(promClient, namespace, componentGraph, componentPodMap, duration, requestCountMap)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting request duration data: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting request duration data: %v", err))
		return
	}
	logging.LogRequestDurationData(requestDurationMap, logFile)

	nodeCPUFrequencyMap, err := metricCollector.CollectNodeCPUFrequency(promClient)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting CPU utilization data: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting CPU utilization data: %v", err))
		return
	}
	logging.LogNodeCpuFrequencies(nodeCPUFrequencyMap, logFile)
}

// handleGetNodeCpuHz는 /get-node-cpu-hz 요청을 처리하고 각 노드의 CPU 주파수를 응답합니다.
func handleGetNodeCpuHz(w http.ResponseWriter, promClient *prometheusClient.PrometheusClient) {
	logFile := logging.GenerateLogFile()
	defer logFile.Close()

	nodeCPUFrequencyMap, err := metricCollector.CollectNodeCPUFrequency(promClient)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting CPU utilization data: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting CPU utilization data: %v", err))
		return
	}
	logging.LogNodeCpuFrequencies(nodeCPUFrequencyMap, logFile)

	log.Printf("Node CPU frequencies successfully collected and sent.")
}

func handleTrafficSchedule(w http.ResponseWriter, r *http.Request, promClient *prometheusClient.PrometheusClient) {
	namespace := r.URL.Query().Get("namespace")
	duration := r.URL.Query().Get("duration")

	if namespace == "" || duration == "" {
		logging.Alert(w, "fail\n")
		return
	}

	logFile := logging.GenerateLogFile()
	defer logFile.Close()

	componentGraph, err := graphGenerator.GenerateGraph(promClient, namespace)
	if err != nil {
		logMessage := fmt.Sprintf("Failed to generate graph: %v\n", err)
		logging.LogMessage(logFile, logMessage)
		logging.Alert(w, "fail\n")
		return
	}
	logging.LogComponentGraph(componentGraph, logFile)

	componentPodMap, err := metricCollector.MapComponentToPodInfo(promClient, namespace)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Failed to generate component pod map: %v\n", err))
		log.Print(fmt.Sprintf("Failed to generate component pod map: %v", err))
		return
	}
	logging.LogComponentPodMap(componentPodMap, logFile)

	requestCountMap, err := metricCollector.CollectRequestCountPerPod(promClient, namespace, componentGraph, componentPodMap, duration)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting request count data: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting request count data: %v\n", err))
		return
	}
	logging.LogRequestCountPerPod(requestCountMap, logFile)

	requestDurationMap, err := metricCollector.CollectRequestDurationPerPod(promClient, namespace, componentGraph, componentPodMap, duration, requestCountMap)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting request duration data: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting request duration data: %v", err))
		return
	}
	logging.LogRequestDurationData(requestDurationMap, logFile)

	cpuUtilizationMap, err := metricCollector.CollectCpuUtilizationPerPod(promClient, namespace, componentPodMap, duration)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting CPU utilization data: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting CPU utilization data: %v", err))
		return
	}
	logging.LogCpuUtilizationPerPod(cpuUtilizationMap, logFile)

	nodeCPUFrequencyMap, err := metricCollector.CollectNodeCPUFrequency(promClient)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting CPU utilization data: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting CPU utilization data: %v", err))
		return
	}
	logging.LogNodeCpuFrequencies(nodeCPUFrequencyMap, logFile)

	// 결과를 콘솔에 출력
	fmt.Printf("Pod Info Data: %+v\n", componentGraph)

	fmt.Printf("Request Duration Data: %+v\n", requestDurationMap)
	fmt.Printf("CPU Utilization Data: %+v\n", cpuUtilizationMap)
	fmt.Printf("CPU Hz Data: %+v\n", nodeCPUFrequencyMap)
}
