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
type Object json.RawMessage

var EmptyJSONObject = Object("{}")

func StructToJsonObject(v interface{}) (Object, error) {
	j, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return j, nil
}

// MarshalJSON returns the *j as the JSON encoding of j.
func (j Object) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return EmptyJSONObject, nil
	}
	return j, nil
}

// UnmarshalJSON sets *j to a copy of Repo
func (j *Object) UnmarshalJSON(data []byte) error {
	if j == nil {
		return fmt.Errorf("json.Object: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// Value returns j as a value.  This does a validating unmarshal into another
// RawMessage.  If j is invalid json, it returns an error.
func (j Object) Value() (driver.Value, error) {
	var m json.RawMessage
	var err = j.Unmarshal(&m)
	if err != nil {
		return []byte{}, err
	}
	return []byte(j), nil
}

// Scan stores the src in *j.  No validation is done.
func (j *Object) Scan(src interface{}) error {
	var source []byte
	switch t := src.(type) {
	case string:
		source = []byte(t)
	case []byte:
		if len(t) == 0 {
			source = EmptyJSONObject
		} else {
			source = t
		}
	case nil:
		*j = EmptyJSONObject
	default:
		return fmt.Errorf("incompatible type for json.Object")
	}
	*j = append((*j)[0:0], source...)
	return nil
}

// Unmarshal unmarshal's the json in j to v, as in json.Unmarshal.
func (j *Object) Unmarshal(v interface{}) error {
	if len(*j) == 0 {
		*j = EmptyJSONObject
	}
	return json.Unmarshal(*j, v)
}

func (j *Object) AsMap() (map[string]interface{}, error) {
	hash := make(map[string]interface{})
	err := j.Unmarshal(&hash)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// AsMapOrEmpty returns a map or if it can't unmarshal it, it returns an empty map
func (j *Object) AsMapOrEmpty() map[string]interface{} {
	m, err := j.AsMap()
	if err != nil {
		return map[string]interface{}{}
	} else {
		return m
	}
}

func (j *Object) AsStringMap() (map[string]string, error) {
	hash := make(map[string]string)
	err := j.Unmarshal(&hash)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (j *Object) AsStringMapOrEmpty() map[string]string {
	m, err := j.AsStringMap()
	if err != nil {
		return map[string]string{}
	} else {
		return m
	}
}

// String supports pretty printing for JSONText types.
func (j Object) String() string {
	return string(j)
}

func FromMap(m map[string]interface{}) (Object, error) {
	if m == nil {
		return EmptyJSONObject, nil
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func FromMapOrEmpty(m map[string]interface{}) Object {
	b, err := FromMap(m)
	if err != nil {
		return EmptyJSONObject
	}
	return b
}

func FromStringMap(m map[string]string) (Object, error) {
	if m == nil {
		return EmptyJSONObject, nil
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func FromStringMapOrEmpty(m map[string]string) Object {
	b, err := FromStringMap(m)
	if err != nil {
		return EmptyJSONObject
	}
	return b
}
