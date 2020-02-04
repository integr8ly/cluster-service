package sendgrid

//SubUser A SendGrid sub user, from https://sendgrid.com/docs/API_Reference/Web_API_v3/subusers.html
type SubUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Disabled bool   `json:"disabled"`
}

//APIKey A SendGrid API key, from https://sendgrid.com/docs/API_Reference/Web_API_v3/API_Keys/index.html
type APIKey struct {
	ID     string   `json:"api_key_id"`
	Key    string   `json:"api_key"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

//IPAddress A SendGrid IP address, from https://sendgrid.com/docs/API_Reference/Web_API_v3/IP_Management/ip_addresses.html
type IPAddress struct {
	IP        string   `json:"ip"`
	Warmup    bool     `json:"warmup"`
	StartDate int      `json:"start_date"`
	SubUsers  []string `json:"subusers"`
	RDNS      string   `json:"rdns"`
	Pools     []string `json:"pools"`
}
