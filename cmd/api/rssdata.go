package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/blue-davinci/aggregate/internal/validator"
)

// startRssFeedScraperHandler() Is the entry point of our scraper function
// It Uses noofroutines and fetchinterval settings from our config then
// Proceeds to get the feeds to fetch, summoning the Main scraper.
func (app *application) startRssFeedScraperHandler() {
	goroutines := app.config.scraper.noofroutines
	interval := app.config.scraper.fetchinterval
	app.logger.PrintInfo("Starting RSS Feed Scraper", map[string]string{
		"No of Go Routines": fmt.Sprintf("%d", goroutines),
		"Interval":          fmt.Sprintf("%ds", interval),
		"No of Retries":     fmt.Sprintf("%d", app.config.scraper.scraperclient.retrymax),
		"Client Timeout":    fmt.Sprintf("%d", app.config.scraper.scraperclient.timeout),
	})
	// start the scraper
	// convert the interval to seconds
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	for ; ; <-ticker.C {
		feeds, err := app.models.RSSFeedData.GetNextFeedsToFetch(goroutines, interval)
		// if we get an error, we log it and continue wuth our work
		if err != nil {
			app.logger.PrintError(err, map[string]string{
				"Error Getting Feeds From DB": "GetNextFeedsToFetch",
			})
			continue
		}

		// For each particular feed, we pass the data to our main Scraping
		// function which launches  seperate go routines for the work.
		app.logger.PrintInfo("Starting scraping workers", map[string]string{
			"Executing workers": fmt.Sprintf("Getting %d feeds", len(feeds)),
		})
		for _, feed := range feeds {
			app.rssFeedScraper(feed)
		}
	}

}

// rssFeedScraper() is the main method which performs scraping for each
// individual feed. It takes in an indvidiual Feed, updates its last fetched
// using MarkFeedAsFetched() and then saved the data to our DB
func (app *application) rssFeedScraper(feed database.Feed) {
	// we want to fetch each of the feeds concurrently, so we make a wait group
	// using our app.background(func(){}) through a for loop to iterate over the feeds starting a routine for each feed
	app.background(func() {
		// get the feed data
		err := app.models.RSSFeedData.MarkFeedAsFetched(feed.ID)
		if err != nil {
			app.logger.PrintError(err, map[string]string{
				"Error Marking Feed As Fetched": "MarkFeedAsFetched",
				"Feed Name":                     feed.Name,
				"Feed ID":                       feed.ID.String(),
			})
			return
		}
		// call our GetRSSFeeds to return all feeds for each specific URL
		rssFeeds, err := app.models.RSSFeedData.GetRSSFeeds(
			app.config.scraper.scraperclient.retrymax,
			app.config.scraper.scraperclient.timeout,
			feed.Url)
		if err != nil {
			switch {
			case err == data.ErrContextDeadline:
				app.logger.PrintInfo(err.Error(), map[string]string{
					"Feed": feed.Name,
					"URL":  feed.Url,
				})
			case err == data.ErrUnableToDetectFeedType:
				app.logger.PrintInfo(err.Error(), map[string]string{
					"Feed": feed.Name,
					"URL":  feed.Url,
				})
			default:
				app.logger.PrintError(err, map[string]string{
					"Error Fetching RSS Feeds": "GetRSSFeeds",
					"URL":                      feed.Url,
				})
			}
			if err == data.ErrContextDeadline {

				return
			}
		}
		// store the fetched data into our DB
		err = app.models.RSSFeedData.CreateRssFeedPost(&rssFeeds, &feed.ID)
		if err != nil {
			app.logger.PrintError(err, map[string]string{
				"Error Creating Rss Feed Post": "CreateRssFeedPost",
				"Feed Name":                    feed.Name})
			return
		}

		/*app.logger.PrintInfo("Finished collecting feeds for: ", map[string]string{
			"Name":   feed.Name,
			"Posts:": fmt.Sprintf("%d", len(rssFeeds.Channel.Item)),
		})*/
	})
}

// Handler for out GetAllPost
func (app *application) GetFollowedRssPostsForUserHandler(w http.ResponseWriter, r *http.Request) {
	// make a struct to hold what we would want from the queries
	var input struct {
		data.Filters
	}
	//validate if queries are provided
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	//get the page & pagesizes as ints and set to the embedded struct
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// We don't use any sort for this endpoint
	input.Filters.Sort = app.readString(qs, "", "")
	// None of the sort values are supported for this endpoint
	input.Filters.SortSafelist = []string{"", ""}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// We are good, now we call our models getposts to get the rss posts
	userRssFollowedPosts, metadata, err := app.models.RSSFeedData.GetFollowedRssPostsForUser(app.contextGetUser(r).ID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Return the feeds in the response body
	err = app.writeJSON(w, http.StatusOK, envelope{"followed_rss_posts": userRssFollowedPosts, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) GetRSSFavoritePostsForUserHandler(w http.ResponseWriter, r *http.Request) {
	// make a struct to hold what we would want from the queries
	var input struct {
		data.Filters
	}
	//validate if queries are provided
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	//get the page & pagesizes as ints and set to the embedded struct
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// We don't use any sort for this endpoint
	input.Filters.Sort = app.readString(qs, "", "")
	// None of the sort values are supported for this endpoint
	input.Filters.SortSafelist = []string{"", ""}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// We are good, now we call our models getposts to get the rss posts
	userRssFavoritePosts, err := app.models.RSSFeedData.GetRSSFavoritePostsForUser(app.contextGetUser(r).ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Return the feeds in the response body
	err = app.writeJSON(w, http.StatusOK, envelope{"favorite_rss_posts": userRssFavoritePosts}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
