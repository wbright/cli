package api

import (
	"cf/configuration"
	"cf/net"
	"cf/terminal"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type AuthenticationRepository interface {
	Authenticate(email string, password string) (apiStatus ApiStatus)
	RefreshAuthToken() (updatedToken string, err error)
}

type UAAAuthenticationRepository struct {
	configRepo configuration.ConfigurationRepository
	config     *configuration.Configuration
	gateway    net.Gateway
}

func NewUAAAuthenticationRepository(gateway net.Gateway, configRepo configuration.ConfigurationRepository) (uaa UAAAuthenticationRepository) {
	uaa.gateway = gateway
	uaa.configRepo = configRepo
	uaa.config, _ = configRepo.Get()
	return
}

func (uaa UAAAuthenticationRepository) Authenticate(email string, password string) (apiStatus ApiStatus) {
	data := url.Values{
		"username":   {email},
		"password":   {password},
		"grant_type": {"password"},
		"scope":      {""},
	}

	errResponse, err := uaa.getAuthToken(data)
	if err != nil && errResponse.StatusCode == 401 {
		apiStatus = NewApiStatusWithMessage("Password is incorrect, please try again.")
	}
	return
}

func (uaa UAAAuthenticationRepository) RefreshAuthToken() (updatedToken string, err error) {
	data := url.Values{
		"refresh_token": {uaa.config.RefreshToken},
		"grant_type":    {"refresh_token"},
		"scope":         {""},
	}

	_, err = uaa.getAuthToken(data)
	updatedToken = uaa.config.AccessToken

	if err != nil {
		fmt.Printf("%s\n\n", terminal.NotLoggedInText())
		os.Exit(1)
	}

	return
}

func (uaa UAAAuthenticationRepository) getAuthToken(data url.Values) (errResponse net.ErrorResponse, err error) {
	type uaaErrorResponse struct {
		Code        string `json:"error"`
		Description string `json:"error_description"`
	}

	type AuthenticationResponse struct {
		AccessToken  string           `json:"access_token"`
		TokenType    string           `json:"token_type"`
		RefreshToken string           `json:"refresh_token"`
		Error        uaaErrorResponse `json:"error"`
	}

	path := fmt.Sprintf("%s/oauth/token", uaa.config.AuthorizationEndpoint)
	request, err := uaa.gateway.NewRequest("POST", path, "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:")), strings.NewReader(data.Encode()))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := new(AuthenticationResponse)
	_, errResponse, err = uaa.gateway.PerformRequestForJSONResponse(request, &response)
	if err != nil {
		return
	}

	if response.Error.Code != "" {
		err = newError("Error setting configuration: %s", response.Error.Description)
		return
	}

	uaa.config.AccessToken = fmt.Sprintf("%s %s", response.TokenType, response.AccessToken)
	uaa.config.RefreshToken = response.RefreshToken
	err = uaa.configRepo.Save()
	if err != nil {
		err = wrapError("Error setting configuration", err)
	}

	return
}
