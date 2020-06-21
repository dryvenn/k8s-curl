package main

import (
	"strings"
	"testing"
)

func TestFetch(t *testing.T) {
	tests := []struct {
		name    string
		fetcher PageFetcher
		// map of key/page-substring
		success map[string]string
	}{
		{
			name: "simple good",
			fetcher: PageFetcher{
				"google": "https://google.com",
			},
			success: map[string]string{"google": "google"},
		},
		{
			name: "simple bad",
			fetcher: PageFetcher{
				"gogole": "https://gogole.com",
			},
			success: map[string]string{},
		},
		{
			name: "simple multi",
			fetcher: PageFetcher{
				"google": "https://google.com",
				"gogole": "https://gogole.com",
			},
			success: map[string]string{"google": "google"},
		},
		{
			name: "multi",
			fetcher: PageFetcher{
				"google":      "https://google.com",
				"datanet":     "http://data.net",
				"datanets":    "https://data.net",
				"cloudflare":  "http://1.1.1.1",
				"cloudflares": "https://1.1.1.1",

				"something else": "invalid url",
			},
			success: map[string]string{
				"google":      "google",
				"datanet":     "Now, now, we've all seen a web server before.",
				"datanets":    "Now, now, we've all seen a web server before.",
				"cloudflare":  "makes your Internet faster",
				"cloudflares": "makes your Internet faster",
			},
		},
		{
			name: "no scheme",
			fetcher: PageFetcher{
				"datanet": "data.net",
			},
			success: map[string]string{
				"datanet": "Now, now, we've all seen a web server before.",
			},
		},
		{
			name: "joke",
			fetcher: PageFetcher{
				"joke": "curl-a-joke.herokuapp.com",
				"poke": "curl-a-poke.herokuapp.com",
			},
			success: map[string]string{
				"joke": "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := test.fetcher.Fetch()
			// Each positive result is a key-page couple.
			// They key must be in 'success', and the page must contain the 'success'
			// substring.
			for key, page := range res {
				if _, ok := test.fetcher[key]; !ok {
					t.Errorf("unknown key %s", key)
				}
				if sub, ok := test.success[key]; !ok {
					t.Errorf("%s key not in success", key)
				} else if !strings.Contains(page, sub) {
					t.Errorf("'%s' not in %s page", sub, key)
				}
			}
			// Test all positive results are accounted for.
			if len(test.success) != len(res) {
				t.Error("missing success results")
			}
			// Test error is correctly nil if applicable.
			if len(test.fetcher)-len(test.success) == 0 {
				if err != nil {
					t.Fatal("expected nil error")
				}
				return
			}
			fetchErr, ok := err.(FetchError)
			if !ok {
				t.Fatalf("%v is not a FetchError", err)
			}
			// Each negative result is a key-error couple.
			// They key must not be in 'success'.
			for key, _ := range fetchErr {
				if _, ok := test.fetcher[key]; !ok {
					t.Errorf("unknown key %s", key)
				}
				if _, ok := test.success[key]; ok {
					t.Errorf("%s key in success", key)
				}
			}
		})
	}
}

func TestPageFetcherFromString(t *testing.T) {
	tests := []struct {
		input   string
		output  PageFetcher
		success bool
	}{
		{
			input:   "key=val",
			output:  PageFetcher{"key": "val"},
			success: true,
		},
		{
			input:   "key1=val1 key2=val2",
			output:  PageFetcher{"key1": "val1", "key2": "val2"},
			success: true,
		},
		{
			input:   "key1=val1  key2=val2",
			output:  PageFetcher{"key1": "val1", "key2": "val2"},
			success: true,
		},
		{
			input:   "key1",
			output:  PageFetcher{},
			success: false,
		},
		{
			input:   "key1 key2",
			output:  PageFetcher{},
			success: false,
		},
		{
			input:   "key1 = key2",
			output:  PageFetcher{},
			success: false,
		},
		{
			input:   "key1=val1 key2",
			output:  PageFetcher{"key1": "val1"},
			success: false,
		},
		{
			input:   "key1=val1=val11 key2=val2",
			output:  PageFetcher{"key1": "val1=val11", "key2": "val2"},
			success: true,
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			res, err := PageFetcherFromString(test.input)
			if test.success && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(res) != len(test.output) {
				t.Fatalf("expected %d couples, got %d %v", len(test.output), len(res), res)
			}
			for k, v := range res {
				if out, ok := test.output[k]; !ok {
					t.Errorf("unexpected key %s", k)
				} else if out != v {
					t.Errorf("unexpected val %s", v)
				}
			}
		})
	}
}
