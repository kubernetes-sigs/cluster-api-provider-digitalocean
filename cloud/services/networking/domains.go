/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package networking

import (
	"fmt"
	"net/http"

	"github.com/digitalocean/godo"
)

// GetDomainRecord retrieves a single domain record from DO.
func (s *Service) GetDomainRecord(domain, name, rType string) (*godo.DomainRecord, error) {
	fqdn := fmt.Sprintf("%s.%s", name, domain)
	records, resp, err := s.scope.Domains.RecordsByTypeAndName(s.ctx, domain, rType, fqdn, nil)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	switch len(records) {
	case 0:
		return nil, nil
	case 1:
		return &records[0], nil
	default:
		return nil, fmt.Errorf("multiple DNS records (%d) found for '%s.%s' type %s",
			len(records), name, domain, rType)
	}
}

// UpsertDomainRecord creates or updates a DO domain record.
func (s *Service) UpsertDomainRecord(domain, name, rType, data string) error {
	record, err := s.GetDomainRecord(domain, name, rType)
	if err != nil {
		return fmt.Errorf("unable to get current DNS record from API: %s", err)
	}
	recordReq := &godo.DomainRecordEditRequest{
		Type: rType,
		Name: name,
		Data: data,
		TTL:  30,
	}
	if record == nil {
		_, _, err = s.scope.Domains.CreateRecord(s.ctx, domain, recordReq)
	} else {
		_, _, err = s.scope.Domains.EditRecord(s.ctx, domain, record.ID, recordReq)
	}
	return err
}

// DeleteDomainRecord removes a DO domain record.
func (s *Service) DeleteDomainRecord(domain, name, rType string) error {
	record, err := s.GetDomainRecord(domain, name, rType)
	if err != nil {
		return fmt.Errorf("unable to get current DNS record from API: %s", err)
	}
	if record == nil {
		return nil
	}
	_, err = s.scope.Domains.DeleteRecord(s.ctx, domain, record.ID)
	return err
}
