/*
Copyright 2017 The Kubernetes Authors.

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

package gce

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"k8s.io/klog/v2"
	"k8s.io/kops/protokube/pkg/gossip"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

type SeedProvider struct {
	computeClient *gophercloud.ServiceClient
	projectID     string
	clusterName   string
}

var _ gossip.SeedProvider = &SeedProvider{}

func (p *SeedProvider) GetSeeds() ([]string, error) {
	var seeds []string

	err := servers.List(p.computeClient, servers.ListOpts{
		TenantID: p.projectID,
	}).EachPage(context.TODO(), func(ctx context.Context, page pagination.Page) (bool, error) {
		var s []servers.Server
		err := servers.ExtractServersInto(page, &s)
		if err != nil {
			return false, err
		}

		for _, server := range s {
			if clusterName, ok := server.Metadata[openstack.TagClusterName]; ok {
				// verify that the instance is from the same cluster
				if clusterName != p.clusterName {
					continue
				}

				var err error
				// find kopsNetwork from metadata, fallback to clustername
				ifName := clusterName
				if val, ok := server.Metadata[openstack.TagKopsNetwork]; ok {
					ifName = val
				}
				addr, err := openstack.GetServerFixedIP(&server, ifName)
				if err != nil {
					klog.Warningf("Failed to list seeds: %v", err)
					continue
				}
				seeds = append(seeds, addr)
			}
		}
		return true, nil
	})
	if err != nil {
		return seeds, fmt.Errorf("Failed to list servers while retrieving seeds: %v", err)
	}

	return seeds, nil
}

func NewSeedProvider(computeClient *gophercloud.ServiceClient, clusterName string, projectID string) (*SeedProvider, error) {
	return &SeedProvider{
		computeClient: computeClient,
		clusterName:   clusterName,
		projectID:     projectID,
	}, nil
}
