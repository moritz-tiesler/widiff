package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"
	"widiff/assert"
	_ "widiff/db"
	wiki "widiff/wiki"
	wikiapi "widiff/wiki_api"
)

var tick = 10 * time.Second

func main() {
	// db.TestDb()
	f, err := os.OpenFile("assert.log", os.O_WRONLY|os.O_CREATE, 0644)
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create("assert.log")
	}
	defer f.Close()
	if err != nil {
		panic(err)
	}
	assert.ToWriter(f)

	topDiff := ""

	go func() {
		startingFrom := time.Now().Add(-1 * time.Minute)
		newTopDiff, err := wikiapi.TopDiff(startingFrom)
		if err != nil {
			log.Printf("failed to initialize diff: %s", err)
		}
		topDiff = newTopDiff

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
				topDiff = newTopDiff
			}
		}
	}()

	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/diff", func(w http.ResponseWriter, r *http.Request) {
		var b bytes.Buffer
		diffRes := wiki.DiffResponse{Minute: topDiff}
		err := json.NewEncoder(&b).Encode(diffRes)
		assert.NoError(err, "encoding error", diffRes)
		w.Write(b.Bytes())
	})
	serveMux.Handle("/view/", http.StripPrefix("/view/", http.FileServer(http.Dir("./static"))))

	http.ListenAndServe(":8080", serveMux)
}
