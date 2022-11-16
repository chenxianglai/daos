package support

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"os/exec"
	"strings"
	"io"
	"io/ioutil"
	"bytes"

	"github.com/pkg/errors"

	"github.com/daos-stack/daos/src/control/server/config"
	"github.com/daos-stack/daos/src/control/lib/control"
	"github.com/daos-stack/daos/src/control/logging"
	"github.com/daos-stack/daos/src/control/common"
	"github.com/daos-stack/daos/src/control/lib/hardware/hwprov"
	"github.com/daos-stack/daos/src/control/lib/hardware"
)

// Folder structure to copy logs and configs
const (
	dmgSystemLogFolder = "DmgSystemLog" // Copy the dmg command output specific to full DAOS system
	dmgNodeLogFolder = "DmgNodeLog" 	// Copy the dmg command output specific to the node storage to individual node log folder
	systemInfo = "SysInfo"				// Copy the system related information
	serverLogs = "ServerLogs"			// Copy the server/conrol and helper logs
	daosConfig = "ServerConfig"			// Copy the server config
)

func getRunningConf() (string, bool) {
	_, err := exec.Command("bash", "-c", "pidof daos_engine").Output()
    if err != nil {
        fmt.Println("daos_engine is not running on server")
        return "", false
    }	

	cmd := "ps -eo args | grep daos_engine | head -n 1 | grep -oP '(?<=-d )[^ ]*'"
	stdout, err := exec.Command("bash", "-c", cmd).Output()
	running_config := filepath.Join(strings.TrimSpace(string(stdout)), config.ConfigOut)

	return running_config, true
}

func getServerConf() string{
	conf, err :=  getRunningConf()

	if err == true {
        return conf
    }

	// Return the default config
	serverConfig := config.DefaultServer()
	default_path := filepath.Join(serverConfig.SocketDir, config.ConfigOut)

	return default_path
}

func cpFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	log_file_name := filepath.Base(src)

	fmt.Printf(" -- SAMIR INFO -- Copy File %s to %s\n", log_file_name, dst)
	out, err := os.Create(filepath.Join(dst, log_file_name))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

// Check if file or directory that starts with . which is hidden
func IsHidden(filename string) (bool) {
	if filename[0:1] == "." {
		return true
	} else {
		return false
	}
	return false
}

func CopyServerConfig(src, dst string) error {
	err := cpFile(src, dst)
	if err != nil {
		return err
	}

	// Rename the file if it's hidden
	result := IsHidden(filepath.Base(src))
	if result == true{
		hiddenConf := filepath.Join(dst, filepath.Base(src))
		nonhiddenConf := filepath.Join(dst, filepath.Base(src)[1:])
		os.Rename(hiddenConf, nonhiddenConf)
	}

	return nil
}

func createFolder(target string) error {
	// Create the folder if it's not exist
	if _, err := os.Stat(target); os.IsNotExist(err) {
		fmt.Println("Log folder does not Exists, so creating folder ", target)

		if err := os.MkdirAll(target, 0700); err != nil && !os.IsExist(err) {
			return errors.Wrapf(err, "failed to create log directory %s", target)
		}
    }

	return nil
}

func cpOutputToFile(cmd string, target string) (string, error) {
	// Run command and copy output to the file
	// executing as subshell enables pipes in cmd string
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "Error running command %s with %s", cmd, out)
	}

	if err := ioutil.WriteFile(filepath.Join(target, cmd), out, 0644); err != nil {
		return "", errors.Wrapf(err, "failed to write %s", filepath.Join(target, cmd))
	}

	return string(out), nil
}

func ArchiveLogs(target string) error {
	var buf bytes.Buffer
	err := common.FolderCompress(target, &buf)
	if err != nil {
		return err
	}

	// write to the the .tar.gzip
	tarFileName := fmt.Sprintf("%s.tar.gz", target)
	fmt.Println(" -- SAMIR - INFO -  Archiving the log file ", tarFileName)
	fileToWrite, err := os.OpenFile(tarFileName, os.O_CREATE|os.O_RDWR, os.FileMode(0755))
	if err != nil {
			return err
	}
	if _, err := io.Copy(fileToWrite, &buf); err != nil {
		return err
	}

	return nil
}

func CollectDmgSysteminfo(dst string, configPath string) error {
	targetDmgLog := filepath.Join(dst, dmgSystemLogFolder)
	err := createFolder(targetDmgLog)
	if err != nil {
		return err
	}

	for _, dmgCommand := range control.DmgLogCollectCmd {
		dmgCommand = strings.Join([]string{dmgCommand, "-o", configPath}, " ")
		_, err = cpOutputToFile(dmgCommand, targetDmgLog)
		if err != nil {
			return err
		}
	}

	return nil
}

func CollectDmgNodeinfo(dst string, configPath string) error {
	cmd := strings.Join([]string{"dmg", "system", "query", "-v", "-o", configPath}, " ")
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		err = errors.Wrapf(
			err, "Error running command %s with %s", cmd, out)
	}
	temp := strings.Split(string(out), "\n")

	for _, v := range temp[2:len(temp)-2] {
		// List the device health info from server
		hostName := strings.Fields(v)[3][1:]
		dmgCommand := strings.Join([]string{control.DmgListDeviceCmd, "-o", configPath, "-l", hostName}, " ")
		targetDmgLog := filepath.Join(dst, hostName, dmgNodeLogFolder)
		output, err := cpOutputToFile(dmgCommand, targetDmgLog)
		if err != nil {
			return err
		}

		// List the device health info from each server
        for _, v1 := range strings.Split(output, "\n") {
			if strings.Contains(v1, "UUID"){
				device := strings.Fields(v1)[0][5:]
				deviceHealthcmd := strings.Join([]string{
					control.DmgDeviceHealthCmd, "-u", device, "-l", hostName, "-o", configPath}, " ")
					fmt.Println(" -- SAMIR DMG System Command -- ", deviceHealthcmd)
				output, err = cpOutputToFile(deviceHealthcmd, targetDmgLog)
				if err != nil {
					return err
				}
            }
        }
	}

	return nil
}

func CollectServerLog(dst string) error {
	hn, err := os.Hostname()
	if err != nil {
		return err
	}

	targetLocation := filepath.Join(dst, hn)
	err = createFolder(targetLocation)
	if err != nil {
		return err
	}

	// Get the server config
	cfgPath := getServerConf()
	serverConfig := config.DefaultServer()
	serverConfig.SetPath(cfgPath)
	serverConfig.Load()	

	// Copy server config file
	targetConfig := filepath.Join(targetLocation, daosConfig)
	err = createFolder(targetConfig)
	if err != nil {
		return err
	}
	err = CopyServerConfig(cfgPath, targetConfig)
	if err != nil {
		return err
	}

	// Copy DAOS server engine log files
	targetServerLogs := filepath.Join(targetLocation, serverLogs)
	err = createFolder(targetServerLogs)
	if err != nil {
		return err
	}
	for i := range serverConfig.Engines {
		// Find the matching file incase of log file is based on PID or it has backup
		matches, _ := filepath.Glob(serverConfig.Engines[i].LogFile + "*")
		for _, logfile := range matches {
			err := cpFile(logfile, targetServerLogs)
			if err != nil {
				return err
			}
		}
	}

	// Copy DAOS Control log file
	err = cpFile(serverConfig.ControlLogFile, targetServerLogs)
	if err != nil {
		return err
	}

	// Copy DAOS Helper log file
	err = cpFile(serverConfig.HelperLogFile, targetServerLogs)
	if err != nil {
		return err
	}

	// Copy daos_metrics log
	dmgNodeLocation := filepath.Join(targetLocation, dmgNodeLogFolder)
	err = createFolder(dmgNodeLocation)
	if err != nil {
		return err
	}
	for i := range serverConfig.Engines {
		engineId := fmt.Sprintf("%d", i)
		cmd := strings.Join([]string{"daos_metrics", "-S",  engineId}, " ")

		_, err = cpOutputToFile(cmd, dmgNodeLocation)
		if err != nil {
			return err
		}
	}

	// Collect dump-topology output
	log := logging.NewCommandLineLogger()
	hwProv := hwprov.DefaultTopologyProvider(log)
	topo, err := hwProv.GetTopology(context.Background())
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dmgNodeLocation, "daos_server dump-topology"))
    if err != nil {
        return err
    }
    defer f.Close()
	hardware.PrintTopology(topo, f)

	// Collect system related information
	targetSysinfo := filepath.Join(targetLocation, systemInfo)
	err = createFolder(targetSysinfo)
	if err != nil {
		return err
	}
	for _, sysCommand := range control.SysInfoCmd {
		_, err = cpOutputToFile(sysCommand, targetSysinfo)
		if err != nil {
			return err
		}
	}

	return nil

}