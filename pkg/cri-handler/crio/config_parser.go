package crio

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/openshift-virtualization/wasp-agent/pkg/log"
	"os"
)

const (
	defaultRuntime  = "crun"
	configLogPrefix = "Updating config from "
	// CrioConfigPath is the default location for the conf file.
	CrioConfigPath = "/host/etc/crio/crio.conf"
	// CrioConfigDropInPath is the default location for the drop-in config files.
	CrioConfigDropInPath = "/host/etc/crio/crio.conf.d"
)

// tomlConfig is another way of looking at a Config, which is
// TOML-friendly (it has all of the explicit tables). It's just used for
// conversions.
type tomlConfig struct {
	Crio struct {
		Runtime struct{ RuntimeConfig } `toml:"runtime"`
	} `toml:"crio"`
}

func (t *tomlConfig) toConfig(c *Config) {
	c.RuntimeConfig = t.Crio.Runtime.RuntimeConfig
}

func (t *tomlConfig) fromConfig(c *Config) {
	t.Crio.Runtime.RuntimeConfig = c.RuntimeConfig
}

// Config represents the entire set of configuration values that can be set for
// the server. This is intended to be loaded from a toml-encoded config file.
type Config struct {
	RuntimeConfig
}

// RuntimeConfig represents the "crio.runtime" TOML config table.
type RuntimeConfig struct {
	// DefaultRuntime is the _name_ of the OCI runtime to be used as the default.
	// The name is matched against the Runtimes map below.
	DefaultRuntime string `toml:"default_runtime"`
}

// DefaultConfig returns the default configuration for crio.
func DefaultConfig() *Config {
	return &Config{
		RuntimeConfig: RuntimeConfig{
			DefaultRuntime: defaultRuntime,
		},
	}
}

/*
Re-build CRI-O configuration in order to determine the runtime binary.
The following order takes place:
1. the last file from all drop-in directories /etc/crio/crio.conf.d/*
2. Root crio config file.
3. Default config.
*/

type CRIOConfigParserImpl struct {
	config  *Config
	fileOps FileOperations
}

func (cp CRIOConfigParserImpl) Parse() error {
	var err error
	// get default config
	cp.config = DefaultConfig()

	// update config from root CRIO config file
	if err = cp.UpdateFromFile(CrioConfigPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Log.Infof("Skipping not-existing config file %s", CrioConfigPath)
		} else {
			return fmt.Errorf("update config from file: %w", err)
		}

	}

	//Update config from drop-in configs
	if err = cp.UpdateFromPath(CrioConfigDropInPath); err != nil {
		return err
	}

	return nil
}

func (cp CRIOConfigParserImpl) UpdateFromPath(path string) error {
	if _, err := cp.fileOps.Stat(path); err != nil && os.IsNotExist(err) {
		log.Log.Infof("Skipping not-existing drop-in directory %s", CrioConfigDropInPath)
		return nil
	}

	log.Log.Infof(configLogPrefix+"path: %s", path)

	if err := cp.fileOps.Walk(path,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			return cp.UpdateFromFile(p)
		}); err != nil {
		return fmt.Errorf("walk path: %w", err)
	}
	return nil
}

func (cp CRIOConfigParserImpl) UpdateFromFile(path string) error {
	data, err := cp.fileOps.ReadFile(path)
	if err != nil {
		return err
	}

	log.Log.Infof(configLogPrefix+" file: %s", path)

	t := new(tomlConfig)
	t.fromConfig(cp.config)

	_, err = toml.Decode(string(data), t)
	if err != nil {
		return fmt.Errorf("unable to decode configuration %v: %w", path, err)
	}

	t.toConfig(cp.config)
	return nil
}

func (cp CRIOConfigParserImpl) GetDefaultRuntimeConfig() string {
	return cp.config.RuntimeConfig.DefaultRuntime
}
