package rpcd

import (
	"io"

	"github.com/Symantec/Dominator/hypervisor/manager"
	"github.com/Symantec/Dominator/lib/log"
	"github.com/Symantec/Dominator/lib/srpc"
)

type srpcType struct {
	manager *manager.Manager
	logger  log.DebugLogger
}

type htmlWriter srpcType

func (hw *htmlWriter) WriteHtml(writer io.Writer) {
	hw.writeHtml(writer)
}

func Setup(manager *manager.Manager, logger log.DebugLogger) (
	*htmlWriter, error) {
	srpcObj := &srpcType{
		manager: manager,
		logger:  logger,
	}
	srpc.RegisterNameWithOptions("Hypervisor", srpcObj, srpc.ReceiverOptions{
		PublicMethods: []string{
			"AcknowledgeVm",
			"ChangeVmOwnerUsers",
			"ChangeVmTags",
			"CreateVm",
			"DestroyVm",
			"DiscardVmOldImage",
			"DiscardVmOldUserData",
			"GetUpdates",
			"GetVmInfo",
			"ListVMs",
			"ProbeVmPort",
			"ReplaceVmImage",
			"ReplaceVmUserData",
			"RestoreVmImage",
			"RestoreVmUserData",
			"StartVm",
			"StopVm",
			"TraceVmMetadata",
		}})
	return (*htmlWriter)(srpcObj), nil
}