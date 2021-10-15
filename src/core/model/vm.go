package model

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/cloud-barista/cb-mcks/src/utils/config"
	ssh "github.com/cloud-barista/cb-spider/cloud-control-manager/vm-ssh"

	logger "github.com/sirupsen/logrus"
)

type VM struct {
	Name            string          `json:"name"`
	Config          string          `json:"connectionName"`
	VPC             string          `json:"vNetId"`
	Subnet          string          `json:"subnetId"`
	Firewall        []string        `json:"securityGroupIds"`
	SSHKey          string          `json:"sshKeyId"`
	Image           string          `json:"imageId"`
	Spec            string          `json:"specId"`
	UserAccount     string          `json:"vmUserAccount"`
	UserPassword    string          `json:"vmUserPassword"`
	Description     string          `json:"description"`
	PublicIP        string          `json:"publicIP"`        // output
	PrivateIP       string          `json:"privateIP"`       // output
	Status          config.VMStatus `json:"status"`          // output
	SystemMessage   string          `json:"systemMessage"`   // output
	Region          RegionInfo      `json:"region"`          // output
	CspViewVmDetail VMDetail        `json:"cspViewVmDetail"` // output
	Credential      string          // private
	Role            string          `json:"role"`
	Csp             config.CSP      `json:"csp"`
	IsCPLeader      bool            `json:"isCPLeader"`
}

type VMInfo struct {
	Name       string     `json:"name"`
	Credential string     // private
	Role       string     `json:"role"`
	Csp        config.CSP `json:"csp"`
	IsCPLeader bool       `json:"isCPLeader"`
}

type RegionInfo struct {
	Region string
	Zone   string
}

type VMDetail struct {
	VMSpecName string
}

const (
	remoteTargetPath = "/tmp"
)

func (self *VM) CheckConnectivity(sshInfo *ssh.SSHInfo) error {
	deadline := 10
	timeout := time.Second * time.Duration(deadline)
	conn, err := net.DialTimeout("tcp", sshInfo.ServerPort, timeout)
	if err != nil {
		logger.Infof(fmt.Sprintf("check connectivity failed.. retry (name=%s, server=%s, cause=%v)", self.Name, sshInfo.ServerPort, err))
		return err
	}
	if conn != nil {
		defer conn.Close()
		return nil
	}

	return errors.New(fmt.Sprintf("Conn is nil (name=%s, server=%s)", self.Name, sshInfo.ServerPort))
}

func (self *VM) CheckConnectSSH(sshInfo *ssh.SSHInfo) error {
	cmd := "/bin/hostname"
	_, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		logger.Infof(fmt.Sprintf("check connect ssh failed.. retry (server=%s, cause=%s)", sshInfo.ServerPort, err))
		return err
	}
	return nil
}

func (self *VM) ConnectionTest(sshInfo *ssh.SSHInfo) error {
	retryCheck := 10
	for i := 0; i < retryCheck; i++ {
		err := self.CheckConnectivity(sshInfo)
		if err == nil {
			logger.Infof(fmt.Sprintf("check connectivity passed (name=%s, server=%s)", self.Name, sshInfo.ServerPort))

			err = self.CheckConnectSSH(sshInfo)
			if err == nil {
				logger.Infof(fmt.Sprintf("check connect ssh passed (name=%s, server=%s)", self.Name, sshInfo.ServerPort))
				break
			}
		}
		if i == retryCheck-1 {
			return errors.New(fmt.Sprintf("Cannot do ssh, the port is not opened (name=%s, server=%s)", self.Name, sshInfo.ServerPort))
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}

func (self *VM) CopyScripts(sshInfo *ssh.SSHInfo, networkCni string) error {
	sourcePath := fmt.Sprintf("%s/src/scripts", *config.Config.AppRootPath)
	sourceFiles := []string{config.BOOTSTRAP_FILE, config.SYSTEMD_SERVICE_FILE}
	if self.Role == config.CONTROL_PLANE && self.IsCPLeader {
		sourceFiles = append(sourceFiles, config.INIT_FILE)
		sourceFiles = append(sourceFiles, config.HA_PROXY_FILE)

		err := self.CreateAddonsDirectory(sshInfo, networkCni)
		if err != nil {
			return errors.New(fmt.Sprintf("create addons directory error (name=%s, cause=%v)", self.Name, err))
		}
		cniFiles := getCniFiles(networkCni)
		sourceFiles = append(sourceFiles, cniFiles...)
	}
	if networkCni == config.NETWORKCNI_CANAL {
		sourceFiles = append(sourceFiles, config.MCKS_BOOTSTRAP_CANAL_FILE)
	} else {
		sourceFiles = append(sourceFiles, config.MCKS_BOOTSTRAP_KILO_FILE)
	}

	logger.Infof("start script file copy (vm=%s, src=%s, dest=%s)\n", self.Name, sourcePath, remoteTargetPath)
	for _, f := range sourceFiles {
		src := fmt.Sprintf("%s/%s", sourcePath, f)
		dest := fmt.Sprintf("%s/%s", remoteTargetPath, f)
		if err := ssh.SSHCopy(*sshInfo, src, dest); err != nil {
			return errors.New(fmt.Sprintf("copy scripts error (server=%s, cause=%s)", sshInfo.ServerPort, err))
		}
	}
	logger.Infof("end script file copy (vm=%s, server=%s)\n", self.Name, sshInfo.ServerPort)
	return nil
}

func (self *VM) CreateAddonsDirectory(sshInfo *ssh.SSHInfo, networkCni string) error {
	addonsPath := fmt.Sprintf("%s/addons/%s", remoteTargetPath, networkCni)

	logger.Infof("create addons directory (vm=%s, path=%s)\n", self.Name, addonsPath)

	cmd := fmt.Sprintf("mkdir -p %s", addonsPath)
	logger.Infof("[CreateAddonsDirectory] %s $ %s", self.Name, cmd)
	_, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (self *VM) SetSystemd(sshInfo *ssh.SSHInfo, networkCni string) error {
	var bsFile string
	if networkCni == config.NETWORKCNI_CANAL {
		bsFile = config.MCKS_BOOTSTRAP_CANAL_FILE
	} else {
		bsFile = config.MCKS_BOOTSTRAP_KILO_FILE
	}

	cmd := fmt.Sprintf("cd %s;./%s %s", remoteTargetPath, bsFile, self.PublicIP)
	logger.Infof("[SetSystemd] %s $ %s", self.Name, cmd)
	_, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		return errors.New(fmt.Sprintf("create mcks-bootstrap error (name=%s)", self.Name))
	}

	cmd = fmt.Sprintf("cd %s;./%s", remoteTargetPath, config.SYSTEMD_SERVICE_FILE)
	logger.Infof("[SetSystemd] %s $ %s", self.Name, cmd)
	_, err = ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		return errors.New(fmt.Sprintf("set systemd service error (name=%s)", self.Name))
	}
	return nil
}

func (self *VM) Bootstrap(sshInfo *ssh.SSHInfo) error {
	cmd := fmt.Sprintf("cd %s;./%s %s", remoteTargetPath, config.BOOTSTRAP_FILE, self.PublicIP)

	logger.Infof("[Bootstrap] %s $ %s", self.Name, cmd)
	result, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		return errors.New("k8s bootstrap error")
	}
	if strings.Contains(result, "kubectl set on hold") {
		return nil
	} else {
		return errors.New(fmt.Sprintf("k8s bootstrap failed (name=%s)", self.Name))
	}
}

func (self *VM) InstallHAProxy(sshInfo *ssh.SSHInfo, IPs []string) error {
	var servers string
	for i, ip := range IPs {
		servers += fmt.Sprintf("  server  api%d  %s:6443  check", i+1, ip)
		if i < len(IPs)-1 {
			servers += "\\n"
		}
	}
	cmd := fmt.Sprintf("sudo sed 's/^{{SERVERS}}/%s/g' %s/%s", servers, remoteTargetPath, config.HA_PROXY_FILE)
	logger.Infof("[InstallHAProxy] %s $ %s", self.Name, cmd)
	result, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		logger.Warnf("get haproxy command error (name=%s, cause=%v)", self.Name, err)
		return err
	}
	logger.Infof("[InstallHAProxy] %s $ %s", self.Name, result)
	_, err = ssh.SSHRun(*sshInfo, result)
	if err != nil {
		logger.Warnf("install haproxy error (name=%s, cause=%v)", self.Name, err)
		return err
	}
	return nil
}

func (self *VM) ControlPlaneInit(sshInfo *ssh.SSHInfo, reqKubernetes Kubernetes) ([]string, string, error) {
	var joinCmd []string

	cmd := fmt.Sprintf("cd %s;./%s %s %s %s %s", remoteTargetPath, config.INIT_FILE, reqKubernetes.PodCidr, reqKubernetes.ServiceCidr, reqKubernetes.ServiceDnsDomain, self.PublicIP)
	logger.Infof("[ControlPlaneInit] %s $ %s", self.Name, cmd)
	cpInitResult, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		logger.Warnf("control plane init error (name=%s, cause=%v)", self.Name, err)
		return nil, "", errors.New("k8s control plane node init error")
	}
	if strings.Contains(cpInitResult, "Your Kubernetes control-plane has initialized successfully") {
		joinCmd = getJoinCmd(cpInitResult)
	} else {
		return nil, "", errors.New(fmt.Sprintf("control palne init failed (name=%s)", self.Name))
	}

	cmd = "sudo cat /etc/kubernetes/admin.conf"
	clusterConfig, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		logger.Errorf("Error while running cmd %s (vm=%s, cause=%v)", cmd, self.Name, err)
	}

	return joinCmd, clusterConfig, nil
}

func (self *VM) InstallNetworkCNI(sshInfo *ssh.SSHInfo, networkCni string) error {
	var cmd string
	cniFiles := getCniFiles(networkCni)

	for _, file := range cniFiles {
		cmd += fmt.Sprintf("sudo kubectl apply -f %s/%s --kubeconfig=/etc/kubernetes/admin.conf;\n", remoteTargetPath, file)
	}

	logger.Infof("[InstallNetworkCNI] %s $ %s", self.Name, cmd)
	_, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		logger.Warnf("networkCNI install failed (name=%s, cause=%v)", self.Name, err)
		return errors.New("NetworkCNI Install error")
	}
	return nil
}

func (self *VM) ControlPlaneJoin(sshInfo *ssh.SSHInfo, CPJoinCmd *string) error {
	if *CPJoinCmd == "" {
		return errors.New("control-plane node join command empty")
	}
	cmd := fmt.Sprintf("sudo %s", *CPJoinCmd)
	logger.Infof("[ControlPlaneJoin] %s $ %s", self.Name, cmd)
	result, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		logger.Warnf("control-plane join error (name=%s, cause=%v)", self.Name, err)
		return errors.New("control-plane node join error")
	}

	if strings.Contains(result, "This node has joined the cluster") {
		_, err = ssh.SSHRun(*sshInfo, "sudo systemctl restart mcks-bootstrap")
		if err != nil {
			logger.Warnf("mcks-bootstrap restart error (name=%s, cause=%v)", self.Name, err)
		}
		return nil
	} else {
		logger.Warnf("control-plane join failed (name=%s)", self.Name)
		return errors.New(fmt.Sprintf("control-plane join failed (name=%s)", self.Name))
	}
}

func (self *VM) WorkerJoin(sshInfo *ssh.SSHInfo, workerJoinCmd *string) error {
	if *workerJoinCmd == "" {
		return errors.New("worker node join command empty")
	}
	cmd := fmt.Sprintf("sudo %s", *workerJoinCmd)
	logger.Infof("[WorkerJoin] %s $ %s", self.Name, cmd)
	result, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		logger.Warnf("worker join error (name=%s, cause=%v)", self.Name, err)
		return errors.New(fmt.Sprintf("worker node join error (name=%s)", self.Name))
	}
	if strings.Contains(result, "This node has joined the cluster") {
		_, err = ssh.SSHRun(*sshInfo, "sudo systemctl restart mcks-bootstrap")
		if err != nil {
			logger.Warnf("mcks-bootstrap restart error (name=%s, cause=%v)", self.Name, err)
		}
		return nil
	} else {
		logger.Warnf("worker join failed (name=%s)", self.Name)
		return errors.New(fmt.Sprintf("worker node join failed (name=%s)", self.Name))
	}
}

func (self *VM) AddNodeLabels(sshInfo *ssh.SSHInfo) error {
	cmd := "/bin/hostname"
	hostName, err := ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		return errors.New(fmt.Sprintf("ssh connection error (server=%s, cause=%s)", sshInfo.ServerPort, err))
	}
	hostName = strings.ToLower(hostName)

	configFile := "admin.conf"
	if self.Role == config.WORKER {
		configFile = "kubelet.conf"
	}

	infos := map[string]interface{}{
		"topology.cloud-barista.github.io/csp": self.Csp,
		"topology.kubernetes.io/region":        self.Region.Region,
	}
	if self.Csp != config.CSP_AZURE {
		infos["topology.kubernetes.io/zone"] = self.Region.Zone
	}

	labels := ""
	for key, value := range infos {
		labels += fmt.Sprintf("%s=%s ", key, value)
	}

	cmd = fmt.Sprintf("sudo kubectl label nodes %s %s --kubeconfig=/etc/kubernetes/%s;", hostName, labels, configFile)
	logger.Infof("[AddNodeLabels] %s $ %s", self.Name, cmd)
	_, err = ssh.SSHRun(*sshInfo, cmd)
	if err != nil {
		return err
	}

	return nil
}

func getJoinCmd(cpInitResult string) []string {
	var join1, join2, join3 string
	joinRegex, _ := regexp.Compile("kubeadm\\sjoin\\s(.*?)\\s--token\\s(.*?)\\n")
	joinRegex2, _ := regexp.Compile("--discovery-token-ca-cert-hash\\ssha256:(.*?)\\n")
	joinRegex3, _ := regexp.Compile("--control-plane --certificate-key(.*?)\\n")

	if joinRegex.MatchString(cpInitResult) {
		join1 = joinRegex.FindString(cpInitResult)
	}
	if joinRegex2.MatchString(cpInitResult) {
		join2 = joinRegex2.FindString(cpInitResult)
	}
	if joinRegex3.MatchString(cpInitResult) {
		join3 = joinRegex3.FindString(cpInitResult)
	}

	return []string{fmt.Sprintf("%s %s %s", join1, join2, join3), fmt.Sprintf("%s %s", join1, join2)}
}

func getCniFiles(cni string) (cniFiles []string) {
	if cni == config.NETWORKCNI_CANAL {
		cniFiles = append(cniFiles, config.CNI_CANAL_FILE)
	} else {
		cniFiles = append(cniFiles, config.CNI_KILO_CRDS_FILE)
		cniFiles = append(cniFiles, config.CNI_KILO_KUBEADM_FILE)
		cniFiles = append(cniFiles, config.CNI_KILO_FLANNEL_FILE)
	}
	return
}
