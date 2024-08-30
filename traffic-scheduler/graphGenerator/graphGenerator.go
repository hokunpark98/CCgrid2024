package graphGenerator

import (
	"fmt"

	"traffic-scheduler/prometheusClient"

	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type ComponentGraph struct {
	Components []string `json:"components"`
	Links      []Link   `json:"links"`
}

type Link struct {
	UC string `json:"uc"`
	DC string `json:"dc"`
}

func GenerateGraph(promClient *prometheusClient.PrometheusClient, namespace string) (*ComponentGraph, error) {
	query := fmt.Sprintf(`istio_requests_total{kubernetes_namespace="%s"}`, namespace)

	result, err := promClient.Query(query)
	if err != nil {
		return nil, err
	}

	graph := simple.NewDirectedGraph()
	nodeMap := make(map[string]int64)
	nodeNames := make(map[int64]string)

	for _, sample := range result {
		source := string(sample.Metric["source_app"])
		dest := string(sample.Metric["destination_app"])

		if source == "unknown" || dest == "unknown" {
			continue
		}

		if _, exists := nodeMap[source]; !exists {
			node := graph.NewNode()
			nodeMap[source] = node.ID()
			nodeNames[node.ID()] = source
			graph.AddNode(node)
		}

		if _, exists := nodeMap[dest]; !exists {
			node := graph.NewNode()
			nodeMap[dest] = node.ID()
			nodeNames[node.ID()] = dest
			graph.AddNode(node)
		}

		graph.SetEdge(graph.NewEdge(graph.Node(nodeMap[source]), graph.Node(nodeMap[dest])))
	}

	sortedNodes, err := topo.Sort(graph)
	if err != nil {
		return nil, fmt.Errorf("Graph is not a DAG")
	}

	var components []string
	for _, node := range sortedNodes {
		components = append(components, nodeNames[node.ID()])
	}

	var links []Link
	it := graph.Edges()
	for it.Next() {
		edge := it.Edge()
		fromName := nodeNames[edge.From().ID()]
		toName := nodeNames[edge.To().ID()]
		links = append(links, Link{UC: fromName, DC: toName})
	}

	return &ComponentGraph{
		Components: components,
		Links:      links,
	}, nil
}
