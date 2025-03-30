package main

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"widiff/assert"
	_ "widiff/db"
	"widiff/feed"
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

	// feed := feed.New(
	// 	time.Duration(30*time.Second),
	// 	time.Duration(30*time.Second),
	// )

	feed := feed.Test()

	serveMux := http.NewServeMux()

	serveMux.Handle("/view/",
		http.StripPrefix("/view/", http.FileServer(http.Dir("./static"))),
	)

	serveMux.HandleFunc("/diff",
		func(w http.ResponseWriter, r *http.Request) {
			initial := feed.Top()
			var b bytes.Buffer
			err := initial.ToJson(&b)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(b.Bytes())
		})

	serveMux.HandleFunc("/notify", feed.Notify)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	if err := http.ListenAndServe(":8080", serveMux); err != nil {
		log.Fatal(err)
	}
}
