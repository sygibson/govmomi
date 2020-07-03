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

package alarm

import (
	"context"
	"flag"
	"fmt"
	"text/tabwriter"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type info struct {
	*flags.DatacenterFlag

	triggered bool
	declared  bool
}

func init() {
	cli.Register("alarm.info", &info{})
}

func (cmd *info) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.DatacenterFlag, ctx = flags.NewDatacenterFlag(ctx)
	cmd.DatacenterFlag.Register(ctx, f)

	f.BoolVar(&cmd.triggered, "t", true, "Show triggered alarms")
	f.BoolVar(&cmd.declared, "d", false, "Show declared alarms")
}

func (cmd *info) Usage() string {
	return "[PATH]..."
}

func (cmd *info) Description() string {
	return `Alarm info for managed objects.

Examples:
  govc alarm.info /dc1/host/cluster1`
}

func (cmd *info) Process(ctx context.Context) error {
	if err := cmd.DatacenterFlag.Process(ctx); err != nil {
		return err
	}
	return nil
}

func (cmd *info) Run(ctx context.Context, f *flag.FlagSet) error {
	c, err := cmd.Client()
	if err != nil {
		return err
	}

	finder, err := cmd.Finder()
	if err != nil {
		return err
	}

	var props []string

	if cmd.triggered {
		props = append(props, "triggeredAlarmState")
	}
	if cmd.declared {
		props = append(props, "declaredAlarmState")
	}

	var objs []types.ManagedObjectReference

	paths := map[types.ManagedObjectReference]string{
		c.ServiceContent.RootFolder.Reference(): "/",
	}

	if f.NArg() == 0 {
		objs = append(objs, c.ServiceContent.RootFolder)
	} else {
		elts, ferr := finder.ManagedObjectList(ctx, f.Arg(0))
		if ferr != nil {
			return ferr
		}

		for _, e := range elts {
			ref := e.Object.Reference()
			paths[ref] = e.Path
			objs = append(objs, ref)
		}
	}

	var entites []mo.ManagedEntity

	pc := property.DefaultCollector(c)
	err = pc.Retrieve(ctx, objs, props, &entites)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(cmd.Out, 2, 0, 2, ' ', 0)

	for _, e := range entites {
		objs = nil

		alarms := e.DeclaredAlarmState
		alarms = append(alarms, e.TriggeredAlarmState...)
		for i := range alarms {
			objs = append(objs, alarms[i].Alarm)
		}

		if len(objs) == 0 {
			continue
		}

		var info []mo.Alarm

		err = pc.Retrieve(ctx, objs, []string{"info"}, &info)
		if err != nil {
			return err
		}

		for i, alarm := range alarms {
			p, ok := paths[alarm.Entity]
			if !ok {
				e, err := finder.Element(ctx, alarm.Entity)
				if err != nil {
					return err
				}
				p = e.Path
				paths[alarm.Entity] = p
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\n", alarm.OverallStatus, p, info[i].Info.Name)
		}
	}

	return tw.Flush()
}
