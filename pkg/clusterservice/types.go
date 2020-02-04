package clusterservice

//Action Descriptor of an action
type Action string

//ActionStatus Descriptor of the current status of an action
type ActionStatus string

const (
	//ActionDelete Deletion would be performed
	ActionDelete Action = "delete"
	//ActionStatusDryRun Action will not be performed
	ActionStatusDryRun ActionStatus = "dry run"
	//ActionStatusInProgress Action is being performed currently
	ActionStatusInProgress ActionStatus = "in progress"
	//ActionStatusEmpty Blank status of action
	c ActionStatus = ""
)

//Report Information about what resources are found in the AWS account related to the cluster
type Report struct {
	Items []*ReportItem
}

//ReportItem Information about a specific AWS resource
type ReportItem struct {
	ID           string
	Name         string
	Action       Action
	ActionStatus ActionStatus
}
