package trafficAllocator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"traffic-scheduler/graphGenerator"
	"traffic-scheduler/metricCollector"
)

// DestinationPodProportionData는 각 소스 파드에 대해 대상 파드에 할당된 트래픽 비율을 나타냅니다.
type DestinationPodProportionData struct {
	DestinationPodName string `json:"destination_pod_name"`
	Proportion         int    `json:"proportion"` // 퍼센트 단위
}

// SourcePodData는 각 소스 파드에 대해 여러 대상 파드에 대한 트래픽 비율을 저장합니다.
type SourcePodData struct {
	SourcePodName   string                         `json:"source_pod_name"`
	ProportionDatas []DestinationPodProportionData `json:"proportion_datas"`
}

// ProportionMap은 소스 컴포넌트에 대해 속한 모든 소스 파드와 그에 대한 대상 파드의 비율을 저장합니다.
type ProportionMap struct {
	Components map[string]map[string][]SourcePodData `json:"components"`
}

// calProportionComponentPair는 각 링크에서 소스 컴포넌트와 대상 컴포넌트에 대해 proportion을 계산하여 반환
func calProportionComponentPair(sourceComponentName, destinationComponentName string, componentPodMap *metricCollector.ComponentPodMap) []SourcePodData {
	sourcePods := componentPodMap.Components[sourceComponentName]
	destinationPods := componentPodMap.Components[destinationComponentName]

	numDestinationPods := len(destinationPods)
	proportionPerPod := 100 / numDestinationPods
	remainder := 100 % numDestinationPods

	var sourcePodDataList []SourcePodData

	for _, sourcePod := range sourcePods {
		var proportionDataList []DestinationPodProportionData
		accumulatedProportion := 0

		for i, destinationPod := range destinationPods {
			proportion := proportionPerPod
			if i < remainder {
				proportion += 1 // 남은 퍼센트를 앞쪽 파드들에게 배분
			}
			accumulatedProportion += proportion

			proportionDataList = append(proportionDataList, DestinationPodProportionData{
				DestinationPodName: destinationPod.PodName,
				Proportion:         accumulatedProportion,
			})
		}

		sourcePodDataList = append(sourcePodDataList, SourcePodData{
			SourcePodName:   sourcePod.PodName,
			ProportionDatas: proportionDataList,
		})
	}

	return sourcePodDataList
}

// TrafficAllocation은 componentGraph의 링크를 순회하며 proportion을 계산하여 ProportionMap에 저장하고 반환합니다.
func TrafficAllocation(componentGraph *graphGenerator.ComponentGraph, componentPodMap *metricCollector.ComponentPodMap, namespace string) *ProportionMap {
	// ProportionMap 구조체 초기화
	proportionMap := &ProportionMap{
		Components: make(map[string]map[string][]SourcePodData),
	}

	// componentGraph의 각 링크에 대해 반복문 실행
	for sourceComponent, destinationComponents := range componentGraph.Components {
		for _, destinationComponent := range destinationComponents {
			// 각 링크에 대해 calProportionComponentPair를 실행하여 proportion 계산
			sourcePodDataList := calProportionComponentPair(sourceComponent, destinationComponent, componentPodMap)

			// 결과를 proportion map에 업데이트
			if proportionMap.Components[sourceComponent] == nil {
				proportionMap.Components[sourceComponent] = make(map[string][]SourcePodData)
			}
			proportionMap.Components[sourceComponent][destinationComponent] = sourcePodDataList
		}
	}

	GenerateEnvoyLuaScript(namespace, proportionMap, componentPodMap)
	// ProportionMap 반환
	return proportionMap
}

// GenerateEnvoyLuaScript는 트래픽 할당 결과에 따라 Lua 스크립트를 생성합니다.
func GenerateEnvoyLuaScript(namespace string, proportionMap *ProportionMap, componentPodMap *metricCollector.ComponentPodMap) {
	for sourceComponent, destMap := range proportionMap.Components {
		var sb strings.Builder

		sb.WriteString(fmt.Sprintf(
			"apiVersion: networking.istio.io/v1alpha3\n"+
				"kind: EnvoyFilter\n"+
				"metadata:\n"+
				"  name: %s-filter\n"+
				"  namespace: %s\n"+
				"spec:\n"+
				"  workloadSelector:\n"+
				"    labels:\n"+
				"      app: %s\n"+
				"  configPatches:\n"+
				"  - applyTo: HTTP_FILTER\n"+
				"    match:\n"+
				"      context: SIDECAR_OUTBOUND\n"+
				"      listener:\n"+
				"        filterChain:\n"+
				"          filter:\n"+
				"            name: envoy.filters.network.http_connection_manager\n"+
				"    patch:\n"+
				"      operation: INSERT_BEFORE\n"+
				"      value:\n"+
				"        name: envoy.filters.http.lua\n"+
				"        typed_config:\n"+
				"          \"@type\": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua\n"+
				"          inline_code: |\n"+
				"            local pod_ip = nil\n\n"+
				"            function envoy_on_request(request_handle)\n"+
				"              if not pod_ip then\n"+
				"                local handle = io.popen(\"hostname -i\")\n"+
				"                pod_ip = handle:read(\"*a\"):match(\"^%%s*(.-)%%s*$\")\n"+
				"                handle:close()\n"+
				"              end\n\n"+
				"              local rand = math.random(0, 100)\n"+
				"              local destination = request_handle:headers():get(\":authority\")\n"+
				"              local domain = destination:match(\"^([^:]+)\")\n\n",
			sourceComponent, namespace, sourceComponent))

		for destinationComponent, sourcePodDataList := range destMap {
			sb.WriteString(fmt.Sprintf("              if domain == \"%s\" then\n", destinationComponent))
			for i, sourcePodData := range sourcePodDataList {
				if i == 0 {
					sb.WriteString(fmt.Sprintf("                  if pod_ip == \"%s\" then\n", componentPodMap.Pods[sourcePodData.SourcePodName].PodIP))
				} else {
					sb.WriteString(fmt.Sprintf("                  elseif pod_ip == \"%s\" then\n", componentPodMap.Pods[sourcePodData.SourcePodName].PodIP))
				}
				for j, proportionData := range sourcePodData.ProportionDatas {
					if j == 0 {
						sb.WriteString(fmt.Sprintf("                      if rand <= %d then\n                        destination = \"%s\" .. destination:match(\"(:.*)$\")\n", proportionData.Proportion, componentPodMap.Pods[proportionData.DestinationPodName].PodIP))
					} else {
						sb.WriteString(fmt.Sprintf("                      elseif rand <= %d then\n                        destination = \"%s\" .. destination:match(\"(:.*)$\")\n", proportionData.Proportion, componentPodMap.Pods[proportionData.DestinationPodName].PodIP))
					}
				}
				sb.WriteString("                      end\n")

			}
			sb.WriteString("                  end\n")
			sb.WriteString("              end\n")
		}

		sb.WriteString("              request_handle:headers():replace(\":authority\", destination)\n            end\n")

		// 파일 경로 생성
		fileName := fmt.Sprintf("%s.yaml", sourceComponent)
		filePath := filepath.Join("envoyfilterResults", fileName)

		// 파일 쓰기
		err := os.WriteFile(filePath, []byte(sb.String()), 0644)
		if err != nil {
			fmt.Printf("Error writing file %s: %v\n", filePath, err)
		} else {
			fmt.Printf("File %s successfully written.\n", filePath)
		}
	}
}
