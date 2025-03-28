package feed

import (
	"log"
	"time"
	_ "widiff/wiki"
	wikiapi "widiff/wiki_api"
)

type Data struct {
	Minute string
	Hour   string
	Day    string
}

var tick = 60 * time.Second

func New() <-chan Data {
	readChan := make(chan Data)

	go func() {
		startingFrom := time.Now().Add(-1 * time.Minute)
		newTopDiff, err := wikiapi.TopDiff(startingFrom)
		if err != nil {
			log.Printf("failed to initialize diff: %s", err)
		}
		readChan <- Data{Minute: newTopDiff, Hour: newTopDiff, Day: newTopDiff}

		ticker := time.NewTicker(tick)
		for {
			select {
			case <-ticker.C:
				startingFrom = time.Now().Add(-1 * time.Minute)
				newTopDiff, err = wikiapi.TopDiff(startingFrom)
				if err != nil {
					break
				}
				log.Println("top diff updated")
				readChan <- Data{Minute: newTopDiff, Hour: newTopDiff, Day: newTopDiff}
			}
		}
	}()
	return readChan
}
