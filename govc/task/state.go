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

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
)

type state struct {
	*flags.ClientFlag

	state    string
	progress int
}

func init() {
	cli.Register("task.state.set", &state{})
}

func (cmd *state) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.ClientFlag, ctx = flags.NewClientFlag(ctx)
	cmd.ClientFlag.Register(ctx, f)

	f.IntVar(&cmd.progress, "p", 0, "Set task progress percent")
}

func (cmd *state) Description() string {
	return `Set task state.

Examples:
  govc TODO:`
}

func (cmd *state) Usage() string {
	return "ID [STATE]"
}

func (cmd *state) Process(ctx context.Context) error {
	if err := cmd.ClientFlag.Process(ctx); err != nil {
		return err
	}
	return nil
}

func (cmd *state) Run(ctx context.Context, f *flag.FlagSet) error {
	c, err := cmd.Client()
	if err != nil {
		return err
	}

	this := mor(f.Arg(0))

	if cmd.progress > 0 {
		_, err = methods.UpdateProgress(ctx, c, &types.UpdateProgress{
			This:        this,
			PercentDone: int32(cmd.progress),
		})

		if err != nil {
			return err
		}
	}

	if f.NArg() < 2 {
		return nil
	}

	_, err = methods.SetTaskState(ctx, c, &types.SetTaskState{
		This:   this,
		State:  types.TaskInfoState(f.Arg(1)),
		Result: nil, // TODO:
		Fault:  nil, // TODO:
	})

	return err
}
