package node_test

import (
	"encoding/json"
	"fmt"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"reflect"
	"testing"
)

func TestSpec_Marshal(t *testing.T) {

	examples := []*struct {
		description string
		spec        node.Spec
	}{
		{
			description: "first",
			spec:        node.Spec{ElectionKey: "key-001", Name: "node-001", Address: "192.168.1.1", Port: 8080},
		},
		{
			description: "second",
			spec:        node.Spec{ElectionKey: "key-561", Name: "node-781", Address: "node-781.cluster.priv", Port: 80},
		},
		{
			description: "third",
			spec:        node.Spec{ElectionKey: "key-871", Name: "node-785", Address: "10.10.12.65", Port: 8989},
		},
	}

	for _, example := range examples {
		t.Run(example.description, func(t *testing.T) {
			expected, _ := json.Marshal(example.spec)
			actual := example.spec.Marshal()
			fmt.Println(actual)
			if string(expected) != actual {
				t.Fatalf("\t\t%s FAIL: expected <%s>, actual <%s>", testutils.Failed, string(expected), actual)
			}
			t.Logf("\t\t%s It must marshal the spec to json.", testutils.Succeed)
		})

	}

}

func TestUnmarshal(t *testing.T) {
	examples := []struct {
		description string
		json        string
		expected    node.Spec
	}{
		{
			description: "first",
			json:        "{\"ElectionKey\":\"key-001\",\"Name\":\"node-001\",\"Address\":\"192.168.1.1\",\"Port\":8080}",
			expected:    node.Spec{ElectionKey: "key-001", Name: "node-001", Address: "192.168.1.1", Port: 8080},
		},
		{
			description: "second",
			json:        "{\"ElectionKey\":\"key-561\",\"Name\":\"node-781\",\"Address\":\"node-781.cluster.priv\",\"Port\":80}",
			expected:    node.Spec{ElectionKey: "key-561", Name: "node-781", Address: "node-781.cluster.priv", Port: 80},
		},
		{
			description: "third",
			json:        "{\"ElectionKey\":\"key-871\",\"Name\":\"node-785\",\"Address\":\"10.10.12.65\",\"Port\":8989}",
			expected:    node.Spec{ElectionKey: "key-871", Name: "node-785", Address: "10.10.12.65", Port: 8989},
		},
	}

	for _, example := range examples {
		actual, err := node.Unmarshal([]byte(example.json))
		if err != nil {
			t.Fatalf("\t\t%s FAIL: failed to unmarshel %v", testutils.Failed, err)
		}

		if example.expected.ElectionKey != actual.GetElectionKey() {
			t.Fatalf("\t\t%s FAIL: expected <%s>, actual <%s>", testutils.Failed, example.expected.ElectionKey, actual.GetName())
		}
		t.Logf("\t\t%s The election key must match.", testutils.Succeed)

		if example.expected.Name != actual.GetName() {
			t.Fatalf("\t\t%s FAIL: expected <%s>, actual <%s>", testutils.Failed, example.expected.Name, actual.GetName())
		}
		t.Logf("\t\t%s The name must match.", testutils.Succeed)

		if example.expected.Address != actual.GetAddress() {
			t.Fatalf("\t\t%s FAIL: expected <%s>, actual <%s>", testutils.Failed, example.expected.Address, actual.GetAddress())
		}
		t.Logf("\t\t%s The address must match.", testutils.Succeed)

		if example.expected.Port != actual.GetPort() {
			t.Fatalf("\t\t%s FAIL: expected <%d>, actual <%d>", testutils.Failed, example.expected.Port, actual.GetPort())
		}
		t.Logf("\t\t%s The port must match.", testutils.Succeed)

		if !reflect.DeepEqual(example.expected.GetSpec(), actual.GetSpec()) {
			t.Fatalf("\t\t%s FAIL: expected to return the same spec. Actually not.", testutils.Failed)
		}
		t.Logf("\t\t%s It must return the same spec.", testutils.Succeed)

	}
}
