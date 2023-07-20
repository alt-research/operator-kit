package numstr

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	"golang.org/x/exp/constraints"
)

// +k8s:openapi-gen=true
type NumOrString struct {
	Type     Type
	IntVal   big.Int
	FloatVal big.Float
	StrVal   string
}

// Type represents the stored type of NumOrString.
type Type int64

const (
	Int    Type = iota // The NumOrString holds an int.
	Float              // The NumOrString holds an float.
	String             // The NumOrString holds a string.
)

var (
	BigMaxInt   = big.NewInt(math.MaxInt64)
	BigMaxFloat = big.NewFloat(math.MaxFloat64)
)

func FromInt[T constraints.Integer](val T) NumOrString {
	return NumOrString{Type: Int, IntVal: *big.NewInt(int64(val))}
}

func FromFloat[T constraints.Float](val T) NumOrString {
	return NumOrString{Type: Int, FloatVal: *big.NewFloat(float64(val))}
}

// FromString creates an NumOrString object with a string value.
func FromString(val string) NumOrString {
	return NumOrString{Type: String, StrVal: val}
}

// Parse the given string and try to convert it to an integer before
// setting it as a string value.
func Parse(s string) NumOrString {
	if i, ok := big.NewInt(0).SetString(s, 10); ok {
		return NumOrString{Type: Int, IntVal: *i}
	} else if f, ok := big.NewFloat(0).SetString(s); ok {
		return NumOrString{Type: Float, FloatVal: *f}
	} else {
		return NumOrString{Type: String, StrVal: s}
	}
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (ns *NumOrString) UnmarshalJSON(value []byte) error {
	if value[0] == '"' {
		ns.Type = String
		return json.Unmarshal(value, &ns.StrVal)
	}
	n := Parse(string(value))
	ns.Type = n.Type
	ns.FloatVal = n.FloatVal
	ns.IntVal = n.IntVal
	return nil
}

// String returns the string value, or the Itoa of the int value.
func (ns *NumOrString) String() string {
	if ns == nil {
		return "<nil>"
	}
	switch ns.Type {
	case String:
		return ns.StrVal
	case Float:
		return ns.FloatVal.String()
	case Int:
		return ns.IntVal.String()
	default:
		panic("unsupported type")
	}
}

// MarshalJSON implements the json.Marshaller interface.
func (ns NumOrString) MarshalJSON() ([]byte, error) {
	switch ns.Type {
	case Int:
		if ns.IntVal.IsInt64() {
			return json.Marshal(ns.IntVal.Int64())
		}
		return json.Marshal(ns.IntVal.String())
	case Float:
		f, a := ns.FloatVal.Float64()
		if a == big.Exact {
			return json.Marshal(f)
		}
		return json.Marshal(ns.FloatVal.String())
	case String:
		return json.Marshal(ns.StrVal)
	default:
		return []byte{}, fmt.Errorf("impossible NumOrString.Type")
	}
}

// OpenAPISchemaType is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
//
// See: https://github.com/kubernetes/kube-openapi/tree/master/pkg/generators
func (NumOrString) OpenAPISchemaType() []string { return []string{"string"} }

// OpenAPISchemaFormat is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
func (NumOrString) OpenAPISchemaFormat() string { return "number-or-string" }

// OpenAPIV3OneOfTypes is used by the kube-openapi generator when constructing
// the OpenAPI v3 spec of this type.
func (NumOrString) OpenAPIV3OneOfTypes() []string { return []string{"number", "integer", "string"} }
