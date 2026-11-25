package daemon

import (
	"context"
	"fmt"

	"github.com/containerd/typeurl/v2"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/events"
	libcontainerdtypes "github.com/moby/moby/v2/daemon/internal/libcontainerd/types"
	"github.com/moby/moby/v2/errdefs"
	"google.golang.org/protobuf/types/known/structpb"
)

// psColumns are the columns of "ps -ef", the default of docker top.
var psColumns = []string{"UID", "PID", "PPID", "C", "STIME", "TTY", "TIME", "CMD"}

// ContainerTop lists the processes of a running container. On android they
// run in the runtime's kernel, which no ps on the host can see, so the
// runtime describes each process of the task in its process info: a
// google.protobuf.Struct with a string for each of the columns of "ps -ef".
// That is the only format; other ps arguments are rejected.
func (daemon *Daemon) ContainerTop(name string, psArgs string) (*container.TopResponse, error) {
	if psArgs != "" && psArgs != "-ef" {
		return nil, errdefs.InvalidParameter(fmt.Errorf("ps arguments %q are not supported on android, only -ef", psArgs))
	}

	ctr, err := daemon.GetContainer(name)
	if err != nil {
		return nil, err
	}

	tsk, err := func() (libcontainerdtypes.Task, error) {
		ctr.Lock()
		defer ctr.Unlock()

		tsk, err := ctr.GetRunningTask()
		if err != nil {
			return nil, err
		}
		if ctr.State.Restarting {
			return nil, errContainerIsRestarting(ctr.ID)
		}
		return tsk, nil
	}()
	if err != nil {
		return nil, err
	}

	infos, err := tsk.Pids(context.Background())
	if err != nil {
		return nil, err
	}
	procList := &container.TopResponse{Titles: psColumns}
	for _, p := range infos {
		if p.Info == nil {
			return nil, errdefs.System(fmt.Errorf("process %d has no info", p.Pid))
		}
		v, err := typeurl.UnmarshalAny(p.Info)
		if err != nil {
			return nil, errdefs.System(fmt.Errorf("info of process %d: %w", p.Pid, err))
		}
		info, ok := v.(*structpb.Struct)
		if !ok {
			return nil, errdefs.System(fmt.Errorf("info of process %d is %T, not a google.protobuf.Struct", p.Pid, v))
		}
		row := make([]string, len(psColumns))
		for i, column := range psColumns {
			row[i] = info.GetFields()[column].GetStringValue()
		}
		procList.Processes = append(procList.Processes, row)
	}
	daemon.LogContainerEvent(ctr, events.ActionTop)
	return procList, nil
}
