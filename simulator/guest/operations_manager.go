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

package guest

import (
	"time"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/toolbox"
	"github.com/vmware/govmomi/toolbox/vix"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

type OperationsManager struct {
	mo.GuestOperationsManager
}

func NewOperationsManager(ref types.ManagedObjectReference) []object.Reference {
	m := &OperationsManager{
		GuestOperationsManager: mo.GuestOperationsManager{
			Self:           ref,
			AuthManager:    &types.ManagedObjectReference{Type: "GuestAuthManager", Value: "guestOperationsAuthManager"},
			FileManager:    &types.ManagedObjectReference{Type: "GuestFileManager", Value: "guestOperationsFileManager"},
			ProcessManager: &types.ManagedObjectReference{Type: "GuestProcessManager", Value: "guestOperationsProcessManager"},
		},
	}

	return []object.Reference{
		m,
		FileManager{mo.GuestFileManager{Self: *m.FileManager}},
		ProcessManager{mo.GuestProcessManager{Self: *m.ProcessManager}, toolbox.NewProcessManager()},
	}
}

type FileManager struct {
	mo.GuestFileManager
}

type ProcessManager struct {
	mo.GuestProcessManager
	*toolbox.ProcessManager
}

func (m ProcessManager) StartProgramInGuest(ctx *simulator.Context, req *types.StartProgramInGuest) soap.HasFault {
	body := new(methods.StartProgramInGuestBody)

	spec := req.Spec.(*types.GuestProgramSpec)
	auth := req.Auth.(*types.NamePasswordAuthentication)

	start := &vix.StartProgramRequest{
		ProgramPath: spec.ProgramPath,
		Arguments:   spec.Arguments,
		WorkingDir:  spec.WorkingDirectory,
		EnvVars:     spec.EnvVariables,
	}

	proc := toolbox.NewProcess()
	proc.Owner = auth.Username

	pid, err := m.Start(start, proc)
	if err != nil {
		panic(err)
	}

	body.Res = &types.StartProgramInGuestResponse{
		Returnval: pid,
	}

	return body
}

func (m ProcessManager) ListProcessesInGuest(ctx *simulator.Context, req *types.ListProcessesInGuest) soap.HasFault {
	body := &methods.ListProcessesInGuestBody{
		Res: new(types.ListProcessesInGuestResponse),
	}

	procs := m.List(req.Pids)

	for _, proc := range procs {
		var end *time.Time
		if proc.EndTime != 0 {
			end = types.NewTime(time.Unix(proc.EndTime, 0))
		}

		body.Res.Returnval = append(body.Res.Returnval, types.GuestProcessInfo{
			Name:      proc.Name,
			Pid:       proc.Pid,
			Owner:     proc.Owner,
			CmdLine:   proc.Name + " " + proc.Args,
			StartTime: time.Unix(proc.StartTime, 0),
			EndTime:   end,
			ExitCode:  proc.ExitCode,
		})
	}

	return body
}

func (m ProcessManager) TerminateProcessInGuest(ctx *simulator.Context, req *types.TerminateProcessInGuest) soap.HasFault {
	body := new(methods.TerminateProcessInGuestBody)

	if m.Kill(req.Pid) {
		body.Res = new(types.TerminateProcessInGuestResponse)
	} // else TODO

	return body
}

func (m FileManager) ListFilesInGuest(ctx *simulator.Context, req *types.ListFilesInGuest) soap.HasFault {
	body := new(methods.ListFilesInGuestBody)

	body.Res = new(types.ListFilesInGuestResponse)

	return body
}
