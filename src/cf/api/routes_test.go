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

var firstPageRoutesResponse = testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "next_url": "/v2/routes?inline-relations-depth=1&page=2",
  "resources": [
    {
      "metadata": {
        "guid": "route-1-guid"
      },
      "entity": {
        "host": "route-1-host",
        "domain": {
          "metadata": {
            "guid": "domain-1-guid"
          },
          "entity": {
            "name": "cfapps.io"
          }
        },
        "space": {
          "metadata": {
            "guid": "space-1-guid"
          },
          "entity": {
            "name": "space-1"
          }
        },
        "apps": [
       	  {
       	    "metadata": {
              "guid": "app-1-guid"
            },
            "entity": {
              "name": "app-1"
       	    }
       	  }
        ]
      }
    }
  ]
}`}

var secondPageRoutesResponse = testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "route-2-guid"
      },
      "entity": {
        "host": "route-2-host",
        "domain": {
          "metadata": {
            "guid": "domain-2-guid"
          },
          "entity": {
            "name": "example.com"
          }
        },
        "space": {
          "metadata": {
            "guid": "space-2-guid"
          },
          "entity": {
            "name": "space-2"
          }
        },
        "apps": [
       	  {
       	    "metadata": {
              "guid": "app-2-guid"
            },
            "entity": {
              "name": "app-2"
       	    }
       	  },
       	  {
       	    "metadata": {
              "guid": "app-3-guid"
            },
            "entity": {
              "name": "app-3"
       	    }
       	  }
        ]
      }
    }
  ]
}`}

func TestRoutesListRoutes(t *testing.T) {
	firstRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/routes?inline-relations-depth=1",
		Response: firstPageRoutesResponse,
	})

	secondRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/routes?inline-relations-depth=1&page=2",
		Response: secondPageRoutesResponse,
	})

	ts, handler, repo, _ := createRoutesRepo(t, firstRequest, secondRequest)
	defer ts.Close()

	stopChan := make(chan bool)
	defer close(stopChan)
	routesChan, statusChan := repo.ListRoutes(stopChan)
	space_Auto := cf.Space{}
	space_Auto.Name = "space-1"
	space_Auto.Guid = "space-1-guid"
	space_Auto2 := cf.Space{}
	space_Auto2.Name = "space-2"
	space_Auto2.Guid = "space-2-guid"
	domain_Auto := cf.Domain{}
	domain_Auto.Name = "cfapps.io"
	domain_Auto.Guid = "domain-1-guid"
	domain_Auto2 := cf.Domain{}
	domain_Auto2.Name = "example.com"
	domain_Auto2.Guid = "domain-2-guid"
	expectedRoutes := []cf.Route{
		{
			Guid:   "route-1-guid",
			Host:   "route-1-host",
			Domain: domain_Auto,

			Space: space_Auto,

			AppNames: []string{"app-1"},
		},
		{
			Guid:   "route-2-guid",
			Host:   "route-2-host",
			Domain: domain_Auto2,

			Space: space_Auto2,

			AppNames: []string{"app-2", "app-3"},
		},
	}

	routes := []cf.Route{}
	for chunk := range routesChan {
		routes = append(routes, chunk...)
	}
	apiResponse := <-statusChan

	assert.Equal(t, routes, expectedRoutes)
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestRoutesListRoutesWithNoRoutes(t *testing.T) {
	emptyRoutesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/routes?inline-relations-depth=1",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
	})

	ts, handler, repo, _ := createRoutesRepo(t, emptyRoutesRequest)
	defer ts.Close()

	stopChan := make(chan bool)
	defer close(stopChan)
	routesChan, statusChan := repo.ListRoutes(stopChan)

	_, ok := <-routesChan
	apiResponse := <-statusChan

	assert.False(t, ok)
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

var findRouteByHostResponse = testnet.TestResponse{Status: http.StatusCreated, Body: `
{ "resources": [
    {
    	"metadata": {
        	"guid": "my-route-guid"
    	},
    	"entity": {
       	     "host": "my-cool-app"
    	}
    }
]}`}

func TestFindByHost(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/routes?q=host%3Amy-cool-app",
		Response: findRouteByHostResponse,
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()

	route, apiResponse := repo.FindByHost("my-cool-app")

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, route.Host, "my-cool-app")
	assert.Equal(t, route.Guid, "my-route-guid")
}

func TestFindByHostWhenHostIsNotFound(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/routes?q=host%3Amy-cool-app",
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: ` { "resources": [ ]}`},
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()

	_, apiResponse := repo.FindByHost("my-cool-app")

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsNotSuccessful())
}

func TestFindByHostAndDomain(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/routes?q=host%3Amy-cool-app%3Bdomain_guid%3Amy-domain-guid",
		Response: findRouteByHostResponse,
	})

	ts, handler, repo, domainRepo := createRoutesRepo(t, request)
	defer ts.Close()
	domain_Auto3 = cf.Domain{}
	domain_Auto3.Guid = "my-domain-guid"
	route, apiResponse := repo.FindByHostAndDomain("my-cool-app", "my-domain.com")

	assert.False(t, apiResponse.IsNotSuccessful())
	assert.True(t, handler.AllRequestsCalled())
	assert.Equal(t, domainRepo.FindByNameName, "my-domain.com")
	assert.Equal(t, route.Host, "my-cool-app")
	assert.Equal(t, route.Guid, "my-route-guid")
	assert.Equal(t, route.Domain, domainRepo.FindByNameDomain)
}

func TestFindByHostAndDomainWhenRouteIsNotFound(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/routes?q=host%3Amy-cool-app%3Bdomain_guid%3Amy-domain-guid",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [ ] }`},
	})

	ts, handler, repo, domainRepo := createRoutesRepo(t, request)
	defer ts.Close()
	domain_Auto4 = cf.Domain{}
	domain_Auto4.Guid = "my-domain-guid"
	_, apiResponse := repo.FindByHostAndDomain("my-cool-app", "my-domain.com")

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestCreateInSpace(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "POST",
		Path:    "/v2/routes",
		Matcher: testnet.RequestBodyMatcher(`{"host":"my-cool-app","domain_guid":"my-domain-guid","space_guid":"my-space-guid"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
{
  "metadata": { "guid": "my-route-guid" },
  "entity": { "host": "my-cool-app" }
}`},
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()
	domain := cf.Domain{}
	domain.Guid = "my-domain-guid"
	newRoute := cf.Route{}
	newRoute.Host = "my-cool-app"
	space := cf.Space{}
	space.Guid = "my-space-guid"

	createdRoute, apiResponse := repo.CreateInSpace(newRoute, domain, space)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	route_Auto := cf.Route{}
	route_Auto.Host = "my-cool-app"
	route_Auto.Guid = "my-route-guid"
	route_Auto.Domain = domain
	assert.Equal(t, createdRoute, route_Auto)
}

func TestCreateRoute(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "POST",
		Path:    "/v2/routes",
		Matcher: testnet.RequestBodyMatcher(`{"host":"my-cool-app","domain_guid":"my-domain-guid","space_guid":"my-space-guid"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
{
  "metadata": { "guid": "my-route-guid" },
  "entity": { "host": "my-cool-app" }
}`},
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()
	domain := cf.Domain{}
	domain.Guid = "my-domain-guid"
	newRoute := cf.Route{}
	newRoute.Host = "my-cool-app"

	createdRoute, apiResponse := repo.Create(newRoute, domain)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	route_Auto2 := cf.Route{}
	route_Auto2.Host = "my-cool-app"
	route_Auto2.Guid = "my-route-guid"
	route_Auto2.Domain = domain
	assert.Equal(t, createdRoute, route_Auto2)
}

func TestBind(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/apps/my-cool-app-guid/routes/my-cool-route-guid",
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()
	route := cf.Route{}
	route.Guid = "my-cool-route-guid"
	app := cf.Application{}
	app.Guid = "my-cool-app-guid"

	apiResponse := repo.Bind(route, app)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestUnbind(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/v2/apps/my-cool-app-guid/routes/my-cool-route-guid",
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()
	route := cf.Route{}
	route.Guid = "my-cool-route-guid"
	app := cf.Application{}
	app.Guid = "my-cool-app-guid"

	apiResponse := repo.Unbind(route, app)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestDelete(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/v2/routes/my-cool-route-guid",
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()
	route := cf.Route{}
	route.Guid = "my-cool-route-guid"

	apiResponse := repo.Delete(route)
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func createRoutesRepo(t *testing.T, requests ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo CloudControllerRouteRepository, domainRepo *testapi.FakeDomainRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)
	space_Auto4 := cf.Space{}
	space_Auto4.Guid = "my-space-guid"
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       space_Auto4,
	}

	gateway := net.NewCloudControllerGateway()
	domainRepo = &testapi.FakeDomainRepository{}

	repo = NewCloudControllerRouteRepository(config, gateway, domainRepo)
	return
}
