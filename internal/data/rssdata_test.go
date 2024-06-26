package data

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

// mockReadCloser creates a ReadCloser from a string for testing purposes.
func mockReadCloser(s string) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBufferString(s))
}
func TestRssFeedDecoder(t *testing.T) {
	testData := `
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
	20 rules of thumb for writing better software. Optimize for simplicity first Write code for humans, not computers Reading is more important than writing Any style is fine, as long as it’s black There should be one way to do it, but seriously this time Hide the sharp knives Changing the rules is better than adding exceptions Libraries are better than frameworks Transitive dependencies are a problem Dynamic runtime dependencies are a bigger problem API surface area is a liability Returning early is a good thing Use more plain text Compiler errors are better than runtime errors Runtime errors are better than bugs Tooling is better than documentation Documentation is better than nothing Configuration sucks, but so does convention The cost of building a feature is its smallest cost Types are one honking great idea – let’s do more of those!
	</description>
	</item>
	</channel>
	`
	resp := &http.Response{
		Body: mockReadCloser(testData),
	}
	type args struct {
		rssFeed *RSSFeed
		resp    *http.Response
		url     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"Test1", args{&RSSFeed{}, resp, "https://www.montreuxdocument.org/celebrating-15-years/atom.xml"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RssFeedDecoder(tt.args.url, tt.args.rssFeed, tt.args.resp); (err != nil) != tt.wantErr {
				t.Errorf("RssFeedDecoder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
