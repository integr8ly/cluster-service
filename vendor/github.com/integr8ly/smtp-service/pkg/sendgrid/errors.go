package sendgrid

//AlreadyExistsError Error to indicate an API key already exists
type AlreadyExistsError struct {
	Message string
}

//Error String representation of error
func (e *AlreadyExistsError) Error() string {
	return e.Message
}

//IsAlreadyExistsError Compare check for AlreadyExistsError
func IsAlreadyExistsError(err error) bool {
	_, ok := err.(*AlreadyExistsError)
	return ok
}

//NotExistError Error to indicate an API key does not exist
type NotExistError struct {
	Message string
}

//Error String representation of error
func (e *NotExistError) Error() string {
	return e.Message
}

//IsNotExistError Compare check for NotExistError
func IsNotExistError(err error) bool {
	_, ok := err.(*NotExistError)
	return ok
}
