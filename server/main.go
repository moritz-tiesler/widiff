package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
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
	User       string `json:"user"`
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
			Minute: Diff{
				DiffString: feedResult.Minute.DiffString,
				Comment:    feedResult.Minute.Comment,
				User:       feedResult.Minute.User,
			},
			Hour: Diff{
				DiffString: feedResult.Hour.DiffString,
				Comment:    feedResult.Hour.Comment,
				User:       feedResult.Hour.User,
			},
			Day: Diff{
				DiffString: feedResult.Day.DiffString,
				Comment:    feedResult.Day.Comment,
				User:       feedResult.Day.User,
			},
		}
		err := json.NewEncoder(&b).Encode(diffs)
		assert.NoError(err, "encoding error", diffs)
		w.Write(b.Bytes())
	})

	serveMux.Handle("/view/", http.StripPrefix("/view/", http.FileServer(http.Dir("./static"))))

	id := 0

	serveMux.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		id++
		rid := id
		ctx := r.Context()

		close := make(chan struct{})

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Create a flusher
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					log.Printf("messaging client id=%d\n", rid)
					fmt.Fprintf(w, "data: Message %d\n\n", rid)
					flusher.Flush()
				case <-ctx.Done():
					log.Printf("client disconnect id=%d\n", rid)
					close <- struct{}{}
					return

				}
			}
		}()
		<-close
		log.Printf("closing client=%d", rid)
	})

	if err := http.ListenAndServe(":8080", serveMux); err != nil {
		log.Fatal(err)
	}
}
