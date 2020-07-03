/*
Copyright (c) 2018 VMware, Inc. All Rights Reserved.

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

package env

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/vim25/mo"
)

type cd struct {
	*flags.DatacenterFlag
}

func init() {
	//cli.Register("env.cd", &cd{})
}

func (cmd *cd) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.DatacenterFlag, ctx = flags.NewDatacenterFlag(ctx)
	cmd.DatacenterFlag.Register(ctx, f)
}

func (cmd *cd) Usage() string {
	return "PATH"
}

func (cmd *cd) Description() string {
	return `TODO:`
}

func (cmd *cd) Run(ctx context.Context, f *flag.FlagSet) error {
	c, err := cmd.Client()
	if err != nil {
		return err
	}

	finder, err := cmd.Finder()
	if err != nil {
		return err
	}

	pwd := os.Getenv("GOVC_PWD")
	arg := f.Arg(0)
	if pwd != "" {
		if !path.IsAbs(arg) {
			arg = path.Join(pwd, arg)
		}
	}

	objs, err := finder.ManagedObjectList(ctx, arg)
	if err != nil {
		return err
	}

	if len(objs) != 1 {
		return flag.ErrHelp
	}

	var output []string

	vars := map[string]string{
		"Datacenter":                  "GOVC_DATACENTER",
		"ResourcePool":                "GOVC_RESOURCE_POOL",
		"VirtualApp":                  "GOVC_RESOURCE_POOL",
		"HostSystem":                  "GOVC_HOST",
		"OpaqueNetwork":               "GOVC_NETWORK",
		"Network":                     "GOVC_NETWORK",
		"DistributedVirtualPortgroup": "GOVC_NETWORK",
		"ClusterComputeResource":      "GOVC_CLUSTER",
		"VirtualMachine":              "GOVC_VM",
		"Datastore":                   "GOVC_DATASTORE",
	}

	env := map[string]string{
		"GOVC_PWD": objs[0].Path,
	}

	for _, key := range vars {
		env[key] = "" // unset
	}

	entities, err := mo.Ancestors(ctx, c, c.ServiceContent.PropertyCollector, objs[0].Object.Reference())
	if err != nil {
		return err
	}

	for _, e := range entities {
		key, ok := vars[e.Reference().Type]
		if ok {
			env[key] = e.Name
		}
	}

	for key, val := range env {
		output = append(output, fmt.Sprintf("%s='%s'", key, val))
	}

	return cmd.WriteResult(envResult(output))
}
