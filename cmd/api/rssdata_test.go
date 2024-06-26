package main

import (
	"fmt"
	"testing"

	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/google/uuid"
)

func Test_application_rssFeedScraper(t *testing.T) {
	type args struct {
		feed database.Feed
	}
	tests := []struct {
		name string
		app  *application
		args args
	}{
		// TODO: Add test cases.
		{"Test1", &application{}, args{database.Feed{ID: uuid.New(), Url: "http://newsrss.bbc.co.uk/rss/newsonline_world_edition/help/rss/rss.xml", Name: "bbc"}}},
		{"Test2", &application{}, args{database.Feed{ID: uuid.New(), Url: "https://wagslane.dev/index.xml", Name: "Lane's"}}},
		{"Test3", &application{}, args{database.Feed{ID: uuid.New(), Url: "https://feeds.megaphone.fm/newheights", Name: "Megaphone"}}},
		{"Test4", &application{}, args{database.Feed{ID: uuid.New(), Url: "http://rss.art19.com/the-daily", Name: "Daily Podcast"}}},
		{"Test5", &application{}, args{database.Feed{ID: uuid.New(), Url: "https://www.engadget.com/rss.xml", Name: "Endagadget"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("Fetching: ", tt.args.feed.Name)
			tt.app.rssFeedScraper(tt.args.feed)
		})
	}
}
