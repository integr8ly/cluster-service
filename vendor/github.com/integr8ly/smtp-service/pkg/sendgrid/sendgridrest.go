package sendgrid

import (
	"github.com/sendgrid/rest"
	sg "github.com/sendgrid/sendgrid-go"
	"github.com/sirupsen/logrus"
)

var _ RESTClient = &BackendRESTClient{}

//RESTClient Thin wrapper around the SendGrid package
//go:generate moq -out sendgridrest_moq.go . RESTClient
type RESTClient interface {
	BuildRequest(endpoint string, method rest.Method) rest.Request
	InvokeRequest(request rest.Request) (*rest.Response, error)
}

//BackendRESTClient Thin wrapper around the SendGrid library
type BackendRESTClient struct {
	apiHost string
	apiKey  string
	logger  *logrus.Entry
}

//NewBackendRESTClient Create a new BackendAPIClient with default logger labels
func NewBackendRESTClient(apiHost, apiKey string, logger *logrus.Entry) *BackendRESTClient {
	return &BackendRESTClient{
		apiHost: apiHost,
		apiKey:  apiKey,
		logger:  logger.WithField(LogFieldAPIClient, ProviderName),
	}
}

//BuildRequest Create a REST request that can be sent through the SendGrid API
func (c *BackendRESTClient) BuildRequest(endpoint string, method rest.Method) rest.Request {
	c.logger.Debugf("getting request with details, key=%s endpoint=%s host=%s", c.apiKey, endpoint, c.apiHost)
	req := sg.GetRequest(c.apiKey, endpoint, c.apiHost)
	req.Method = method
	return req
}

//InvokeRequest Invoke a REST request against the SendGrid API
func (c *BackendRESTClient) InvokeRequest(request rest.Request) (*rest.Response, error) {
	c.logger.Debugf("performing api request with details, url=%s method=%s body=%s", request.BaseURL, request.Method, string(request.Body))
	return sg.API(request)
}
