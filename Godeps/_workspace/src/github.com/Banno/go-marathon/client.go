package marathon

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	url        *url.URL
	httpClient *http.Client
}

type ClientError struct {
	HttpStatusCode int
}

func (ce ClientError) Error() string {
	return fmt.Sprintf("Got HTTP Status Code: %d", ce.HttpStatusCode)
}

func NewClient(hostname string, port int) *Client {
	return NewClientForUrl(fmt.Sprintf("http://%s:%d", hostname, port))
}

func NewClientForUrl(rawurl string) *Client {
	url, err := url.Parse(rawurl)

	if err != nil {
		panic(err)
	}

	c := &Client{
		url:        url,
		httpClient: &http.Client{},
	}

	return c
}

func (c *Client) getFullUrl(apiEndpoint string) string {
	fullUrl, err := c.url.Parse(apiEndpoint)
	if err != nil {
		panic(err)
	}
	return fullUrl.String()
}

func (c *Client) getJson(apiEndpoint string) ([]byte, error) {
	// https://github.com/mesosphere/marathon/issues/1357
	req, _ := http.NewRequest("GET", c.getFullUrl(apiEndpoint), nil)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if statusCodeErr := checkSuccessFullStatusCode(resp); statusCodeErr != nil {
		return nil, statusCodeErr
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return body, nil
}

func (c *Client) postJson(apiEndpoint string, json []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", c.getFullUrl(apiEndpoint), bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if statusCodeErr := checkSuccessFullStatusCode(resp); statusCodeErr != nil {
		return nil, statusCodeErr
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return body, nil
}

func (c *Client) putJson(apiEndpoint string, json []byte) ([]byte, error) {
	req, err := http.NewRequest("PUT", c.getFullUrl(apiEndpoint), bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if statusCodeErr := checkSuccessFullStatusCode(resp); statusCodeErr != nil {
		return nil, statusCodeErr
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return body, nil
}

func (c *Client) ping() error {
	resp, err := c.httpClient.Get(c.getFullUrl("/ping"))
	if err != nil {
		return err
	}
	if statusCodeErr := checkSuccessFullStatusCode(resp); statusCodeErr != nil {
		return statusCodeErr
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	bodyStr := string(body)
	if strings.Contains(bodyStr, "pong") != true {
		return fmt.Errorf("/ping didn't return pong but did return: %s", bodyStr)
	}

	return nil
}

func checkSuccessFullStatusCode(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ClientError{resp.StatusCode}
	}
	return nil
}
