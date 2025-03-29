package feed

import (
	"cmp"
	"log"
	"slices"
	"time"
	"widiff/assert"
	_ "widiff/wiki"
	wikiapi "widiff/wiki_api"
)

type Buffers struct {
	Minute RingBuffer[wikiapi.Diff]
	Hour   RingBuffer[wikiapi.Diff]
	Day    RingBuffer[wikiapi.Diff]
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

func InitData(d wikiapi.Diff) Data {
	return Data{
		Minute: d,
		Hour:   d,
		Day:    d,
	}
}

var tick = 60 * time.Second

func New() <-chan Data {
	buffs := NewBuffers()
	readChan := make(chan Data)

	go func() {
		// Todo: look a bit further back (1 Minute + 5 seconds)

		startingFrom := time.Now().Add(-1 * time.Minute)
		newTopDiff, err := wikiapi.TopDiff(startingFrom)
		if err != nil {
			log.Printf("failed to initialize diff: %s", err)
		}
		log.Println("top diff updated")
		buffs.Update(newTopDiff)
		readChan <- buffs.Report()

		ticker := time.NewTicker(tick)
		for {
			select {
			case <-ticker.C:
				startingFrom = time.Now().Add(-1 * time.Minute)
				newTopDiff, err = wikiapi.TopDiff(startingFrom)
				if err != nil {
					break
				}
				buffs.Update(newTopDiff)
				log.Println("top diff updated")
				readChan <- buffs.Report()
			}
		}
	}()
	return readChan
}

func maxDiff(diffs ...wikiapi.Diff) wikiapi.Diff {
	longest := slices.MaxFunc(diffs, func(a, b wikiapi.Diff) int {
		return cmp.Compare(a.Size, b.Size)
	})
	return longest
}
