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

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type Manager struct {
	object.Common

	pc *property.Collector
}

// GetManager wraps NewManager, returning ErrNotSupported
// when the client is not connected to a vCenter instance.
func GetManager(c *vim25.Client) (*Manager, error) {
	if c.ServiceContent.AlarmManager == nil {
		return nil, object.ErrNotSupported
	}
	return NewManager(c), nil
}

func NewManager(c *vim25.Client) *Manager {
	m := Manager{
		Common: object.NewCommon(c, *c.ServiceContent.AlarmManager),
		pc:     property.DefaultCollector(c),
	}

	return &m
}

func (m Manager) AcknowledgeAlarm(ctx context.Context, alarm types.ManagedObjectReference, entity object.Reference) error {
	req := types.AcknowledgeAlarm{
		This:   m.Reference(),
		Alarm:  alarm,
		Entity: entity.Reference(),
	}

	_, err := methods.AcknowledgeAlarm(ctx, m.Client(), &req)

	return err
}

func (m Manager) GetAlarm(ctx context.Context, entity object.Reference) ([]mo.Alarm, error) {
	req := types.GetAlarm{
		This: m.Reference(),
	}

	if entity != nil {
		ref := entity.Reference()
		req.Entity = &ref
	}

	res, err := methods.GetAlarm(ctx, m.Client(), &req)
	if err != nil {
		return nil, err
	}

	if len(res.Returnval) == 0 {
		return nil, nil
	}

	alarms := make([]mo.Alarm, 0, len(res.Returnval))

	err = m.pc.Retrieve(ctx, res.Returnval, []string{"info"}, &alarms)
	if err != nil {
		return nil, err
	}

	return alarms, nil
}

func (m Manager) GetAlarmState(ctx context.Context, entity object.Reference) ([]types.AlarmState, error) {
	req := types.GetAlarmState{
		This:   m.Reference(),
		Entity: entity.Reference(),
	}

	res, err := methods.GetAlarmState(ctx, m.Client(), &req)
	if err != nil {
		return nil, err
	}

	return res.Returnval, nil
}
