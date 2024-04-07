package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/Permify/permify-cli/core/logger"
)

// ReadErrorResponse reads the error returned from the server if any
func ReadErrorResponse(StatusCode int, body []byte) error {
	logger.Log.Error("api call failed with", "status_code", StatusCode)
	if StatusCode != 500 {
		msg := ErrorResponse{StatusCode: StatusCode}
		err := json.Unmarshal(body, &msg)
		if err != nil {
			logger.Log.Error("failed to marshal the response body")
			return err
		}
		logger.Log.Error("Below is the error message from the server")
		return errors.New(msg.Message)
	} 

	return errors.New(string(body))
}

// Get request implementation
func Get(host, path string, query map[string]string, respModel interface{}) error {
	finalURL, err := url.JoinPath(host, path)
	if err != nil {
		logger.Log.Error("failed to join url")
		return err
	}
	req, err := http.NewRequest("GET", finalURL, nil)
	if err != nil {
		logger.Log.Error("failed to create request object")
		return err
	}
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	logger.Log.Info("requesting ", "url", req.URL.String())
	client := &http.Client{
		// CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// 	return http.ErrUseLastResponse
		// },
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error("failed to make the request")
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("failed to read body message")
		return err
	}
	if resp.StatusCode == 200 {
		return json.Unmarshal(body, respModel)
	}
	return ReadErrorResponse(resp.StatusCode, body)
}

// Post request implementation
func Post(host, path string, query map[string]string, requestModel, respModel interface{}) error {
	finalURL, err := url.JoinPath(host, path)
	if err != nil {
		logger.Log.Error("failed to join url")
		return err
	}
	marshalParams, err := json.Marshal(requestModel)
	if err != nil {
		logger.Log.Error("failed to create request parameters")
		return err
	}
	req, err := http.NewRequest(http.MethodPost, finalURL, bytes.NewBuffer(marshalParams))
	if err != nil {
		logger.Log.Error("failed to create request")
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	logger.Log.Info("requesting ", "url", req.URL.String())
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error("failed to make the request")
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("failed to read body message")
		return err
	}
	if resp.StatusCode == 200 {
		return json.Unmarshal(body, respModel)
	}
	return ReadErrorResponse(resp.StatusCode, body)
}

// Delete request implementation
func Delete(host, path string, query map[string]string, respModel interface{}) error {
	finalURL, err := url.JoinPath(host, path)
	if err != nil {
		logger.Log.Error("failed to join url")
		return err
	}
	req, err := http.NewRequest("DELETE", finalURL, nil)
	if err != nil {
		logger.Log.Error("failed to create request object")
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	logger.Log.Info("requesting ", "url", req.URL.String())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error("failed to make the request")
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("failed to read body message")
		return err
	}
	if resp.StatusCode == 200 {
		return json.Unmarshal(body, respModel)
	}
	return ReadErrorResponse(resp.StatusCode, body)
}

// Put request implementation
func Put(host, path string, query map[string]string, requestParams, respModel interface{}) error {
	finalURL, err := url.JoinPath(host, path)
	if err != nil {
		logger.Log.Error("failed to join url")
		return err
	}
	marshalParams, err := json.Marshal(requestParams)
	if err != nil {
		logger.Log.Error("failed to create request parameters")
		return err
	}
	req, err := http.NewRequest(http.MethodPut, finalURL, bytes.NewBuffer(marshalParams))
	if err != nil {
		logger.Log.Error("failed to create request")
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	logger.Log.Info("requesting ", "url", req.URL.String())
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error("failed to make the request")
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("failed to read body message")
		return err
	}
	if resp.StatusCode == 200 {
		return json.Unmarshal(body, respModel)
	}
	return ReadErrorResponse(resp.StatusCode, body)
}
