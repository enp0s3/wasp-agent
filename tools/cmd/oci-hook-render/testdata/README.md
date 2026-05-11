# Test Data for OCI Hook Render Tool

This directory contains fake CRI-O configuration fixtures used for testing the OCI hook script rendering.

## Directory Structure

```
testdata/
├── crun/                          # CRI-O configured with crun runtime
│   ├── crio.conf
│   └── crio.conf.d/
│       └── 00-default.conf
├── runc/                          # CRI-O configured with runc runtime
│   ├── crio.conf
│   └── crio.conf.d/
│       └── 00-default.conf
└── dropin-override/               # Main config sets runc, drop-in overrides to crun
    ├── crio.conf
    └── crio.conf.d/
        └── 10-override-runtime.conf
```

## Fixtures

- **crun/** — Both main config and drop-in set `default_runtime = "crun"`.
- **runc/** — Both main config and drop-in set `default_runtime = "runc"`.
- **dropin-override/** — Main config sets `runc`, but the drop-in file overrides the runtime to `crun`. This validates that drop-in files take precedence over the main configuration.

## Usage

```bash
go run ./tools/cmd/oci-hook-render/ \
  -crio-config-dir tools/cmd/oci-hook-render/testdata/crun/crio.conf.d \
  -o /tmp/oci-hook.sh
```
