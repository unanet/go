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

	var entries = []map[string]interface{}{map1,map2}

	var m1 interface{}
	for _,i := range entries {
		if len(destMap) == 0 {
			m1 = Merge(destMap, i)
		}else{
			m1 = Merge(destMap, i)
		}

	}

	//m1 := Merge(make(map[string]interface{}), map1)
	// PrettyPrint(m1)
	//m2 := Merge(m1, map2)
	PrettyPrint(m1)
}
