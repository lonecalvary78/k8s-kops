/*
Copyright 2019 The Kubernetes Authors.

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

package designate

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"k8s.io/klog/v2"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	// ProviderName is the name of this DNS provider
	ProviderName = "openstack-designate"
)

func init() {
	dnsprovider.RegisterDNSProvider(ProviderName, func(config io.Reader) (dnsprovider.Interface, error) {
		return newDesignate(config)
	})
}

func newDesignate(_ io.Reader) (*Interface, error) {
	oc := vfs.OpenstackConfig{}
	region, err := oc.GetRegion()
	if err != nil {
		return nil, fmt.Errorf("error finding openstack region: %v", err)
	}

	ao, err := oc.GetCredential()
	if err != nil {
		return nil, err
	}

	provider, err := openstack.NewClient(ao.IdentityEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building openstack provider client: %v", err)
	}
	ua := gophercloud.UserAgent{}
	ua.Prepend("kops/designate")
	provider.UserAgent = ua
	klog.V(4).Infof("Using user-agent %s", ua.Join())

	tlsconfig := &tls.Config{}
	tlsconfig.InsecureSkipVerify = true
	transport := &http.Transport{TLSClientConfig: tlsconfig}
	provider.HTTPClient = http.Client{
		Transport: transport,
	}

	klog.V(2).Info("authenticating to keystone")

	err = openstack.Authenticate(context.TODO(), provider, ao)
	if err != nil {
		return nil, fmt.Errorf("error building openstack authenticated client: %v", err)
	}

	sc, err := openstack.NewDNSV2(provider, gophercloud.EndpointOpts{
		Type:   "dns",
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating a ServiceClient: %v", err)
	}
	return New(sc), nil
}
