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

func TestCreateServiceBinding(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "POST",
		Path:     "/v2/service_bindings",
		Matcher:  testnet.RequestBodyMatcher(`{"app_guid":"my-app-guid","service_instance_guid":"my-service-instance-guid"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createServiceBindingRepo(t, req)
	defer ts.Close()
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Guid = "my-service-instance-guid"
	app := cf.Application{}
	app.Guid = "my-app-guid"
	apiResponse := repo.Create(serviceInstance, app)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestCreateServiceBindingIfError(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "POST",
		Path:    "/v2/service_bindings",
		Matcher: testnet.RequestBodyMatcher(`{"app_guid":"my-app-guid","service_instance_guid":"my-service-instance-guid"}`),
		Response: testnet.TestResponse{
			Status: http.StatusBadRequest,
			Body:   `{"code":90003,"description":"The app space binding to service is taken: 7b959018-110a-4913-ac0a-d663e613cdea 346bf237-7eef-41a7-b892-68fb08068f09"}`,
		},
	})

	ts, handler, repo := createServiceBindingRepo(t, req)
	defer ts.Close()
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Guid = "my-service-instance-guid"
	app := cf.Application{}
	app.Guid = "my-app-guid"
	apiResponse := repo.Create(serviceInstance, app)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, apiResponse.ErrorCode, "90003")
}

var deleteBindingReq = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method:   "DELETE",
	Path:     "/v2/service_bindings/service-binding-2-guid",
	Response: testnet.TestResponse{Status: http.StatusOK},
})

func TestDeleteServiceBinding(t *testing.T) {
	ts, handler, repo := createServiceBindingRepo(t, deleteBindingReq)
	defer ts.Close()
	binding_Auto := cf.ServiceBinding{}
	binding_Auto.Url = "/v2/service_bindings/service-binding-1-guid"
	binding_Auto.AppGuid = "app-1-guid"
	binding_Auto2 := cf.ServiceBinding{}
	binding_Auto2.Url = "/v2/service_bindings/service-binding-2-guid"
	binding_Auto2.AppGuid = "app-2-guid"
	serviceBindings := []cf.ServiceBinding{binding_Auto, binding_Auto2}
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Guid = "my-service-instance-guid"
	serviceInstance.ServiceBindings = serviceBindings
	app := cf.Application{}
	app.Guid = "app-2-guid"
	found, apiResponse := repo.Delete(serviceInstance, app)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.True(t, found)
}

func TestDeleteServiceBindingWhenBindingDoesNotExist(t *testing.T) {
	ts, handler, repo := createServiceBindingRepo(t, deleteBindingReq)
	defer ts.Close()

	serviceBindings := []cf.ServiceBinding{}
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Guid = "my-service-instance-guid"
	serviceInstance.ServiceBindings = serviceBindings
	app := cf.Application{}
	app.Guid = "app-2-guid"
	found, apiResponse := repo.Delete(serviceInstance, app)

	assert.False(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.False(t, found)
}

func createServiceBindingRepo(t *testing.T, req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceBindingRepository) {
	ts, handler = testnet.NewTLSServer(t, []testnet.TestRequest{req})
	space_Auto := cf.Space{}
	space_Auto.Guid = "my-space-guid"
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Space:       space_Auto,
		Target:      ts.URL,
	}

	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerServiceBindingRepository(config, gateway)
	return
}
