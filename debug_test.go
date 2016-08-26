package debug

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func assertContains(t *testing.T, str, substr string) {
	if !strings.Contains(str, substr) {
		t.Fatalf("expected %q to contain %q", str, substr)
	}
}

func assertNotContains(t *testing.T, str, substr string) {
	if strings.Contains(str, substr) {
		t.Fatalf("expected %q to not contain %q", str, substr)
	}
}

func TestDefault(t *testing.T) {
	var b []byte
	buf := bytes.NewBuffer(b)
	SetWriter(buf)

	debug := Debug("foo")
	debug("something")
	debug("here")
	debug("whoop")

	if buf.Len() != 0 {
		t.Fatalf("buffer should be empty")
	}
}

func TestEnable(t *testing.T) {
	var b []byte
	buf := bytes.NewBuffer(b)
	SetWriter(buf)

	Enable("foo")

	debug := Debug("foo")
	debug("something")
	debug("here")
	debug("whoop")

	if buf.Len() == 0 {
		t.Fatalf("buffer should have output")
	}

	str := string(buf.Bytes())
	assertContains(t, str, "something")
	assertContains(t, str, "here")
	assertContains(t, str, "whoop")
}

func TestMultipleOneEnabled(t *testing.T) {
	var b []byte
	buf := bytes.NewBuffer(b)
	SetWriter(buf)

	Enable("foo")

	foo := Debug("foo")
	foo("foo")

	bar := Debug("bar")
	bar("bar")

	if buf.Len() == 0 {
		t.Fatalf("buffer should have output")
	}

	str := string(buf.Bytes())
	assertContains(t, str, "foo")
	assertNotContains(t, str, "bar")
}

func TestMultipleEnabled(t *testing.T) {
	var b []byte
	buf := bytes.NewBuffer(b)
	SetWriter(buf)

	Enable("foo,bar")

	foo := Debug("foo")
	foo("foo")

	bar := Debug("bar")
	bar("bar")

	if buf.Len() == 0 {
		t.Fatalf("buffer should have output")
	}

	str := string(buf.Bytes())
	assertContains(t, str, "foo")
	assertContains(t, str, "bar")
}

func TestEnableDisable(t *testing.T) {
	var b []byte
	buf := bytes.NewBuffer(b)
	SetWriter(buf)

	Enable("foo,bar")
	Disable()

	foo := Debug("foo")
	foo("foo")

	bar := Debug("bar")
	bar("bar")

	if buf.Len() != 0 {
		t.Fatalf("buffer should not have output")
	}
}

func TestExcludes(t *testing.T) {
	var b []byte
	buf := bytes.NewBuffer(b)
	SetWriter(buf)

	Enable("*,-foo")

	foo := Debug("foo")
	foo("foo")

	bar := Debug("bar")
	bar("bar")

	if buf.Len() == 0 {
		t.Fatalf("buffer should have output")
	}

	str := string(buf.Bytes())
	assertNotContains(t, str, "foo")
	assertContains(t, str, "bar")
}

func TestSplitPattern(t *testing.T) {
	var testCases = []struct {
		input    string
		includes string
		excludes string
	}{
		{"*", "*", ""},
		{"*,-foo", "*", "foo"},
		{"*,-foo,-bar:baz,-one:two:three", "*", "foo,bar:baz,one:two:three"},
		{"-one,two,-three,four,-five,six", "two,four,six", "one,three,five"},
	}

	for _, testCase := range testCases {
		includes, excludes := splitPattern(testCase.input)

		if includes != testCase.includes {
			t.Errorf("splitPattern includes(%s): expected %s, actual %s", testCase.input, testCase.includes, includes)
		}

		if excludes != testCase.excludes {
			t.Errorf("splitPattern excludes(%s): expected %s, actual %s", testCase.input, testCase.excludes, excludes)
		}
	}
}

func TestPatternToRegex(t *testing.T) {
	var testCases = []struct {
		input  string
		output string
	}{
		{"\\*", "^(.*?)$"},
		{"a,b", "^(a|b)$"},
		{"a:\\*,b:\\*", "^(a:.*?|b:.*?)$"},
	}

	for _, testCase := range testCases {
		actual := patternToRegex(testCase.input)

		if actual != testCase.output {
			t.Errorf("patternToRegex(%s): expected %s, actual %s", testCase.input, testCase.output, actual)
		}
	}
}

func ExampleEnable() {
	Enable("mongo:connection")
	Enable("mongo:*")
	Enable("foo,bar,baz")
	Enable("*")
}

func ExampleDebug() {
	var debug = Debug("single")

	for {
		debug("sending mail")
		debug("send email to %s", "tobi@segment.io")
		debug("send email to %s", "loki@segment.io")
		debug("send email to %s", "jane@segment.io")
		time.Sleep(500 * time.Millisecond)
	}
}

func BenchmarkDisabled(b *testing.B) {
	debug := Debug("something")
	for i := 0; i < b.N; i++ {
		debug("stuff")
	}
}

func BenchmarkNonMatch(b *testing.B) {
	debug := Debug("something")
	Enable("nonmatch")
	for i := 0; i < b.N; i++ {
		debug("stuff")
	}
}
