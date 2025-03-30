package feed

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
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

func (d Data) ToJson() ([]byte, error) {
	var b bytes.Buffer
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
	err := json.NewEncoder(&b).Encode(diffs)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), err
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
	current      Data
	notifyTicker time.Ticker
}

func New(updateEvery, notifyEver time.Duration) *Feed {
	f := &Feed{}
	f.initStream(updateEvery)
	f.notifyTicker = *time.NewTicker(notifyEver)
	return f
}

func Test() *Feed {
	feed := New(
		time.Duration(10*time.Second),
		time.Duration(5*time.Second),
	)
	return feed
}

func fork[T any](in <-chan T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for v := range in {
			out <- v
		}
	}()
	return out
}

func (f *Feed) Top() Data {
	return f.current
}

func (f *Feed) Notify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ticker := fork(f.notifyTicker.C)

	ctx := r.Context()
	close := make(chan struct{})

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!",
			http.StatusInternalServerError,
		)
		return
	}

	go func() {
		for {
			select {
			case <-ticker:
				feedResult := f.Top()
				json, err := feedResult.ToJson()
				assert.NoError(err, "encoding error", feedResult)
				fmt.Fprintf(w, "data:  %s\n\n", json)
				flusher.Flush()
			case <-ctx.Done():
				log.Printf("client disconnect\n")
				close <- struct{}{}
				return
			}
		}
	}()
	<-close
	log.Printf("closing client")
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
		log.Println("top diff updated")
		buffs.Update(newTopDiff)
		feedUpdate := buffs.Report()
		f.current = feedUpdate

		// push new data periodically
		for {
			select {
			case <-ticker.C:
				startingFrom := time.Now().Add(-1 * time.Minute).Add(-10 * time.Second)
				newTopDiff, err = wikiapi.TopDiff(startingFrom)
				newTopDiff, err = wikiapi.TopDiff(startingFrom)
				if err != nil {
					break
				}
				buffs.Update(newTopDiff)
				log.Println("top diff updated")
				feedUpdate := buffs.Report()
				f.current = feedUpdate
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
