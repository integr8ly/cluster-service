package errors

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

//WrapLog Wrap an existing error with a message and log the provided message
func WrapLog(err error, msg string, logger *logrus.Entry) error {
	logger.Error(msg)
	return fmt.Errorf("%s: %w", msg, err)
}
