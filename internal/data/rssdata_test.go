package data

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/atom"
)

// mockReadCloser creates a ReadCloser from a string for testing purposes.
func mockReadCloser(s string) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBufferString(s))
}

func TestRssFeedDecoder(t *testing.T) {
	rssTestData := `
	<rss>
	<channel>
	<title>Lane's Blog</title>
	<link>https://wagslane.dev/</link>
	<description>Recent content on Lane's Blog</description>
	<generator>Hugo -- gohugo.io</generator>
	<language>en-us</language>
	<lastBuildDate>Sun, 08 Jan 2023 00:00:00 +0000</lastBuildDate>
	<atom:link href="https://wagslane.dev/index.xml" rel="self" type="application/rss+xml"/>
	<item>
	<title>The Zen of Proverbs</title>
	<link>https://wagslane.dev/posts/zen-of-proverbs/</link>
	<pubDate>Sun, 08 Jan 2023 00:00:00 +0000</pubDate>
	<guid>https://wagslane.dev/posts/zen-of-proverbs/</guid>
	<description>
	20 rules of thumb for writing better software.
	</description>
	</item>
	</channel>
	</rss>
	`

	atomTestData := `
	<feed xmlns="http://www.w3.org/2005/Atom">
	<title>Example Atom Feed</title>
	<link href="http://example.org/"/>
	<updated>2003-12-13T18:30:02Z</updated>
	<author>
	<name>John Doe</name>
	</author>
	<id>urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6</id>
	<entry>
	<title>Atom-Powered Robots Run Amok</title>
	<link href="http://example.org/2003/12/13/atom03"/>
	<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
	<updated>2003-12-13T18:30:02Z</updated>
	<summary>Some text.</summary>
	</entry>
	</feed>
	`

	// Create mock HTTP responses for RSS and Atom feeds
	rssResp := &http.Response{
		Body: mockReadCloser(rssTestData),
	}
	atomResp := &http.Response{
		Body: mockReadCloser(atomTestData),
	}

	type args struct {
		rssFeed   *RSSFeed
		resp      *http.Response
		url       string
		sanitizer *bluemonday.Policy
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Valid RSS Feed",
			args:    args{&RSSFeed{}, rssResp, "https://www.example.com/rss.xml", bluemonday.UGCPolicy()},
			want:    "Lane&#39;s Blog",
			wantErr: false,
		},
		{
			name:    "Valid Atom Feed",
			args:    args{&RSSFeed{}, atomResp, "https://www.example.com/atom.xml", bluemonday.UGCPolicy()},
			want:    "Example Atom Feed",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RssFeedDecoderDecider(tt.args.url, tt.args.rssFeed, tt.args.sanitizer, tt.args.resp); (err != nil) != tt.wantErr {
				t.Errorf("RssFeedDecoder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.args.rssFeed.Channel.Title != tt.want {
				t.Errorf("Atom Feed Title = %v, want %v", tt.args.rssFeed.Channel.Title, tt.want)
			}
		})
	}
}

// Test function for converting gofeed.Feed to RSSFeed with sanitization
func Test_convertGofeedToRSSFeed(t *testing.T) {
	// Example feed data to be used in the tests
	feedWithBasicInfo := &gofeed.Feed{
		Title:       "Test Blog",
		Link:        "https://testblog.com",
		Description: "This is a test blog feed.",
		Items: []*gofeed.Item{
			{
				Title:       "First Post",
				Link:        "https://testblog.com/first-post",
				Description: "<p>This is the first post.</p>",
				Content:     "This is the content of the first post.",
			},
		},
	}

	emptyFeed := &gofeed.Feed{}

	feedWithSpecialCharacters := &gofeed.Feed{
		Title:       "Special <Feed>",
		Link:        "https://specialfeed.com",
		Description: "This feed contains <b>HTML</b> & special characters!",
		Items: []*gofeed.Item{
			{
				Title:       "Special & Unique Post",
				Link:        "https://specialfeed.com/post",
				Description: "<p>This post contains & symbols.</p>",
				Content:     "<p>This content has <b>bold</b> and <i>italic</i> text.</p>",
			},
		},
	}

	feedWithXSSAttempt := &gofeed.Feed{
		Title:       "XSS Test Feed",
		Link:        "https://xsstest.com",
		Description: "This feed contains an XSS attempt!",
		Items: []*gofeed.Item{
			{
				Title:       "XSS Post",
				Link:        "https://xsstest.com/post",
				Description: `<p>This post tries to execute <script>alert('XSS')</script></p>`,
				Content:     `<p>This content tries to execute <img src="x" onerror="alert('XSS')"></p>`,
			},
		},
	}

	// Initialize the sanitizer policy
	sanitizer := bluemonday.UGCPolicy()

	tests := []struct {
		name            string
		feed            *gofeed.Feed
		wantTitle       string
		wantItemTitle   string
		wantItemContent string
	}{
		{
			name:            "Basic RSS Feed Conversion",
			feed:            feedWithBasicInfo,
			wantTitle:       "Test Blog",
			wantItemTitle:   "First Post",
			wantItemContent: "This is the content of the first post.",
		},
		{
			name:            "Empty Feed",
			feed:            emptyFeed,
			wantTitle:       "",
			wantItemTitle:   "",
			wantItemContent: "",
		},
		{
			name:            "Feed with Special Characters",
			feed:            feedWithSpecialCharacters,
			wantTitle:       "Special ",
			wantItemTitle:   "Special &amp; Unique Post",
			wantItemContent: "<p>This content has <b>bold</b> and <i>italic</i> text.</p>",
		},
		{
			name:            "Feed with XSS Attempt",
			feed:            feedWithXSSAttempt,
			wantTitle:       "XSS Test Feed",
			wantItemTitle:   "XSS Post",
			wantItemContent: `<p>This content tries to execute <img src="x"></p>`, // Sanitized output
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an empty RSSFeed object
			rssFeed := &RSSFeed{}

			// Convert gofeed.Feed to RSSFeed with sanitization
			convertGofeedToRSSFeed(rssFeed, tt.feed, sanitizer)

			// Check the title of the RSS feed
			if rssFeed.Channel.Title != tt.wantTitle {
				t.Errorf("RSSFeed Title = %v, want %v", rssFeed.Channel.Title, tt.wantTitle)
			}

			// If there are items in the RSS feed, check the title and content of the first item
			if len(rssFeed.Channel.Item) > 0 {
				item := rssFeed.Channel.Item[0]
				if item.Title != tt.wantItemTitle {
					t.Errorf("RSSFeed Item Title = %v, want %v", item.Title, tt.wantItemTitle)
				}
				if item.Content != tt.wantItemContent {
					t.Errorf("RSSFeed Item Content = %v, want %v", item.Content, tt.wantItemContent)
				}
			} else if tt.wantItemTitle != "" || tt.wantItemContent != "" {
				t.Errorf("Expected items but found none.")
			}
		})
	}
}

func Test_convertAtomfeedToRSSFeed(t *testing.T) {
	// Example feed data to be used in the tests
	feedWithBasicInfo := &atom.Feed{
		Title:    "Test Atom Blog",
		Subtitle: "This is a test atom blog feed.",
		Links: []*atom.Link{
			{Href: "https://testatomblog.com"},
		},
		Entries: []*atom.Entry{
			{
				Title:   "First Atom Post",
				Links:   []*atom.Link{{Href: "https://testatomblog.com/first-post"}},
				Summary: "<p>This is the first atom post.</p>",
				Content: &atom.Content{
					Type:  "html",
					Value: "<p>This is the content of the first atom post.</p>",
				},
			},
		},
	}

	emptyFeed := &atom.Feed{}

	feedWithSpecialCharacters := &atom.Feed{
		Title:    "Special <Feed>",
		Subtitle: "This feed contains <b>HTML</b> & special characters!",
		Links: []*atom.Link{
			{Href: "https://specialfeed.com"},
		},
		Entries: []*atom.Entry{
			{
				Title:   "Special & Unique Post",
				Summary: "<p>This post contains & symbols.</p>",
				Content: &atom.Content{Type: "html", Value: "<p>This content has <b>bold</b> and <i>italic</i> text.</p>"},
				Links: []*atom.Link{
					{Href: "https://specialfeed.com/post"},
				},
				Published: "2023-08-02T00:00:00Z",
			},
		},
	}

	feedWithXSSAttempt := &atom.Feed{
		Title:    "XSS Test Feed",
		Subtitle: "This feed contains an XSS attempt!",
		Links: []*atom.Link{
			{Href: "https://xsstest.com"},
		},
		Entries: []*atom.Entry{
			{
				Title:   "XSS Post",
				Summary: `<p>This post tries to execute <script>alert('XSS')</script></p>`,
				Content: &atom.Content{Type: "html", Value: `<p>This content tries to execute <img src="x" onerror="alert('XSS')"></p>`},
				Links: []*atom.Link{
					{Href: "https://xsstest.com/post"},
				},
				Published: "2023-08-03T00:00:00Z",
			},
		},
	}

	// Initialize the sanitizer policy
	sanitizer := bluemonday.UGCPolicy()

	tests := []struct {
		name            string
		feed            *atom.Feed
		wantTitle       string
		wantItemTitle   string
		wantItemContent string
	}{
		{
			name:            "Basic Atom Feed Conversion",
			feed:            feedWithBasicInfo,
			wantTitle:       "Test Atom Blog",
			wantItemTitle:   "First Atom Post",
			wantItemContent: "<p>This is the content of the first atom post.</p>",
		},
		{
			name:            "Empty Feed",
			feed:            emptyFeed,
			wantTitle:       "",
			wantItemTitle:   "",
			wantItemContent: "",
		},
		{
			name:            "Feed with Special Characters",
			feed:            feedWithSpecialCharacters,
			wantTitle:       "Special ",
			wantItemTitle:   "Special &amp; Unique Post",
			wantItemContent: "<p>This content has <b>bold</b> and <i>italic</i> text.</p>",
		},
		{
			name:            "Feed with XSS Attempt",
			feed:            feedWithXSSAttempt,
			wantTitle:       "XSS Test Feed",
			wantItemTitle:   "XSS Post",
			wantItemContent: `<p>This content tries to execute <img src="x"></p>`, // Sanitized output
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an empty RSSFeed object
			rssFeed := &RSSFeed{}

			// Convert atom.Feed to RSSFeed with sanitization
			convertAtomfeedToRSSFeed(rssFeed, tt.feed, sanitizer)

			// Check the title of the RSS feed
			if rssFeed.Channel.Title != tt.wantTitle {
				t.Errorf("RSSFeed Title = %v, want %v", rssFeed.Channel.Title, tt.wantTitle)
			}

			// If there are items in the RSS feed, check the title and content of the first item
			if len(rssFeed.Channel.Item) > 0 {
				item := rssFeed.Channel.Item[0]
				if item.Title != tt.wantItemTitle {
					t.Errorf("RSSFeed Item Title = %v, want %v", item.Title, tt.wantItemTitle)
				}
				if item.Content != tt.wantItemContent {
					t.Errorf("RSSFeed Item Content = %v, want %v", item.Content, tt.wantItemContent)
				}
			} else if tt.wantItemTitle != "" || tt.wantItemContent != "" {
				t.Errorf("Expected items but found none.")
			}
		})
	}
}
