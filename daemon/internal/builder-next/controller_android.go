//go:build android

package buildkit

import (
	"github.com/moby/buildkit/solver/pb"
	"github.com/moby/buildkit/util/apicaps"
)

func initSnapshotterCaps() {
	pb.Caps.Init(apicaps.Cap{
		ID:                pb.CapMergeOp,
		Enabled:           false,
		DisabledReasonMsg: "merge op is not supported on android",
	})
	pb.Caps.Init(apicaps.Cap{
		ID:                pb.CapDiffOp,
		Enabled:           false,
		DisabledReasonMsg: "diff op is not supported on android",
	})
}
