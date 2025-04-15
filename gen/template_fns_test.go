package gen

import (
	"reflect"
	"testing"
)

func TestZip(t *testing.T) {
	tests := []struct {
		name   string
		a      []string
		b      []string
		fmtStr string
		want   []string
	}{
		{
			name:   "basic zip",
			a:      []string{"a", "b", "c"},
			b:      []string{"1", "2", "3"},
			fmtStr: "%s:%s",
			want:   []string{"a:1", "b:2", "c:3"},
		},
		{
			name:   "empty lists",
			a:      []string{},
			b:      []string{},
			fmtStr: "%s-%s",
			want:   []string{},
		},
		{
			name:   "different format",
			a:      []string{"x", "y"},
			b:      []string{"foo", "bar"},
			fmtStr: "%s(%s)",
			want:   []string{"x(foo)", "y(bar)"},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := zip(tt.a, tt.b, tt.fmtStr)
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("zip() = %v, want %v", got, tt.want)
				}
			})
	}

	// Test panic case
	t.Run(
		"unequal lengths", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("zip() did not panic with unequal length inputs")
				}
			}()
			zip([]string{"a", "b"}, []string{"1"}, "%s:%s")
		})
}

func TestJoinl(t *testing.T) {
	tests := []struct {
		name string
		sep  string
		a    []string
		want string
	}{
		{
			name: "comma separator",
			sep:  ", ",
			a:    []string{"a", "b", "c"},
			want: "a, b, c",
		},
		{
			name: "empty list",
			sep:  "-",
			a:    []string{},
			want: "",
		},
		{
			name: "single item",
			sep:  ":",
			a:    []string{"solo"},
			want: "solo",
		},
		{
			name: "newline separator",
			sep:  "\n",
			a:    []string{"line1", "line2", "line3"},
			want: "line1\nline2\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := joinl(tt.sep, tt.a)
				if got != tt.want {
					t.Errorf("joinl() = %v, want %v", got, tt.want)
				}
			})
	}
}
