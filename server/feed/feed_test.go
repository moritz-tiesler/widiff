package feed

import (
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
	f := New(&testWikiApi{counter: 100}, 100*time.Millisecond, gem.Test())

	for range 1 * 60 {
		newTopDiff, _ := f.FetchDiff()
		buffs.Update(newTopDiff)
	}

	f = New(&testWikiApi{}, 100*time.Millisecond, gem.Test())
	for range 1 * 60 {
		newTopDiff, _ := f.FetchDiff()
		buffs.Update(newTopDiff)
	}

	actual := buffs.Report()
	expected := Data{
		Minute: wiki_api.Diff{Size: 60},
		Hour:   wiki_api.Diff{Size: 60},
		Day:    wiki_api.Diff{Size: 160},
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("wrong report data, expected=%+v, got=%+v", expected, actual)
	}
}
