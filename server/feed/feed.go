package feed

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"time"
	"widiff/assert"
	"widiff/gem"
	wikiapi "widiff/wiki_api"
)

type WikiSource interface {
	TopDiff(startingFrom time.Time) (wikiapi.Diff, error)
}

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
			Review:     d.Minute.Review,
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
	Review     string `json:"review"`
}

type Feed struct {
	Source    WikiSource
	push      chan Data
	stop      chan struct{}
	generator gem.Generator
}

func (f *Feed) Pull() chan Data {
	return f.push
}

func New(source WikiSource, updateEvery time.Duration, generator gem.Generator) *Feed {
	f := &Feed{Source: source}
	f.initStream(updateEvery)
	f.push = make(chan Data, 1)
	f.generator = generator
	return f
}

// TODO: implement corresponding Start method
func (f *Feed) Stop() {
	close(f.stop)
}

func Test(source WikiSource) *Feed {
	feed := New(
		source,
		time.Duration(10*time.Second),
		gem.Test(),
	)
	return feed
}

func (f *Feed) judgeDiff(diff wikiapi.Diff) (string, error) {
	prompt := buildPrompt(diff.DiffString, diff.Comment)
	judged, err := f.generator.Generate(prompt)
	return judged, err
}

// TODO: pass ctx to wiki requests
func (f *Feed) initStream(interval time.Duration) {
	buffs := NewBuffers()
	ticker := time.NewTicker(interval)
	go func() {
		// populate feed with initial value
		newTopDiff, _ := f.FetchDiff()
		buffs.Update(newTopDiff)
		feedUpdate := buffs.Report()
		review, err := f.judgeDiff(feedUpdate.Minute)
		if err == nil {
			feedUpdate.Minute.Review = review
		}
		f.push <- feedUpdate
		for {
			select {
			case <-f.stop:
				return
			// push new data periodically
			case <-ticker.C:
				newTopDiff, err := f.FetchDiff()
				if err != nil {
					continue
				}
				// todo judge diff before buffs.Upadate to have review
				// also for hourly and daily top
				buffs.Update(newTopDiff)
				feedUpdate := buffs.Report()
				review, err := f.judgeDiff(feedUpdate.Minute)
				if err == nil {
					feedUpdate.Minute.Review = review
				}
				f.push <- feedUpdate
			}
		}
	}()
}

func (f *Feed) FetchDiff() (wikiapi.Diff, error) {
	startingFrom := time.Now().Add(-1 * time.Minute).Add(-10 * time.Second)
	newTopDiff, err := f.Source.TopDiff(startingFrom)
	return newTopDiff, err
}

func maxDiff(diffs ...wikiapi.Diff) wikiapi.Diff {
	longest := slices.MaxFunc(diffs, func(a, b wikiapi.Diff) int {
		return cmp.Compare(a.Size, b.Size)
	})
	return longest
}

func buildPrompt(diff, comment string) string {
	return diff + fmt.Sprintf("\ncomment: %s", comment)
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
