package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	test "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/testcommon"
	"github.com/awalterschulze/gographviz"
)

func main() {
	// callGraph := test.CreateGraphTestOneAsyncCall()
	// callGraph := test.CreateGraphTestTwoAsyncCalls()
	// callGraph := test.CreateGraphTestAsyncCallsAsync()
	// callGraph := test.CreateGraphTestAsyncCallsAsync2() // not allowed to run!
	// callGraph := test.CreateGraphTestDifferentTypeOfCallsToSameFunction()

	// callGraph := test.CreateGraphTestCallbackCallsAsync()
	// callGraph := test.CreateGraphTestSimpleSyncAndAsync1()
	// callGraph := test.CreateGraphTestSimpleSyncAndAsync2()
	// callGraph := test.CreateGraphTest1()
	callGraph := test.CreateGraphTest2()

	///////////////////

	graphviz := toGraphviz(callGraph, true)
	createSvg("1 call-graph", graphviz)

	executionGraph := callGraph.CreateExecutionGraphFromCallGraph()
	graphviz = toGraphviz(executionGraph, true)
	createSvg("2 execution-graph", graphviz)

	gasGraph := executionGraph.CreateGasGraphFromExecutionGraph()
	graphviz = toGraphviz(gasGraph, false)
	createSvg("3 tree-call-graph", graphviz)

	gasGraph.ComputeRemainingGasBeforeCallbacks()
	graphviz = toGraphviz(gasGraph, false)
	createSvg("4 gas-graph-gasbeforecallbacks", graphviz)

	gasGraph.ComputeGasStepByStep(func(graph *test.TestCallGraph, step int) {
		graphviz = toGraphviz(gasGraph, false)
		createSvg(fmt.Sprintf("step %d", step), graphviz)
	})
}

func createSvg(file string, graphviz *gographviz.Graph) {
	location := os.Args[1]

	destDot := location + file + ".dot"

	output := graphviz.String()
	err := ioutil.WriteFile(destDot, []byte(output), 0644)
	if err != nil {
		panic(err)
	}

	out, err := exec.Command("dot" /*"-extent 800x1500",*/, "-Tsvg", destDot).Output()
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(location+file+".svg", out, 0644)
	if err != nil {
		panic(err)
	}
}

func toGraphviz(graph *test.TestCallGraph, showGasEdgeLabels bool) *gographviz.Graph {
	graphviz := gographviz.NewGraph()
	graphviz.Directed = true
	graphName := "G"
	graphviz.Attrs["nodesep"] = "1.5"

	nodeCounters := make(map[string]int)
	for _, node := range graph.Nodes {
		node.Label, node.VisualLabel = getGraphvizNodeLabel(node, nodeCounters)
	}

	for _, node := range graph.Nodes {
		attrs := make(map[string]string)
		if node.IsStartNode {
			attrs["shape"] = "box"
		}
		if node.Visited {
			attrs["penwidth"] = "4"
		}
		setGasLabel(node, attrs)
		if !node.IsEndOfSyncExecutionNode {
			attrs["bgcolor"] = "grey"
			attrs["style"] = "filled"
			attrs["label"] = node.VisualLabel
		}
		from := node.Label
		graphviz.AddNode(graphName, from, attrs)
		for _, edge := range node.GetEdges() {
			to := edge.To.Label
			attrs := make(map[string]string)
			if edge.Label != "" {
				attrs["label"] = edge.Label
				if showGasEdgeLabels && edge.Type != test.Callback && edge.Type != test.GroupCallback {
					attrs["label"] += "\n" +
						"P" + strconv.Itoa(int(edge.GasLimit)) +
						"/U" + strconv.Itoa(int(edge.GasUsed))
					if edge.Type == test.Async {
						attrs["label"] += "/CU" + strconv.Itoa(int(edge.GasUsedByCallback))
					}
				}
				attrs["label"] = strconv.Quote(attrs["label"])
			}
			if edge.Color != "" {
				attrs["color"] = edge.Color
			} else {
				attrs["color"] = "black"
			}
			graphviz.AddEdge(from, to, true, attrs)
		}
	}

	return graphviz
}

func getGraphvizNodeLabel(node *test.TestCallNode, nodeCounters map[string]int) (string, string) {
	if nodeCounters == nil {
		return node.Label, node.Label
	}
	if node.VisualLabel != "" {
		return node.Label, node.VisualLabel
	}

	var prefix string
	if node.Call.FunctionName == test.SpecialLabel {
		prefix = test.SpecialLabel
	} else {
		prefix, _ = strconv.Unquote(node.Label)
	}

	counter, present := nodeCounters[prefix]
	if !present {
		counter = 0
	}
	counter++
	nodeCounters[prefix] = counter

	suffix := ""
	if counter > 1 {
		suffix = "_" + strconv.Itoa(counter)
	}
	return strconv.Quote(prefix + suffix), strconv.Quote(prefix)
}

const gasFontStart = "<<font color='green'>"
const gasFontEnd = "</font>>"

func setGasLabel(node *test.TestCallNode, attrs map[string]string) {
	if node.GasLimit == 0 && node.GasUsed == 0 {
		// special label for end nodes without gas info
		if node.IsEndOfSyncExecutionNode {
			attrs["label"] = strconv.Quote("*")
		}
		return
	}

	gasLimit := strconv.Itoa(int(node.GasLimit))
	gasUsed := strconv.Itoa(int(node.GasUsed))
	gasRemaining := strconv.Itoa(int(node.GasRemaining))
	gasRemainingAfterCallback := strconv.Itoa(int(node.GasRemainingAfterCallback))
	gasLocked := strconv.Itoa(int(node.GasLocked))
	var xlabel string
	if node.IsEndOfSyncExecutionNode {
		attrs["label"] = gasFontStart + gasUsed + gasFontEnd
	} else {
		// display only gas locked for uncomputed gas values (for group callbacks and context callbacks)
		if node.GasLimit == 0 {
			// xlabel += gasFontStart + "L" + gasLocked + gasFontEnd
			// attrs["xlabel"] = xlabel
			return
		}
		xlabel = gasFontStart
		xlabel += "P" + gasLimit
		if node.GasLocked != 0 {
			xlabel += "/L" + gasLocked
		}

		// TODO matei-p only for debug
		// xlabel += "/U" + strconv.Itoa(int(node.GasUsed))

		xlabel += "<br/>R" + gasRemaining
		if node.GasRemainingAfterCallback != 0 {
			xlabel += "<br/>A" + gasRemainingAfterCallback
		}
		xlabel += gasFontEnd
		attrs["xlabel"] = xlabel
	}
}
