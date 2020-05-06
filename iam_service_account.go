package main

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iam/v1"
)

type IamServiceAccountClient struct {
	service *iam.Service
	ctx     *context.Context
}

func NewIamClient(ctx context.Context) *IamServiceAccountClient {
	service, err := iam.NewService(ctx)
	if err != nil {
		log.Fatalf("iam.NewService: %v", err)
	}
	return &IamServiceAccountClient{service: service}
}

// createKey creates a service account key.
func (i *IamServiceAccountClient) createKey(serviceAccountEmail string) (*iam.ServiceAccountKey, error) {
	resource := "projects/-/serviceAccounts/" + serviceAccountEmail
	request := &iam.CreateServiceAccountKeyRequest{}
	key, err := i.service.Projects.ServiceAccounts.Keys.Create(resource, request).Do()
	if err != nil {
		return nil, fmt.Errorf("Projects.ServiceAccounts.Keys.Create: %v", err)
	}
	log.Infof("Created key: %v", key.Name)
	return key, nil
}

// deleteKey deletes a service account key.
func (i *IamServiceAccountClient) deleteKey(fullKeyName string) error {
	var err error
	_, err = i.service.Projects.ServiceAccounts.Keys.Delete(fullKeyName).Do()
	if err != nil {
		return fmt.Errorf("Projects.ServiceAccounts.Keys.Delete: %v", err)
	}
	log.Infof("Deleted key: %v", fullKeyName)
	return nil
}

// listKey lists a service account's keys.
func (i *IamServiceAccountClient) listKeys(serviceAccountEmail string) ([]*iam.ServiceAccountKey, error) {
	resource := "projects/-/serviceAccounts/" + serviceAccountEmail
	response, err := i.service.Projects.ServiceAccounts.Keys.List(resource).Do()
	if err != nil {
		return nil, fmt.Errorf("Projects.ServiceAccounts.Keys.List: %v", err)
	}
	return response.Keys, nil
}

// rotateKey, and remove old keys if exists
func (i *IamServiceAccountClient) rotateKey(serviceAccountEmail string) (*iam.ServiceAccountKey, error) {
	keys, err := i.listKeys(serviceAccountEmail)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		log.Debugf("service account key: (%v) ValidAfterTime: (%v)  ValidBeforeTime: (%v)", key.Name, key.ValidAfterTime, key.ValidBeforeTime)
		if i.isSystemMangedKey(key) {
			continue
		}

		err = i.deleteKey(key.Name)
		if err != nil {
			return nil, err
		}
	}

	key, err := i.createKey(serviceAccountEmail)
	if err != nil {
		return nil, err
	}
	return key, nil

}

func (i *IamServiceAccountClient) isSystemMangedKey(key *iam.ServiceAccountKey) bool {
	/*
		user managed keys have a valid before time of forever
		ValidAfterTime: (2020-05-05T13:34:26Z)  ValidBeforeTime: (9999-12-31T23:59:59Z)

		system managed keys have a valid before time less then two years + a few days
		ValidAfterTime: (2020-05-04T13:18:54Z)  ValidBeforeTime: (2022-05-08T04:58:36Z)
	*/
	beforeTs, err := time.Parse(time.RFC3339, key.ValidBeforeTime)
	if err != nil {
		return false
	}

	return beforeTs.Year() != 9999
}

func (i *IamServiceAccountClient) keyToDelete(keys []*iam.ServiceAccountKey) []*iam.ServiceAccountKey {
	maxKeys := 3
	var toDelete []*iam.ServiceAccountKey

	if len(keys) <= maxKeys {
		return toDelete
	}
	// create a list of keys that is ordered by oldest to latest valid after time.

	for _, key := range keys {
		log.Debugf("service account key: (%v) ValidAfterTime: (%v)  ValidBeforeTime: (%v)", key.Name, key.ValidAfterTime, key.ValidBeforeTime)
		if i.isSystemMangedKey(key) {
			continue
		}

	}

	return toDelete
}
