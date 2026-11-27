//go:build !windows && !android

package daemon

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	runcoptions "github.com/containerd/containerd/api/types/runc/options"
	"github.com/containerd/log"
	"github.com/moby/moby/api/types/system"
	"github.com/pkg/errors"
)

func populateRuncCommit(v *system.Commit, cfg *configStore) error {
	_, _, commit, err := parseDefaultRuntimeVersion(&cfg.Runtimes)
	if err != nil {
		return err
	}
	v.ID = commit
	return nil
}

func (daemon *Daemon) populateInitCommit(ctx context.Context, v *system.Info, cfg *configStore) error {
	v.InitBinary = cfg.GetInitPath()
	initBinary, err := cfg.LookupInitPath()
	if err != nil {
		log.G(ctx).WithError(err).Warnf("Failed to find docker-init")
		return nil
	}

	rv, err := exec.CommandContext(ctx, initBinary, "--version").Output()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		log.G(ctx).WithError(err).Warnf("Failed to retrieve %s version", initBinary)
		return nil
	}

	_, commit, err := parseInitVersion(string(rv))
	if err != nil {
		log.G(ctx).WithError(err).Warnf("failed to parse %s version", initBinary)
		return nil
	}
	v.InitCommit.ID = commit
	return nil
}

func parseDefaultRuntimeVersion(rts *runtimes) (runtime, version, commit string, _ error) {
	shim, opts, err := rts.Get(rts.Default)
	if err != nil {
		return "", "", "", err
	}
	shimopts, ok := opts.(*runcoptions.Options)
	if !ok {
		return "", "", "", fmt.Errorf("%s: retrieving version not supported", shim)
	}
	rt := shimopts.BinaryName
	if rt == "" {
		rt = defaultRuntimeName
	}
	rv, err := exec.Command(rt, "--version").Output()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to retrieve %s version: %w", rt, err)
	}
	runtime, version, commit, err = parseRuntimeVersion(string(rv))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse %s version: %w", rt, err)
	}
	return runtime, version, commit, nil
}

func populateRuncVersion(cfg *configStore, v *system.VersionResponse) error {
	_, ver, commit, err := parseDefaultRuntimeVersion(&cfg.Runtimes)
	if err != nil {
		return err
	}
	v.Components = append(v.Components, system.ComponentVersion{
		Name:    cfg.Runtimes.Default,
		Version: ver,
		Details: map[string]string{
			"GitCommit": commit,
		},
	})
	return nil
}

func populateInitVersion(ctx context.Context, cfg *configStore, v *system.VersionResponse) error {
	initBinary, err := cfg.LookupInitPath()
	if err != nil {
		log.G(ctx).WithError(err).Warn("Failed to find docker-init")
		return nil
	}

	rv, err := exec.CommandContext(ctx, initBinary, "--version").Output()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		log.G(ctx).WithError(err).Warnf("Failed to retrieve %s version", initBinary)
		return nil
	}

	ver, commit, err := parseInitVersion(string(rv))
	if err != nil {
		log.G(ctx).WithError(err).Warnf("failed to parse %s version", initBinary)
		return nil
	}
	v.Components = append(v.Components, system.ComponentVersion{
		Name:    filepath.Base(initBinary),
		Version: ver,
		Details: map[string]string{
			"GitCommit": commit,
		},
	})
	return nil
}
