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

	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
)

type Manager struct {
	types.ManagedObjectReference

	c *vim25.Client
}

func NewManager(c *vim25.Client) *Manager {
	return &Manager{
		ManagedObjectReference: *c.ServiceContent.TaskManager,
		c: c,
	}
}

func (m *Manager) Client() *vim25.Client {
	return m.c
}

func (m *Manager) CreateCollector(ctx context.Context, filter types.TaskFilterSpec) (*HistoryCollector, error) {
	req := types.CreateCollectorForTasks{
		This:   m.Reference(),
		Filter: filter,
	}

	res, err := methods.CreateCollectorForTasks(ctx, m.c, &req)
	if err != nil {
		return nil, err
	}

	return NewHistoryCollector(m.c, res.Returnval), nil
}
