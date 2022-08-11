package jmerge

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func PrettyPrint(m interface{}) {
	pretty, _ := json.MarshalIndent(m, "", "    ")
	fmt.Println(string(pretty))
}

func TestMerge(t *testing.T) {
	b1, err := os.ReadFile("./test_files/file_3.json")
	require.NoError(t, err)
	var map1 = make(map[string]interface{})
	err = json.Unmarshal(b1, &map1)
	require.NoError(t, err)

	b2, err := os.ReadFile("./test_files/file_4.json")
	require.NoError(t, err)
	var map2 = make(map[string]interface{})
	err = json.Unmarshal(b2, &map2)
	require.NoError(t, err)

	var destMap = make(map[string]interface{})

	var entries = []map[string]interface{}{map1, map2}

	for _, i := range entries {
		destMap = Merge(destMap, i)
	}

	//m1 := Merge(make(map[string]interface{}), map1)
	// PrettyPrint(m1)
	//m2 := Merge(m1, map2)
	PrettyPrint(destMap)
}
