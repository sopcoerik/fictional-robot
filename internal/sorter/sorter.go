package sorter

import (
	"fmt"
	"sort"

	"github.com/sopcoerik/fictional-robot/internal/parser"
)

func SortServices(config *parser.Config) ([]string, error) {

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
            if _, ok := config.Services[dep]; !ok {
                return nil, fmt.Errorf("service %q depends on %q, which is not defined", name, dep)
            }
            inDegree[name]++
            dependentsOf[dep] = append(dependentsOf[dep], name)
        }
    }

    // 2. Seed the queue with services that have no dependencies, in a stable
    //    (alphabetical) order so the startup order is deterministic instead of
    //    depending on Go's random map iteration
    names := make([]string, 0, len(inDegree))
    for name := range inDegree {
        names = append(names, name)
    }
    sort.Strings(names)

    for _, name := range names {
        if inDegree[name] == 0 {
            readyQueue = append(readyQueue, name)
        }
    }

    // 3. Process the queue (BFS)
    // We use a simple slice as a queue, iterating by index
    for i := 0; i < len(readyQueue); i++ {
        current := readyQueue[i]

        // process dependents in a stable order too, so newly-ready services
        // are queued deterministically
        sort.Strings(dependentsOf[current])

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
        return nil, fmt.Errorf("circular dependency detected: a cycle exists in your service graph")
    }

    return readyQueue, nil
}
