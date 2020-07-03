/*
Copyright (c) 2020 VMware, Inc. All Rights Reserved.

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

package dvs

import (
	"context"
	"flag"
	"fmt"

	"github.com/kr/pretty"
	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
)

type check struct {
	*flags.DatacenterFlag

	path string
}

func init() {
	cli.Register("dvs.check", &check{})
}

func (cmd *check) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.DatacenterFlag, ctx = flags.NewDatacenterFlag(ctx)
	cmd.DatacenterFlag.Register(ctx, f)

	f.StringVar(&cmd.path, "dvs", "", "DVS path")
}

func (cmd *check) Usage() string {
	return "HOST..."
}

func (cmd *check) Description() string {
	return `Check hosts from DVS.

Examples:
  govc dvs.check -dvs dvsName hostA hostB hostC`
}

func (cmd *check) Run(ctx context.Context, f *flag.FlagSet) error {
	if f.NArg() == 0 {
		return flag.ErrHelp
	}

	finder, err := cmd.Finder()
	if err != nil {
		return err
	}

	net, err := finder.Network(ctx, cmd.path)
	if err != nil {
		return err
	}

	dvs, ok := net.(*object.DistributedVirtualSwitch)
	if !ok {
		return fmt.Errorf("%s (%T) is not of type %T", cmd.path, net, dvs)
	}

	c, err := cmd.Client()
	if err != nil {
		return err
	}

	obj, err := cmd.ManagedObject(ctx, f.Arg(0))
	if err != nil {
		return err
	}

	req := types.QueryDvsCheckCompatibility{
		This: *c.ServiceContent.DvSwitchManager,
		HostContainer: types.DistributedVirtualSwitchManagerHostContainer{
			Container: obj.Reference(),
		},
	}

	res, err := methods.QueryDvsCheckCompatibility(ctx, c, &req)
	if err != nil {
		return err
	}

	pretty.Printf("RES=%# v\n", res)
	return nil
}
