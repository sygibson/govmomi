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

package guest

import (
	"context"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

// TODO: this should be vim25.Client.Service.Content.GuestCustomizationManager
var customizationManager = types.ManagedObjectReference{
	Type:  "VirtualMachineGuestCustomizationManager",
	Value: "GuestCustomizationManager",
}

type customizeGuestRequest struct {
	This types.ManagedObjectReference  `xml:"_this"`
	Vm   types.ManagedObjectReference  `xml:"vm"`
	Auth types.BaseGuestAuthentication `xml:"auth,typeattr"`
	Spec types.CustomizationSpec       `xml:"spec"`
}

type customizeGuestResponse struct {
	Returnval types.ManagedObjectReference `xml:"returnval"`
}

type customizeGuestBody struct {
	Req    *customizeGuestRequest  `xml:"urn:vim25 CustomizeGuest_Task,omitempty"`
	Res    *customizeGuestResponse `xml:"CustomizeGuest_TaskResponse,omitempty"`
	Fault_ *soap.Fault             `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
}

func (b *customizeGuestBody) Fault() *soap.Fault { return b.Fault_ }

func customizeGuestTask(ctx context.Context, r soap.RoundTripper, req *customizeGuestRequest) (*customizeGuestResponse, error) {
	var reqBody, resBody customizeGuestBody

	reqBody.Req = req

	if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
		return nil, err
	}

	return resBody.Res, nil
}

type CustomizationManager struct {
	c  *vim25.Client
	vm types.ManagedObjectReference
}

func NewCustomizationManager(c *vim25.Client, vm types.ManagedObjectReference) *CustomizationManager {
	return &CustomizationManager{c, vm}
}

func (m *CustomizationManager) Customize(ctx context.Context, auth types.BaseGuestAuthentication, spec types.CustomizationSpec) (*object.Task, error) {
	req := customizeGuestRequest{
		This: customizationManager,
		Vm:   m.vm,
		Auth: auth,
		Spec: spec,
	}

	res, err := customizeGuestTask(ctx, m.c, &req)
	if err != nil {
		return nil, err
	}

	return object.NewTask(m.c, res.Returnval), nil
}
