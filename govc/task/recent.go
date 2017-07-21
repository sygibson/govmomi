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
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/types"
)

type recent struct {
	*flags.DatacenterFlag

	Follow bool
	ref    bool
}

func init() {
	cli.Register("tasks", &recent{})
}

func (cmd *recent) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.DatacenterFlag, ctx = flags.NewDatacenterFlag(ctx)
	cmd.DatacenterFlag.Register(ctx, f)

	f.BoolVar(&cmd.Follow, "f", false, "Follow recent task updates")
	f.BoolVar(&cmd.ref, "i", false, "Print the task managed object reference")
}

func (cmd *recent) Description() string {
	return `Display info for recent tasks.

Examples:
  govc tasks
  govc tasks -f
  govc tasks -f /dc1/host/cluster1`
}

func (cmd *recent) Usage() string {
	return "[PATH]"
}

func (cmd *recent) Process(ctx context.Context) error {
	if err := cmd.DatacenterFlag.Process(ctx); err != nil {
		return err
	}
	return nil
}

func chop(s string) string {
	if len(s) < 40 {
		return s
	}

	return s[:39] + "*"
}

func (cmd *recent) Run(ctx context.Context, f *flag.FlagSet) error {
	c, err := cmd.Client()
	if err != nil {
		return err
	}

	m := c.ServiceContent.TaskManager

	watch := *m

	if f.NArg() == 1 {
		refs, merr := cmd.ManagedObjects(ctx, f.Args())
		if merr != nil {
			return nil
		}
		watch = refs[0]
	}

	v, err := view.NewManager(c).CreateTaskView(ctx, &watch)
	if err != nil {
		return nil
	}

	defer v.Destroy(context.Background())

	v.Follow = cmd.Follow

	stamp := "15:04:05"

	tmpl := "%-40s %-40s %11s %11s %11s %11s %s\n"
	header := []interface{}{"ID", "Target", "Initiator", "Queued", "Started", "Result", "Completed"}
	if cmd.ref {
		header = append([]interface{}{"Task"}, header...)
		tmpl = "%-40s " + tmpl
	}
	fmt.Fprintf(cmd.Out, tmpl, header...)

	var last string

	return v.Collect(ctx, func(tasks []types.TaskInfo) {
		for _, info := range tasks {
			var user string

			switch x := info.Reason.(type) {
			case *types.TaskReasonUser:
				user = x.UserName
			}

			if info.EntityName == "" || user == "" {
				continue
			}

			queued := info.QueueTime.Format(stamp)
			start := "-"
			end := start

			if info.StartTime != nil {
				start = info.StartTime.Format(stamp)
				queued = info.StartTime.Sub(info.QueueTime).String()
			}

			var result string

			if info.CompleteTime == nil {
				result = fmt.Sprintf("%d%%", info.Progress)
			} else {
				result = string(info.State)
				end = fmt.Sprintf("%s (%s)", info.CompleteTime.Format(stamp), info.CompleteTime.Sub(*info.StartTime).String())
			}

			var items []interface{}
			if cmd.ref {
				items = append(items, info.Task.Value)
			}

			items = append(items, chop(info.DescriptionId), chop(info.EntityName), user, queued, start, result, end)
			item := fmt.Sprintf(tmpl, items...)

			if item == last {
				continue
			}
			fmt.Fprint(cmd.Out, item)
			last = item
		}
	})
}
