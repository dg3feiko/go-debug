package debug

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	writer      io.Writer = os.Stderr
	regIncludes *regexp.Regexp
	regExcludes *regexp.Regexp
	m           sync.Mutex
	enabled     = false
)

// Debugger function.
type DebugFunction func(string, ...interface{})

// Terminal colors used at random.
var colors []string = []string{
	"31",
	"32",
	"33",
	"34",
	"35",
	"36",
}

// Initialize with DEBUG environment variable.
func init() {
	env := os.Getenv("DEBUG")

	if "" != env {
		Enable(env)
	}
}

// SetWriter replaces the default of os.Stderr with `w`.
func SetWriter(w io.Writer) {
	m.Lock()
	defer m.Unlock()
	writer = w
}

// Disable all pattern matching. This function is thread-safe.
func Disable() {
	m.Lock()
	defer m.Unlock()
	enabled = false
}

// Enable the given debug `pattern`. Patterns take a glob-like form,
// for example if you wanted to enable everything, just use "*", or
// if you had a library named mongodb you could use "mongodb:connection",
// or "mongodb:*". Multiple matches can be made with a comma, for
// example "mongo*,redis*".
//
// This function is thread-safe.
func Enable(pattern string) {
	m.Lock()
	defer m.Unlock()
	pattern = regexp.QuoteMeta(pattern)

	includes, excludes := splitPattern(pattern)

	regIncludes = regexp.MustCompile(patternToRegex(includes))
	regExcludes = regexp.MustCompile(patternToRegex(excludes))

	enabled = true
}

// Debug creates a debug function for `name` which you call
// with printf-style arguments in your application or library.
func Debug(name string) DebugFunction {
	prevGlobal := time.Now()
	color := colors[rand.Intn(len(colors))]
	prev := time.Now()

	return func(format string, args ...interface{}) {
		if !enabled {
			return
		}

		if regExcludes.MatchString(name) {
			return
		}

		if !regIncludes.MatchString(name) {
			return
		}

		d := deltas(prevGlobal, prev, color)
		fmt.Fprintf(writer, d+" \033["+color+"m"+name+"\033[0m - "+format+"\n", args...)
		prevGlobal = time.Now()
		prev = time.Now()
	}
}

// Return formatting for deltas.
func deltas(prevGlobal, prev time.Time, color string) string {
	now := time.Now()
	global := now.Sub(prevGlobal).Nanoseconds()
	delta := now.Sub(prev).Nanoseconds()
	ts := now.UTC().Format("15:04:05.000")
	deltas := fmt.Sprintf("%s %-6s \033["+color+"m%-6s", ts, humanizeNano(global), humanizeNano(delta))
	return deltas
}

// Humanize nanoseconds to a string.
func humanizeNano(n int64) string {
	var suffix string

	switch {
	case n > 1e9:
		n /= 1e9
		suffix = "s"
	case n > 1e6:
		n /= 1e6
		suffix = "ms"
	case n > 1e3:
		n /= 1e3
		suffix = "us"
	default:
		suffix = "ns"
	}

	return strconv.Itoa(int(n)) + suffix
}

// filterSlice is a standard array filter implementation
func filterSlice(strings []string, fn func(string) bool) []string {
	var result []string
	for _, val := range strings {
		if fn(val) {
			result = append(result, val)
		}
	}
	return result
}

// mapSlice is a standard array map implementation
func mapSlice(strings []string, fn func(string) string) []string {
	result := make([]string, len(strings))
	for i, val := range strings {
		result[i] = fn(val)
	}
	return result
}

// splitPattern takes in a debug pattern and splits the comma-list into
// inclusions and exclusions, depending on whether the token begins with a dash
func splitPattern(pattern string) (string, string) {
	tokens := strings.Split(pattern, ",")

	includes := filterSlice(tokens, func(str string) bool {
		return !strings.HasPrefix(str, "-")
	})

	excludes := filterSlice(tokens, func(str string) bool {
		return strings.HasPrefix(str, "-")
	})
	excludes = mapSlice(excludes, func(str string) string {
		return str[1:]
	})

	return strings.Join(includes, ","), strings.Join(excludes, ",")
}

// patternToRegex takes a debug string pattern and formats it as a regex string
func patternToRegex(pattern string) string {
	pattern = strings.Replace(pattern, "\\*", ".*?", -1)
	pattern = strings.Replace(pattern, ",", "|", -1)
	return "^(" + pattern + ")$"
}
