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

package task

import (
	"context"
	"flag"
	"fmt"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
)

type create struct {
	*flags.ClientFlag

	types.CreateTask
	obj string
}

func init() {
	cli.Register("task.create", &create{})
}

func (cmd *create) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.ClientFlag, ctx = flags.NewClientFlag(ctx)
	cmd.ClientFlag.Register(ctx, f)

	f.StringVar(&cmd.obj, "o", "/", "ManagedObject with which Task will be associated")
	f.BoolVar(&cmd.Cancelable, "c", true, "True if the task should be cancelable, false otherwise")
	f.StringVar(&cmd.InitiatedBy, "user", "", "The name of the user on whose behalf the Extension is creating the task")
}

func (cmd *create) Description() string {
	return `Create tasks.

Examples:
  govc task.create foo.Bar`
}

func (cmd *create) Usage() string {
	return "ID"
}

func (cmd *create) Process(ctx context.Context) error {
	if err := cmd.ClientFlag.Process(ctx); err != nil {
		return err
	}
	return nil
}

func (cmd *create) Run(ctx context.Context, f *flag.FlagSet) error {
	if f.NArg() != 1 {
		return flag.ErrHelp
	}

	c, err := cmd.Client()
	if err != nil {
		return err
	}

	cmd.This = *c.ServiceContent.TaskManager
	cmd.TaskTypeId = f.Arg(0)

	if cmd.obj == "/" {
		cmd.Obj = c.ServiceContent.RootFolder
	} else {
		if !cmd.Obj.FromString(cmd.obj) {
			return fmt.Errorf("invalid object id: %s", cmd.obj) // TODO: inventory path
		}
	}

	res, err := methods.CreateTask(ctx, c, &cmd.CreateTask)
	if err != nil {
		return err
	}

	fmt.Println(res.Returnval.Task.String())

	return nil
}
