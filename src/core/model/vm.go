package model

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cloud-barista/cb-ladybug/src/utils/config"
	"github.com/cloud-barista/cb-ladybug/src/utils/lang"
	ssh "github.com/cloud-barista/cb-spider/cloud-control-manager/vm-ssh"

	logger "github.com/sirupsen/logrus"
)

type VM struct {
	Name         string   `json:"name"`
	Config       string   `json:"connectionName"`
	VPC          string   `json:"vNetId"`
	Subnet       string   `json:"subnetId"`
	Firewall     []string `json:"securityGroupIds"`
	SSHKey       string   `json:"sshKeyId"`
	Image        string   `json:"imageId"`
	Spec         string   `json:"specId"`
	UserAccount  string   `json:"vmUserAccount"`
	UserPassword string   `json:"vmUserPassword"`
	Description  string   `json:"description"`
	PublicIP     string   `json:"publicIP"` // output
	Credential   string   // private
	UId          string   `json:"uid"`
	Role         string   `json:"role"`
}

const (
	remoteTargetPath = "/tmp"
)

func (self *VM) ConnectionTest(sshInfo *ssh.SSHInfo) error {
	cmd := "/bin/hostname"
	_, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (self *VM) CopyScripts(sshInfo *ssh.SSHInfo) error {
	sourcePath := fmt.Sprintf("%s/src/scripts", *config.Config.AppRootPath)
	sourceFile := []string{config.BOOTSTRAP_FILE}
	if self.Role == config.CONTROL_PLANE {
		sourceFile = append(sourceFile, config.INIT_FILE)
	}

	logger.Infof("start script file copy (vm=%s, src=%s, dest=%s)\n", self.Name, sourcePath, remoteTargetPath)
	for _, f := range sourceFile {
		src := fmt.Sprintf("%s/%s", sourcePath, f)
		dest := fmt.Sprintf("%s/%s", remoteTargetPath, f)
		if err := ssh.SSHCopy(*sshInfo, src, dest); err != nil {
			return errors.New(fmt.Sprintf("copy scripts error (server=%s, cause=%s)", sshInfo.ServerPort, err))
		}
	}
	return nil
}

func (self *VM) Bootstrap(sshInfo *ssh.SSHInfo) (bool, error) {
	cmd := fmt.Sprintf("cd %s;./%s", remoteTargetPath, config.BOOTSTRAP_FILE)

	result, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		return false, errors.New("k8s bootstrap error")
	}
	if strings.Contains(result, "kubectl set on hold") {
		return true, nil
	} else {
		return false, nil
	}
}

func (self *VM) ControlPlaneInit(sshInfo *ssh.SSHInfo, ip string) (string, string, error) {
	var workerJoinCmd string

	cmd := fmt.Sprintf("cd %s;./%s", remoteTargetPath, config.INIT_FILE)
	cpInitResult, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		return "", "", errors.New("k8s control plane node init error")
	}
	if strings.Contains(cpInitResult, "Your Kubernetes control-plane has initialized successfully") {
		workerJoinCmd = lang.GetWorkerJoinCmd(cpInitResult)
	} else {
		return "", "", nil
	}

	cmd = fmt.Sprintf("sudo sed '5s/.*/    server: https:\\/\\/%s:6443/g' /etc/kubernetes/admin.conf", ip)
	clusterConfig, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		logger.Errorf("Error while running cmd %s (vm=%s, cause=%v)", cmd, self.Name, err)
	}

	return workerJoinCmd, clusterConfig, nil
}

func (self *VM) WorkerJoin(sshInfo *ssh.SSHInfo, workerJoinCmd *string) (bool, error) {
	if *workerJoinCmd == "" {
		return false, errors.New("worker node join command empty")
	}
	cmd := *workerJoinCmd
	result, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		return false, errors.New("k8s worker node join error")
	}
	if strings.Contains(result, "This node has joined the cluster") {
		return true, nil
	} else {
		return false, errors.New("worker node join failed")
	}
}

func (self *VM) WorkerJoinForAddNode(sshInfo *ssh.SSHInfo, workerJoinCmd *string) (bool, error) {
	if *workerJoinCmd == "" {
		return false, errors.New("worker node join command empty")
	}
	cmd := fmt.Sprintf("sudo %s", *workerJoinCmd)
	result, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		return false, errors.New("k8s worker node join error")
	}
	if strings.Contains(result, "This node has joined the cluster") {
		return true, nil
	} else {
		return false, errors.New("worker node join failed")
	}
}
