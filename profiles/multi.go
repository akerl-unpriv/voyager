package profiles

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

// MultiStore is a storage backend which tries a series of backends
type MultiStore struct {
	Backends []Store
}

// Lookup looks up creds from the list of backends
func (m *MultiStore) Lookup(profile string) (credentials.Value, error) {
	var err error
	var writer WritableStore
	var creds credentials.Value

	for _, item := range m.Backends {
		creds, err = item.Lookup(profile)
		if err == nil {
			break
		}
		if writer == nil {
			writer = item.(WritableStore)
		}
		logger.DebugMsg(fmt.Sprintf("backend failed with error: %s", err))
	}
	if creds.AccessKeyID == "" {
		return credentials.Value{}, fmt.Errorf("all backends failed to return creds")
	}

	if writer != nil {
		err := writer.Write(profile, creds)
		if err != nil {
			return credentials.Value{}, err
		}
	}

	return creds, nil
}

// Check returns true if any backend has the credentials cached
func (m *MultiStore) Check(profile string) bool {
	for _, item := range m.Backends {
		if item.Check(profile) {
			return true
		}
	}
	return false
}

// WritableStore defines a backend which can save credentials
type WritableStore interface {
	Write(string, credentials.Value) error
}
