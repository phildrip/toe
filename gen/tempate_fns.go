package gen

import (
	"fmt"
	"strings"
)

// zip combines two lists of strings into a list of strings, with the
// elements of both lists combined with fmtStr.
// eg zip([]string{"a", "b"}, []string{"1", "2"}, "%s:%s") => []string{"a:1", "b:2"}
func zip(a []string, b []string, fmtStr string) []string {
	if len(a) != len(b) {
		panic("unequal length")
	}
	var zipped []string
	for i := range a {
		zipped = append(zipped, fmt.Sprintf(fmtStr, a[i], b[i]))
	}
	return zipped
}

// joinl joins a list of strings with a separator, with arguments reversed
// compared to strings.Join.
// It's useful in go templates, so we can pipe a list of strings into a template.
func joinl(sep string, a []string) string {
	return strings.Join(a, sep)
}
