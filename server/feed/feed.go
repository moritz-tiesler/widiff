package feed

import (
	"cmp"
	"log"
	"slices"
	"time"
	_ "widiff/wiki"
	wikiapi "widiff/wiki_api"
)

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
	readChan := make(chan Data)

	fd := Data{}
	go func() {
		startingFrom := time.Now().Add(-1 * time.Minute)
		newTopDiff, err := wikiapi.TopDiff(startingFrom)
		if err != nil {
			log.Printf("failed to initialize diff: %s", err)
		}
		log.Println("top diff updated")
		fd = InitData(newTopDiff)
		readChan <- fd

		ticker := time.NewTicker(tick)
		for {
			select {
			case <-ticker.C:
				startingFrom = time.Now().Add(-1 * time.Minute)
				newTopDiff, err = wikiapi.TopDiff(startingFrom)
				if err != nil {
					break
				}
				fd = updateFeedData(fd, newTopDiff)
				log.Println("top diff updated")
				readChan <- fd
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

func updateFeedData(old Data, diff wikiapi.Diff) Data {
	var updated Data
	updated.Minute = diff
	updated.Hour = maxDiff(old.Hour, diff)
	updated.Day = maxDiff(old.Day, diff)

	return updated
}
