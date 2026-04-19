package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func json_from_request[T any](web string) (T, error) {
	var furl, _ = url.Parse(web)

	req, _ := http.NewRequest("GET", furl.String(), nil)
	req.Header.Set("User-Agent", "test-go-app")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		var zero T
		return zero, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	var mod_data T
	json.Unmarshal(body, &mod_data)
	return mod_data, err
}

func isValidUrl(toTest string) bool {
	// Basic structure check
	if _, err := url.ParseRequestURI(toTest); err != nil {
		return false
	}

	// Detailed check for scheme and host
	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}
