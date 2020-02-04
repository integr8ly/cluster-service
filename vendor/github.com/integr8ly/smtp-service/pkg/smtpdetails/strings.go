package smtpdetails

const (
	//LogFieldDetailProvider Logging key for specifying the SMTP details provider e.g. SendGrid
	LogFieldDetailProvider = "smtp_service_detail_provider"
	//SecretKeyHost Default secret data key for SMTP host
	SecretKeyHost = "host"
	//SecretKeyPort Default secret data key for SMTP port
	SecretKeyPort = "port"
	//SecretKeyTLS Default secret data key for SMTP TLS
	SecretKeyTLS = "tls"
	//SecretKeyUsername Default secret data key for SMTP auth username
	SecretKeyUsername = "username"
	//SecretKeyPassword Default secret data key for SMTP auth password
	SecretKeyPassword = "password"
	//SecretGVKKind GVK Kind of an OpenShift/Kubernetes Secret
	SecretGVKKind = "Secret"
	//SecretGVKVersion GVK Version of an OpenShift/Kubernetes Secret
	SecretGVKVersion = "v1"
)
