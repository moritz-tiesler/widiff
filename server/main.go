package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"widiff/assert"
	_ "widiff/db"
	"widiff/feed"
	wiki "widiff/wiki"
)

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

	diffFeed := feed.New()

	var feedResult feed.Data
	go func() {
		for {
			select {
			case feedResult = <-diffFeed:
			}
		}
	}()

	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/diff", func(w http.ResponseWriter, r *http.Request) {
		var b bytes.Buffer
		diffResponse := wiki.DiffResponse{
			Minute: feedResult.Minute,
			Hour:   feedResult.Hour,
			Day:    feedResult.Day,
		}
		err := json.NewEncoder(&b).Encode(diffResponse)
		assert.NoError(err, "encoding error", diffResponse)
		w.Write(b.Bytes())
	})
	serveMux.Handle("/view/", http.StripPrefix("/view/", http.FileServer(http.Dir("./static"))))

	http.ListenAndServe(":8080", serveMux)
}
