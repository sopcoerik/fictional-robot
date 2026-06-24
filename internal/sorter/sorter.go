package sorter

import (
	"log"

	"github.com/sopcoerik/fictional-robot/internal/parser"
)

func SortServices(config *parser.Config) []string {

    // inDegree: how many dependencies this service is waiting on
    inDegree := make(map[string]int)

    // dependentsOf: list of services that depend on this one
    dependentsOf := make(map[string][]string)
    
    var readyQueue []string

    // 1. Build the graph and calculate in-degrees
    for name, data := range config.Services {
		
        // Initialize every service with 0 in-degree
        if _, exists := inDegree[name]; !exists {
            inDegree[name] = 0
        }

        for _, dep := range data.DependsOn {
            inDegree[name]++
            dependentsOf[dep] = append(dependentsOf[dep], name)
        }
    }

    // 2. Seed the queue with services that have no dependencies
    for name, degree := range inDegree {
        if degree == 0 {
            readyQueue = append(readyQueue, name)
        }
    }

    // 3. Process the queue (BFS)
    // We use a simple slice as a queue, iterating by index
    for i := 0; i < len(readyQueue); i++ {
        current := readyQueue[i]

        // For every service waiting on the current one...
        for _, dependent := range dependentsOf[current] {
            inDegree[dependent]-- // One dependency is now satisfied

            // If it has no more pending dependencies, it's ready
            if inDegree[dependent] == 0 {
                readyQueue = append(readyQueue, dependent)
            }
        }
    }

    // 4. Check for cycles
    if len(readyQueue) != len(config.Services) {
        log.Fatal("Circular dependency detected. A cycle exists in your service graph.")
    }

    return readyQueue
}
