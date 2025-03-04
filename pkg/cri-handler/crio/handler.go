package crio

import (
	"bufio"
	"fmt"
	"github.com/openshift-virtualization/wasp-agent/pkg/log"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FileOperations interface {
	Open(fileName string) (*os.File, error)
	Create(fileName string) (*os.File, error)
	Copy(dst io.Writer, src io.Reader) error
	Chmod(fileName string, mode os.FileMode) error
	ReadFile(name string) ([]byte, error)
	Walk(root string, fn filepath.WalkFunc) error
	Stat(name string) (os.FileInfo, error)
	MkdirAll(path string, perm os.FileMode) error
	Symlink(oldname, newname string) error
	Dir(path string) string
}

type CRIOConfigParser interface {
	Parse() error
	GetDefaultRuntimeConfig() string
}

type HookConfiguration struct {
	SourceScriptPath string
	DstScriptPath    string
	SourceConfigPath string
	DstConfigPath    string
}

type CRIOHandler struct {
	runtime           string
	HookConfig        HookConfiguration
	fileOps           FileOperations
	configParser      CRIOConfigParser
	temporaryHookPath string
}

func New(conf HookConfiguration) *CRIOHandler {
	var handler CRIOHandler

	handler.HookConfig = conf
	handler.fileOps = FileOperationsImpl{}
	handler.configParser = CRIOConfigParserImpl{
		fileOps: handler.fileOps,
	}

	return &handler
}

func (c *CRIOHandler) Setup() error {
	var err error

	if err = c.configParser.Parse(); err != nil {
		return err
	}

	c.runtime = c.configParser.GetDefaultRuntimeConfig()
	log.Log.Infof("Detected default runtime: %s", c.runtime)
	if err = c.UpdateHookScript(); err != nil {
		return err
	}

	if err = c.InstallOCIHook(); err != nil {
		return err
	}

	if err = c.SetCrioSocketSymLink(); err != nil {
		return err
	}

	return nil
}

func (c *CRIOHandler) UpdateHookScript() error {
	const temporaryHookPath = "/tmp/hook.sh.tmp"
	srcFile, err := c.fileOps.Open(c.HookConfig.SourceScriptPath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return err
	}
	defer srcFile.Close()

	dstFile, err := c.fileOps.Create(temporaryHookPath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return err
	}
	defer dstFile.Close()

	scanner := bufio.NewScanner(srcFile)
	writer := bufio.NewWriter(dstFile)
	for scanner.Scan() {
		line := strings.Replace(scanner.Text(), "runc", c.runtime, -1)
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error reading source file: %v\n", err)

	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Error flushing content to destination file: %v\n", err)
	}

	c.temporaryHookPath = temporaryHookPath
	return nil
}

func (c *CRIOHandler) InstallOCIHook() error {
	log.Log.Info("Installing OCI hook script")
	err := c.MoveFile(c.temporaryHookPath, c.HookConfig.DstScriptPath)
	if err != nil {
		return err
	}

	log.Log.Info("Installing OCI hook configuration")
	err = c.MoveFile(c.HookConfig.SourceConfigPath, c.HookConfig.DstConfigPath)
	if err != nil {
		return err
	}

	return nil
}

func (c *CRIOHandler) MoveFile(sourcePath, destPath string) error {
	inputFile, err := c.fileOps.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %v", err)
	}
	defer inputFile.Close()

	outputFile, err := c.fileOps.Create(destPath)
	if err != nil {
		return fmt.Errorf("Couldn't open dest file: %v", err)
	}
	defer outputFile.Close()

	err = c.fileOps.Copy(outputFile, inputFile)
	if err != nil {
		return fmt.Errorf("Couldn't copy to dest from source: %v", err)
	}

	inputFile.Close()

	// Set file permissions to make it executable
	err = c.fileOps.Chmod(destPath, 0755)
	if err != nil {
		return fmt.Errorf("Couldn't set file permissions: %v", err)
	}

	return nil

}
func (c *CRIOHandler) SetCrioSocketSymLink() error {
	const CRIOSocketPathOnContainer = "/var/run/crio/crio.sock"
	const CRIOSocketPathOnHost = "/host/var/run/crio/crio.sock"

	socketDir := c.fileOps.Dir(CRIOSocketPathOnContainer)
	log.Log.Infof("Creating CRIO socket directory %s", socketDir)
	err := c.fileOps.MkdirAll(socketDir, 0755)
	if err != nil {
		return err
	}

	log.Log.Infof(
		"Creating symlink source: %s destination: %s", CRIOSocketPathOnHost, CRIOSocketPathOnContainer)
	c.fileOps.Symlink(CRIOSocketPathOnHost, CRIOSocketPathOnContainer)
	if err != nil {
		return err
	}

	return nil
}
