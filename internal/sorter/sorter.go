package sorter

import (
	"log"

	"github.com/sopcoerik/fictional-robot/internal/parser"
)

func SortServices(config *parser.Config) ([]string) {
	indegree := make(map[string]int)
	dependencyGraph := make(map[string][]string)
	var queue []string

	for sName, sData := range(config.Services) {
		indegree[sName] = 0

		for _, service := range(sData.DependsOn) {
			indegree[sName] += 1
			dependencyGraph[service] = append(dependencyGraph[service], sName)
		}

		if indegree[sName] == 0 {
			queue = append(queue, sName)
		}
	}

	var orderedQueue []string

	for i := 0; i < len(queue); i += 1 {
		sName := queue[i]

		orderedQueue = append(orderedQueue, sName)

		for _, s := range(dependencyGraph[sName]) {

			indegree[s] -= 1

			if indegree[s] == 0 {
				queue = append(queue, s)

			}
		}
	}

	if len(orderedQueue) != len(config.Services) {
		log.Fatal("Circular dependency detected. Quitting program...")
	}

	return orderedQueue
}
