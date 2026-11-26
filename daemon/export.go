package daemon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"

	"github.com/moby/moby/v2/errdefs"
)

// ContainerExport writes the contents of the container to the given
// writer. An error is returned if the container cannot be found.
func (daemon *Daemon) ContainerExport(ctx context.Context, name string, out io.Writer) error {
	ctr, err := daemon.GetContainer(name)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" && ctr.ImagePlatform.OS == "windows" {
		return errors.New("the daemon on this operating system does not support exporting Windows containers")
	}

	if ctr.State.IsDead() {
		err := fmt.Errorf("You cannot export container %s which is Dead", ctr.ID)
		return errdefs.Conflict(err)
	}

	if ctr.State.IsRemovalInProgress() {
		err := fmt.Errorf("You cannot export container %s which is being removed", ctr.ID)
		return errdefs.Conflict(err)
	}

	err = daemon.containerExport(ctx, ctr, out)
	if err != nil {
		return fmt.Errorf("Error exporting container %s: %v", name, err)
	}

	return nil
}
