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
	"flag"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"golang.org/x/net/context"
)

type vm struct {
	*flags.SearchFlag
	*flags.HostSystemFlag
	*flags.ResourcePoolFlag
}

func init() {
	cli.Register("vm.markas.vm", &vm{})
}

func (cmd *vm) Register(f *flag.FlagSet) {
	cmd.SearchFlag = flags.NewSearchFlag(flags.SearchVirtualMachines)
}

func (cmd *vm) Process() error { return nil }

func (cmd *vm) Usage() string {
	return "VM..."
}

func (cmd *vm) Description() string {
	return `Mark VM template as a virtual machine.
govc vm.markas.vm -pool Resources -host h1 foo
`
}

func (cmd *vm) Run(f *flag.FlagSet) error {
	ctx := context.TODO()

	vms, err := cmd.VirtualMachines(f.Args())
	if err != nil {
		return err
	}

	pool, err := cmd.ResourcePool()
	if err != nil {
		return err
	}

	host, err := cmd.HostSystemIfSpecified()
	if err != nil {
		return err
	}

	for _, vm := range vms {
		err := vm.MarkAsVirtualMachine(ctx, *pool, host)
		if err != nil {
			return err
		}
	}

	return nil
}
