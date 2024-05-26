package depgraph_test

import (
	"depgraph"
	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, g.AddLink("", "Perform CRMS Validations", "Order Fulfilment"))
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
		{order: "8.6", node: "Perform CRMS Validations"},
		{order: "8.7", node: "Order Fulfilment"},
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
		assert.Equal(t, x.node, actual[i].Node)
		assert.Equal(t, x.order, actual[i].Step)
	}
	////t.Log(sorted)
	//for _ = range nodes {
	//}
	//for i, node := range nodes {
	//	assert.Equal(t, sorted[i], node.node)
	//}
}

func TestStress(t *testing.T) {
	for range 1000 {
		TestTopologicalSort001(t)
		TestTopologicalSort002(t)
		TestTopologicalSort003(t)
		TestTopologicalSort004(t)
		TestTopologicalSort005(t)
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

func TestTopologicalSort005(t *testing.T) {
	g := depgraph.New()
	g.AddNode("Activity_0fs7ehp", 4780.000000, 1514.000000)
	g.AddNode("Activity_14d0wi6", 4780.000000, 2246.000000)
	assert.NoError(t, g.AddLink("", "Activity_0fs7ehp", "Activity_14d0wi6")) // Post Port-in cancelation --> Process request to cancel Port-in
	g.AddNode("Gateway_1icfqwu", 4075.000000, 1381.000000)
	g.AddNode("Gateway_0r20x7i", 4445.000000, 1381.000000)
	assert.NoError(t, g.AddLink("", "Gateway_1icfqwu", "Gateway_0r20x7i")) // Order requires Logistics SIM Card delivery? --> endif
	g.AddNode("Id_4a187b58-f35c-4cdd-8ac2-2f90a9425d3f", 3938.000000, 1366.000000)
	g.AddNode("Gateway_1icfqwu", 4075.000000, 1381.000000)
	assert.NoError(t, g.AddLink("", "Id_4a187b58-f35c-4cdd-8ac2-2f90a9425d3f", "Gateway_1icfqwu")) // Decompose, Orchestrate Order --> Order requires Logistics SIM Card delivery?
	g.AddNode("Gateway_1uv7159", 6475.000000, 961.000000)
	g.AddNode("Gateway_19ib8qw", 6575.000000, 961.000000)
	assert.NoError(t, g.AddLink("", "Gateway_1uv7159", "Gateway_19ib8qw")) // endif --> Order includes Device?
	g.AddNode("Activity_1qj1e0v", 4780.000000, 946.000000)
	g.AddNode("Activity_1cofeb7", 4780.000000, 1127.000000)
	assert.NoError(t, g.AddLink("", "Activity_1qj1e0v", "Activity_1cofeb7")) // Post Port-in cancelation request --> Cancel Port-in Order
	g.AddNode("Event_156e4wi", 3692.000000, 968.000000)
	g.AddNode("Id_9803dd09-4618-4d4f-9a36-7fb2247d1e74", 3785.000000, 946.000000)
	assert.NoError(t, g.AddLink("", "Event_156e4wi", "Id_9803dd09-4618-4d4f-9a36-7fb2247d1e74")) //  --> Post Order Submission
	g.AddNode("Activity_1176p0t", 5930.000000, 1366.000000)
	g.AddNode("Activity_1hcsk28", 5930.000000, 1514.000000)
	assert.NoError(t, g.AddLink("", "Activity_1176p0t", "Activity_1hcsk28")) // Sync Port-in RFS --> Post Port-in RFS
	g.AddNode("Activity_0wyrzg9", 4320.000000, 1640.000000)
	g.AddNode("Activity_15shrim", 4320.000000, 1514.000000)
	assert.NoError(t, g.AddLink("", "Activity_0wyrzg9", "Activity_15shrim")) // Notify delivery Completion - pSIM delivered --> Post delivery completion - pSIM delivered
	g.AddNode("Activity_07y35my", 6080.000000, 946.000000)
	g.AddNode("Gateway_1iqkz3r", 6235.000000, 961.000000)
	assert.NoError(t, g.AddLink("", "Activity_07y35my", "Gateway_1iqkz3r")) // Receive Callback for Inventory update and device delivery --> SIM type?
	g.AddNode("Activity_0gc946n", 6330.000000, 946.000000)
	g.AddNode("Id_7c12b81c-70df-419e-bdea-bbbe4e2cc408", 6330.000000, 1996.000000)
	assert.NoError(t, g.AddLink("", "Activity_0gc946n", "Id_7c12b81c-70df-419e-bdea-bbbe4e2cc408")) // Post SIM card in use --> Mark SIM card as used
	g.AddNode("Id_af69b80c-765a-4ae7-be40-22cb244b611b", 6080.000000, 1366.000000)
	g.AddNode("Event_0imzxiq", 6942.000000, 1388.000000)
	assert.NoError(t, g.AddLink("", "Id_af69b80c-765a-4ae7-be40-22cb244b611b", "Event_0imzxiq")) // Order Completion Processing -->
	g.AddNode("Gateway_19ib8qw", 6575.000000, 961.000000)
	g.AddNode("Gateway_0y26mfn", 6815.000000, 961.000000)
	assert.NoError(t, g.AddLink("", "Gateway_19ib8qw", "Gateway_0y26mfn")) // Order includes Device? --> endif
	g.AddNode("Gateway_1iqkz3r", 6235.000000, 961.000000)
	g.AddNode("Gateway_1uv7159", 6475.000000, 961.000000)
	assert.NoError(t, g.AddLink("", "Gateway_1iqkz3r", "Gateway_1uv7159")) // SIM type? --> endif
	g.AddNode("Gateway_15nz4h8", 5351.000000, 1278.000000)
	g.AddNode("Gateway_0m3neg2", 5351.000000, 1142.000000)
	assert.NoError(t, g.AddLink("", "Gateway_15nz4h8", "Gateway_0m3neg2")) // Next action? --> Port-in canceled
	g.AddNode("Activity_1176p0t", 5930.000000, 1366.000000)
	g.AddNode("Id_af69b80c-765a-4ae7-be40-22cb244b611b", 6080.000000, 1366.000000)
	assert.NoError(t, g.AddLink("", "Activity_1176p0t", "Id_af69b80c-765a-4ae7-be40-22cb244b611b")) // Sync Port-in RFS --> Order Completion Processing
	g.AddNode("Activity_0er1yq6", 5170.000000, 1366.000000)
	g.AddNode("Gateway_0u6v33s", 5351.000000, 1381.000000)
	assert.NoError(t, g.AddLink("", "Activity_0er1yq6", "Gateway_0u6v33s")) // Process feedback to Port-in request --> Port-in response?
	g.AddNode("Id_33b7ee99-afd9-40b7-bee3-a0637ee9f340", 5637.000000, 1514.000000)
	g.AddNode("Id_05abb9f9-43e9-4a25-be75-0ca2dfd38502", 5637.000000, 1766.000000)
	assert.NoError(t, g.AddLink("", "Id_33b7ee99-afd9-40b7-bee3-a0637ee9f340", "Id_05abb9f9-43e9-4a25-be75-0ca2dfd38502")) // Synchronize with Billing --> Billing Synchronization
	g.AddNode("Activity_1r47mqk", 4170.000000, 1366.000000)
	g.AddNode("Activity_0wsuoyg", 4170.000000, 1514.000000)
	assert.NoError(t, g.AddLink("", "Activity_1r47mqk", "Activity_0wsuoyg")) // Request for Logistics SIM Card Delivery --> Post Logistics Order request to deliver SIM Card
	g.AddNode("Activity_0fu3e4x", 4520.000000, 1366.000000)
	g.AddNode("Activity_0z6vbvm", 4520.000000, 2246.000000)
	assert.NoError(t, g.AddLink("", "Activity_0fu3e4x", "Activity_0z6vbvm")) // Send Port-in request to Regulator --> Process request to Port-in
	g.AddNode("Activity_1vtvz75", 6680.000000, 946.000000)
	g.AddNode("Activity_1swfmss", 6680.000000, 1640.000000)
	assert.NoError(t, g.AddLink("", "Activity_1vtvz75", "Activity_1swfmss")) // Post request to deliver Device --> Fulfill Logistics Order to deliver Device
	g.AddNode("Activity_1vtvz75", 6680.000000, 946.000000)
	g.AddNode("Gateway_0y26mfn", 6815.000000, 961.000000)
	assert.NoError(t, g.AddLink("", "Activity_1vtvz75", "Gateway_0y26mfn")) // Post request to deliver Device --> endif
	g.AddNode("Activity_1qrayp8", 5080.000000, 946.000000)
	g.AddNode("Activity_1sgwv9e", 5220.000000, 946.000000)
	assert.NoError(t, g.AddLink("", "Activity_1qrayp8", "Activity_1sgwv9e")) // Validate Port-in Order (resubmission) --> Post Port-in resubmission request
	g.AddNode("Id_9803dd09-4618-4d4f-9a36-7fb2247d1e74", 3785.000000, 946.000000)
	g.AddNode("Id_9bcb3e93-3bf8-49eb-a1b1-e9cc8d9c8e1d", 3785.000000, 1366.000000)
	assert.NoError(t, g.AddLink("", "Id_9803dd09-4618-4d4f-9a36-7fb2247d1e74", "Id_9bcb3e93-3bf8-49eb-a1b1-e9cc8d9c8e1d")) // Post Order Submission --> Validate,Create,Enrich Order, Record Cust data
	g.AddNode("Id_355a3ed4-2455-4ae7-b5e0-b486af3a9105", 5487.000000, 1366.000000)
	g.AddNode("Id_8a402122-eeb2-46b1-a5f2-9f695d2e8a88", 5487.000000, 1876.000000)
	assert.NoError(t, g.AddLink("", "Id_355a3ed4-2455-4ae7-b5e0-b486af3a9105", "Id_8a402122-eeb2-46b1-a5f2-9f695d2e8a88")) // OSS Service Fulfillment - Activation --> Provision Network
	g.AddNode("Activity_0zvoww1", 4170.000000, 1640.000000)
	g.AddNode("Activity_0wyrzg9", 4320.000000, 1640.000000)
	assert.NoError(t, g.AddLink("", "Activity_0zvoww1", "Activity_0wyrzg9")) // Fulfill Logistics Order to deliver SIM Card --> Notify delivery Completion - pSIM delivered
	g.AddNode("Id_9bcb3e93-3bf8-49eb-a1b1-e9cc8d9c8e1d", 3785.000000, 1366.000000)
	g.AddNode("Id_4a187b58-f35c-4cdd-8ac2-2f90a9425d3f", 3938.000000, 1366.000000)
	assert.NoError(t, g.AddLink("", "Id_9bcb3e93-3bf8-49eb-a1b1-e9cc8d9c8e1d", "Id_4a187b58-f35c-4cdd-8ac2-2f90a9425d3f")) // Validate,Create,Enrich Order, Record Cust data --> Decompose, Orchestrate Order
	g.AddNode("Activity_0gc946n", 6330.000000, 946.000000)
	g.AddNode("Gateway_1uv7159", 6475.000000, 961.000000)
	assert.NoError(t, g.AddLink("", "Activity_0gc946n", "Gateway_1uv7159")) // Post SIM card in use --> endif
	g.AddNode("Activity_1cofeb7", 4780.000000, 1127.000000)
	g.AddNode("Activity_0fs7ehp", 4780.000000, 1514.000000)
	assert.NoError(t, g.AddLink("", "Activity_1cofeb7", "Activity_0fs7ehp")) // Cancel Port-in Order --> Post Port-in cancelation
	g.AddNode("Activity_1cofeb7", 4780.000000, 1127.000000)
	g.AddNode("Gateway_0m3neg2", 5351.000000, 1142.000000)
	assert.NoError(t, g.AddLink("", "Activity_1cofeb7", "Gateway_0m3neg2")) // Cancel Port-in Order --> Port-in canceled
	g.AddNode("Activity_1qrayp8", 5080.000000, 946.000000)
	g.AddNode("Activity_0ku6ql7", 5080.000000, 1189.000000)
	assert.NoError(t, g.AddLink("", "Activity_1qrayp8", "Activity_0ku6ql7")) // Validate Port-in Order (resubmission) --> Perform CRMS validations (resubmission)
	g.AddNode("Activity_14y8a5o", 5080.000000, 626.000000)
	g.AddNode("Activity_1qydfap", 5080.000000, 766.000000)
	assert.NoError(t, g.AddLink("", "Activity_14y8a5o", "Activity_1qydfap")) // Resubmit Port-in request --> Capture Port-in resubmission request
	g.AddNode("Id_3f5cd1e3-761d-4613-90ae-792af851635d", 5790.000000, 1366.000000)
	g.AddNode("Activity_1176p0t", 5930.000000, 1366.000000)
	assert.NoError(t, g.AddLink("", "Id_3f5cd1e3-761d-4613-90ae-792af851635d", "Activity_1176p0t")) // Update Inventory with Port-in MSISDN --> Sync Port-in RFS
	g.AddNode("Activity_0wsuoyg", 4170.000000, 1514.000000)
	g.AddNode("Activity_0zvoww1", 4170.000000, 1640.000000)
	assert.NoError(t, g.AddLink("", "Activity_0wsuoyg", "Activity_0zvoww1")) // Post Logistics Order request to deliver SIM Card --> Fulfill Logistics Order to deliver SIM Card
	g.AddNode("Gateway_0r20x7i", 4445.000000, 1381.000000)
	g.AddNode("Activity_0fu3e4x", 4520.000000, 1366.000000)
	assert.NoError(t, g.AddLink("", "Gateway_0r20x7i", "Activity_0fu3e4x")) // endif --> Send Port-in request to Regulator
	g.AddNode("Activity_0at6g8t", 5220.000000, 1263.000000)
	g.AddNode("Activity_0fu3e4x", 4520.000000, 1366.000000)
	assert.NoError(t, g.AddLink("", "Activity_0at6g8t", "Activity_0fu3e4x")) // Order ResubmittedX --> Send Port-in request to Regulator
	g.AddNode("Id_40e7d082-7ecb-4720-989c-ca65c390f2a6", 5637.000000, 1366.000000)
	g.AddNode("Id_3f5cd1e3-761d-4613-90ae-792af851635d", 5790.000000, 1366.000000)
	assert.NoError(t, g.AddLink("", "Id_40e7d082-7ecb-4720-989c-ca65c390f2a6", "Id_3f5cd1e3-761d-4613-90ae-792af851635d")) // BSS Service Fulfillment & Activation --> Update Inventory with Port-in MSISDN
	g.AddNode("Activity_0fu3e4x", 4520.000000, 1366.000000)
	g.AddNode("Event_1ryfqic", 5082.000000, 1388.000000)
	assert.NoError(t, g.AddLink("", "Activity_0fu3e4x", "Event_1ryfqic")) // Send Port-in request to Regulator --> Waiting for Port-in response
	g.AddNode("Activity_0rbroim", 4780.000000, 766.000000)
	g.AddNode("Activity_1qj1e0v", 4780.000000, 946.000000)
	assert.NoError(t, g.AddLink("", "Activity_0rbroim", "Activity_1qj1e0v")) // Capture Port-in cancelation request --> Post Port-in cancelation request
	g.AddNode("Id_af69b80c-765a-4ae7-be40-22cb244b611b", 6080.000000, 1366.000000)
	g.AddNode("Id_0efd2669-b950-4d2d-aa2e-edc02424d928", 6080.000000, 2126.000000)
	assert.NoError(t, g.AddLink("", "Id_af69b80c-765a-4ae7-be40-22cb244b611b", "Id_0efd2669-b950-4d2d-aa2e-edc02424d928")) // Order Completion Processing --> Publish Kafka event
	g.AddNode("Id_40e7d082-7ecb-4720-989c-ca65c390f2a6", 5637.000000, 1366.000000)
	g.AddNode("Id_33b7ee99-afd9-40b7-bee3-a0637ee9f340", 5637.000000, 1514.000000)
	assert.NoError(t, g.AddLink("", "Id_40e7d082-7ecb-4720-989c-ca65c390f2a6", "Id_33b7ee99-afd9-40b7-bee3-a0637ee9f340")) // BSS Service Fulfillment & Activation --> Synchronize with Billing
	g.AddNode("Activity_00qw565", 5050.000000, 2246.000000)
	g.AddNode("Event_1ryfqic", 5082.000000, 1388.000000)
	assert.NoError(t, g.AddLink("", "Activity_00qw565", "Event_1ryfqic")) // Reply Port-in response --> Waiting for Port-in response
	g.AddNode("Activity_15shrim", 4320.000000, 1514.000000)
	g.AddNode("Activity_0r87n5x", 4320.000000, 1366.000000)
	assert.NoError(t, g.AddLink("", "Activity_15shrim", "Activity_0r87n5x")) // Post delivery completion - pSIM delivered --> Receives Logistics Call Back - pSIM delivered
	g.AddNode("Activity_1qydfap", 5080.000000, 766.000000)
	g.AddNode("Activity_1qrayp8", 5080.000000, 946.000000)
	assert.NoError(t, g.AddLink("", "Activity_1qydfap", "Activity_1qrayp8")) // Capture Port-in resubmission request --> Validate Port-in Order (resubmission)
	g.AddNode("Gateway_15nz4h8", 5351.000000, 1278.000000)
	g.AddNode("Activity_0at6g8t", 5220.000000, 1263.000000)
	assert.NoError(t, g.AddLink("", "Gateway_15nz4h8", "Activity_0at6g8t")) // Next action? --> Order ResubmittedX
	g.AddNode("Event_1ryfqic", 5082.000000, 1388.000000)
	g.AddNode("Activity_0er1yq6", 5170.000000, 1366.000000)
	assert.NoError(t, g.AddLink("", "Event_1ryfqic", "Activity_0er1yq6")) // Waiting for Port-in response --> Process feedback to Port-in request
	g.AddNode("Gateway_0u6v33s", 5351.000000, 1381.000000)
	g.AddNode("Gateway_15nz4h8", 5351.000000, 1278.000000)
	assert.NoError(t, g.AddLink("", "Gateway_0u6v33s", "Gateway_15nz4h8")) // Port-in response? --> Next action?
	g.AddNode("Activity_0ctpaup", 4780.000000, 626.000000)
	g.AddNode("Activity_0rbroim", 4780.000000, 766.000000)
	assert.NoError(t, g.AddLink("", "Activity_0ctpaup", "Activity_0rbroim")) // Submit Port-in cancelation request --> Capture Port-in cancelation request
	g.AddNode("Id_355a3ed4-2455-4ae7-b5e0-b486af3a9105", 5487.000000, 1366.000000)
	g.AddNode("Id_40e7d082-7ecb-4720-989c-ca65c390f2a6", 5637.000000, 1366.000000)
	assert.NoError(t, g.AddLink("", "Id_355a3ed4-2455-4ae7-b5e0-b486af3a9105", "Id_40e7d082-7ecb-4720-989c-ca65c390f2a6")) // OSS Service Fulfillment - Activation --> BSS Service Fulfillment & Activation
	g.AddNode("Activity_1hcsk28", 5930.000000, 1514.000000)
	g.AddNode("Activity_0n17894", 5930.000000, 2246.000000)
	assert.NoError(t, g.AddLink("", "Activity_1hcsk28", "Activity_0n17894")) // Post Port-in RFS --> Sync Port-in RFS with other Operators
	g.AddNode("Gateway_1icfqwu", 4075.000000, 1381.000000)
	g.AddNode("Activity_1r47mqk", 4170.000000, 1366.000000)
	assert.NoError(t, g.AddLink("", "Gateway_1icfqwu", "Activity_1r47mqk")) // Order requires Logistics SIM Card delivery? --> Request for Logistics SIM Card Delivery
	g.AddNode("Id_af69b80c-765a-4ae7-be40-22cb244b611b", 6080.000000, 1366.000000)
	g.AddNode("Activity_07y35my", 6080.000000, 946.000000)
	assert.NoError(t, g.AddLink("", "Id_af69b80c-765a-4ae7-be40-22cb244b611b", "Activity_07y35my")) // Order Completion Processing --> Receive Callback for Inventory update and device delivery
	g.AddNode("Activity_0r87n5x", 4320.000000, 1366.000000)
	g.AddNode("Gateway_0r20x7i", 4445.000000, 1381.000000)
	assert.NoError(t, g.AddLink("", "Activity_0r87n5x", "Gateway_0r20x7i")) // Receives Logistics Call Back - pSIM delivered --> endif
	g.AddNode("Activity_1sgwv9e", 5220.000000, 946.000000)
	g.AddNode("Activity_0at6g8t", 5220.000000, 1263.000000)
	assert.NoError(t, g.AddLink("", "Activity_1sgwv9e", "Activity_0at6g8t")) // Post Port-in resubmission request --> Order ResubmittedX
	g.AddNode("Gateway_0u6v33s", 5351.000000, 1381.000000)
	g.AddNode("Id_355a3ed4-2455-4ae7-b5e0-b486af3a9105", 5487.000000, 1366.000000)
	assert.NoError(t, g.AddLink("", "Gateway_0u6v33s", "Id_355a3ed4-2455-4ae7-b5e0-b486af3a9105")) // Port-in response? --> OSS Service Fulfillment - Activation
	g.AddNode("Gateway_0y26mfn", 6815.000000, 961.000000)
	g.AddNode("Event_1b30bns", 6932.000000, 968.000000)
	assert.NoError(t, g.AddLink("", "Gateway_0y26mfn", "Event_1b30bns")) // endif -->
	g.AddNode("Gateway_19ib8qw", 6575.000000, 961.000000)
	g.AddNode("Activity_1vtvz75", 6680.000000, 946.000000)
	assert.NoError(t, g.AddLink("", "Gateway_19ib8qw", "Activity_1vtvz75")) // Order includes Device? --> Post request to deliver Device
	g.AddNode("Gateway_1iqkz3r", 6235.000000, 961.000000)
	g.AddNode("Activity_0gc946n", 6330.000000, 946.000000)
	assert.NoError(t, g.AddLink("", "Gateway_1iqkz3r", "Activity_0gc946n")) // SIM type? --> Post SIM card in use
	g.AddNode("Gateway_0m3neg2", 5351.000000, 1142.000000)
	g.AddNode("Event_0me7i5f", 6942.000000, 1149.000000)
	assert.NoError(t, g.AddLink("", "Gateway_0m3neg2", "Event_0me7i5f")) // Port-in canceled -->
	{
		expect := []struct {
			order string
			node  string
		}{
			{order: "0001", node: "Event_156e4wi"},                                     //
			{order: "0002", node: "Id_9803dd09-4618-4d4f-9a36-7fb2247d1e74"},           // Post Order Submission
			{order: "0003", node: "Id_9bcb3e93-3bf8-49eb-a1b1-e9cc8d9c8e1d"},           // Validate,Create,Enrich Order, Record Cust data
			{order: "0004", node: "Id_4a187b58-f35c-4cdd-8ac2-2f90a9425d3f"},           // Decompose, Orchestrate Order
			{order: "0005", node: "Gateway_1icfqwu"},                                   // Order requires Logistics SIM Card delivery?
			{order: "0006", node: "Activity_1r47mqk"},                                  // Request for Logistics SIM Card Delivery
			{order: "0007", node: "Activity_0wsuoyg"},                                  // Post Logistics Order request to deliver SIM Card
			{order: "0008", node: "Activity_0zvoww1"},                                  // Fulfill Logistics Order to deliver SIM Card
			{order: "0009", node: "Activity_0wyrzg9"},                                  // Notify delivery Completion - pSIM delivered
			{order: "0010", node: "Activity_15shrim"},                                  // Post delivery completion - pSIM delivered
			{order: "0011", node: "Activity_0r87n5x"},                                  // Receives Logistics Call Back - pSIM delivered
			{order: "0012", node: "Gateway_0r20x7i"},                                   // endif
			{order: "0013", node: "Activity_0fu3e4x"},                                  // Send Port-in request to Regulator
			{order: "0013.0001", node: "Activity_0z6vbvm"},                             // Process request to Port-in
			{order: "0014", node: "Event_1ryfqic"},                                     // Waiting for Port-in response
			{order: "0015", node: "Activity_0er1yq6"},                                  // Process feedback to Port-in request
			{order: "0016", node: "Gateway_0u6v33s"},                                   // Port-in response?
			{order: "0016.0001", node: "Gateway_15nz4h8"},                              // Next action?
			{order: "0016.0001.0001", node: "Activity_0at6g8t"},                        // Order ResubmittedX
			{order: "0016.0002", node: "Gateway_0m3neg2"},                              // Port-in canceled
			{order: "0016.0003", node: "Event_0me7i5f"},                                //
			{order: "0017", node: "Id_355a3ed4-2455-4ae7-b5e0-b486af3a9105"},           // OSS Service Fulfillment - Activation
			{order: "0017.0001", node: "Id_8a402122-eeb2-46b1-a5f2-9f695d2e8a88"},      // Provision Network
			{order: "0018", node: "Id_40e7d082-7ecb-4720-989c-ca65c390f2a6"},           // BSS Service Fulfillment & Activation
			{order: "0018.0001", node: "Id_33b7ee99-afd9-40b7-bee3-a0637ee9f340"},      // Synchronize with Billing
			{order: "0018.0002", node: "Id_05abb9f9-43e9-4a25-be75-0ca2dfd38502"},      // Billing Synchronization
			{order: "0019", node: "Id_3f5cd1e3-761d-4613-90ae-792af851635d"},           // Update Inventory with Port-in MSISDN
			{order: "0020", node: "Activity_1176p0t"},                                  // Sync Port-in RFS
			{order: "0020.0001", node: "Activity_1hcsk28"},                             // Post Port-in RFS
			{order: "0020.0002", node: "Activity_0n17894"},                             // Sync Port-in RFS with other Operators
			{order: "0021", node: "Id_af69b80c-765a-4ae7-be40-22cb244b611b"},           // Order Completion Processing
			{order: "0021.0001.0001", node: "Id_0efd2669-b950-4d2d-aa2e-edc02424d928"}, // Publish Kafka event
			{order: "0021.0002.0001", node: "Event_0imzxiq"},                           //
			{order: "0022", node: "Activity_07y35my"},                                  // Receive Callback for Inventory update and device delivery
			{order: "0023", node: "Gateway_1iqkz3r"},                                   // SIM type?
			{order: "0024", node: "Activity_0gc946n"},                                  // Post SIM card in use
			{order: "0024.0001", node: "Id_7c12b81c-70df-419e-bdea-bbbe4e2cc408"},      // Mark SIM card as used
			{order: "0025", node: "Gateway_1uv7159"},                                   // endif
			{order: "0026", node: "Gateway_19ib8qw"},                                   // Order includes Device?
			{order: "0027", node: "Activity_1vtvz75"},                                  // Post request to deliver Device
			{order: "0027.0001", node: "Activity_1swfmss"},                             // Fulfill Logistics Order to deliver Device
			{order: "0028", node: "Gateway_0y26mfn"},                                   // endif
			{order: "0029", node: "Event_1b30bns"},                                     //
			{order: "A.0001", node: "Activity_0ctpaup"},                                // Submit Port-in cancelation request
			{order: "A.0002", node: "Activity_0rbroim"},                                // Capture Port-in cancelation request
			{order: "A.0003", node: "Activity_1qj1e0v"},                                // Post Port-in cancelation request
			{order: "A.0004", node: "Activity_1cofeb7"},                                // Cancel Port-in Order
			{order: "A.0005", node: "Activity_0fs7ehp"},                                // Post Port-in cancelation
			{order: "A.0006", node: "Activity_14d0wi6"},                                // Process request to cancel Port-in
			{order: "B.0001", node: "Activity_00qw565"},                                // Reply Port-in response
			{order: "C.0001", node: "Activity_14y8a5o"},                                // Resubmit Port-in request
			{order: "C.0002", node: "Activity_1qydfap"},                                // Capture Port-in resubmission request
			{order: "C.0003", node: "Activity_1qrayp8"},                                // Validate Port-in Order (resubmission)
			{order: "C.0003.0001", node: "Activity_1sgwv9e"},                           // Post Port-in resubmission request
			{order: "C.0004", node: "Activity_0ku6ql7"},                                // Perform CRMS validations (resubmission)
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
			assert.Equal(t, x.node, actual[i].Node)
			assert.Equal(t, x.order, actual[i].SortedStep)
		}
		////t.Log(sorted)
		//for _ = range nodes {
		//}
		//for i, node := range nodes {
		//	assert.Equal(t, sorted[i], node.node)
		//}
	}
}

func TestTopologicalSort006(t *testing.T) {
	g := depgraph.New()
	g.AddNode("Activity_1c6h4mm", 430.000000, -220.000000)
	g.AddNode("Activity_0h922nq", 600.000000, -220.000000)
	assert.NoError(t, g.AddLink("", "Activity_1c6h4mm", "Activity_0h922nq")) // Snapshot Account Views --> Refresh
	g.AddNode("Gateway_0utu2zc", 745.000000, -283.000000)
	g.AddNode("Activity_139t6xx", 950.000000, -298.000000)
	assert.NoError(t, g.AddLink("", "Gateway_0utu2zc", "Activity_139t6xx")) // Any Issues? --> Account Data Viewed
	g.AddNode("Gateway_0utu2zc", 745.000000, -283.000000)
	g.AddNode("Activity_1y5gsi9", 820.000000, -210.000000)
	assert.NoError(t, g.AddLink("", "Gateway_0utu2zc", "Activity_1y5gsi9")) // Any Issues? --> Raise Service Request
	g.AddNode("Activity_1y5gsi9", 820.000000, -210.000000)
	g.AddNode("Activity_139t6xx", 950.000000, -298.000000)
	assert.NoError(t, g.AddLink("", "Activity_1y5gsi9", "Activity_139t6xx")) // Raise Service Request --> Account Data Viewed
	g.AddNode("Event_1fy56rv", 82.000000, -276.000000)
	g.AddNode("Activity_1e7q5j6", 160.000000, -298.000000)
	assert.NoError(t, g.AddLink("", "Event_1fy56rv", "Activity_1e7q5j6")) //  --> Account View
	g.AddNode("Gateway_0tnoiya", 315.000000, -283.000000)
	g.AddNode("Activity_0furyas", 550.000000, -110.000000)
	assert.NoError(t, g.AddLink("", "Gateway_0tnoiya", "Activity_0furyas")) // Accordion View --> Transaction Account Views
	g.AddNode("Gateway_0tnoiya", 315.000000, -283.000000)
	g.AddNode("Gateway_0utu2zc", 745.000000, -283.000000)
	assert.NoError(t, g.AddLink("", "Gateway_0tnoiya", "Gateway_0utu2zc")) // Accordion View --> Any Issues?
	g.AddNode("Activity_1e7q5j6", 160.000000, -298.000000)
	g.AddNode("Gateway_0tnoiya", 315.000000, -283.000000)
	assert.NoError(t, g.AddLink("", "Activity_1e7q5j6", "Gateway_0tnoiya")) // Account View --> Accordion View
	g.AddNode("Gateway_0tnoiya", 315.000000, -283.000000)
	g.AddNode("Activity_1c6h4mm", 430.000000, -220.000000)
	assert.NoError(t, g.AddLink("", "Gateway_0tnoiya", "Activity_1c6h4mm")) // Accordion View --> Snapshot Account Views
	g.AddNode("Activity_08e5gsv", 710.000000, -110.000000)
	g.AddNode("Activity_0furyas", 550.000000, -110.000000)
	assert.NoError(t, g.AddLink("", "Activity_08e5gsv", "Activity_0furyas")) // Refresh --> Transaction Account Views
	g.AddNode("Activity_0h922nq", 600.000000, -220.000000)
	g.AddNode("Activity_1c6h4mm", 430.000000, -220.000000)
	assert.NoError(t, g.AddLink("", "Activity_0h922nq", "Activity_1c6h4mm")) // Refresh --> Snapshot Account Views
	g.AddNode("Activity_0furyas", 550.000000, -110.000000)
	g.AddNode("Activity_08e5gsv", 710.000000, -110.000000)
	assert.NoError(t, g.AddLink("", "Activity_0furyas", "Activity_08e5gsv")) // Transaction Account Views --> Refresh
	g.AddNode("Activity_139t6xx", 950.000000, -298.000000)
	g.AddNode("Event_1o2dsrx", 1092.000000, -276.000000)
	assert.NoError(t, g.AddLink("", "Activity_139t6xx", "Event_1o2dsrx")) // Account Data Viewed -->
	{
		expect := []struct {
			order string
			node  string
		}{
			{order: "0001", node: "Event_1fy56rv"},              //
			{order: "0002", node: "Activity_1e7q5j6"},           // Account View
			{order: "0003", node: "Gateway_0tnoiya"},            // Accordion View
			{order: "0003.0001.0001", node: "Activity_1c6h4mm"}, // Snapshot Account Views
			{order: "0003.0001.0002", node: "Activity_0h922nq"}, // Refresh
			{order: "0003.0002.0001", node: "Activity_0furyas"}, // Transaction Account Views
			{order: "0003.0002.0002", node: "Activity_08e5gsv"}, // Refresh
			{order: "0004", node: "Gateway_0utu2zc"},            // Any Issues?
			{order: "0005", node: "Activity_1y5gsi9"},           // Raise Service Request
			{order: "0006", node: "Activity_139t6xx"},           // Account Data Viewed
			{order: "0007", node: "Event_1o2dsrx"},              //
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
			assert.Equal(t, x.node, actual[i].Node)
			assert.Equal(t, x.order, actual[i].SortedStep)
		}
	}

}

func TestTopologicalSort007(t *testing.T) {
	g := depgraph.New()
	g.AddNode("Activity_0l71uiq", 140.000000, 910.000000)
	g.AddNode("Activity_1xyli2s", 270.000000, 910.000000)
	assert.NoError(t, g.AddLink("", "Activity_0l71uiq", "Activity_1xyli2s")) // Request internal Blacklist Validation --> Request external Blacklist Validation
	g.AddNode("Activity_1dbuz2n", 735.000000, 790.000000)
	g.AddNode("Activity_0kpif64", 735.000000, 1450.000000)
	assert.NoError(t, g.AddLink("", "Activity_1dbuz2n", "Activity_0kpif64")) // Request Billing Validation --> Billing Validation
	g.AddNode("Activity_0o84rnf", 735.000000, 640.000000)
	g.AddNode("Activity_1dbuz2n", 735.000000, 790.000000)
	assert.NoError(t, g.AddLink("", "Activity_0o84rnf", "Activity_1dbuz2n")) // Check Billing --> Request Billing Validation
	g.AddNode("Activity_0o84rnf", 735.000000, 640.000000)
	g.AddNode("Gateway_0ko9dvk", 910.000000, 655.000000)
	assert.NoError(t, g.AddLink("", "Activity_0o84rnf", "Gateway_0ko9dvk")) // Check Billing --> Pass Billing Check?
	g.AddNode("Activity_1xyli2s", 270.000000, 910.000000)
	g.AddNode("Activity_1gn4p38", 270.000000, 1050.000000)
	assert.NoError(t, g.AddLink("", "Activity_1xyli2s", "Activity_1gn4p38")) // Request external Blacklist Validation --> External Blacklist Validation
	g.AddNode("Activity_0kpp56m", 140.000000, 640.000000)
	g.AddNode("Gateway_1on7znk", 305.000000, 655.000000)
	assert.NoError(t, g.AddLink("", "Activity_0kpp56m", "Gateway_1on7znk")) // Synchronous Blacklist Validation --> Pass Blacklist?
	g.AddNode("Activity_1la4k4y", 510.000000, 530.000000)
	g.AddNode("Gateway_1174vjb", 650.000000, 655.000000)
	assert.NoError(t, g.AddLink("", "Activity_1la4k4y", "Gateway_1174vjb")) // Capture Release Letter --> endif
	g.AddNode("Activity_1e9uv0g", 885.000000, 530.000000)
	g.AddNode("Activity_173n75w", 885.000000, 380.000000)
	assert.NoError(t, g.AddLink("", "Activity_1e9uv0g", "Activity_173n75w")) // Returns Billing Check Result --> Request for Payment regularization
	g.AddNode("Activity_0kpp56m", 140.000000, 640.000000)
	g.AddNode("Activity_0pb6zdt", 140.000000, 790.000000)
	assert.NoError(t, g.AddLink("", "Activity_0kpp56m", "Activity_0pb6zdt")) // Synchronous Blacklist Validation --> Perform Blacklist Validation
	g.AddNode("Activity_1la4k4y", 510.000000, 530.000000)
	g.AddNode("Activity_0i2xgyh", 510.000000, 790.000000)
	assert.NoError(t, g.AddLink("", "Activity_1la4k4y", "Activity_0i2xgyh")) // Capture Release Letter --> Request to Store Release Letter
	g.AddNode("Activity_19pmjpu", 510.000000, 380.000000)
	g.AddNode("Activity_1la4k4y", 510.000000, 530.000000)
	assert.NoError(t, g.AddLink("", "Activity_19pmjpu", "Activity_1la4k4y")) // Authorized Agent scans the Release Letter --> Capture Release Letter
	g.AddNode("Activity_0i2xgyh", 510.000000, 790.000000)
	g.AddNode("Activity_0cr870w", 510.000000, 1320.000000)
	assert.NoError(t, g.AddLink("", "Activity_0i2xgyh", "Activity_0cr870w")) // Request to Store Release Letter --> Store Release Letter
	g.AddNode("Gateway_1174vjb", 650.000000, 655.000000)
	g.AddNode("Activity_0o84rnf", 735.000000, 640.000000)
	assert.NoError(t, g.AddLink("", "Gateway_1174vjb", "Activity_0o84rnf")) // endif --> Check Billing
	g.AddNode("Activity_0scu270", 280.000000, 530.000000)
	g.AddNode("Activity_0x5lv85", 280.000000, 380.000000)
	assert.NoError(t, g.AddLink("", "Activity_0scu270", "Activity_0x5lv85")) // Returns Blacklist Checking Result --> Informs Customer about Blacklist check result
	g.AddNode("Activity_0x5lv85", 280.000000, 380.000000)
	g.AddNode("Activity_19pmjpu", 510.000000, 380.000000)
	assert.NoError(t, g.AddLink("", "Activity_0x5lv85", "Activity_19pmjpu")) // Informs Customer about Blacklist check result --> Authorized Agent scans the Release Letter
	g.AddNode("Activity_0l71uiq", 140.000000, 910.000000)
	g.AddNode("Activity_0bydgx6", 140.000000, 1180.000000)
	assert.NoError(t, g.AddLink("", "Activity_0l71uiq", "Activity_0bydgx6")) // Request internal Blacklist Validation --> Internal Blacklist Validation
	g.AddNode("Gateway_1on7znk", 305.000000, 655.000000)
	g.AddNode("Gateway_1174vjb", 650.000000, 655.000000)
	assert.NoError(t, g.AddLink("", "Gateway_1on7znk", "Gateway_1174vjb")) // Pass Blacklist? --> endif
	g.AddNode("Gateway_0ko9dvk", 910.000000, 655.000000)
	g.AddNode("Event_1gnl54n", 1132.000000, 662.000000)
	assert.NoError(t, g.AddLink("", "Gateway_0ko9dvk", "Event_1gnl54n")) // Pass Billing Check? -->
	g.AddNode("Gateway_0ko9dvk", 910.000000, 655.000000)
	g.AddNode("Activity_1e9uv0g", 885.000000, 530.000000)
	assert.NoError(t, g.AddLink("", "Gateway_0ko9dvk", "Activity_1e9uv0g")) // Pass Billing Check? --> Returns Billing Check Result
	g.AddNode("Gateway_1on7znk", 305.000000, 655.000000)
	g.AddNode("Activity_0scu270", 280.000000, 530.000000)
	assert.NoError(t, g.AddLink("", "Gateway_1on7znk", "Activity_0scu270")) // Pass Blacklist? --> Returns Blacklist Checking Result
	g.AddNode("Event_0jm34t5", 22.000000, 662.000000)
	g.AddNode("Activity_0kpp56m", 140.000000, 640.000000)
	assert.NoError(t, g.AddLink("", "Event_0jm34t5", "Activity_0kpp56m")) //  --> Synchronous Blacklist Validation
	g.AddNode("Activity_0pb6zdt", 140.000000, 790.000000)
	g.AddNode("Activity_0l71uiq", 140.000000, 910.000000)
	assert.NoError(t, g.AddLink("", "Activity_0pb6zdt", "Activity_0l71uiq")) // Perform Blacklist Validation --> Request internal Blacklist Validation
	{
		expect := []struct {
			order string
			node  string
		}{
			{order: "0001", node: "Event_0jm34t5"},              //
			{order: "0002", node: "Activity_0kpp56m"},           // Synchronous Blacklist Validation
			{order: "0002.0001", node: "Activity_0pb6zdt"},      // Perform Blacklist Validation
			{order: "0002.0002", node: "Activity_0l71uiq"},      // Request internal Blacklist Validation
			{order: "0002.0002.0001", node: "Activity_0bydgx6"}, // Internal Blacklist Validation
			{order: "0002.0003", node: "Activity_1xyli2s"},      // Request external Blacklist Validation
			{order: "0002.0004", node: "Activity_1gn4p38"},      // External Blacklist Validation
			{order: "0003", node: "Gateway_1on7znk"},            // Pass Blacklist?
			{order: "0004", node: "Activity_0scu270"},           // Returns Blacklist Checking Result
			{order: "0005", node: "Activity_0x5lv85"},           // Informs Customer about Blacklist check result
			{order: "0006", node: "Activity_19pmjpu"},           // Authorized Agent scans the Release Letter
			{order: "0007", node: "Activity_1la4k4y"},           // Capture Release Letter
			{order: "0007.0001", node: "Activity_0i2xgyh"},      // Request to Store Release Letter
			{order: "0007.0002", node: "Activity_0cr870w"},      // Store Release Letter
			{order: "0008", node: "Gateway_1174vjb"},            // endif
			{order: "0009", node: "Activity_0o84rnf"},           // Check Billing
			{order: "0009.0001", node: "Activity_1dbuz2n"},      // Request Billing Validation
			{order: "0009.0002", node: "Activity_0kpif64"},      // Billing Validation
			{order: "0010", node: "Gateway_0ko9dvk"},            // Pass Billing Check?
			{order: "0010.0001", node: "Event_1gnl54n"},         //
			{order: "0011", node: "Activity_1e9uv0g"},           // Returns Billing Check Result
			{order: "0012", node: "Activity_173n75w"},           // Request for Payment regularization
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
			assert.Equal(t, x.node, actual[i].Node)
			assert.Equal(t, x.order, actual[i].SortedStep)
		}
	}
}
