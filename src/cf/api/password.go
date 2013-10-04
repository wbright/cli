package api

import (
	"cf/configuration"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type PasswordRepository interface {
	GetScore(password string) (string, ApiStatus)
	UpdatePassword(old string, new string) ApiStatus
}

type CloudControllerPasswordRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway

	infoResponse InfoResponse
}

func NewCloudControllerPasswordRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerPasswordRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

type ScoreResponse struct {
	Score         int
	RequiredScore int
}

type InfoResponse struct {
	TokenEndpoint string `json:"token_endpoint"`
	UserGuid      string `json:"user"`
}

func (repo CloudControllerPasswordRepository) GetScore(password string) (score string, apiStatus ApiStatus) {
	infoResponse, apiStatus := repo.getTargetInfo()
	if apiStatus.NotSuccessful() {
		return
	}

	scorePath := fmt.Sprintf("%s/password/score", infoResponse.TokenEndpoint)
	scoreBody := url.Values{
		"password": []string{password},
	}

	scoreRequest, apiStatus := newRequest(repo.gateway, "POST", scorePath, repo.config.AccessToken, strings.NewReader(scoreBody.Encode()))
	if apiStatus.NotSuccessful() {
		return
	}
	scoreRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	scoreResponse := ScoreResponse{}

	_, apiStatus = performRequestForJSONResponse(repo.gateway, scoreRequest, &scoreResponse)
	if apiStatus.NotSuccessful() {
		return
	}

	score = translateScoreResponse(scoreResponse)
	return
}

func (repo CloudControllerPasswordRepository) UpdatePassword(old string, new string) (apiStatus ApiStatus) {
	infoResponse, apiStatus := repo.getTargetInfo()
	if apiStatus.NotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/Users/%s/password", infoResponse.TokenEndpoint, repo.config.UserGuid())
	body := fmt.Sprintf(`{"password":"%s","oldPassword":"%s"}`, new, old)
	request, apiStatus := newRequest(repo.gateway, "PUT", path, repo.config.AccessToken, strings.NewReader(body))
	if apiStatus.NotSuccessful() {
		return
	}

	request.Header.Set("Content-Type", "application/json")

	apiStatus = performRequest(repo.gateway, request)
	return
}

func (repo *CloudControllerPasswordRepository) getTargetInfo() (response InfoResponse, apiStatus ApiStatus) {
	if repo.infoResponse.UserGuid == "" {
		path := fmt.Sprintf("%s/info", repo.config.Target)
		var request *net.Request
		request, apiStatus = newRequest(repo.gateway, "GET", path, repo.config.AccessToken, nil)
		if apiStatus.NotSuccessful() {
			return
		}

		response = InfoResponse{}

		_, apiStatus = performRequestForJSONResponse(repo.gateway, request, &response)

		repo.infoResponse = response
	}

	response = repo.infoResponse

	return
}

func translateScoreResponse(response ScoreResponse) string {
	if response.Score == 10 {
		return "strong"
	}

	if response.Score >= response.RequiredScore {
		return "good"
	}

	return "weak"
}
