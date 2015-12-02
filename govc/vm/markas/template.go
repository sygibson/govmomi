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

type template struct {
	*flags.SearchFlag
}

func init() {
	cli.Register("vm.markas.template", &template{})
}

func (cmd *template) Register(f *flag.FlagSet) {
	cmd.SearchFlag = flags.NewSearchFlag(flags.SearchVirtualMachines)
}

func (cmd *template) Process() error { return nil }

func (cmd *template) Usage() string {
	return "VM..."
}

func (cmd *template) Description() string {
	return `Mark VM as a template.`
}

func (cmd *template) Run(f *flag.FlagSet) error {
	ctx := context.TODO()

	vms, err := cmd.VirtualMachines(f.Args())
	if err != nil {
		return err
	}

	for _, vm := range vms {
		err := vm.MarkAsTemplate(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
