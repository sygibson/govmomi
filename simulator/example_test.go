/*
Copyright (c) 2017 VMware, Inc. All Rights Reserved.

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

package simulator_test

import (
	"context"
	"fmt"
	"log"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

// Example boilerplate for starting a simulator initialized with an ESX model.
func ExampleESX() {
	ctx := context.Background()

	// ESXi model + initial set of objects (VMs, network, datastore)
	model := simulator.ESX()

	defer model.Remove()
	err := model.Create()
	if err != nil {
		log.Fatal(err)
	}

	s := model.Service.NewServer()
	defer s.Close()

	c, _ := govmomi.NewClient(ctx, s.URL, true)

	fmt.Printf("%s with %d host", c.Client.ServiceContent.About.ApiType, model.Count().Host)
	// Output: HostAgent with 1 host
}

// Example for starting a simulator with empty inventory, similar to a fresh install of vCenter.
func ExampleModel() {
	ctx := context.Background()

	model := simulator.VPX()
	model.Datacenter = 0 // No DC == no inventory

	defer model.Remove()
	err := model.Create()
	if err != nil {
		log.Fatal(err)
	}

	s := model.Service.NewServer()
	defer s.Close()

	c, _ := govmomi.NewClient(ctx, s.URL, true)

	fmt.Printf("%s with %d hosts", c.Client.ServiceContent.About.ApiType, model.Count().Host)
	// Output: VirtualCenter with 0 hosts
}

// Example boilerplate for starting a simulator initialized with a vCenter model.
func ExampleVPX() {
	ctx := context.Background()

	// vCenter model + initial set of objects (cluster, hosts, VMs, network, datastore, etc)
	model := simulator.VPX()

	defer model.Remove()
	err := model.Create()
	if err != nil {
		log.Fatal(err)
	}

	s := model.Service.NewServer()
	defer s.Close()

	c, _ := govmomi.NewClient(ctx, s.URL, true)

	fmt.Printf("%s with %d hosts", c.Client.ServiceContent.About.ApiType, model.Count().Host)
	// Output: VirtualCenter with 4 hosts
}

// Run simplifies startup/cleanup of a simulator instance for example or testing purposes.
func ExampleModel_Run() {
	err := simulator.VPX().Run(func(ctx context.Context, c *vim25.Client) error {
		// Client has connected and logged in to a new simulator instance.
		// Server.Close and Model.Remove are called when this func returns.
		s, err := session.NewManager(c).UserSession(ctx)
		if err != nil {
			return err
		}
		fmt.Print(s.UserName)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	// Output: user
}

// Test simplifies startup/cleanup of a simulator instance for testing purposes.
func ExampleTest() {
	simulator.Test(func(ctx context.Context, c *vim25.Client) {
		// Client has connected and logged in to a new simulator instance.
		// Server.Close and Model.Remove are called when this func returns.
		s, err := session.NewManager(c).UserSession(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(s.UserName)
	})
	// Output: user
}

type PropertyCollectorRetrievePropertiesOverride struct {
	simulator.PropertyCollector
}

// RetrieveProperties overrides simulator.PropertyCollector.RetrieveProperties, returning a custom value for the "config.description" field
func (pc *PropertyCollectorRetrievePropertiesOverride) RetrieveProperties(ctx *simulator.Context, req *types.RetrieveProperties) soap.HasFault {
	body := &methods.RetrievePropertiesBody{
		Res: &types.RetrievePropertiesResponse{
			Returnval: []types.ObjectContent{
				{
					Obj: req.SpecSet[0].ObjectSet[0].Obj,
					PropSet: []types.DynamicProperty{
						{
							Name: "config.description",
							Val:  "This property overridden by vcsim test",
						},
					},
				},
			},
		},
	}

	for _, spec := range req.SpecSet {
		for _, prop := range spec.PropSet {
			for _, path := range prop.PathSet {
				if path == "config.description" {
					return body
				}
			}
		}
	}

	return pc.PropertyCollector.RetrieveProperties(ctx, req)
}

func ExamplePropertyCollector_RetrieveProperties() {
	simulator.Test(func(ctx context.Context, c *vim25.Client) {
		ref := simulator.Map.Any("DistributedVirtualPortgroup").Reference()

		pc := new(PropertyCollectorRetrievePropertiesOverride)
		pc.Self = c.ServiceContent.PropertyCollector
		simulator.Map.Put(pc)

		pg := object.NewDistributedVirtualPortgroup(c, ref)

		var content []types.ObjectContent

		err := pg.Properties(ctx, pg.Reference(), []string{"config.description"}, &content)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(content[0].PropSet[0].Val)
	})
	// Output: This property overridden by vcsim test
}
