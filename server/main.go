package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"widiff/assert"
	_ "widiff/db"
	"widiff/feed"
)

type Diffs struct {
	Minute Diff `json:"minute"`
	Hour   Diff `json:"hour"`
	Day    Diff `json:"day"`
}

type Diff struct {
	DiffString string `json:"diffstring"`
	Comment    string `json:"comment"`
}

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
		diffs := Diffs{
			Minute: Diff{DiffString: feedResult.Minute.DiffString, Comment: feedResult.Minute.Comment},
			Hour:   Diff{DiffString: feedResult.Hour.DiffString, Comment: feedResult.Hour.Comment},
			Day:    Diff{DiffString: feedResult.Day.DiffString, Comment: feedResult.Day.Comment},
		}
		err := json.NewEncoder(&b).Encode(diffs)
		assert.NoError(err, "encoding error", diffs)
		w.Write(b.Bytes())
	})
	serveMux.Handle("/view/", http.StripPrefix("/view/", http.FileServer(http.Dir("./static"))))

	if err := http.ListenAndServe(":8080", serveMux); err != nil {
		log.Fatal(err)
	}
}
