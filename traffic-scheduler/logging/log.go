package logging

import (
	"fmt"
	"log"
	"os"
	"traffic-scheduler/graphGenerator"
	"traffic-scheduler/metricCollector"
)

func LogComponentGraph(componentGraph *graphGenerator.ComponentGraph, logFile *os.File) {
	logFile.WriteString("Component Graph:\n")
	log.Print("Component Graph:")
	logFile.WriteString("  Components:\n")
	log.Print("  Components:")

	for component := range componentGraph.Components {
		logFile.WriteString("    - " + component + "\n")
		log.Print("    - " + component)

		logFile.WriteString("  Links:\n")
		log.Print("  Links:")
		for _, link := range componentGraph.Components[component] {
			logFile.WriteString("    - [" + component + " -> " + link + "]\n")
			log.Print("    - [" + component + " -> " + link + "]")
		}
		logFile.WriteString(fmt.Sprintf("\n"))
		log.Print("")
	}
}

func LogComponentPodMap(componentPodMap *metricCollector.ComponentPodMap, logFile *os.File) {
	logFile.WriteString("Component to Pod Map:\n")
	log.Print("Component to Pod Map:\n")
	for component, pods := range componentPodMap.Components {
		logFile.WriteString("  Comonent:" + component + "\n")
		log.Print("  Comonent:" + component + "\n")

		for _, pod := range pods {
			logFile.WriteString("    [" + pod.PodName + ", " + pod.PodIP + ", " +
				pod.HostName + ", " + pod.HostIP + "]\n")
			log.Print("    [" + pod.PodName + ", " + pod.PodIP + ", " +
				pod.HostName + ", " + pod.HostIP + "]\n")
		}
	}
	logFile.WriteString(fmt.Sprintf("\n"))
	log.Print("")
}

func LogRequestCountPerPod(requestCountMap *metricCollector.RequestCountMap, logFile *os.File) {
	logFile.WriteString(fmt.Sprintf("Request Count Per Pod:\n"))
	log.Print("Request Count Per Pod:")
	for _, component := range requestCountMap.Components {
		logFile.WriteString(fmt.Sprintf("  Component: %s\n", component.ComponentName))
		log.Print(fmt.Sprintf("  Component: %s", component.ComponentName))
		for _, podRequest := range component.PodRequests {
			logFile.WriteString(fmt.Sprintf("    Pod: %s, Request Count: %d\n", podRequest.PodName, podRequest.RequestCount))
			log.Print(fmt.Sprintf("    Pod: %s, Request Count: %d", podRequest.PodName, podRequest.RequestCount))
		}
	}
	logFile.WriteString(fmt.Sprintf("\n"))
	log.Print("")
}

func LogRequestDurationData(requestDurationMap *metricCollector.RequestDurationMap, logFile *os.File) {
	logFile.WriteString(fmt.Sprintf("Request Duration Per Pod:\n"))
	log.Print("Request Duration Per Pod:")
	for _, component := range requestDurationMap.Components {
		logFile.WriteString(fmt.Sprintf("  Component: %s\n", component.ComponentName))
		log.Print(fmt.Sprintf("  Component: %s", component.ComponentName))
		for _, podDuration := range component.PodDurations {
			logFile.WriteString(fmt.Sprintf("    Pod: %s, Request Duration: %d ms\n", podDuration.PodName, podDuration.RequestDuration))
			log.Print(fmt.Sprintf("    Pod: %s, Request Duration: %d ms", podDuration.PodName, podDuration.RequestDuration))
		}
	}
	logFile.WriteString(fmt.Sprintf("\n"))
	log.Print("")
}

func LogCpuUtilizationPerPod(cpuUtilizationMap *metricCollector.CpuUtilizationMap, logFile *os.File) {
	logFile.WriteString(fmt.Sprintf("CPU Utilization Per Pod:\n"))
	log.Print("CPU Utilization Per Pod:")
	for _, component := range cpuUtilizationMap.Components {
		logFile.WriteString(fmt.Sprintf("  Component: %s\n", component.ComponentName))
		log.Print(fmt.Sprintf("  Component: %s", component.ComponentName))
		for _, podCpu := range component.PodCpuUsage {
			logFile.WriteString(fmt.Sprintf("    Pod: %s, CPU Utilization: %d%%\n", podCpu.PodName, podCpu.CpuUtilization))
			log.Print(fmt.Sprintf("    Pod: %s, CPU Utilization: %d%%", podCpu.PodName, podCpu.CpuUtilization))
		}
	}
	logFile.WriteString(fmt.Sprintf("\n"))
	log.Print("")
}

func LogNodeCpuFrequencies(nodeCPUFrequencyMap *metricCollector.NodeCPUFrequencyMap, logFile *os.File) {
	logFile.WriteString("Node CPU Frequencies:\n")
	log.Print("Node CPU Frequencies:")
	for node, nodeFreq := range nodeCPUFrequencyMap.Nodes {
		logFile.WriteString(fmt.Sprintf("  NodeName: %s, Hertz: %d\n", node, nodeFreq.Hertz))
		log.Print(fmt.Sprintf("  NodeName: %s, Hertz: %d", node, nodeFreq.Hertz))
	}
	logFile.WriteString(fmt.Sprintf("\n"))
	log.Print("")
}
