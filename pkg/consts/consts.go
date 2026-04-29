package consts

const (
	HookTemplateFile     = "/app/OCI-hook/hookscript.template"
	HookScriptPath       = "/host/opt/oci-hook-swap.sh"
	HookConfigSource     = "/app/OCI-hook/swap-for-burstable.json"
	HookConfigPath       = "/host/run/containers/oci/hooks.d/swap-for-burstable.json"
	CrioConfigPath       = "/host/etc/crio/crio.conf"
	CrioConfigDropInPath = "/host/etc/crio/crio.conf.d"
)
