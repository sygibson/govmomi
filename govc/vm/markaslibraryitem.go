/*
Copyright (c) 2014-2015 VMware, Inc. All Rights Reserved.

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

package vm

import (
	"context"
	"flag"
	"log"

	"github.com/kr/pretty"
	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/internal"
	"github.com/vmware/govmomi/vim25/types"
)

type markaslibraryitem struct {
	*flags.SearchFlag
	*flags.ResourcePoolFlag
}

func init() {
	cli.Register("vm.markaslibraryitem", &markaslibraryitem{})
}

func (cmd *markaslibraryitem) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.SearchFlag, ctx = flags.NewSearchFlag(ctx, flags.SearchVirtualMachines)
	cmd.SearchFlag.Register(ctx, f)

	cmd.ResourcePoolFlag, ctx = flags.NewResourcePoolFlag(ctx)
	cmd.ResourcePoolFlag.Register(ctx, f)
}

func (cmd *markaslibraryitem) Process(ctx context.Context) error {
	if err := cmd.SearchFlag.Process(ctx); err != nil {
		return err
	}
	if err := cmd.ResourcePoolFlag.Process(ctx); err != nil {
		return err
	}
	return nil
}

func (cmd *markaslibraryitem) Usage() string {
	return "VM..."
}

func (cmd *markaslibraryitem) Description() string {
	return `Mark VM as a virtual machine libraryitem.

Examples:
  govc vm.markaslibraryitem $name`
}

func (cmd *markaslibraryitem) Run(ctx context.Context, f *flag.FlagSet) error {
	vms, err := cmd.VirtualMachines(f.Args())
	if err != nil {
		return err
	}

	pool, err := cmd.ResourcePool()
	if err != nil {
		return err
	}

	for _, vm := range vms {
		if true {
			var content []types.ObjectContent
			err = vm.Properties(ctx, vm.Reference(), []string{"config.contentLibItemInfo"}, &content)
			if err != nil {
				return err
			}
			pretty.Printf("%# v\n", content)
			return nil
		}
		req := internal.MarkAsLibraryItem_TaskRequest{
			This: vm.Reference(),
			Pool: pool.Reference(),
			Info: internal.VirtualMachineContentLibraryItemInfo{
				ContentLibraryItemUuid:    "d85429e7-651d-4d21-8601-c18005ebf50f",
				ContentLibraryItemVersion: "version-1",
			},
			SnapshotName:        "govc",
			SnapshotDescription: "Created by govc",
		}

		res, err := internal.MarkAsLibraryItem_Task(ctx, vm.Client(), &req)
		if err != nil {
			return err
		}
		log.Printf("T=%s", res.Returnval)
	}

	return nil
}
