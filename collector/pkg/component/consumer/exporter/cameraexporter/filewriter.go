package cameraexporter

import (
	"encoding/json"
	"fmt"
	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

type fileWriter struct {
	config *fileConfig
	logger *component.TelemetryLogger
}

func newFileWriter(cfg *fileConfig, logger *component.TelemetryLogger) (*fileWriter, error) {
	// StoragePath must be an absolute path
	if !path.IsAbs(cfg.StoragePath) {
		return nil, fmt.Errorf("storage_path must be an absolute path")
	}
	// Check if there is the directory.
	// If the directory is not found, we create one
	if err := os.MkdirAll(cfg.StoragePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create fileWriter: %w", err)
	}
	return &fileWriter{
		config: cfg,
		logger: logger,
	}, nil
}

func (fw *fileWriter) write(group *model.DataGroup) {
	groupName := group.Name
	switch groupName {
	case constnames.SingleNetRequestMetricGroup:
		fw.writeTrace(group)
	case constnames.CameraEventGroupName:
		fw.writeCpuEvents(group)
	}
}

func (fw *fileWriter) pidFilePath(pid int64) string {
	return path.Join(fw.config.StoragePath, strconv.FormatInt(pid, 10))
}

func fileName(podName string, isServer bool, timestamp uint64, protocol string) string {
	if len(podName) == 0 {
		podName = "null"
	}
	var isServerString string
	if isServer {
		isServerString = "true"
	} else {
		isServerString = "false"
	}
	return podName + "_" + isServerString + "_" + protocol + "_" + strconv.FormatUint(timestamp, 10)
}

func (fw *fileWriter) writeTrace(group *model.DataGroup) {
	pid := group.Labels.GetIntValue(constlabels.Pid)
	// Create the directory first for saving its profile files.
	// If there has been such a directory, it will do nothing and return nil.
	baseDir := fw.pidFilePath(pid)
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		fw.logger.Errorf("Failed to create pid directory: %v", err)
		return
	}

	timestamp := group.Timestamp
	isServer := group.Labels.GetBoolValue(constlabels.IsServer)
	var podName string
	if isServer {
		podName = group.Labels.GetStringValue(constlabels.DstPod)
	} else {
		podName = group.Labels.GetStringValue(constlabels.SrcPod)
	}
	protocol := group.Labels.GetStringValue(constlabels.Protocol)
	fileName := fileName(podName, isServer, timestamp, protocol)
	// Check whether we need to roll over the files
	err := fw.writeFile(baseDir, fileName, group)
	if err != nil {
		fw.logger.Errorf("Failed to write trace to file: %v", err)
		return
	}
}

func (fw *fileWriter) writeFile(baseDir string, fileName string, group *model.DataGroup) error {
	// Check whether the count of files is greater than MaxCount
	err := fw.rotateFiles(baseDir)
	if err != nil {
		fw.logger.Infof("can't rotate files in %s: %v", baseDir, err)
	}
	filePath := filepath.Join(baseDir, fileName)
	f, err := os.Create(filePath)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("can't create new file: %w", err)
	}
	bytes, err := json.Marshal(group)
	if err != nil {
		return fmt.Errorf("can't marshal DataGroup: %w", err)
	}
	_, err = f.Write(bytes)
	return err
}

func (fw *fileWriter) rotateFiles(baseDir string) error {
	// No constrains set
	if fw.config.MaxFileCount <= 0 {
		return nil
	}
	// Get all files path
	toBeRotated, err := getFilesName(baseDir)
	if err != nil {
		return fmt.Errorf("can't get files list: %w", err)
	}
	// No need to rotate
	if len(toBeRotated) < fw.config.MaxFileCount {
		return nil
	}
	// TODO Remove the older files
	toBeRotated = toBeRotated[:len(toBeRotated)-fw.config.MaxFileCount+1]
	// Remove the stale files asynchronously
	go func() {
		for _, file := range toBeRotated {
			_ = os.Remove(filepath.Join(baseDir, file))
		}
	}()
	return nil
}

func getFilesName(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	files, err := f.Readdirnames(-1)
	return files, err
}

const dividingLine = "\n---\n"

func (fw *fileWriter) writeCpuEvents(group *model.DataGroup) {
	// triggerKey is the file name of trace file
	triggerKey := group.Labels.GetStringValue("trigger_key")
	if triggerKey == "" {
		fw.logger.Infof("Trigger_key of cpu_events is empty")
		return
	}
	pid := group.Labels.GetIntValue(constlabels.Pid)
	baseDir := fw.pidFilePath(pid)
	filePath := filepath.Join(baseDir, triggerKey)
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0)
	defer f.Close()
	if err != nil {
		// Just return if we can't find the exported file
		return
	}
	_, err = f.Write([]byte(dividingLine))
	if err != nil {
		fw.logger.Infof("Failed to append CpuEvents to the file %s: %v", filePath, err)
		return
	}
	eventsBytes, _ := json.Marshal(group)
	_, err = f.Write(eventsBytes)
	if err != nil {
		fw.logger.Infof("Failed to append CpuEvents to the file %s: %v", filePath, err)
		return
	}
}

func (fw *fileWriter) name() string {
	return storageFile
}
