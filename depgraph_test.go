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
	//assert.Error(t, g.DependOn("y", "x"))
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
	//assert.Error(t, g.DependOn("z", "x"))
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

func TestStress(t *testing.T) {
	for _ = range 1000 {
		TestTopologicalSort001(t)
		TestTopologicalSort002(t)
		TestTopologicalSort003(t)
		TestTopologicalSort004(t)
	}
}

func TestTopologicalSort002(t *testing.T) {
	g := depgraph.New()
	assert.NoError(t, g.AddLink("", "Activity_0o84rnf", "Activity_1jelc91")) // Check Billing --> Check Credit Score
	assert.NoError(t, g.AddLink("", "Activity_0z3bka9", "Activity_0i96zzv")) // Request Credit Score --> Retrieve Credit Score
	assert.NoError(t, g.AddLink("", "Activity_0l71uiq", "Activity_0bydgx6")) // Request Fraud Validation --> Internal Fraud Validation
	assert.NoError(t, g.AddLink("", "Activity_0kpp56m", "Activity_0o84rnf")) // Check Fraud --> Check Billing
	assert.NoError(t, g.AddLink("", "Activity_1jelc91", "Activity_0z3bka9")) // Check Credit Score --> Request Credit Score
	assert.NoError(t, g.AddLink("", "Activity_03xsx2d", "Activity_1xyli2s")) // Check Blacklist --> Request Blacklist Validation
	assert.NoError(t, g.AddLink("", "Event_0jm34t5", "Activity_03xsx2d"))    //  --> Check Blacklist
	assert.NoError(t, g.AddLink("", "Activity_1id12f4", "Event_1gnl54n"))    // Feedback Risk Assessment Result -->
	assert.NoError(t, g.AddLink("", "Activity_0kpp56m", "Activity_0l71uiq")) // Check Fraud --> Request Fraud Validation
	assert.NoError(t, g.AddLink("", "Activity_0o84rnf", "Activity_1dbuz2n")) // Check Billing --> Request Billing Validation
	assert.NoError(t, g.AddLink("", "Activity_1xyli2s", "Activity_1gn4p38")) // Request Blacklist Validation --> External Blacklist Validation
	assert.NoError(t, g.AddLink("", "Activity_03xsx2d", "Activity_0kpp56m")) // Check Blacklist --> Check Fraud
	assert.NoError(t, g.AddLink("", "Activity_1dbuz2n", "Activity_0kpif64")) // Request Billing Validation --> Billing Validation
	assert.NoError(t, g.AddLink("", "Activity_1jelc91", "Activity_1id12f4")) // Check Credit Score --> Feedback Risk Assessment Result
	expect := []struct {
		order string
		node  string
	}{
		{order: "1", node: "Event_0jm34t5"},      //
		{order: "2", node: "Activity_03xsx2d"},   // Check Blacklist
		{order: "2.1", node: "Activity_1xyli2s"}, // Request Blacklist Validation
		{order: "2.2", node: "Activity_1gn4p38"}, // External Blacklist Validation
		{order: "3", node: "Activity_0kpp56m"},   // Check Fraud
		{order: "3.1", node: "Activity_0l71uiq"}, // Request Fraud Validation
		{order: "3.2", node: "Activity_0bydgx6"}, // Internal Fraud Validation
		{order: "4", node: "Activity_0o84rnf"},   // Check Billing
		{order: "4.1", node: "Activity_1dbuz2n"}, // Request Billing Validation
		{order: "4.2", node: "Activity_0kpif64"}, // Billing Validation
		{order: "5", node: "Activity_1jelc91"},   // Check Credit Score
		{order: "5.1", node: "Activity_0z3bka9"}, // Request Credit Score
		{order: "5.2", node: "Activity_0i96zzv"}, // Retrieve Credit Score
		{order: "6", node: "Activity_1id12f4"},   // Feedback Risk Assessment Result
		{order: "7", node: "Event_1gnl54n"},      //
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
}

func TestTopologicalSort003(t *testing.T) {
	g := depgraph.New()
	assert.NoError(t, g.AddLink("", "Activity_02ppuc4", "Activity_1ph2hdx"))   // Device Eligibility & Subscription CHN --> Contract Eligibility & Subscription CHN
	assert.NoError(t, g.AddLink("", "Gateway_0vcjsia", "Activity_06y4ors"))    // Customer Instruction --> Business Fee Calculation & Payment AC
	assert.NoError(t, g.AddLink("", "Activity_0001xto", "Activity_1c5tycp"))   // _Add eSIM to Device AC --> Notification
	assert.NoError(t, g.AddLink("", "Activity_1y77qve", "Activity_039w2cd"))   // Plan Eligibility & Subscription CHN --> Addon Eligibility & Subscription CHN
	assert.NoError(t, g.AddLink("", "Activity_039w2cd", "Activity_02ppuc4"))   // Addon Eligibility & Subscription CHN --> Device Eligibility & Subscription CHN
	assert.NoError(t, g.AddLink("", "Activity_1g4ob8p", "Gateway_0vcjsia"))    // Resource Allocation AC --> Customer Instruction
	assert.NoError(t, g.AddLink("", "Activity_1tverkt", "Activity_106dfdj"))   // KYC Processing AC --> Customer Info Capturing & Processing CHN
	assert.NoError(t, g.AddLink("", "Activity_1g6wphe", "Activity_1oxtn1o"))   // Customer Risk Assessment CHN --> Plan Subscription Pre-Validation CHN
	assert.NoError(t, g.AddLink("", "Activity_1c5tycp", "Event_0r30wdr"))      // Notification --> End
	assert.NoError(t, g.AddLink("", "Activity_06y4ors", "Activity_1fyetyd"))   // Business Fee Calculation & Payment AC --> Fulfillment & Activation AC
	assert.NoError(t, g.AddLink("", "StartEvent_04ty3ep", "Activity_1tverkt")) // start --> KYC Processing AC
	assert.NoError(t, g.AddLink("", "Activity_106dfdj", "Activity_1g6wphe"))   // Customer Info Capturing & Processing CHN --> Customer Risk Assessment CHN
	assert.NoError(t, g.AddLink("", "Activity_1oxtn1o", "Activity_1y77qve"))   // Plan Subscription Pre-Validation CHN --> Plan Eligibility & Subscription CHN
	assert.NoError(t, g.AddLink("", "Activity_1ph2hdx", "Activity_1g4ob8p"))   // Contract Eligibility & Subscription CHN --> Resource Allocation AC
	assert.NoError(t, g.AddLink("", "Gateway_0vcjsia", "Activity_1y77qve"))    // Customer Instruction --> Plan Eligibility & Subscription CHN
	assert.NoError(t, g.AddLink("", "Activity_1fyetyd", "Activity_0001xto"))   // Fulfillment & Activation AC --> _Add eSIM to Device AC
	{
		expect := []struct {
			order string
			node  string
		}{
			{order: "1", node: "StartEvent_04ty3ep"}, // start
			{order: "2", node: "Activity_1tverkt"},   // KYC Processing AC
			{order: "3", node: "Activity_106dfdj"},   // Customer Info Capturing & Processing CHN
			{order: "4", node: "Activity_1g6wphe"},   // Customer Risk Assessment CHN
			{order: "5", node: "Activity_1oxtn1o"},   // Plan Subscription Pre-Validation CHN
			{order: "6", node: "Activity_1y77qve"},   // Plan Eligibility & Subscription CHN
			{order: "7", node: "Activity_039w2cd"},   // Addon Eligibility & Subscription CHN
			{order: "8", node: "Activity_02ppuc4"},   // Device Eligibility & Subscription CHN
			{order: "9", node: "Activity_1ph2hdx"},   // Contract Eligibility & Subscription CHN
			{order: "10", node: "Activity_1g4ob8p"},  // Resource Allocation AC
			{order: "11", node: "Gateway_0vcjsia"},   // Customer Instruction
			{order: "12", node: "Activity_06y4ors"},  // Business Fee Calculation & Payment AC
			{order: "13", node: "Activity_1fyetyd"},  // Fulfillment & Activation AC
			{order: "14", node: "Activity_0001xto"},  // _Add eSIM to Device AC
			{order: "15", node: "Activity_1c5tycp"},  // Notification
			{order: "16", node: "Event_0r30wdr"},     // End
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
}

func TestTopologicalSort004(t *testing.T) {
	g := depgraph.New()
	assert.NoError(t, g.AddLink("", "StartEvent_04ty3ep", "Activity_1tverkt")) // start --> KYC Processing AC
	assert.NoError(t, g.AddLink("", "Activity_1tverkt", "Activity_106dfdj"))   // KYC Processing AC --> Customer Info Capturing & Processing CHN
	assert.NoError(t, g.AddLink("", "Activity_106dfdj", "Activity_1g6wphe"))   // Customer Info Capturing & Processing CHN --> Customer Risk Assessment CHN
	assert.NoError(t, g.AddLink("", "Activity_1oxtn1o", "Activity_1y77qve"))   // Plan Subscription Pre-Validation CHN --> Plan Eligibility & Subscription CHN
	assert.NoError(t, g.AddLink("", "Gateway_0vcjsia", "Activity_06y4ors"))    // Customer Instruction --> Business Fee Calculation & Payment AC
	assert.NoError(t, g.AddLink("", "Activity_1y77qve", "Activity_039w2cd"))   // Plan Eligibility & Subscription CHN --> Addon Eligibility & Subscription CHN
	assert.NoError(t, g.AddLink("", "Activity_039w2cd", "Activity_02ppuc4"))   // Addon Eligibility & Subscription CHN --> Device Eligibility & Subscription CHN
	assert.NoError(t, g.AddLink("", "Activity_02ppuc4", "Activity_1ph2hdx"))   // Device Eligibility & Subscription CHN --> Contract Eligibility & Subscription CHN
	assert.NoError(t, g.AddLink("", "Gateway_0vcjsia", "Activity_1y77qve"))    // Customer Instruction --> Plan Eligibility & Subscription CHN
	assert.NoError(t, g.AddLink("", "Activity_06y4ors", "Activity_1fyetyd"))   // Business Fee Calculation & Payment AC --> Fulfillment & Activation AC
	assert.NoError(t, g.AddLink("", "Activity_1fyetyd", "Activity_0001xto"))   // Fulfillment & Activation AC --> _Add eSIM to Device AC
	assert.NoError(t, g.AddLink("", "Activity_1g6wphe", "Activity_1oxtn1o"))   // Customer Risk Assessment CHN --> Plan Subscription Pre-Validation CHN
	assert.NoError(t, g.AddLink("", "Activity_1g4ob8p", "Gateway_0vcjsia"))    // Resource Allocation AC --> Customer Instruction
	assert.NoError(t, g.AddLink("", "Activity_0001xto", "Activity_1c5tycp"))   // _Add eSIM to Device AC --> Notification
	assert.NoError(t, g.AddLink("", "Activity_1c5tycp", "Event_0r30wdr"))      // Notification --> End
	assert.NoError(t, g.AddLink("", "Activity_1ph2hdx", "Activity_1g4ob8p"))   // Contract Eligibility & Subscription CHN --> Resource Allocation AC
	{
		expect := []struct {
			order string
			node  string
		}{
			{order: "1", node: "StartEvent_04ty3ep"}, // start
			{order: "2", node: "Activity_1tverkt"},   // KYC Processing AC
			{order: "3", node: "Activity_106dfdj"},   // Customer Info Capturing & Processing CHN
			{order: "4", node: "Activity_1g6wphe"},   // Customer Risk Assessment CHN
			{order: "5", node: "Activity_1oxtn1o"},   // Plan Subscription Pre-Validation CHN
			{order: "6", node: "Activity_1y77qve"},   // Plan Eligibility & Subscription CHN
			{order: "7", node: "Activity_039w2cd"},   // Addon Eligibility & Subscription CHN
			{order: "8", node: "Activity_02ppuc4"},   // Device Eligibility & Subscription CHN
			{order: "9", node: "Activity_1ph2hdx"},   // Contract Eligibility & Subscription CHN
			{order: "10", node: "Activity_1g4ob8p"},  // Resource Allocation AC
			{order: "11", node: "Gateway_0vcjsia"},   // Customer Instruction
			{order: "12", node: "Activity_06y4ors"},  // Business Fee Calculation & Payment AC
			{order: "13", node: "Activity_1fyetyd"},  // Fulfillment & Activation AC
			{order: "14", node: "Activity_0001xto"},  // _Add eSIM to Device AC
			{order: "15", node: "Activity_1c5tycp"},  // Notification
			{order: "16", node: "Event_0r30wdr"},     // End
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
}
