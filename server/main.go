package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"widiff/assert"
	"widiff/broker"
	_ "widiff/db"
	"widiff/feed"
)

func main() {
	// db.TestDb()
	logFile, err := os.OpenFile("assert.log", os.O_WRONLY|os.O_CREATE, 0644)
	if errors.Is(err, os.ErrNotExist) {
		logFile, err = os.Create("assert.log")
	}
	defer logFile.Close()
	if err != nil {
		panic(err)
	}
	assert.ToWriter(logFile)

	// feed := feed.New(
	// 	time.Duration(30*time.Second),
	// 	time.Duration(30*time.Second),
	// )

	wikiFeed := feed.Test()
	broker := broker.New[feed.Data]()
	go broker.Start()

	var init feed.Data

	go func() {
		for {
			select {
			case feedUpate := <-wikiFeed.Pull():
				init = feedUpate
				broker.Publish(feedUpate)
			}
		}
	}()

	serveMux := http.NewServeMux()

	serveMux.Handle("/view/",
		http.StripPrefix("/view/", http.FileServer(http.Dir("./static"))),
	)

	serveMux.HandleFunc("/diff",
		func(w http.ResponseWriter, r *http.Request) {
			var b bytes.Buffer
			err := init.ToJson(&b)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(b.Bytes())
		})

	serveMux.HandleFunc("/notify",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			ctx := r.Context()
			msgCh := broker.Subscribe()

			closeHandler := make(chan struct{})

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
					case update := <-msgCh:
						var b bytes.Buffer
						err := update.ToJson(&b)
						assert.NoError(err, "encoding error", update)
						fmt.Fprintf(w, "data:  %s\n\n", b.Bytes())
						flusher.Flush()
					case <-ctx.Done():
						log.Printf("client disconnect\n")
						broker.Unsubscribe(msgCh)
						close(closeHandler)
						return
					}
				}
			}()
			<-closeHandler
			log.Printf("closing client")
		})

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	if err := http.ListenAndServe(":8080", serveMux); err != nil {
		broker.Stop()
		wikiFeed.Stop()
		log.Fatal(err)
	}
}
