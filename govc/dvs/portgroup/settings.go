/*
Copyright (c) 2019 VMware, Inc. All Rights Reserved.

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

package portgroup

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type settings struct {
	*flags.DatacenterFlag
}

func init() {
	cli.Register("dvs.portgroup.settings", &settings{})
}

func (cmd *settings) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.DatacenterFlag, ctx = flags.NewDatacenterFlag(ctx)
	cmd.DatacenterFlag.Register(ctx, f)
}

func (cmd *settings) Usage() string {
	return "PATH"
}

func (cmd *settings) Run(ctx context.Context, f *flag.FlagSet) error {
	finder, err := cmd.Finder()
	if err != nil {
		return err
	}

	net, err := finder.Network(ctx, f.Arg(0))
	if err != nil {
		return err
	}

	o, ok := net.(*object.DistributedVirtualPortgroup)
	if !ok {
		return fmt.Errorf("%s (%s) is not a DVPG", f.Arg(0), net.Reference().Type)
	}

	var pg mo.DistributedVirtualPortgroup
	err = o.Properties(ctx, o.Reference(), []string{"config.defaultPortConfig"}, &pg)
	if err != nil {
		return err
	}

	return cmd.WriteResult(&settingsResult{pg.Config.DefaultPortConfig.(*types.VMwareDVSPortSetting)})
}

type settingsResult struct {
	*types.VMwareDVSPortSetting
}

func (r *settingsResult) Dump() interface{} {
	return r.VMwareDVSPortSetting
}

func allow(b *types.BoolPolicy) string {
	if b.Value != nil && *b.Value {
		return "Accept"
	}
	return "Reject"
}

func boolPolicy(b *types.BoolPolicy) bool {
	if b.Value != nil && *b.Value {
		return true
	}
	return false
}

var enabled = map[bool]string{
	true:  "Yes",
	false: "No",
}

func join(s []string) string {
	return strings.Join(s, ",")
}

var teamingPolicyMode = map[types.DistributedVirtualSwitchNicTeamingPolicyMode]string{
	types.DistributedVirtualSwitchNicTeamingPolicyModeLoadbalance_ip:        "Route based on IP hash",
	types.DistributedVirtualSwitchNicTeamingPolicyModeLoadbalance_srcmac:    "Route based on source MAC hash",
	types.DistributedVirtualSwitchNicTeamingPolicyModeLoadbalance_srcid:     "Route based on originating virtual port",
	types.DistributedVirtualSwitchNicTeamingPolicyModeFailover_explicit:     "Use explicit failover order",
	types.DistributedVirtualSwitchNicTeamingPolicyModeLoadbalance_loadbased: "Route based on physical NIC load",
}

func teamingPolicy(p string) string {
	return fmt.Sprintf("%s (%s)", p, teamingPolicyMode[types.DistributedVirtualSwitchNicTeamingPolicyMode(p)])
}

func check(f *types.DVSFailureCriteria) string {
	if f.CheckBeacon.Value != nil && *f.CheckBeacon.Value {
		return "Beacon probing"
	}
	return "Link status only"
}

func (r *settingsResult) Write(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	s := r.VMwareDVSPortSetting
	fmt.Fprintln(tw, "Security:")
	fmt.Fprintf(tw, "  Allow promiscuous mode:\t%s\n", allow(s.SecurityPolicy.AllowPromiscuous))
	fmt.Fprintf(tw, "  Allow forged transmits:\t%s\n", allow(s.SecurityPolicy.ForgedTransmits))
	fmt.Fprintf(tw, "  Allow MAC changes:\t%s\n", allow(s.SecurityPolicy.MacChanges))

	fmt.Fprintln(tw, "\nTeaming and failover:")
	fmt.Fprintf(tw, "  Load balancing:\t%s\n", teamingPolicy(s.UplinkTeamingPolicy.Policy.Value))
	fmt.Fprintf(tw, "  Network failure detection:\t%s\n", check(s.UplinkTeamingPolicy.FailureCriteria))
	fmt.Fprintf(tw, "  Notify switches:\t%s\n", enabled[boolPolicy(s.UplinkTeamingPolicy.NotifySwitches)])
	fmt.Fprintf(tw, "  Failback:\t%s\n", enabled[!boolPolicy(s.UplinkTeamingPolicy.RollingOrder)])
	fmt.Fprintf(tw, "  Active uplinks:\t%s\n", join(s.UplinkTeamingPolicy.UplinkPortOrder.ActiveUplinkPort))
	fmt.Fprintf(tw, "  Standby uplinks:\t%s\n", join(s.UplinkTeamingPolicy.UplinkPortOrder.StandbyUplinkPort))

	return tw.Flush()
}
