package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"traffic-scheduler/graphGenerator"
	"traffic-scheduler/prometheusClient"
	"traffic-scheduler/trafficMonitor"
)

// generateLogFile는 요청이 올 때마다 새로운 로그 파일을 생성하고 로그를 기록.
func generateLogFile() *os.File {
	// 한국 시간대 설정
	loc, _ := time.LoadLocation("Asia/Seoul")
	now := time.Now().In(loc)

	// 파일 이름을 현재 시간으로 설정
	filename := filepath.Join("log", now.Format("2006-01-02_15-04-05")+".log")

	// 파일 생성
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}

	// 첫 줄에 현재 시간 기록
	logFile.WriteString("Log start time: " + now.Format("2006-01-02 15:04:05") + "\n")
	logFile.WriteString("---------------------------------------------\n")
	logFile.WriteString(fmt.Sprintf("\n\n"))
	log.Print("Log start time: " + now.Format("2006-01-02 15:04:05"))
	log.Print("---------------------------------------------")
	return logFile
}

// generateComponentGraph는 namespace를 기반으로 구성하고 있는 Component들과 Link를 나타냄.
// Link는 [[uc dc], [uc dc] .... ]로 나타냄
func generateComponentGraph(promClient *prometheusClient.PrometheusClient, namespace string, logFile *os.File) (*graphGenerator.ComponentGraph, error) {
	componentGraph, err := graphGenerator.GenerateGraph(promClient, namespace)
	if err != nil {
		return nil, err
	}

	// Component Graph 로그 포맷팅
	logFile.WriteString("Component Graph:\n")
	log.Print("Component Graph:")
	logFile.WriteString("  Components:\n")
	log.Print("  Components:")
	for _, component := range componentGraph.Components {
		logFile.WriteString("    - " + component + "\n")
		log.Print("    - " + component)
	}

	logFile.WriteString("  Links:\n")
	log.Print("  Links:")
	for _, link := range componentGraph.Links {
		logFile.WriteString("    - [" + link.UC + " -> " + link.DC + "]\n")
		log.Print("    - [" + link.UC + " -> " + link.DC + "]")
	}
	logFile.WriteString(fmt.Sprintf("\n"))
	log.Print("")
	return componentGraph, nil
}

// generateComponentPodMap는 컴포넌트와 endpoint에 해당하는 파드들을 매핑함. {component: [pod1, pod2, ..], ...}
func generateComponentPodMap(promClient *prometheusClient.PrometheusClient, componentGraph *graphGenerator.ComponentGraph, logFile *os.File) map[string][]string {
	componentPodMap := graphGenerator.MapComponentsToPods(promClient, componentGraph)

	// Component to Pod Map 로그 포맷팅
	logFile.WriteString("Component to Pod Map:\n")
	log.Print("Component to Pod Map:")
	for component, pods := range componentPodMap {
		logFile.WriteString("  " + component + ": [" + strings.Join(pods, ", ") + "]\n")
		log.Print("  " + component + ": [" + strings.Join(pods, ", ") + "]")
	}
	logFile.WriteString(fmt.Sprintf("\n"))
	log.Print("")
	return componentPodMap
}

// getRequestCountPerPod는 DC의 각 파드들이 UC로부터 수신한 요청 수를 가져옴
func getRequestCountPerPod(promClient *prometheusClient.PrometheusClient, namespace string, componentGraph *graphGenerator.ComponentGraph, componentPodMap map[string][]string, duration string, logFile *os.File) (*trafficMonitor.RequestCountData, error) {
	requestCountData, err := trafficMonitor.CollectRequestCountPerPod(promClient, namespace, componentGraph, componentPodMap, duration)
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
func getRequestDurationData(promClient *prometheusClient.PrometheusClient, namespace string, componentGraph *graphGenerator.ComponentGraph, componentPodMap map[string][]string, duration string, requestCountData *trafficMonitor.RequestCountData, logFile *os.File) (*trafficMonitor.RequestDurationData, error) {
	requestDurationData, err := trafficMonitor.CollectRequestDurationPerPod(promClient, namespace, componentGraph, componentPodMap, duration, requestCountData)
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
func getCpuUtilizationPerPod(promClient *prometheusClient.PrometheusClient, namespace string, componentGraph *graphGenerator.ComponentGraph, componentPodMap map[string][]string, duration string, logFile *os.File) (*trafficMonitor.CpuUtilizationData, error) {
	cpuUtilizationData, err := trafficMonitor.CollectCpuUtilizationPerPod(promClient, namespace, componentGraph, componentPodMap, duration)
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

func main() {
	// Prometheus 클라이언트 생성
	promClient, err := prometheusClient.NewPrometheusClient("http://10.106.247.30:8080")
	if err != nil {
		log.Fatalf("Failed to create Prometheus client: %v", err)
	}

	// handleTrafficScheduler는 /traffic-scheduler 요청을 처리
	http.HandleFunc("/traffic-scheduler", func(w http.ResponseWriter, r *http.Request) {
		// 로그 파일 생성 및 설정
		logFile := generateLogFile()
		defer logFile.Close()

		namespace := r.URL.Query().Get("namespace")
		duration := r.URL.Query().Get("duration")

		// 필수 파라미터 체크
		if namespace == "" || duration == "" {
			http.Error(w, "namespace and duration parameters are required", http.StatusBadRequest)
			logFile.WriteString("Missing parameters: namespace or duration\n")
			log.Print("Missing parameters: namespace or duration")
			return
		}

		// component 그래프 생성 (component 간 의존성 확인)
		componentGraph, err := generateComponentGraph(promClient, namespace, logFile)
		if err != nil {
			logFile.WriteString(fmt.Sprintf("Failed to generate graph: %v\n", err))
			log.Print(fmt.Sprintf("Failed to generate graph: %v", err))
			http.Error(w, "Failed to generate graph", http.StatusInternalServerError)
			return
		}

		// 트래픽 할당 (component와 Pod 매핑)
		componentPodMap := generateComponentPodMap(promClient, componentGraph, logFile)

		// 요청 데이터를 수집하고 로그 파일에 기록
		requestCountData, err := getRequestCountPerPod(promClient, namespace, componentGraph, componentPodMap, duration, logFile)
		if err != nil {
			logFile.WriteString(fmt.Sprintf("Error collecting requestCount data: %v\n", err))
			log.Print(fmt.Sprintf("Error collecting requestCount data: %v", err))
			http.Error(w, "Error collecting requestCount data", http.StatusInternalServerError)
			return
		}

		// 요청 수 데이터를 기반으로 평균 요청 지속 시간을 수집
		requestDurationData, err := getRequestDurationData(promClient, namespace, componentGraph, componentPodMap, duration, requestCountData, logFile)
		if err != nil {
			logFile.WriteString(fmt.Sprintf("Error collecting request duration data: %v\n", err))
			log.Print(fmt.Sprintf("Error collecting request duration data: %v", err))
			http.Error(w, "Error collecting request duration data", http.StatusInternalServerError)
			return
		}

		// 파드별 CPU 활용도를 수집
		cpuUtilizationData, err := getCpuUtilizationPerPod(promClient, namespace, componentGraph, componentPodMap, duration, logFile)
		if err != nil {
			logFile.WriteString(fmt.Sprintf("Error collecting CPU utilization data: %v\n", err))
			log.Print(fmt.Sprintf("Error collecting CPU utilization data: %v", err))
			http.Error(w, "Error collecting CPU utilization data", http.StatusInternalServerError)
			return
		}

		// 사용자에게 응답
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Traffic scheduling completed successfully\n"))

		// RequestDurationData 및 CpuUtilizationData 사용 예시
		fmt.Printf("Collected request duration data: %+v\n", requestDurationData)
		fmt.Printf("Collected CPU utilization data: %+v\n", cpuUtilizationData)
	})

	log.Println("Server starting on port 13000")
	if err := http.ListenAndServe(":13000", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
