package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mattn/go-gntp"
	"net/http"
	"log"
	"os"
	"time"
)

var server = flag.String("gntp", "127.0.0.1:23053", "The GNTP DSN")

type Build struct {
	RepositoryId int    `json:"repository_id"`
	EventType    string `json:"event_type"`
	FinishedAt   string `json:"finished_at"`
	Number       string `json:"number"`
	State        string `json:"state"`
	Result       int    `json:"result"`
	Branch       string `json:"branch"`
	Duration     int    `json:"duration"`
	Commit       string `json:"commit"`
	Message      string `json:"message"`
	StartedAt    string `json:"started_at"`
	Id           int    `json:"id`
}

func main() {
	flag.Parse()
	idmap := map[int]bool{}


	growl := gntp.NewClient()
	growl.Server = *server
	growl.AppName = "travis notify"
	growl.Register([]gntp.Notification{
		gntp.Notification{
			Event:   "success",
			Enabled: false,
		}, gntp.Notification{
			Event:   "failed",
			Enabled: true,
		},
	})

	for _, proj := range os.Args {
		go func(proj string) {
			first := true
			for {
				r, err := http.Get(fmt.Sprintf("https://api.travis-ci.org/repositories/%s/builds.json", proj))
				if err != nil {
					log.Println(err)
					time.Sleep(30 * time.Second)
					continue
				}
				defer r.Body.Close()

				var builds []Build
				json.NewDecoder(r.Body).Decode(&builds)

				for _, build := range builds {
					if _, ok := idmap[build.Id]; ok {
						continue
					}
					if build.State != "finished" {
						continue
					}
					idmap[build.Id] = true

					if first {
						continue
					}

					event := "success"
					text := "Congratulations!"
					if build.Result != 0 {
						event = "failed"
						text = "So bad!"
					}

					growl.Notify(&gntp.Message{
						Event:    event,
						Title:    proj,
						Text:     text,
						Callback: fmt.Sprintf("https://travis-ci.org/%s/jobs/%d", proj, build.Id),
						//Icon:     icon(event),
					})
				}
				first = false

				time.Sleep(30 * time.Second)
			}
		}(proj)
	}

	select {}
}
