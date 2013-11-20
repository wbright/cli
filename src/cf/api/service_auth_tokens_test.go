package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
	"testing"
)

func TestServiceAuthCreate(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "POST",
		Path:     "/v2/service_auth_tokens",
		Matcher:  testnet.RequestBodyMatcher(`{"label":"a label","provider":"a provider","token":"a token"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createServiceAuthTokenRepo(t, req)
	defer ts.Close()
	authToken_Auto := cf.ServiceAuthToken{}
	authToken_Auto.Label = "a label"
	authToken_Auto.Provider = "a provider"
	authToken_Auto.Token = "a token"
	apiResponse := repo.Create(authToken_Auto)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestServiceAuthFindAll(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/service_auth_tokens",
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body: `{ "resources": [
				{
				  "metadata": {
					"guid": "mysql-core-guid"
				  },
				  "entity": {
					"label": "mysql",
					"provider": "mysql-core"
				  }
				},
				{
				  "metadata": {
					"guid": "postgres-core-guid"
				  },
				  "entity": {
					"label": "postgres",
					"provider": "postgres-core"
				  }
				}
			]}`},
	})

	ts, handler, repo := createServiceAuthTokenRepo(t, req)
	defer ts.Close()

	authTokens, apiResponse := repo.FindAll()
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())

	assert.Equal(t, len(authTokens), 2)

	assert.Equal(t, authTokens[0].Label, "mysql")
	assert.Equal(t, authTokens[0].Provider, "mysql-core")
	assert.Equal(t, authTokens[0].Guid, "mysql-core-guid")

	assert.Equal(t, authTokens[1].Label, "postgres")
	assert.Equal(t, authTokens[1].Provider, "postgres-core")
	assert.Equal(t, authTokens[1].Guid, "postgres-core-guid")
}

func TestServiceAuthFindByLabelAndProvider(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/service_auth_tokens?q=label:a-label;provider:a-provider",
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body: `{"resources": [{
			"metadata": { "guid": "mysql-core-guid" },
			"entity": {
				"label": "mysql",
				"provider": "mysql-core"
			}
		}]}`},
	})

	ts, handler, repo := createServiceAuthTokenRepo(t, req)
	defer ts.Close()

	serviceAuthToken, apiResponse := repo.FindByLabelAndProvider("a-label", "a-provider")

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
	authToken_Auto2 := cf.ServiceAuthToken{}
	authToken_Auto2.Guid = "mysql-core-guid"
	authToken_Auto2.Label = "mysql"
	authToken_Auto2.Provider = "mysql-core"
	assert.Equal(t, serviceAuthToken, authToken_Auto2)
}

func TestServiceAuthFindByLabelAndProviderWhenNotFound(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/service_auth_tokens?q=label:a-label;provider:a-provider",
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body:   `{"resources": []}`},
	})

	ts, handler, repo := createServiceAuthTokenRepo(t, req)
	defer ts.Close()

	_, apiResponse := repo.FindByLabelAndProvider("a-label", "a-provider")

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestServiceAuthUpdate(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/service_auth_tokens/mysql-core-guid",
		Matcher:  testnet.RequestBodyMatcher(`{"token":"a value"}`),
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	ts, handler, repo := createServiceAuthTokenRepo(t, req)
	defer ts.Close()
	authToken_Auto3 := cf.ServiceAuthToken{}
	authToken_Auto3.Guid = "mysql-core-guid"
	authToken_Auto3.Token = "a value"
	apiResponse := repo.Update(authToken_Auto3)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestServiceAuthDelete(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/v2/service_auth_tokens/mysql-core-guid",
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	ts, handler, repo := createServiceAuthTokenRepo(t, req)
	defer ts.Close()
	authToken_Auto4 := cf.ServiceAuthToken{}
	authToken_Auto4.Guid = "mysql-core-guid"
	apiResponse := repo.Delete(authToken_Auto4)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func createServiceAuthTokenRepo(t *testing.T, request testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceAuthTokenRepository) {
	ts, handler = testnet.NewTLSServer(t, []testnet.TestRequest{request})

	config := &configuration.Configuration{
		Target:      ts.URL,
		AccessToken: "BEARER my_access_token",
	}
	gateway := net.NewCloudControllerGateway()

	repo = NewCloudControllerServiceAuthTokenRepository(config, gateway)
	return
}
