package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// PageFetcher is an unordered set of keyword-URL couples.
type PageFetcher map[string]string

// FetchError is a collection of errors that happened during Fetch().
type FetchError map[string]error

func (e FetchError) Error() string {
	if e == nil {
		return ""
	}
	ret := make([]string, 0, len(e))
	for key, err := range e {
		ret = append(ret, fmt.Sprintf("%s='%v'", key, err))
	}
	return strings.Join(ret, " ")
}

// Exclude will remove from PageFetcher all the keys that are in the given
// map.
func (pf PageFetcher) Exclude(excl map[string]string) {
	for k := range excl {
		delete(pf, k)
	}
}

// Fetch will HTTP-Get the content of each keyword's URL and return it.
// It will return a non-nil error when at least one the call failed.
func (pf PageFetcher) Fetch() (map[string]string, error) {
	ret := make(map[string]string)
	errs := make(FetchError)
	for key, url := range pf {
		// Make sure the URL starts with a scheme
		// FIXME: use url.Parse?
		if !strings.HasPrefix(url, "http") {
			url = "http://" + url
		}
		res, err := http.Get(url)
		if err != nil {
			errs[key] = err
			continue
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			errs[key] = err
			continue
		}
		ret[key] = string(body)
	}
	if len(errs) != 0 {
		return ret, errs
	}
	// Make sure a nil type is returned when there is no error.
	return ret, nil
}

// PageFetcherFromString returns a new PageFetcher built from a space-separated
// list of key=url couples.
// It shall return an error if any couple is not parsable.
func PageFetcherFromString(s string) (PageFetcher, error) {
	pf := make(PageFetcher)

	for _, couple := range strings.Split(s, " ") {
		if len(couple) == 0 {
			continue
		}
		keyval := strings.SplitN(couple, "=", 2)
		if len(keyval) != 2 {
			return pf, fmt.Errorf("cannot parse '%s' as key=value", couple)
		}
		pf[keyval[0]] = keyval[1]
	}

	return pf, nil
}
