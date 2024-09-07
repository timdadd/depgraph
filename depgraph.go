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
	id        any
	x         float32
	y         float32
	addOrder  int
	isGateway bool
}

// A node in this graph is just any, so a nodeMap is a map whose
// keys are the nodes that are present.  Int can be a weighting if everything else is equal
type nodeMap map[any]*node

// dependencyMap tracks the nodes that have some dependency relationship to
// some other node, represented by the key of the map.
type dependencyMap map[any]nodeMap

type TopologyOrder struct {
	Node       any
	FromLinkID string
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
	linkMap map[any]map[any]string // [from][to]ID

	orderedTopology []*TopologyOrder
	handled         map[any]*TopologyOrder
}

func New() *Graph {
	return &Graph{
		dependencyMap: make(dependencyMap, 20),
		dependentMap:  make(dependencyMap, 20),
		nodes:         make(nodeMap, 20),
		linkMap:       make(map[any]map[any]string, 20),
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

// AddNode adds/updates a node on the graph.  isGateway used in AllPaths algorithm
func (g *Graph) AddNode(id any, x, y float32, isGateway ...bool) (err error) {
	// If the node already exists it will be updated
	var gateway bool
	if len(isGateway) > 0 {
		gateway = isGateway[0]
	}
	if n, ok := g.nodes[id]; ok {
		if (x + y) > 0 { // If 0 passed then assume only the gateway is changing
			n.x = x
			n.y = y
		}
		n.isGateway = gateway
	} else {
		g.nodes[id] = &node{
			id:        id,
			x:         x,
			y:         y,
			addOrder:  len(g.nodes),
			isGateway: gateway,
		}
	}
	return
}

// AddLink adds a link between two nodes and records the linkID, only one linkID allowed between nodes
func (g *Graph) AddLink(linkID string, from, to any) (err error) {
	if err = g.DependOn(to, from); err != nil || linkID == "" {
		return
	}
	if linkFromMap, inFromMap := g.linkMap[from]; !inFromMap {
		g.linkMap[from] = map[any]string{to: linkID}
	} else if id, inToMap := linkFromMap[to]; !inToMap {
		linkFromMap[to] = linkID
	} else {
		if linkID != id {
			return fmt.Errorf("link %v and %v both link node %v and %v", linkID, id, from, to)
		}
	}
	return
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
		linkMap:       g.linkMap, // This can be a pointer as it doesn't get mangled
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

// Sorted returns all the nodes in the graph sorted by layers
func (g *Graph) Sorted() []any {
	nodeCount := 0
	layers := g.SortedLayers()
	for _, layer := range layers {
		nodeCount += len(layer)
	}

	allNodes := make([]any, 0, nodeCount)
	for _, layer := range layers {
		for _, n := range layer {
			allNodes = append(allNodes, n)
		}
	}

	return allNodes
}

// TopologicalSort tries to prioritise the longest branch and is good for sequence diagrams
// Any off shoots are handled before carrying on
func (g *Graph) TopologicalSort() []*TopologyOrder {
	// Copy the graph, so we can remove things we've visited
	shrinkingGraph := g.clone()
	shrinkingGraph.handled = make(map[any]*TopologyOrder, len(g.nodes))
	shrinkingGraph.sortLeaves("", "", 0, 0, nil, nil)
	sort.Slice(shrinkingGraph.orderedTopology, func(i, j int) bool {
		return shrinkingGraph.orderedTopology[i].SortedStep < shrinkingGraph.orderedTopology[j].SortedStep
	})
	return shrinkingGraph.orderedTopology
}

// sortLeaves is a shrinking graph algorithm, that is, as we deal with something we remove from the graph
// Stops any issues with recursion in the graph
func (g *Graph) sortLeaves(prefix, sortedPrefix string, parent, level int, previousNode any, children nodeMap) {
	rootLeaf := prefix == "" && parent == 0 && level == 0
	var leaves []any
	if children == nil {
		leaves = g.Leaves() // Find all nodes that don't have a dependency
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
	fromNode := previousNode
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
				fromNode = previousNode
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
			FromLinkID: "",
			Step:       fmt.Sprintf("%s%d", prefix, offset),
			SortedStep: fmt.Sprintf("%s%04d", sortedPrefix, offset),
			Level:      level,
		}
		if fromNode != nil && len(g.linkMap) > 0 {
			if toLinkMap, inMap := g.linkMap[fromNode]; inMap {
				to.FromLinkID = toLinkMap[leafNode]
			}
		}
		g.orderedTopology = append(g.orderedTopology, to)
		g.handled[leafNode] = to
		c := g.dependentMap[leafNode]
		g.remove(leafNode)
		// If we're following a path then keep following until the end
		// If this is a singleton root step then don't go down a level
		if c == nil && children != nil || (rootLeaf && c == nil) {
			fromNode = leafNode
			continue
		}
		g.sortLeaves(prefix, sortedPrefix, offset, level, leafNode, c)
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

// AllPaths uses the isGateway flag on nodes to know where multiple paths are needed
// If we have a decision Node A with 2 decisions A' A” then we should get 2 paths
// If we have two decision nodes A & B with A', A”, B', B”, B”' then we should get 6 paths
// First we find all the path options by using the dependency map
// Then we recurse through making a copy of the graph but removing anything from the dependency map that we don't want
// So that only one path is found by the topology sort
func (g *Graph) AllPaths() (pathNames [][]string, allTopologies [][]*TopologyOrder) {
	// First determine all the paths based upon the decision nodes
	var gatewayNodes = make([]any, 0, 10)        // []from
	var gatewayDependents = make([][]any, 0, 10) //[][]to
	totalPaths := 1
	gatewayCount := 0
	for f, n := range g.nodes { // From Node, Node
		if n.isGateway {
			dependents := g.immediateDependents(f)
			if len(dependents) > 1 {
				gatewayNodes = append(gatewayNodes, f)
				gatewayDependents = append(gatewayDependents, make([]any, 0, len(dependents)))
				for to := range dependents {
					gatewayDependents[gatewayCount] = append(gatewayDependents[gatewayCount], to)
				}
				totalPaths *= len(dependents)
				gatewayCount++
			}
		}
	}
	//fmt.Printf("Decision Node: %d, Paths:%d\n", gatewayCount, totalPaths)
	paths := Combinations2D(gatewayDependents)
	allTopologies = make([][]*TopologyOrder, 0, len(paths))
	pathNames = make([][]string, 0, len(paths))
	// Now loop through all the possible combinations
	// Make a copy of the graph
	// Only keep the combination of interest
	// Build the topology
	// do for next combination
	for _, path := range paths {
		pathGraph := g.clone()
		pathName := make([]string, len(path))
		for i, to := range path {
			from := gatewayNodes[i]
			toNode := g.nodes[to]
			// Just replace the dependency map with the choice we want
			// This isn't changing dependency map - might need to add later
			pathGraph.dependentMap[from] = nodeMap{to: g.nodes[to]}
			pathName[i] = fmt.Sprintf("%s", toNode.id)
		}
		// Potentially this comes up with duplicate schemas
		allTopologies = append(allTopologies, pathGraph.TopologicalSort())
		pathNames = append(pathNames, pathName)
	}
	return
}

// Combinations2D provides a list of all combinations of a 2D array
// if array is [2,3,4],[4,5,6] then output is [2,4][2,5][2,6][3,4][3,5][3,6]...
func Combinations2D(array2D [][]any) (combinations [][]any) {
	if len(array2D) == 0 {
		return
	}
	totalCombinations := 1
	for _, array := range array2D {
		totalCombinations *= len(array)
	}
	combinations = make([][]any, 0, totalCombinations)
	// Now we need to go through each combination of gateway nodes and complete the topological sort
	idx := make([]int, len(array2D)) // Index into each node starting at 0
	for {
		// Record the current combination
		var combination = make([]any, len(array2D))
		for i := 0; i < len(array2D); i++ {
			combination[i] = array2D[i][idx[i]]
		}
		combinations = append(combinations, combination)
		// Find the rightmost array that still has an item to consider
		var next int
		for next = len(array2D) - 1; next >= 0; next-- {
			if idx[next] < len(array2D[next])-1 {
				break
			}
		}
		if next < 0 {
			break // All permutations covered
		}
		idx[next]++ // Next Permutation
		// Reset everything again
		for i := next + 1; i < len(array2D); i++ {
			idx[i] = 0
		}
	}
	return
}

// DFS Depth First Search
func (g *Graph) DFS(s, f any) (paths [][]any) {
	visited := make(map[any]bool)
	var path []any
	g.dfs(s, f, visited, path, &paths)
	return paths
}

// DFS Depth First Search
func (g *Graph) dfs(s any, e any, visited map[any]bool, path []any, paths *[][]any) {
	visited[s] = true
	path = append(path, s)

	// Reached the end
	if s == e {
		*paths = append(*paths, path)
	} else {
		tos := g.linkMap[s]
		for to := range tos {
			if !visited[to] {
				g.dfs(to, e, visited, path, paths)
			}
		}
	}

	delete(visited, s)
	path = path[:len(path)-1]
}
