package depgraph

import (
	"errors"
	"fmt"
	"sort"
)

// https://dave.cheney.net/2014/03/25/the-empty-struct
// https://github.com/kendru/darwin/blob/main/go/depgraph/depgraph.go
// TimDadd - modified to use any instead of string and new sort algorithm

type node struct {
	id       any
	x        float32
	y        float32
	addOrder int
}

// A node in this graph is just any, so a nodeMap is a map whose
// keys are the nodes that are present.  Int can be a weighting if everything else is equal
type nodeMap map[any]*node

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

func (g *Graph) AddNode(id any, x, y float32) {
	g.nodes[id] = &node{
		id:       id,
		x:        x,
		y:        y,
		addOrder: len(g.nodes),
	}
	return
}

// AddLink id here for potential future use
func (g *Graph) AddLink(id string, from, to any) error {
	return g.DependOn(to, from)
}

// DependOn sets a dependency between a child and parent
func (g *Graph) DependOn(child, parent any) error {
	if child == parent {
		return errors.New("self-referential dependencyMap not allowed")
	}

	//if g.DependsOn(parent, child) {
	//	return errors.New("circular dependencyMap not allowed")
	//}

	// Add nodes if not already added
	if n := g.nodes[parent]; n == nil {
		g.nodes[parent] = &node{
			id:       parent,
			x:        0,
			y:        0,
			addOrder: len(g.nodes),
		}
	}
	if n := g.nodes[child]; n == nil {
		g.nodes[child] = &node{
			id:       child,
			x:        0,
			y:        0,
			addOrder: len(g.nodes),
		}
	}

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
	for nodeID := range g.nodes {
		if _, ok := g.dependencyMap[nodeID]; !ok {
			leaves = append(leaves, nodeID)
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
	}

	return layers
}

func removeFromDepMap(dm dependencyMap, key, nodeId any) {
	nMap := dm[key]
	if len(nMap) == 1 {
		// The only element in the nodeMap must be `node`, so we
		// can delete the entry entirely.
		delete(dm, key)
	} else {
		// Otherwise, remove the single node from the nodeMap.
		delete(nMap, nodeId)
	}
}

func (g *Graph) remove(nodeID any) {
	// Remove edges from things that depend on `node`.
	for dependent := range g.dependentMap[nodeID] {
		removeFromDepMap(g.dependencyMap, dependent, nodeID)
	}
	delete(g.dependentMap, nodeID)

	// Remove all edges from node to the things it depends on.
	for dependency := range g.dependencyMap[nodeID] {
		removeFromDepMap(g.dependentMap, dependency, nodeID)
	}
	delete(g.dependencyMap, nodeID)

	// Finally, remove the node itself.
	delete(g.nodes, nodeID)
}

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
func (g *Graph) buildTransitive(rootNodeId any, nextFn func(any) nodeMap) nodeMap {
	if _, ok := g.nodes[rootNodeId]; !ok {
		return nil
	}
	out := make(nodeMap)
	searchNext := []any{rootNodeId}
	for len(searchNext) > 0 {
		// List of new nodes from this layer of the dependency graph. This is
		// assigned to `searchNext` at the end of the outer "discovery" loop.
		var discovered []any
		for _, nextNodeId := range searchNext {
			// For each node to discover, find the next nodes.
			for nextNode := range nextFn(nextNodeId) {
				// If we have not seen the node before, add it to the output as well
				// as the list of nodes to traverse in the next iteration.
				if _, ok := out[nextNode]; !ok {
					out[nextNode] = g.nodes[nextNode]
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

func addNodeToNodeset(dm dependencyMap, key, nodeId any) {
	n := &node{
		id: nodeId,
		x:  0,
		y:  0,
	}
	if nodes, ok := dm[key]; !ok {
		n.addOrder = 0
		nodes = nodeMap{nodeId: n} // Initialise the map
		dm[key] = nodes
	} else {
		n.addOrder = len(nodes)
		nodes[nodeId] = n
	}
}

func (g *Graph) SortedWithOrder() []*TopologyOrder {
	// Copy the graph, so we can remove things we've visited
	shrinkingGraph := g.clone()
	shrinkingGraph.handled = make(map[any]*TopologyOrder, len(g.nodes))
	shrinkingGraph.sortLeaves("", "", 0, 0, nil)
	sort.Slice(shrinkingGraph.orderedTopology, func(i, j int) bool {
		return shrinkingGraph.orderedTopology[i].SortedStep < shrinkingGraph.orderedTopology[j].SortedStep
	})
	return shrinkingGraph.orderedTopology
}

// The graph is a shrinking graph, that is, as we deal with something we remove from the graph
// Stops any issues with recursion in the graph
func (g *Graph) sortLeaves(prefix, sortedPrefix string, parent, level int, children nodeMap) {
	rootLeaf := prefix == "" && parent == 0 && level == 0
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
			// Pick dependents over co-ordinates except when the root - then try and start top left
			if dependents[leaves[i]] == dependents[leaves[j]] || rootLeaf {
				nodeI := g.nodes[leaves[i]]
				nodeJ := g.nodes[leaves[j]]
				if nodeI.x != nodeJ.x {
					return nodeI.x < nodeJ.x
				} else if nodeI.y != nodeJ.y {
					return nodeI.y < nodeJ.y
				}
				return nodeI.addOrder < nodeI.addOrder
			}
			return dependents[leaves[i]] > dependents[leaves[j]] // One with the longest path
		})
	}
	offset := parent + 1
	parentPrefix := prefix
	parentSortedPrefix := sortedPrefix
	for i, leafNode := range leaves {
		//stopAts := []string{"Event_1gnl54n", "Activity_0l71uiq"} //Activity_00qw565
		//for _, stopAt := range stopAts {
		//	if leafNode == stopAt {
		//		fmt.Println(stopAt)
		//		break
		//	}
		//}
		// By the time we're here, the leaf may have already been processed in another branch
		if _, ok := g.handled[leafNode]; ok {
			continue
		}
		// Update the prefix if we have more than one leaf
		// If i=0 then this is main route and the parent prefix is used
		if i > 0 {
			offset = 1 // Reset the offset
			if i == 1 {
				level++
			}
			if rootLeaf {
				prefix = fmt.Sprintf("%c.", 64+i) // Use a letter for the top layer - different paths!
				sortedPrefix = prefix
			} else { // Prefix format depends on number of leaves
				switch len(leaves) {
				case 2: // If we only have two leaves then we simplify the second prefix (i.e. 1-1,1-2,1-3)
					prefix = fmt.Sprintf("%s%d.", parentPrefix, parent)
					sortedPrefix = fmt.Sprintf("%s%04d.", parentSortedPrefix, parent)
				default: // More than 2 leaves then full-fat prefix (i.e. 1-1-1, 1-2-1, 1-3-1 etc.)
					prefix = fmt.Sprintf("%s%d.%d.", parentPrefix, parent, i)
					sortedPrefix = fmt.Sprintf("%s%04d.%04d.", parentSortedPrefix, parent, i)
				}
			}
		}
		to := &TopologyOrder{
			Node:       leafNode,
			SortedStep: fmt.Sprintf("%s%04d", sortedPrefix, offset),
			Step:       fmt.Sprintf("%s%d", prefix, offset),
			Level:      level,
		}
		g.orderedTopology = append(g.orderedTopology, to)
		g.handled[leafNode] = to
		c := g.dependentMap[leafNode]
		g.remove(leafNode)
		// If we're following a path then keep following until the end
		// If this is a singleton root step then don't go down a level
		if c == nil && children != nil || (rootLeaf && c == nil) {
			continue
		}
		g.sortLeaves(prefix, sortedPrefix, offset, level, c)
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
