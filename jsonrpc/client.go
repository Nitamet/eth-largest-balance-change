package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	httpClient http.Client
	Endpoint   string
}

type RequestBody struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	Id      int    `json:"id"`
}

type ResponseBody struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  any    `json:"result"`
}

type InvalidResponseCodeError struct {
	Code int
}

func (e InvalidResponseCodeError) Error() string {
	return fmt.Sprintf("invalid response code: %d", e.Code)
}

func (c *Client) Call(ctx context.Context, method string, params []any, out any) (err error) {
	body := RequestBody{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		Id:      1,
	}

	encodedBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.Endpoint, bytes.NewBuffer(encodedBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		bodyCloseErr := response.Body.Close()

		if err == nil {
			err = bodyCloseErr
		}
	}()

	if response.StatusCode != http.StatusOK {
		return InvalidResponseCodeError{
			Code: response.StatusCode,
		}
	}

	var responseBody ResponseBody
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		return err
	}

	result, err := json.Marshal(responseBody.Result)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(result, out); err != nil {
		return err
	}

	return err
}
