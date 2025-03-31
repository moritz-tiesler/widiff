package feed

import (
	"cmp"
	"encoding/json"
	"io"
	"log"
	"slices"
	"sync"
	"time"
	"widiff/assert"
	wikiapi "widiff/wiki_api"
)

type Buffers struct {
	Minute Buffer[wikiapi.Diff]
	Hour   Buffer[wikiapi.Diff]
	Day    Buffer[wikiapi.Diff]
}

func NewBuffers() *Buffers {
	return &Buffers{
		Minute: NewBuffer[wikiapi.Diff](1),
		Hour:   NewBuffer[wikiapi.Diff](60),
		Day:    NewBuffer[wikiapi.Diff](1440),
	}
}

func (bs *Buffers) Report() Data {
	maxMinute := maxDiff(bs.Minute.Items()...)
	maxHour := maxDiff(bs.Hour.Items()...)
	maxDay := maxDiff(bs.Day.Items()...)

	assert.Assert(maxMinute.Size <= maxHour.Size, "minutely greater than hourly")
	assert.Assert(maxHour.Size <= maxDay.Size, "hourly greater than daily")

	return Data{
		Minute: maxMinute,
		Hour:   maxHour,
		Day:    maxDay,
	}
}

func (bs *Buffers) Update(diff wikiapi.Diff) {
	bs.Minute.Write(diff)
	bs.Hour.Write(diff)
	bs.Day.Write(diff)
}

type Data struct {
	Minute wikiapi.Diff
	Hour   wikiapi.Diff
	Day    wikiapi.Diff
}

func (d Data) ToJson(w io.Writer) error {
	diffs := Diffs{
		Minute: Diff{
			DiffString: d.Minute.DiffString,
			Comment:    d.Minute.Comment,
			User:       d.Minute.User,
		},
		Hour: Diff{
			DiffString: d.Hour.DiffString,
			Comment:    d.Hour.Comment,
			User:       d.Hour.User,
		},
		Day: Diff{
			DiffString: d.Day.DiffString,
			Comment:    d.Day.Comment,
			User:       d.Day.User,
		},
	}
	err := json.NewEncoder(w).Encode(diffs)
	return err
}

type Diffs struct {
	Minute Diff `json:"minute"`
	Hour   Diff `json:"hour"`
	Day    Diff `json:"day"`
}

type Diff struct {
	DiffString string `json:"diffstring"`
	Comment    string `json:"comment"`
	User       string `json:"user"`
}

var tick = 60 * time.Second

type Feed struct {
	push    chan Data
	current Data
	sync.RWMutex
}

func (f *Feed) Pull() chan Data {
	return f.push
}

func New(updateEvery, notifyEver time.Duration) *Feed {
	f := &Feed{}
	f.initStream(updateEvery)
	f.push = make(chan Data, 1)
	return f
}

func Test() *Feed {
	feed := New(
		time.Duration(10*time.Second),
		time.Duration(5*time.Second),
	)
	return feed
}

func (f *Feed) initStream(interval time.Duration) {
	buffs := NewBuffers()

	ticker := time.NewTicker(interval)
	go func() {
		// populate feed with initial value
		startingFrom := time.Now().Add(-1 * time.Minute).Add(-10 * time.Second)
		newTopDiff, err := wikiapi.TopDiff(startingFrom)
		if err != nil {
			log.Printf("failed to initialize diff: %s", err)
		}
		buffs.Update(newTopDiff)
		feedUpdate := buffs.Report()
		f.push <- feedUpdate
		// push new data periodically
		for {
			select {
			case <-ticker.C:
				startingFrom := time.Now().Add(-1 * time.Minute).Add(-10 * time.Second)
				newTopDiff, err = wikiapi.TopDiff(startingFrom)
				if err != nil {
					break
				}
				buffs.Update(newTopDiff)
				feedUpdate := buffs.Report()
				f.push <- feedUpdate
			}
		}
	}()
}

func maxDiff(diffs ...wikiapi.Diff) wikiapi.Diff {
	longest := slices.MaxFunc(diffs, func(a, b wikiapi.Diff) int {
		return cmp.Compare(a.Size, b.Size)
	})
	return longest
}

func MultTime(
	duration time.Duration,
	mult float64,
) time.Duration {
	durationFloat := float64(duration.Nanoseconds())

	newDurationNanoseconds := durationFloat * mult

	newDuration := time.Duration(newDurationNanoseconds)

	return newDuration
}
