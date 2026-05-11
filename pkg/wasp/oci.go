package wasp

import (
	"fmt"
	"os"
	"strings"
	"text/template"

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

func setOCIHook(suffix string) error {
	scriptPath := consts.HookScriptPath(suffix)
	configPath := consts.HookConfigPath(suffix)

	if err := setupHookScript(scriptPath); err != nil {
		return err
	}

	if err := renderHookConfig(configPath, scriptPath); err != nil {
		return err
	}

	return nil
}

func cleanupOCIHook(suffix string) {
	cleanupFiles(consts.HookConfigPath(suffix), consts.HookScriptPath(suffix))
}

func cleanupFiles(paths ...string) {
	for _, p := range paths {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			klog.Warningf("failed to remove %s: %v", p, err)
		}
	}
}

func renderHookConfig(configPath, scriptPath string) error {
	return renderHookConfigFromTemplate(consts.HookConfigTemplateFile, configPath, scriptPath)
}

func renderHookConfigFromTemplate(templateFile, configPath, scriptPath string) error {
	nodeScriptPath := strings.TrimPrefix(scriptPath, "/host")

	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return fmt.Errorf("error while parsing hook config template: %v", err)
	}

	dstFile, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create hook config %s: %v", configPath, err)
	}
	defer dstFile.Close()

	data := struct{ HookScriptPath string }{HookScriptPath: nodeScriptPath}
	if err := tmpl.Execute(dstFile, data); err != nil {
		return fmt.Errorf("error while rendering hook config template: %v", err)
	}

	return nil
}

func setupHookScript(scriptPath string) error {
	crioConfig := crioConfiguration(config.New(consts.CrioConfigPath, consts.CrioConfigDropInPath))
	runtime, err := crioConfig.GetRuntime()
	if err != nil {
		return err
	}
	klog.Infof("detected runtime " + runtime)

	renderer := hookRenderer(oci_hook_render.New(consts.HookTemplateFile, scriptPath, runtime))
	if err := renderer.Render(); err != nil {
		return err
	}

	err = os.Chmod(scriptPath, 0755)
	if err != nil {
		return fmt.Errorf("Couldn't set file permissions: %v", err)
	}

	return nil
}
