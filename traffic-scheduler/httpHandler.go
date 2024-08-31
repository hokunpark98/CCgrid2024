package main

import (
	"fmt"
	"log"
	"net/http"
	"traffic-scheduler/prometheusClient"
)

// handleGetGraph는 /get-graph 요청을 처리하고 컴포넌트 그래프를 생성하여 응답
func handleGetGraph(w http.ResponseWriter, r *http.Request, promClient *prometheusClient.PrometheusClient) {
	logFile := generateLogFile()
	defer logFile.Close()

	namespace := r.URL.Query().Get("namespace")

	if namespace == "" {
		http.Error(w, "namespace parameter is required", http.StatusBadRequest)
		logFile.WriteString("Missing parameter: namespace\n")
		log.Print("Missing parameter: namespace")
		return
	}

	componentGraph, err := generateComponentGraph(promClient, namespace, logFile)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Failed to generate graph: %v\n", err))
		log.Print(fmt.Sprintf("Failed to generate graph: %v", err))
		http.Error(w, "Failed to generate graph", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Graph generated successfully for namespace: %s\n", namespace)))

	fmt.Printf("Component Graph: %+v\n", componentGraph)
}

// handleGetMonitoringInfo는 /get-monitoring-info 요청을 처리하여 컴포넌트 그래프를 생성하고 모니터링 정보를 반환
func handleGetMonitoringInfo(w http.ResponseWriter, r *http.Request, promClient *prometheusClient.PrometheusClient) {
	logFile := generateLogFile()
	defer logFile.Close()

	namespace := r.URL.Query().Get("namespace")
	duration := r.URL.Query().Get("duration")

	if namespace == "" || duration == "" {
		http.Error(w, "namespace and duration parameters are required", http.StatusBadRequest)
		logFile.WriteString("Missing parameters: namespace or duration\n")
		log.Print("Missing parameters: namespace or duration")
		return
	}

	componentGraph, err := generateComponentGraph(promClient, namespace, logFile)

	if err != nil {
		logFile.WriteString(fmt.Sprintf("Failed to generate graph: %v\n", err))
		log.Print(fmt.Sprintf("Failed to generate graph: %v", err))
		http.Error(w, "Failed to generate graph", http.StatusInternalServerError)
		return
	}

	componentPodMap := generateComponentPodMap(promClient, componentGraph, logFile)

	requestCountData, err := getRequestCountPerPod(promClient, namespace, componentGraph, componentPodMap, duration, logFile)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting request count data: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting request count data: %v", err))
		http.Error(w, "Error collecting request count data", http.StatusInternalServerError)
		return
	}

	_, err = getRequestDurationData(promClient, namespace, componentGraph, componentPodMap, duration, requestCountData, logFile)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting request duration data: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting request duration data: %v", err))
		http.Error(w, "Error collecting request duration data", http.StatusInternalServerError)
		return
	}

	_, err = getCpuUtilizationPerPod(promClient, namespace, componentGraph, componentPodMap, duration, logFile)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting CPU utilization data: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting CPU utilization data: %v", err))
		http.Error(w, "Error collecting CPU utilization data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Monitoring information collected successfully\n"))
}

// handleGetNodeCpuHz는 /get-node-cpu-hz 요청을 처리하고 각 노드의 CPU 주파수를 응답합니다.
func handleGetNodeCpuHz(w http.ResponseWriter, promClient *prometheusClient.PrometheusClient) {
	logFile := generateLogFile()
	defer logFile.Close()

	_, err := getNodeCpuFrequencies(promClient, logFile)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting node CPU frequencies: %v\n", err))
		http.Error(w, "Error collecting node CPU frequencies", http.StatusInternalServerError)
		return
	}

	log.Printf("Node CPU frequencies successfully collected and sent.")
}

func handleTrafficSchedule(w http.ResponseWriter, r *http.Request, promClient *prometheusClient.PrometheusClient) {
	logFile := generateLogFile()
	defer logFile.Close()

	namespace := r.URL.Query().Get("namespace")
	duration := r.URL.Query().Get("duration")

	if namespace == "" || duration == "" {
		http.Error(w, "namespace and duration parameters are required", http.StatusBadRequest)
		logFile.WriteString("Missing parameters: namespace or duration\n")
		log.Print("Missing parameters: namespace or duration")
		return
	}

	componentGraph, err := generateComponentGraph(promClient, namespace, logFile)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Failed to generate graph: %v\n", err))
		log.Print(fmt.Sprintf("Failed to generate graph: %v", err))
		http.Error(w, "Failed to generate graph", http.StatusInternalServerError)
		return
	}

	componentPodMap := generateComponentPodMap(promClient, componentGraph, logFile)

	podInfoMap, err := getMapPodsToIP(promClient, namespace, logFile)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting pod to IP map: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting pod to IP map: %v", err))
		http.Error(w, "Error collecting pod to IP map", http.StatusInternalServerError)
		return
	}

	requestCountData, err := getRequestCountPerPod(promClient, namespace, componentGraph, componentPodMap, duration, logFile)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting request count data: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting request count data: %v", err))
		http.Error(w, "Error collecting request count data", http.StatusInternalServerError)
		return
	}

	requestDurationData, err := getRequestDurationData(promClient, namespace, componentGraph, componentPodMap, duration, requestCountData, logFile)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting request duration data: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting request duration data: %v", err))
		http.Error(w, "Error collecting request duration data", http.StatusInternalServerError)
		return
	}

	cpuUtilizationData, err := getCpuUtilizationPerPod(promClient, namespace, componentGraph, componentPodMap, duration, logFile)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting CPU utilization data: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting CPU utilization data: %v", err))
		http.Error(w, "Error collecting CPU utilization data", http.StatusInternalServerError)
		return
	}

	nodeFrequencies, err := getNodeCpuFrequencies(promClient, logFile)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Error collecting node CPU frequencies: %v\n", err))
		log.Print(fmt.Sprintf("Error collecting node CPU frequencies: %v", err))
		http.Error(w, "Error collecting node CPU frequencies", http.StatusInternalServerError)
		return
	}

	//exeTrafficAllocator(componentGraph, )

	// 응답 작성
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Traffic scheduling completed successfully\n"))

	// 결과를 콘솔에 출력
	fmt.Printf("Pod Info Data: %+v\n", componentGraph)
	fmt.Printf("Request Duration Data: %+v\n", requestDurationData)
	fmt.Printf("CPU Utilization Data: %+v\n", cpuUtilizationData)
	fmt.Printf("CPU Hz Data: %+v\n", nodeFrequencies)
	fmt.Printf("Pod Info Data: %+v\n", podInfoMap)
}
