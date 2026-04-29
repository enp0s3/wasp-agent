package wasp

import (
	"fmt"
	"os"

	"github.com/openshift-virtualization/wasp-agent/pkg/consts"
	"github.com/openshift-virtualization/wasp-agent/pkg/wasp/config"
	oci_hook_render "github.com/openshift-virtualization/wasp-agent/pkg/wasp/oci-hook-render"
	"k8s.io/klog/v2"
)

type crioConfiguration interface {
	GetRuntime() (string, error)
}

type hookRenderer interface {
	Render() error
}

func setOCIHook() error {
	err := setupHookScript()
	if err != nil {
		return err
	}

	err = moveFile(consts.HookConfigSource, consts.HookConfigPath)
	if err != nil {
		return err
	}

	return nil
}


func setupHookScript() error {
	crioConfig := crioConfiguration(config.New(consts.CrioConfigPath, consts.CrioConfigDropInPath))
	runtime, err := crioConfig.GetRuntime()
	if err != nil {
		return err
	}
	klog.Infof("detected runtime " + runtime)

	renderer := hookRenderer(oci_hook_render.New(consts.HookTemplateFile, consts.HookScriptPath, runtime))
	if err := renderer.Render(); err != nil {
		return err
	}

	err = os.Chmod(consts.HookScriptPath, 0755)
	if err != nil {
		return fmt.Errorf("Couldn't set file permissions: %v", err)
	}

	return nil
}
