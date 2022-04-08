// Package l wraps logging.
package l

import (
	"github.com/KarelKubat/smartlog/client"
	"github.com/KarelKubat/smartlog/client/any"
)

// SetOutput sets the logging destination. See https://github.com/KarelKubat/smartlog.
func SetOutput(o string) (err error) {
	client.DefaultClient, err = any.New(o)
	return err
}

// Infof emits an informational message.
func Infof(msg string, args ...interface{}) error {
	return client.DefaultClient.Infof(msg, args...)
}

// Warnf emits a warning.
func Warnf(msg string, args ...interface{}) error {
	return client.DefaultClient.Warnf(msg, args...)
}

// Fatalf emits an error message and halts.
func Fatalf(msg string, args ...interface{}) error {
	return client.DefaultClient.Fatalf(msg, args...)
}
