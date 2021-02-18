package json

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONText is a json.RawMessage, which is a []byte underneath.
// Value() validates the json format in the source, and returns an error if
// the json is not valid.  Scan does no validation.  JSONText additionally
// implements `Unmarshal`, which unmarshals the json within to an interface{}
type List json.RawMessage

var EmptyJSONList = List("[]")

func StructToJsonList(v interface{}) (List, error) {
	j, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return j, nil
}

func StructToJsonListOrEmpty(v interface{}) List {
	obj, err := StructToJsonList(v)
	if err != nil {
		return EmptyJSONList
	}

	return obj
}

// MarshalJSON returns the *j as the JSON encoding of j.
func (j List) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return EmptyJSONList, nil
	}
	return j, nil
}

// UnmarshalJSON sets *j to a copy of Repo
func (j *List) UnmarshalJSON(data []byte) error {
	if j == nil {
		return fmt.Errorf("json.List: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// Value returns j as a value.  This does a validating unmarshal into another
// RawMessage.  If j is invalid json, it returns an error.
func (j List) Value() (driver.Value, error) {
	var m json.RawMessage
	var err = j.Unmarshal(&m)
	if err != nil {
		return []byte{}, err
	}
	return []byte(j), nil
}

// Scan stores the src in *j.  No validation is done.
func (j *List) Scan(src interface{}) error {
	var source []byte
	switch t := src.(type) {
	case string:
		source = []byte(t)
	case []byte:
		if len(t) == 0 {
			source = EmptyJSONList
		} else {
			source = t
		}
	case nil:
		*j = EmptyJSONList
	default:
		return fmt.Errorf("incompatible type for json.List")
	}
	*j = append((*j)[0:0], source...)
	return nil
}

// Unmarshal unmarshal's the json in j to v, as in json.Unmarshal.
func (j *List) Unmarshal(v interface{}) error {
	if len(*j) == 0 {
		*j = EmptyJSONList
	}
	return json.Unmarshal(*j, v)
}

func (j *List) AsList() ([]string, error) {
	list := make([]string, 0)
	err := j.Unmarshal(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (j *List) AsListOrEmpty() []string {
	l, err := j.AsList()
	if err != nil {
		return []string{}
	}
	return l
}

// String supports pretty printing for JSONText types.
func (j List) String() string {
	return string(j)
}

func FromList(l []string) (List, error) {
	if l == nil {
		return EmptyJSONList, nil
	}

	b, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func FromListOrEmpty(l []string) List {
	list, err := FromList(l)
	if err != nil {
		return EmptyJSONList
	}

	return list
}
