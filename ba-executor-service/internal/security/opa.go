package security

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type OPAClient struct {
	url string
}

func NewOPAClient() *OPAClient {
	url := os.Getenv("OPA_URL")
	if url == "" {
		url = "http://opa:8181/v1/data/ba/agent/system/allow"
	}
	return &OPAClient{url: url}
}

type OPAQuery struct {
	Input interface{} `json:"input"`
}

type OPAResponse struct {
	Result bool `json:"result"`
}

func (c *OPAClient) Check(input interface{}) (bool, error) {
	query := OPAQuery{Input: input}
	body, err := json.Marshal(query)
	if err != nil {
		return false, err
	}

	resp, err := http.Post(c.url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var opaResp OPAResponse
	if err := json.NewDecoder(resp.Body).Decode(&opaResp); err != nil {
		return false, err
	}

	return opaResp.Result, nil
}
