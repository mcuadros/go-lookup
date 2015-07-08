package lookup

import (
	"reflect"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestLookup_Map(c *C) {
	value, err := Lookup(map[string]int{"foo": 42}, "foo")
	c.Assert(err, IsNil)
	c.Assert(value.Int(), Equals, int64(42))
}

func (s *S) TestLookup_Ptr(c *C) {
	value, err := Lookup(&structFixture, "String")
	c.Assert(err, IsNil)
	c.Assert(value.String(), Equals, "foo")
}

func (s *S) TestLookup_StructBasic(c *C) {
	value, err := Lookup(structFixture, "String")
	c.Assert(err, IsNil)
	c.Assert(value.String(), Equals, "foo")
}

func (s *S) TestLookup_StructPlusMap(c *C) {
	value, err := Lookup(structFixture, "Map", "foo")
	c.Assert(err, IsNil)
	c.Assert(value.Int(), Equals, int64(42))
}

func (s *S) TestAggregableLookup_StructIndex(c *C) {
	value, err := Lookup(structFixture, "StructSlice", "Map", "foo")

	c.Assert(err, IsNil)
	c.Assert(value.Interface(), DeepEquals, []int{42, 42})
}

func (s *S) TestAggregableLookup_StructNestedMap(c *C) {
	value, err := Lookup(structFixture, "StructSlice[0]", "String")

	c.Assert(err, IsNil)
	c.Assert(value.Interface(), DeepEquals, "foo")
}

func (s *S) TestAggregableLookup_StructNested(c *C) {
	value, err := Lookup(structFixture, "StructSlice", "StructSlice", "String")

	c.Assert(err, IsNil)
	c.Assert(value.Interface(), DeepEquals, []string{"bar", "foo", "qux", "baz"})
}

func (s *S) TestAggregableLookupString_Complex(c *C) {
	value, err := LookupString(structFixture, "StructSlice.StructSlice[0].String")
	c.Assert(err, IsNil)
	c.Assert(value.Interface(), DeepEquals, []string{"bar", "foo", "qux", "baz"})

	value, err = LookupString(structFixture, "StructSlice[0].Map.foo")
	c.Assert(err, IsNil)
	c.Assert(value.Interface(), DeepEquals, 42)
}

func (s *S) TestMergeValue(c *C) {
	v := mergeValue([]reflect.Value{reflect.ValueOf("qux"), reflect.ValueOf("foo")})
	c.Assert(v.Interface(), DeepEquals, []string{"qux", "foo"})
}

func (s *S) TestMergeValueSlice(c *C) {
	v := mergeValue([]reflect.Value{
		reflect.ValueOf([]string{"foo", "bar"}),
		reflect.ValueOf([]string{"qux", "baz"}),
	})

	c.Assert(v.Interface(), DeepEquals, []string{"foo", "bar", "qux", "baz"})
}

func (s *S) TestMergeValueZero(c *C) {
	v := mergeValue([]reflect.Value{reflect.Value{}, reflect.ValueOf("foo")})
	c.Assert(v.Interface(), DeepEquals, []string{"foo"})
}

func (s *S) TestParseIndex(c *C) {
	key, index, err := parseIndex("foo[42]")
	c.Assert(err, IsNil)
	c.Assert(key, Equals, "foo")
	c.Assert(index, Equals, 42)
}

func (s *S) TestParseIndexNooIndex(c *C) {
	key, index, err := parseIndex("foo")
	c.Assert(err, IsNil)
	c.Assert(key, Equals, "foo")
	c.Assert(index, Equals, -1)
}

func (s *S) TestParseIndexMalFormed(c *C) {
	key, index, err := parseIndex("foo[]")
	c.Assert(err, Equals, ErrMalformedIndex)
	c.Assert(key, Equals, "")
	c.Assert(index, Equals, -1)

	key, index, err = parseIndex("foo[42")
	c.Assert(err, Equals, ErrMalformedIndex)
	c.Assert(key, Equals, "")
	c.Assert(index, Equals, -1)

	key, index, err = parseIndex("foo42]")
	c.Assert(err, Equals, ErrMalformedIndex)
	c.Assert(key, Equals, "")
	c.Assert(index, Equals, -1)
}

type MyStruct struct {
	String      string
	Map         map[string]int
	Nested      *MyStruct
	StructSlice []*MyStruct
}

var mapFixutre = map[string]int{"foo": 42}
var structFixture = MyStruct{
	String: "foo",
	Map:    mapFixutre,
	StructSlice: []*MyStruct{
		{Map: mapFixutre, String: "foo", StructSlice: []*MyStruct{{String: "bar"}, {String: "foo"}}},
		{Map: mapFixutre, String: "qux", StructSlice: []*MyStruct{{String: "qux"}, {String: "baz"}}},
	},
}
