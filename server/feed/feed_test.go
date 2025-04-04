package feed

import (
	"fmt"
	"reflect"
	"testing"
	"time"
	"widiff/gem"
	"widiff/wiki_api"
)

type testWikiApi struct {
	counter int
}

func (twa *testWikiApi) TopDiff(s time.Time) (wiki_api.Diff, error) {
	twa.counter++
	return wiki_api.Diff{Size: twa.counter}, nil
}

func TestMaxValues(t *testing.T) {
	buffs := NewBuffers()

	baseLine := 240
	for range 24 {
		f := New(&testWikiApi{counter: baseLine}, 100*time.Millisecond, gem.Test())
		for range 1 * 60 {
			newTopDiff, _ := f.fetchDiff()
			buffs.Update(newTopDiff)
		}
		baseLine -= 10
	}

	actual := buffs.Report()
	expected := Data{
		Minute: wiki_api.Diff{Size: 60},
		Hour:   wiki_api.Diff{Size: 60},
		Day:    wiki_api.Diff{Size: 160},
	}

	fmt.Printf("Minute:\n, %v\n", buffs.Minute.Items())
	fmt.Printf("Hour:\n, %v\n", buffs.Hour.Items())
	fmt.Printf("Day:\n, %v\n", buffs.Day.Items())

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("wrong report data, expected=%+v, got=%+v", expected, actual)
	}
}
