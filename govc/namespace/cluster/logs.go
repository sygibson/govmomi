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

package cluster

import (
	"context"
	"errors"
	"flag"
	"io"
	"os"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/vapi/namespace"
)

type logs struct {
	*flags.ClusterFlag
}

func init() {
	cli.Register("namespace.cluster.logs.download", &logs{})
}

func (cmd *logs) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.ClusterFlag, ctx = flags.NewClusterFlag(ctx)
	cmd.ClusterFlag.Register(ctx, f)
}

func (cmd *logs) Description() string {
	return `Download namespace cluster support logs.

See also: govc logs.download

Examples:
  govc namespace.cluster.logs.download - | tar -xvf -`
}

func (cmd *logs) Run(ctx context.Context, f *flag.FlagSet) error {
	c, err := cmd.RestClient()
	if err != nil {
		return err
	}

	cluster, err := cmd.Cluster()
	if err != nil {
		return err
	}

	id := cluster.Reference().Value

	var w io.Writer

	switch f.Arg(0) {
	case "-":
		w = os.Stdout
	default:
		return errors.New("TODO")
	}

	m := namespace.NewManager(c)

	bundle, err := m.CreateSupportBundle(ctx, id)
	if err != nil {
		return err
	}

	return m.GetSupportBundle(ctx, bundle, w)
}
