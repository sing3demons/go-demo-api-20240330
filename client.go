package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

func makeRequest(method, url string, payload io.Reader) ([]byte, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func HttpClientGet[T any](url string) (T, error) {
	var response T
	body, err := makeRequest(http.MethodGet, url, nil)
	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return response, err
	}

	return response, nil
}

func HttpClientPost[T any](url string, payload any) (T, error) {
	var response T

	payloadData, err := json.Marshal(payload)
	if err != nil {
		return response, err
	}
	body, err := makeRequest(http.MethodPost, url, bytes.NewReader(payloadData))
	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return response, err
	}

	return response, nil
}
