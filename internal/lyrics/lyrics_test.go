package lyrics

import (
	"reflect"
	"testing"
)

func TestParseLRCSortsTimestampedLines(t *testing.T) {
	lines := parseLRC("[00:10.50] Later\n[00:01.25] First\nnot a lyric\n[01:02.03] Last")

	want := []Line{
		{Time: 1.25, Text: "First"},
		{Time: 10.5, Text: "Later"},
		{Time: 62.03, Text: "Last"},
	}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("parseLRC() = %#v, want %#v", lines, want)
	}
}

func TestParsePlainTrimsLines(t *testing.T) {
	lines := parsePlain(" first line \n\nsecond line")

	want := []Line{
		{Text: "first line"},
		{Text: ""},
		{Text: "second line"},
	}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("parsePlain() = %#v, want %#v", lines, want)
	}
}

func TestIsSynced(t *testing.T) {
	if IsSynced(nil) {
		t.Fatal("nil lyrics should not be synced")
	}
	if IsSynced([]Line{{Text: "plain"}}) {
		t.Fatal("plain lyrics should not be synced")
	}
	if !IsSynced([]Line{{Time: 0, Text: "intro"}, {Time: 2.5, Text: "line"}}) {
		t.Fatal("timestamped lyrics should be synced")
	}
}

func TestActiveIndex(t *testing.T) {
	lines := []Line{
		{Time: 1, Text: "one"},
		{Time: 4, Text: "two"},
		{Time: 8, Text: "three"},
	}

	tests := []struct {
		name string
		pos  float64
		want int
	}{
		{name: "before first", pos: 0, want: 0},
		{name: "at first", pos: 1, want: 0},
		{name: "between lines", pos: 6, want: 1},
		{name: "after last", pos: 12, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ActiveIndex(lines, tt.pos); got != tt.want {
				t.Fatalf("ActiveIndex() = %d, want %d", got, tt.want)
			}
		})
	}
}
