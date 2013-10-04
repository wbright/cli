package api

import (
	"cf/net"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func newRequest(gateway net.Gateway, method, path, accessToken string, body io.Reader) (req *net.Request, apiStatus ApiStatus) {
	req, err := gateway.NewRequest(method, path, accessToken, body)
	if err != nil {
		apiStatus = NewApiStatusWithMessage(err.Error())
	}
	return
}

func performRequest(gateway net.Gateway, request *net.Request) (apiStatus ApiStatus) {
	errResponse, err := gateway.PerformRequest(request)
	if err != nil {
		apiStatus = NewApiStatusWithErrorResponse(errResponse)
	}
	return
}

func performRequestForResponseBytes(gateway net.Gateway, request *net.Request) (bytes []byte, headers http.Header, apiStatus ApiStatus) {
	bytes, headers, errResponse, err := gateway.PerformRequestForResponseBytes(request)
	if err != nil {
		apiStatus = NewApiStatusWithErrorResponse(errResponse)
	}
	return
}

func performRequestForTextResponse(gateway net.Gateway, request *net.Request) (response string, headers http.Header, apiStatus ApiStatus) {
	response, headers, errResponse, err := gateway.PerformRequestForTextResponse(request)
	if err != nil {
		apiStatus = NewApiStatusWithErrorResponse(errResponse)
	}
	return
}

func performRequestForJSONResponse(gateway net.Gateway, request *net.Request, response interface{}) (headers http.Header, apiStatus ApiStatus) {
	headers, errResponse, err := gateway.PerformRequestForJSONResponse(request, &response)
	if err != nil {
		apiStatus = NewApiStatusWithErrorResponse(errResponse)
	}
	return
}

func wrapError(msg string, err error) error {
	return newError("%s: %s", msg, err.Error())
}

func newError(msg string, args ...interface{}) error {
	return errors.New(fmt.Sprintf(msg, args...))
}
