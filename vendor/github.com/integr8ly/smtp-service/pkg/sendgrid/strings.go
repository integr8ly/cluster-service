package sendgrid

const (
	//ProviderName Standardised name of the SendGrid provider
	ProviderName = "sendgrid"
	//EnvAPIKey Name of the env var to retrieve the SendGrid API key
	EnvAPIKey = "SENDGRID_API_KEY"
	//APIHost SendGrid API default host
	APIHost = "https://api.sendgrid.com"
	//APIRouteSubUsers SendGrid v3 API endpoint for sub user management
	APIRouteSubUsers = "/v3/subusers"
	//APIRouteAPIKeys SendGrid v3 API endpoint for api key management
	APIRouteAPIKeys = "/v3/api_keys"
	//APIRouteIPAddresses SendGrid v3 API endpoint for ip address management
	APIRouteIPAddresses = "/v3/ips"
	//HeaderOnBehalfOf SendGrid v3 header for declaring an action is on behalf of a sub user
	HeaderOnBehalfOf = "on-behalf-of"
	//LogFieldAPIClient Logging field name for a description of the API client
	LogFieldAPIClient = "sendgrid_service_api_client"
	//ConnectionDetailsHost Default SendGrid host
	ConnectionDetailsHost = "smtp.sendgrid.net"
	//ConnectionDetailsPort Default SendGrid port
	ConnectionDetailsPort = 587
	//ConnectionDetailsTLS Default SendGrid TLS setting
	ConnectionDetailsTLS = true
	//ConnectionDetailsUsername Default SendGrid SMTP auth username
	ConnectionDetailsUsername = "apikey"
)
