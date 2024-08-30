package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"traffic-scheduler/graphGenerator"
	"traffic-scheduler/prometheusClient"
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
