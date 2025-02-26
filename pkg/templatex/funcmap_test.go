package templatex

import (
	"testing"

	"github.com/iancoleman/strcase"
	"github.com/stretchr/testify/assert"
)

func TestUpperCamel(t *testing.T) {
	testData := []struct {
		input  string
		expect string
	}{
		{input: "", expect: ""},
		{input: "foo", expect: "Foo"},
		{input: "foo bar", expect: "FooBar"},
		{input: "foo_bar", expect: "FooBar"},
		{input: "foo-bar", expect: "FooBar"},
		{input: "_foobar", expect: "Foobar"},
		{input: "_foobar_", expect: "Foobar"},
	}
	for _, v := range testData {
		actual := strcase.ToCamel(v.input)
		assert.Equal(t, v.expect, actual)
	}
}

func TestLowerCamel(t *testing.T) {
	testData := []struct {
		input  string
		expect string
	}{
		{input: "", expect: ""},
		{input: "foo", expect: "foo"},
		{input: "Foo bar", expect: "fooBar"},
		{input: "Foo_bar", expect: "fooBar"},
		{input: "Foo-bar", expect: "fooBar"},
		{input: "_foobar", expect: "Foobar"},
		{input: "_foobar_", expect: "Foobar"},
		{input: "FooBar", expect: "fooBar"},
		{input: "Foo_Bar", expect: "fooBar"},
	}
	for _, v := range testData {
		actual := strcase.ToLowerCamel(v.input)
		assert.Equal(t, v.expect, actual)
	}
}

func TestLineComment(t *testing.T) {
	testData := []struct {
		input  string
		expect string
	}{
		{input: "", expect: ""},
		{input: "foo", expect: "foo"},
		{input: "foo\nbar", expect: "foo\n// bar"},
		{input: "foo\nbar\n", expect: "foo\n// bar"},
		{input: "\nfoo\nbar\n", expect: "foo\n// bar"},
	}
	for _, v := range testData {
		actual := lineComment(v.input)
		assert.Equal(t, v.expect, actual)
	}
}
