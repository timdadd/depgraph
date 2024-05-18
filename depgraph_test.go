package depgraph_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/timdadd/depgraph"
	"testing"
)

func TestImmediateDependencies(t *testing.T) {
	g := depgraph.New()

	assert.NoError(t, g.DependOn("x", "y"))

	assert.True(t, g.DependsOn("x", "y"))
	assert.True(t, g.HasDependent("y", "x"))
	assert.False(t, g.DependsOn("y", "x"))
	assert.False(t, g.HasDependent("x", "y"))

	// No self-dependencyMap.
	assert.Error(t, g.DependOn("z", "z"))
	// No bidirectional dependencyMap.
	assert.Error(t, g.DependOn("y", "x"))
}

func TestTransitiveDependencies(t *testing.T) {
	g := depgraph.New()

	assert.NoError(t, g.DependOn("x", "y"))
	assert.NoError(t, g.DependOn("y", "z"))

	assert.True(t, g.DependsOn("x", "z"))
	assert.True(t, g.HasDependent("z", "x"))
	assert.False(t, g.DependsOn("z", "x"))
	assert.False(t, g.HasDependent("x", ""))

	// No circular dependencyMap.
	assert.Error(t, g.DependOn("z", "x"))
}

func TestLeaves(t *testing.T) {
	g := depgraph.New()
	assert.NoError(t, g.DependOn("cake", "eggs"))
	assert.NoError(t, g.DependOn("cake", "flour"))
	assert.NoError(t, g.DependOn("eggs", "chickens"))
	assert.NoError(t, g.DependOn("flour", "grain"))
	assert.NoError(t, g.DependOn("chickens", "feed"))
	assert.NoError(t, g.DependOn("chickens", "grain"))
	assert.NoError(t, g.DependOn("grain", "soil"))

	leaves := g.Leaves()
	assert.ElementsMatch(t, leaves, []string{"feed", "soil"})
}

func TestTopologicalSort(t *testing.T) {
	g := depgraph.New()
	assert.NoError(t, g.DependOn("cake", "eggs"))
	assert.NoError(t, g.DependOn("cake", "flour"))
	assert.NoError(t, g.DependOn("eggs", "chickens"))
	assert.NoError(t, g.DependOn("flour", "grain"))
	assert.NoError(t, g.DependOn("chickens", "grain"))
	assert.NoError(t, g.DependOn("grain", "soil"))

	sorted := g.SortedWithOrder()
	//for i, l := range sorted {
	//	t.Logf("Sequence:%d,Node:%s", i, l.Node)
	//}
	//for i, l := range sorted {
	//	t.Logf("Sequence:%d,Node:%s", i, l.Node)
	//}
	pairs := []struct {
		before string
		after  string
	}{
		{
			before: "soil",
			after:  "grain",
		},
		{
			before: "grain",
			after:  "chickens",
		},
		{
			before: "grain",
			after:  "flour",
		},
		{
			before: "chickens",
			after:  "eggs",
		},
		{
			before: "flour",
			after:  "cake",
		},
		{
			before: "eggs",
			after:  "cake",
		},
	}
	comesBefore := func(before, after interface{}) bool {
		iBefore := -1
		iAfter := -1
		for i, elem := range sorted {
			if elem.Node == before {
				iBefore = i
			} else if elem.Node == after {
				iAfter = i
			}
		}
		return iBefore < iAfter
	}
	for _, pair := range pairs {
		assert.True(t, comesBefore(pair.before, pair.after))
	}
}

func TestLayeredTopologicalSort(t *testing.T) {
	g := depgraph.New()

	assert.NoError(t, g.DependOn("web", "database"))
	assert.NoError(t, g.DependOn("web", "aggregator"))
	assert.NoError(t, g.DependOn("aggregator", "database"))
	assert.NoError(t, g.DependOn("web", "logger"))
	assert.NoError(t, g.DependOn("web", "config"))
	assert.NoError(t, g.DependOn("web", "metrics"))
	assert.NoError(t, g.DependOn("database", "config"))
	assert.NoError(t, g.DependOn("metrics", "config"))
	/*
		   /--------------\
		web - aggregator - database
		   \_ logger               \
		    \___________________ config
		     \________ metrics _/
	*/

	layers := g.SortedLayers()
	//for i, l := range layers {
	//	t.Logf("Layer:%d,Nodes:%v", i, l)
	//}
	assert.Len(t, layers, 4)
	assert.ElementsMatch(t, []string{"config", "logger"}, layers[0])
	assert.ElementsMatch(t, []string{"database", "metrics"}, layers[1])
	assert.ElementsMatch(t, []string{"aggregator"}, layers[2])
	assert.ElementsMatch(t, []string{"web"}, layers[3])
}

func TestTopologicalSort001(t *testing.T) {
	g := depgraph.New()
	assert.NoError(t, g.AddLink("", "Order Submitted", "SIM Type?"))
	assert.NoError(t, g.AddLink("eSIM", "SIM Type?", "Prompt for email address"))
	assert.NoError(t, g.AddLink("", "Prompt for email address", "Enter email address"))
	assert.NoError(t, g.AddLink("", "Enter email address", "Capture email address"))
	assert.NoError(t, g.AddLink("SIM", "SIM Type?", "SIM Type Known"))
	assert.NoError(t, g.AddLink("", "Capture email address", "SIM Type Known"))
	assert.NoError(t, g.AddLink("", "SIM Type Known", "Submit & Display Order"))
	assert.NoError(t, g.AddLink("", "Submit & Display Order", "Review Order Confirmation"))
	assert.NoError(t, g.AddLink("", "Submit & Display Order", "In Parallel"))
	assert.NoError(t, g.AddLink("", "In Parallel", "Require Logistics Order?"))
	assert.NoError(t, g.AddLink("", "In Parallel", "Is eSIM?"))
	assert.NoError(t, g.AddLink("Yes", "Require Logistics Order?", "Fulfil Logistics Order"))
	assert.NoError(t, g.AddLink("No", "Require Logistics Order?", "Logistics Handled"))
	assert.NoError(t, g.AddLink("", "Fulfil Logistics Order", "Logistics Handled"))
	assert.NoError(t, g.AddLink("", "Logistics Handled", "Submit CRM Order"))
	assert.NoError(t, g.AddLink("", "Submit CRM Order", "Validate Order"))
	assert.NoError(t, g.AddLink("", "Validate Order", "Perform CRMS Validations"))
	assert.NoError(t, g.AddLink("", "Validate Order", "Order Fulfilment"))
	assert.NoError(t, g.AddLink("eSIM", "Is eSIM?", "Request Confirm & Release of eSIM"))
	assert.NoError(t, g.AddLink("pSIM", "Is eSIM?", "xSIM Handled"))
	assert.NoError(t, g.AddLink("", "Request Confirm & Release of eSIM", "Generate & Display eSIM QR Code"))
	assert.NoError(t, g.AddLink("", "Generate & Display eSIM QR Code", "Show & Download eSIM Profile"))
	assert.NoError(t, g.AddLink("", "Generate & Display eSIM QR Code", "Send eMail"))
	assert.NoError(t, g.AddLink("", "Send eMail", "xSIM Handled"))
	assert.NoError(t, g.AddLink("", "Send eMail", "Wait for Download"))
	assert.NoError(t, g.AddLink("", "Wait for Download", "Mark eSIM as Downloaded"))
	assert.NoError(t, g.AddLink("", "Mark eSIM as Downloaded", "Mark eSIM as Installed"))
	assert.NoError(t, g.AddLink("", "Mark eSIM as Installed", "eSIM Installed"))
	assert.NoError(t, g.AddLink("", "Request Confirm & Release of eSIM", "Post eSIM Confirm & Release Request"))
	assert.NoError(t, g.AddLink("", "Post eSIM Confirm & Release Request", "Mark eSIM as Confirmed & Released"))

	expect := []struct {
		order string
		node  string
	}{
		{order: "1", node: "Order Submitted"},
		{order: "2", node: "SIM Type?"},
		{order: "3", node: "Prompt for email address"},
		{order: "4", node: "Enter email address"},
		{order: "5", node: "Capture email address"},
		{order: "6", node: "SIM Type Known"},
		{order: "7", node: "Submit & Display Order"},
		{order: "7.1", node: "Review Order Confirmation"},
		{order: "8", node: "In Parallel"},
		{order: "8.1", node: "Require Logistics Order?"},
		{order: "8.2", node: "Fulfil Logistics Order"},
		{order: "8.3", node: "Logistics Handled"},
		{order: "8.4", node: "Submit CRM Order"},
		{order: "8.5", node: "Validate Order"},
		{order: "8.5.1", node: "Perform CRMS Validations"},
		{order: "8.6", node: "Order Fulfilment"},
		{order: "9", node: "Is eSIM?"},
		{order: "10", node: "Request Confirm & Release of eSIM"},
		{order: "10.1", node: "Post eSIM Confirm & Release Request"},
		{order: "10.2", node: "Mark eSIM as Confirmed & Released"},
		{order: "11", node: "Generate & Display eSIM QR Code"},
		{order: "11.1", node: "Show & Download eSIM Profile"},
		{order: "12", node: "Send eMail"},
		{order: "12.1", node: "xSIM Handled"},
		{order: "13", node: "Wait for Download"},
		{order: "14", node: "Mark eSIM as Downloaded"},
		{order: "15", node: "Mark eSIM as Installed"},
		{order: "16", node: "eSIM Installed"},
	}
	assert.Len(t, g.Nodes(), len(expect))
	actual := g.SortedWithOrder()
	assert.Len(t, actual, len(expect))
	check := make(map[any]bool, len(actual))
	if len(actual) > len(expect) {
		for _, x := range actual {
			check[x] = false
		}
		for _, x := range expect {
			check[x.node] = true
		}
		for k, v := range check {
			if !v {
				t.Logf("Where is %s", k)
			}
		}
	} else if len(expect) > len(actual) {
		for _, x := range expect {
			check[x.node] = false
		}
		for _, y := range actual {
			check[y.Node] = true
		}
		for k, v := range check {
			if !v {
				t.Logf("Where is %s", k)
			}
		}

	}

	for i, x := range expect {
		//t.Logf("Expected:%s, Got:%s (%d-%s), Equal:%t", x.node, actual[i].Node, actual[i].Level, actual[i].Step, x.node == actual[i].Node)
		assert.Equal(t, actual[i].Node, x.node)
		assert.Equal(t, actual[i].Step, x.order)
	}
	////t.Log(sorted)
	//for _ = range nodes {
	//}
	//for i, node := range nodes {
	//	assert.Equal(t, sorted[i], node.node)
	//}
}
