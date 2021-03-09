package jmerge

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var input1 = []byte("{ \"spec\": { \"containers\": [ { \"livenessProbe\": { \"httpGet\": { \"path\": \"/liveliness\", \"port\": 3000 }, \"periodSeconds\": 30, \"initialDelaySeconds\": 10 } } ] } }")
var input2 = []byte("{ \"spec\": { \"containers\": [ { \"resources\": { \"requests\": { \"cpu\": \"200m\", \"memory\": \"50M\" } } }  ] } }")

func PrettyPrint(m interface{}) {
	pretty, _ := json.MarshalIndent(m, "", "    ")
	fmt.Println(string(pretty))
}

func TestMerge(t *testing.T) {
	var map1 = make(map[string]interface{})
	err := json.Unmarshal(input1, &map1)
	require.NoError(t, err)

	var map2 = make(map[string]interface{})
	err = json.Unmarshal(input2, &map2)
	require.NoError(t, err)

	var destMap = make(map[string]interface{})

	m1 := Merge(destMap, map1)
	PrettyPrint(m1)
	m2 := Merge(m1, map2)
	PrettyPrint(m2)
}
