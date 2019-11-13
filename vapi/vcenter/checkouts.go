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

package vcenter

import (
	"context"
	"net/http"
	"path"

	"github.com/vmware/govmomi/vapi/internal"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// CheckOut specification
type CheckOut struct {
	Name      string     `json:"name,omitempty"`
	Placement *Placement `json:"placement,omitempty"`
	PoweredOn bool       `json:"powered_on,omitempty"`
}

// CheckIn specification
type CheckIn struct {
	Message string `json:"message"`
}

// CheckOut a library item containing a VM template.
func (c *Manager) CheckOut(ctx context.Context, libraryItemID string, checkout *CheckOut) (*types.ManagedObjectReference, error) {
	url := c.Resource(path.Join(internal.VCenterVMTXLibraryItem, libraryItemID, "check-outs")).WithParam("action", "check-out")
	var res string
	spec := struct {
		Spec *CheckOut `json:"spec"`
	}{checkout}
	err := c.Do(ctx, url.Request(http.MethodPost, spec), &res)
	if err != nil {
		return nil, err
	}
	return &types.ManagedObjectReference{Type: "VirtualMachine", Value: res}, nil
}

// CheckIn a VM into the library item.
func (c *Manager) CheckIn(ctx context.Context, libraryItemID string, vm mo.Reference, checkin *CheckIn) (string, error) {
	p := path.Join(internal.VCenterVMTXLibraryItem, libraryItemID, "check-outs", vm.Reference().Value)
	url := c.Resource(p).WithParam("action", "check-in")
	var res string
	spec := struct {
		Spec *CheckIn `json:"spec"`
	}{checkin}
	return res, c.Do(ctx, url.Request(http.MethodPost, spec), &res)
}
