package consts

import "fmt"

const (
	HookTemplateFile       = "/app/OCI-hook/hookscript.template"
	HookConfigTemplateFile = "/app/OCI-hook/swap-for-burstable.json"
	CrioConfigPath         = "/host/etc/crio/crio.conf"

	CrioConfigDropInPath = "/host/etc/crio/crio.conf.d"

	HookScriptDir = "/host/opt"
	HookConfigDir = "/host/run/containers/oci/hooks.d"
)

func HookScriptPath(suffix string) string {
	return fmt.Sprintf("%s/oci-hook-swap-%s.sh", HookScriptDir, suffix)
}

func HookConfigPath(suffix string) string {
	return fmt.Sprintf("%s/swap-for-burstable-%s.json", HookConfigDir, suffix)
}
