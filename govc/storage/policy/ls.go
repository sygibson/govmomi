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

package policy

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/pbm"
	"github.com/vmware/govmomi/pbm/types"
)

type ls struct {
	*flags.ClientFlag
	*flags.OutputFlag
}

func init() {
	cli.Register("storage.policy.ls", &ls{})
}

func (cmd *ls) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.ClientFlag, ctx = flags.NewClientFlag(ctx)
	cmd.ClientFlag.Register(ctx, f)

	cmd.OutputFlag, ctx = flags.NewOutputFlag(ctx)
	cmd.OutputFlag.Register(ctx, f)
}

func (cmd *ls) Process(ctx context.Context) error {
	if err := cmd.ClientFlag.Process(ctx); err != nil {
		return err
	}
	if err := cmd.OutputFlag.Process(ctx); err != nil {
		return err
	}
	return nil
}

func (cmd *ls) Usage() string {
	return "[NAME]"
}

func (cmd *ls) Description() string {
	return `VM Storage Policy listing.

Examples:
  govc storage.policy.ls`
}

func listProfiles(ctx context.Context, c *pbm.Client, name string) ([]types.BasePbmProfile, error) {
	rtype := types.PbmProfileResourceType{
		ResourceType: string(types.PbmProfileResourceTypeEnumSTORAGE),
	}

	category := types.PbmProfileCategoryEnumREQUIREMENT

	ids, err := c.QueryProfile(ctx, rtype, string(category))
	if err != nil {
		return nil, err
	}

	profiles, err := c.RetrieveContent(ctx, ids)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return profiles, nil
	}

	for _, p := range profiles {
		if p.GetPbmProfile().Name == name {
			return []types.BasePbmProfile{p}, nil
		}
	}

	return c.RetrieveContent(ctx, []types.PbmProfileId{{UniqueId: name}})
}

type lsResult struct {
	Profile []types.BasePbmProfile
}

func (r *lsResult) Write(w io.Writer) error {
	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)

	for i := range r.Profile {
		p := r.Profile[i].GetPbmProfile()
		_, _ = fmt.Fprintf(tw, "%s\t%s\n", p.ProfileId.UniqueId, p.Name)
	}

	return tw.Flush()
}

func (cmd *ls) Run(ctx context.Context, f *flag.FlagSet) error {
	vc, err := cmd.Client()
	if err != nil {
		return err
	}

	c, err := pbm.NewClient(ctx, vc)
	if err != nil {
		return err
	}

	profiles, err := listProfiles(ctx, c, f.Arg(0))
	if err != nil {
		return err
	}

	return cmd.WriteResult(&lsResult{profiles})
}
