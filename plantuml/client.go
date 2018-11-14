package plantuml

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const DefaultClientURL = "http://www.plantuml.com/plantuml"

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type ClientConfig struct {
	Doer Doer
	URL  string `yaml:"url" json:"url"`
}

type Client struct {
	config ClientConfig
}

func NewClient(config ClientConfig) (*Client, error) {
	if config.Doer == nil {
		config.Doer = http.DefaultClient
	}
	if config.URL == "" {
		config.URL = DefaultClientURL
	}
	return &Client{
		config: config,
	}, nil
}

// TODO: cleanup with more informative errors

// Convert takes the input text from uml files and sends it to the plantUML api to
// convert it to the requested format.
func (c *Client) Convert(text []byte, format string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s/%s", c.config.URL, format, Encode(text))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.config.Doer.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}
