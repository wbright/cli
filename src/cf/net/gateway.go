package net

import (
	"bytes"
	"cf"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
)

const INVALID_TOKEN_CODE = "GATEWAY INVALID TOKEN CODE"

type ErrorResponse struct {
	StatusCode  int
	ErrorCode   string
	Description string
}

type errorHandler func(*http.Response) ErrorResponse

type tokenRefresher interface {
	RefreshAuthToken() (string, error)
}

type Request struct {
	*http.Request
}

type Gateway struct {
	authenticator tokenRefresher
	errHandler    errorHandler
}

func newGateway(errHandler errorHandler) (gateway Gateway) {
	gateway.errHandler = errHandler
	return
}

func (gateway *Gateway) SetTokenRefresher(auth tokenRefresher) {
	gateway.authenticator = auth
}

func (gateway Gateway) NewRequest(method, path, accessToken string, body io.Reader) (req *Request, err error) {
	request, err := http.NewRequest(method, path, body)
	if err != nil {
		err = wrapError("Error building request", err)
		return
	}

	if accessToken != "" {
		request.Header.Set("Authorization", accessToken)
	}

	request.Header.Set("accept", "application/json")
	request.Header.Set("User-Agent", "go-cli "+cf.Version+" / "+runtime.GOOS)
	req = &Request{request}
	return
}

func (gateway Gateway) PerformRequest(request *Request) (errResponse ErrorResponse, err error) {
	_, errResponse, err = gateway.doRequestHandlingAuth(request)
	return
}

func (gateway Gateway) PerformRequestForResponseBytes(request *Request) (bytes []byte, headers http.Header, errResponse ErrorResponse, err error) {
	rawResponse, errResponse, err := gateway.doRequestHandlingAuth(request)
	if err != nil {
		return
	}

	bytes, err = ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		err = wrapError("Error reading response", err)
	}
	return
}

func (gateway Gateway) PerformRequestForTextResponse(request *Request) (response string, headers http.Header, errResponse ErrorResponse, err error) {
	bytes, headers, errResponse, err := gateway.PerformRequestForResponseBytes(request)
	response = string(bytes)
	return
}

func (gateway Gateway) PerformRequestForJSONResponse(request *Request, response interface{}) (headers http.Header, errResponse ErrorResponse, err error) {
	bytes, headers, errResponse, err := gateway.PerformRequestForResponseBytes(request)
	if err != nil {
		return
	}

	err = json.Unmarshal(bytes, &response)
	if err != nil {
		err = wrapError("Invalid JSON response from server", err)
	}
	return
}

func (gateway Gateway) doRequestHandlingAuth(request *Request) (response *http.Response, errResponse ErrorResponse, err error) {
	var bodyBytes []byte
	if request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(request.Body)
		request.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	}

	response, err = doRequest(request.Request)

	if err != nil && response == nil {
		err = wrapError("Error performing request", err)
		return
	}

	if response.StatusCode > 299 {
		errResponse = gateway.errHandler(response)
		err = errors.New("Server error")
		return
	}

	if err == nil || gateway.authenticator == nil {
		return
	}

	if errResponse.ErrorCode == INVALID_TOKEN_CODE {
		var newToken string
		newToken, err = gateway.authenticator.RefreshAuthToken()

		if err == nil {
			request.Header.Set("Authorization", newToken)
			if len(bodyBytes) > 0 {
				request.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
			}

			errResponse = ErrorResponse{}
			response, err = doRequest(request.Request)
			return
		}
	}

	return
}

func wrapError(msg string, err error) error {
	return errors.New(fmt.Sprintf("%s: %s", msg, err.Error()))
}
