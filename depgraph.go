package depgraph

import (
	"errors"
	"fmt"
	"sort"
)

// https://dave.cheney.net/2014/03/25/the-empty-struct
// https://github.com/kendru/darwin/blob/main/go/depgraph/depgraph.go
// TimDadd - modified to use any instead of string and new sort algorithm

// A node in this graph is just any, so a nodeMap is a map whose
// keys are the nodes that are present.  Int can be a weighting if everything else is equal
type nodeMap map[any]int

// dependencyMap tracks the nodes that have some dependency relationship to
// some other node, represented by the key of the map.
type dependencyMap map[any]nodeMap

type TopologyOrder struct {
	Node       any
	Step       string
	SortedStep string
	Level      int
}

type Graph struct {
	nodes nodeMap
	// Maintain dependency relationships in both directions. These
	// data structures are the edges of the graph.

	// `dependencyMap` tracks child -> parents.
	dependencyMap dependencyMap
	// `dependentMap` tracks parent -> children.
	dependentMap dependencyMap
	// Keep track of the nodes of the graph themselves.

	orderedTopology []*TopologyOrder
	handled         map[any]*TopologyOrder
}

func New() *Graph {
	return &Graph{
		dependencyMap: make(dependencyMap),
		dependentMap:  make(dependencyMap),
		nodes:         make(nodeMap),
	}
}

func (g *Graph) Nodes() (nodes []any) {
	nodes = make([]any, len(g.nodes))
	i := 0
	for n := range g.nodes {
		nodes[i] = n
		i++
	}
	return nodes
}

func (g *Graph) AddNode(node any) {
	g.nodes[node] = len(g.nodes)
	return
}

// AddLink name here for potential future use
func (g *Graph) AddLink(name string, from, to any) error {
	return g.DependOn(to, from)
}

// DependOn sets a dependency between a child and parent
func (g *Graph) DependOn(child, parent any) error {
	if child == parent {
		return errors.New("self-referential dependencyMap not allowed")
	}

	if g.DependsOn(parent, child) {
		return errors.New("circular dependencyMap not allowed")
	}

	// Add nodes.
	g.nodes[parent] = len(g.nodes)
	g.nodes[child] = len(g.nodes)

	// Add edges.
	addNodeToNodeset(g.dependentMap, parent, child)
	addNodeToNodeset(g.dependencyMap, child, parent)

	return nil
}

// DependsOn returns true if child depends on parent
func (g *Graph) DependsOn(child, parent any) bool {
	deps := g.dependencies(child)
	_, ok := deps[parent]
	return ok
}

// HasDependent returns true if child is dependent on parent
func (g *Graph) HasDependent(parent, child any) bool {
	deps := g.dependents(parent)
	_, ok := deps[child]
	return ok
}

// Leaves finds all nodes that don't have a dependency
func (g *Graph) Leaves() (leaves []any) {
	for node := range g.nodes {
		if _, ok := g.dependencyMap[node]; !ok {
			leaves = append(leaves, node)
		}
	}
	return leaves
}

// SortedLayers returns a slice of graph nodes in topological sort order. That is,
// if `B` depends on `A`, then `A` is guaranteed to come before `B` in the sorted output.
// The graph is guaranteed to be cycle-free because cycles are detected while building the
// graph. Additionally, the output is grouped into "layers", which are guaranteed to not have
// any dependencyMap within each layer. This is useful, e.g. when building an execution plan for
// some DAG, in which case each element within each layer could be executed in parallel. If you
// do not need this layered property, use `Graph.TopoSorted()`, which flattens all elements.
func (g *Graph) SortedLayers() (layers [][]any) {
	// Copy the graph
	shrinkingGraph := g.clone()
	for {
		leaves := shrinkingGraph.Leaves()
		if len(leaves) == 0 {
			break
		}
		if len(leaves) > 1 {
			// Sort the leaves by number of dependentMap
			dependents := make(map[any]int, len(leaves))
			for _, leafNode := range leaves {
				dependents[leafNode] = len(g.dependents(leafNode))
			}
			sort.Slice(leaves, func(i, j int) bool {
				return dependents[leaves[i]] < dependents[leaves[j]]
			})
		}

		layers = append(layers, leaves)
		for _, leafNode := range leaves {
			shrinkingGraph.remove(leafNode)
		}
		if leaves[0] == "Submit & Display Order" {
			fmt.Println(leaves)
		}
	}

	return layers
}

// // SortedMap returns a map[node]sort starting at 1
// // If they are on the same layer then they get the sort number
// // See also `Graph.SortedLayers()`.
//
//	func (g *Graph) SortedMap() (sortedNodeMap map[any]int) {
//		sortedNodeMap = make(map[any]int, len(g.nodes))
//		level := 0
//		// Copy the graph
//		shrinkingGraph := g.clone()
//		for {
//			leaves := shrinkingGraph.Leaves()
//			if len(leaves) == 0 {
//				break
//			}
//			level++
//			for _, leafNode := range leaves {
//				sortedNodeMap[leafNode] = level
//				shrinkingGraph.remove(leafNode)
//			}
//		}
//		return
//	}
func removeFromDepMap(dm dependencyMap, key, node any) {
	nodes := dm[key]
	if len(nodes) == 1 {
		// The only element in the nodeMap must be `node`, so we
		// can delete the entry entirely.
		delete(dm, key)
	} else {
		// Otherwise, remove the single node from the nodeMap.
		delete(nodes, node)
	}
}

func (g *Graph) remove(node any) {
	// Remove edges from things that depend on `node`.
	for dependent := range g.dependentMap[node] {
		removeFromDepMap(g.dependencyMap, dependent, node)
	}
	delete(g.dependentMap, node)

	// Remove all edges from node to the things it depends on.
	for dependency := range g.dependencyMap[node] {
		removeFromDepMap(g.dependentMap, dependency, node)
	}
	delete(g.dependencyMap, node)

	// Finally, remove the node itself.
	delete(g.nodes, node)
}

//// Sorted returns all the nodes in the graph is topological sort order.
//// See also `Graph.SortedLayers()`.
//func (g *Graph) Sorted() (allNodes []any) {
//	nodeCount := 0
//	layers := g.SortedLayers()
//	for _, layer := range layers {
//		nodeCount += len(layer)
//	}
//
//	allNodes = make([]any, 0, nodeCount)
//	for _, layer := range layers {
//		for _, node := range layer {
//			allNodes = append(allNodes, node)
//		}
//	}
//
//	return allNodes
//}
//
//// SortedNodes returns all the nodes in the graph is topological sort order.
//// See also `Graph.SortedLayers()`.
//func (g *Graph) SortedNodes() (nodes []any) {
//	nodeCount := 0
//	layers := g.SortedLayers()
//	for _, layer := range layers {
//		nodeCount += len(layer)
//	}
//
//	nodes = make([]any, 0, nodeCount)
//	for _, layer := range layers {
//		for _, node := range layer {
//			nodes = append(nodes, node)
//		}
//	}
//
//	return nodes
//}

func (g *Graph) dependencies(child any) nodeMap {
	return g.buildTransitive(child, g.immediateDependencies)
}

func (g *Graph) immediateDependencies(node any) nodeMap {
	return g.dependencyMap[node]
}

func (g *Graph) dependents(parent any) nodeMap {
	return g.buildTransitive(parent, g.immediateDependents)
}

func (g *Graph) immediateDependents(node any) nodeMap {
	return g.dependentMap[node]
}

func (g *Graph) clone() *Graph {
	return &Graph{
		dependencyMap: copyDepMap(g.dependencyMap),
		dependentMap:  copyDepMap(g.dependentMap),
		nodes:         copyNodeset(g.nodes),
	}
}

// buildTransitive starts at `root` and continues calling `nextFn` to keep discovering more nodes until
// the graph is exhausted. It returns the set of all discovered nodes.
func (g *Graph) buildTransitive(root any, nextFn func(any) nodeMap) nodeMap {
	if _, ok := g.nodes[root]; !ok {
		return nil
	}

	out := make(nodeMap)
	searchNext := []any{root}
	for len(searchNext) > 0 {
		// List of new nodes from this layer of the dependency graph. This is
		// assigned to `searchNext` at the end of the outer "discovery" loop.
		discovered := []any{}
		for _, node := range searchNext {
			// For each node to discover, find the next nodes.
			for nextNode := range nextFn(node) {
				// If we have not seen the node before, add it to the output as well
				// as the list of nodes to traverse in the next iteration.
				if _, ok := out[nextNode]; !ok {
					out[nextNode] = len(out)
					discovered = append(discovered, nextNode)
				}
			}
		}
		searchNext = discovered
	}

	return out
}

func copyNodeset(s nodeMap) nodeMap {
	out := make(nodeMap, len(s))
	for k, v := range s {
		out[k] = v
	}
	return out
}

func copyDepMap(m dependencyMap) dependencyMap {
	out := make(dependencyMap, len(m))
	for k, v := range m {
		out[k] = copyNodeset(v)
	}
	return out
}

func addNodeToNodeset(dm dependencyMap, key, node any) {
	if nodes, ok := dm[key]; !ok {
		nodes = nodeMap{node: 0} // Initialise the map
		dm[key] = nodes
	} else {
		nodes[node] = len(nodes)
	}
}

//func (g *Graph) topologicalSortUtil(v any, visited map[any]bool, stack *[]any) {
//	visited[v] = true
//
//	for _, u := range g.dependentMap[v] {
//		if !visited[u] {
//			g.topologicalSortUtil(u, visited, stack)
//		}
//	}
//
//	*stack = append([]any{v}, *stack...)
//}
//
//func (g *Graph) TopologicalSort() []any {
//	var stack []any
//	visited := make(map[any]bool)
//
//	for v := range g.nodes {
//		if !visited[v] {
//			g.topologicalSortUtil(v, visited, &stack)
//		}
//	}
//
//	return stack
//}

func (g *Graph) SortedWithOrder() []*TopologyOrder {
	// Copy the graph, so we can remove things we've visited
	shrinkingGraph := g.clone()
	shrinkingGraph.handled = make(map[any]*TopologyOrder, len(g.nodes))
	shrinkingGraph.sortLeaves("", "", 1, 0, nil)
	sort.Slice(shrinkingGraph.orderedTopology, func(i, j int) bool {
		return shrinkingGraph.orderedTopology[i].SortedStep < shrinkingGraph.orderedTopology[j].SortedStep
	})
	return shrinkingGraph.orderedTopology
}

// The graph is a shrinking graph
func (g *Graph) sortLeaves(prefix, sortedPrefix string, step, level int, children nodeMap) {
	var leaves []any
	if children == nil {
		leaves = g.Leaves()
	} else {
		for child := range children {
			if _, handled := g.handled[child]; !handled {
				leaves = append(leaves, child)
			}
		}
	}
	if len(leaves) == 0 {
		return
	}
	if len(leaves) > 1 {
		// Sort the leaves by number of dependentMap, most dependentMap first
		dependents := make(map[any]int, len(leaves))
		for _, leafNode := range leaves {
			dependents[leafNode] = len(g.dependents(leafNode))
		}
		sort.Slice(leaves, func(i, j int) bool {
			if dependents[leaves[i]] == dependents[leaves[j]] {
				return g.nodes[leaves[i]] > g.nodes[leaves[j]]
			}
			return dependents[leaves[i]] > dependents[leaves[j]]
		})
	}
	for i, leafNode := range leaves {
		// By the time we're here, the leaf may have already been processed in another branch
		if _, ok := g.handled[leafNode]; ok {
			continue
		}
		to := &TopologyOrder{
			Node: leafNode,
		}
		// Update the prefix if we have more than one leaf
		if i == 1 {
			prefix = fmt.Sprintf("%s%d.", prefix, step-1)
			sortedPrefix = fmt.Sprintf("%s%04d.", sortedPrefix, step-1)
			step = 0
			level++
		}
		to.SortedStep = fmt.Sprintf("%s%04d", sortedPrefix, step+i)
		to.Step = fmt.Sprintf("%s%d", prefix, step+i)
		to.Level = level
		g.orderedTopology = append(g.orderedTopology, to)
		g.handled[leafNode] = to
		c := g.dependentMap[leafNode]
		g.remove(leafNode)
		//if to.Node == "Validate Order" {
		//	fmt.Println("Next Step:", to.SortedStep)
		//}
		// If we're following a path then keep following until the end
		if c == nil && children != nil {
			continue
		}
		g.sortLeaves(prefix, sortedPrefix, step+i+1, level, c)
	}

}

// unhandledLeaves finds all nodes that don't have a dependency
func (g *Graph) unhandledLeaves() (leaves []any) {
	for node := range g.nodes {
		if _, ok := g.dependencyMap[node]; !ok {
			leaves = append(leaves, node)
		}
	}
	return leaves
}
