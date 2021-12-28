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

const (
	VM_USER_ACCOUNT       = "cb-user"
	REMOTE_TARGET_PATH    = "/tmp"
	CNI_CANAL_FILE        = "addons/canal/canal_v3.20.0.yaml"
	CNI_KILO_CRDS_FILE    = "addons/kilo/crds_v0.3.0.yaml"
	CNI_KILO_KUBEADM_FILE = "addons/kilo/kilo-kubeadm-flannel_v0.3.0.yaml"
	CNI_KILO_FLANNEL_FILE = "addons/kilo/kube-flannel_v0.14.0.yaml"
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

func (self *VM) SSHRun(format string, a ...interface{}) (string, error) {

	address := fmt.Sprintf("%s:22", self.PublicIP)
	command := fmt.Sprintf(format, a...)

	output, err := ssh.SSHRun(
		ssh.SSHInfo{
			UserName:   VM_USER_ACCOUNT,
			PrivateKey: []byte(self.Credential),
			ServerPort: address,
		}, command)
	if err != nil {
		logger.Errorf("[%s] failed to run SSH command (server=%s, cause=%v, command=%s, output=%s)", self.Name, address, err, command, output)
	} else {
		logger.Infof("[%s] ssh execute is completed (server=%s, command=%s)", self.Name, address, command)
	}
	return output, err
}

func (self *VM) SSHCopy(source string, destination string) error {

	address := fmt.Sprintf("%s:22", self.PublicIP)

	err := ssh.SSHCopy(
		ssh.SSHInfo{
			UserName:   VM_USER_ACCOUNT,
			PrivateKey: []byte(self.Credential),
			ServerPort: address,
		}, source, destination)
	if err != nil {
		logger.Errorf("[%s] failed to copying files (server=%s, source=%s, destination=%s, cause=%v)", self.Name, address, source, destination, err)
	} else {
		logger.Infof("[%s] file copy is completed (server=%s, source=%s, destination=%s)", self.Name, address, source, destination)
	}
	return err
}

func (self *VM) checkConnectivity() error {

	address := fmt.Sprintf("%s:22", self.PublicIP)
	timeout := time.Second * time.Duration(10)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		logger.Warnf("[%s] failed to checking connectivity.. retry (cause=%v)", self.Name, err)
		return err
	}
	if conn != nil {
		defer conn.Close()
		logger.Infof("[%s] check connectivity is completed (server=%s)", self.Name, address)
		return nil
	}

	logger.Errorf("[%s] failed to checking connectivity (server=%s)", self.Name, address)
	return errors.New(fmt.Sprintf("Check connectivity failed. (vm=%s, cause=connection is nil)", self.Name))
}

func (self *VM) CheckConnectSSH() error {
	if _, err := self.SSHRun("/bin/hostname"); err != nil {
		return errors.New(fmt.Sprintf("Failed to check connect VM. (vm=%s)", self.Name))
	}
	return nil
}

func (self *VM) ConnectionTest() error {

	logger.Infof("[%s] start the process of 'connection test'", self.Name)
	retryCheck := 15
	for i := 0; i < retryCheck; i++ {
		err := self.checkConnectivity()
		if err == nil {
			err = self.CheckConnectSSH()
			if err == nil {
				break
			}
		}
		if i == retryCheck-1 {
			logger.Errorf("[%s] Connection retry test count exceeded.", self.Name)
			return errors.New(fmt.Sprintf("Cannot do ssh, the port is not opened. (vm=%s, connection retry test count exceeded)", self.Name))
		}
		time.Sleep(2 * time.Second)
	}
	logger.Infof("[%s] completed the 'connection test' process", self.Name)
	return nil
}

/* bootstrap */
func (self *VM) Bootstrap(networkCni string) error {

	logger.Infof("[%s] Start copying files. of 'bootstrap'", self.Name)

	// 1. copy files
	//  - list-up copy bootstrap files
	sourcePath := fmt.Sprintf("%s/src/scripts", *config.Config.AppRootPath)
	sourceFiles := []string{"bootstrap.sh"}

	//  - list-up for leader-node
	if self.Role == config.CONTROL_PLANE && self.IsCPLeader {
		sourceFiles = append(sourceFiles, "haproxy.sh", "k8s-init.sh")
		if _, err := self.SSHRun("mkdir -p %s/addons/%s", REMOTE_TARGET_PATH, networkCni); err != nil {
			return errors.New(fmt.Sprintf("Failed to create a addon directory. (node=%s, path=%s)", self.Name, "addons/"+networkCni))
		}
		if networkCni == config.NETWORKCNI_CANAL {
			sourceFiles = append(sourceFiles, CNI_CANAL_FILE)
		} else {
			sourceFiles = append(sourceFiles, CNI_KILO_CRDS_FILE, CNI_KILO_KUBEADM_FILE, CNI_KILO_FLANNEL_FILE)
		}
	}

	//  - copy list-up files
	logger.Infof("[%s] Start copying files. (files=%v)", self.Name, sourceFiles)
	for _, f := range sourceFiles {
		src := fmt.Sprintf("%s/%s", sourcePath, f)
		dest := fmt.Sprintf("%s/%s", REMOTE_TARGET_PATH, f)
		if err := self.SSHCopy(src, dest); err != nil {
			logger.Errorf("[%s] Failed to copy bootstrap files (source=%s, destination=%s, cause=%v)", self.Name, src, dest, err)
			return errors.New(fmt.Sprintf("Failed to copy bootstrap files. (vm=%s, file=%s)", self.Name, f))
		}
	}

	// 2. execute bootstrap.sh
	if output, err := self.SSHRun("%s/bootstrap.sh %s %s %s %s %s", REMOTE_TARGET_PATH, self.Csp, self.Region.Region, self.Name, self.PublicIP, networkCni); err != nil {
		return errors.New(fmt.Sprintf("Failed to execute bootstrap.sh (vm=%s)", self.Name))
	} else if !strings.Contains(output, "kubectl set on hold") {
		logger.Errorf("[%s] failed to execute bootstrap.sh (cause='kubectl not set on hold')", self.Name)
		return errors.New(fmt.Sprintf("Failed to execute bootstrap.sh shell. (vm=%s, cause='kubectl not set on hold')", self.Name))
	}

	logger.Infof("[%s] completed the'bootstrap process", self.Name)
	return nil

}

/* setup haproxy */
func (self *VM) InstallHAProxy(IPs []string) error {

	logger.Infof("[%s] start the process of 'set up HA'", self.Name)

	var servers string
	for i, ip := range IPs {
		servers += fmt.Sprintf("  server  api%d  %s:6443  check", i+1, ip)
		if i < len(IPs)-1 {
			servers += "\\n"
		}
	}

	if output, err := self.SSHRun("sudo sed 's/^{{SERVERS}}/%s/g' %s/%s", servers, REMOTE_TARGET_PATH, "haproxy.sh"); err != nil {
		return errors.New(fmt.Sprintf("Failed to set up haproxy. (vm=%s, shell=%s)", self.Name, "haproxy.sh"))
	} else {
		if _, err = self.SSHRun(output); err != nil {
			return errors.New(fmt.Sprintf("Failed to set up haproxy. (vm=%s, command='%s')", self.Name, output))
		}
	}

	logger.Infof("[%s] completed the 'set up HA' process", self.Name)
	return nil
}

// coantrol-plane init
func (self *VM) ControlPlaneInit(reqKubernetes Kubernetes) ([]string, string, error) {

	logger.Infof("[%s] start the process of 'control-plane init.'", self.Name)

	var joinCmd []string

	if output, err := self.SSHRun("cd %s;./%s %s %s %s %s", REMOTE_TARGET_PATH, "k8s-init.sh", reqKubernetes.PodCidr, reqKubernetes.ServiceCidr, reqKubernetes.ServiceDnsDomain, self.PublicIP); err != nil {
		return nil, "", errors.New(fmt.Sprintf("Failed to initialize control-plane. (vm=%s, shell=%s)", self.Name, "k8s-init.sh"))
	} else if strings.Contains(output, "Your Kubernetes control-plane has initialized successfully") {
		joinCmd = getJoinCmd(output)
	} else {
		logger.Errorf("[%s] failed to initialize control-plane (the output not contains 'Your Kubernetes control-plane has initialized successfully')")
		return nil, "", errors.New(fmt.Sprintf("Failed to initialize control-plane. (vm=%s)", self.Name))
	}

	ouput, _ := self.SSHRun("sudo cat /etc/kubernetes/admin.conf")

	logger.Infof("[%s] completed the 'control-plane init.' process", self.Name)
	return joinCmd, ouput, nil
}

/* install network-cni */
func (self *VM) InstallNetworkCNI(networkCni string) error {

	logger.Infof("[%s] start the process of 'install network-cni'", self.Name)

	var cmd string
	cniFiles := getCniFiles(networkCni)

	for _, file := range cniFiles {
		cmd += fmt.Sprintf("sudo kubectl apply -f %s/%s --kubeconfig=/etc/kubernetes/admin.conf;\n", REMOTE_TARGET_PATH, file)
	}

	if _, err := self.SSHRun(cmd); err != nil {
		return err
	}

	logger.Infof("[%s] completed the 'install network-cni' process", self.Name)
	return nil
}

func (self *VM) ControlPlaneJoin(CPJoinCmd *string) error {

	logger.Infof("[%s] start the process of 'control-plane join'", self.Name)

	if *CPJoinCmd == "" {
		logger.Errorf("[%s] control-plane join-command is empty", self.Name)
		return errors.New(fmt.Sprintf("The control-plane join-command is empty. (node=%s)", self.Name))
	}
	if output, err := self.SSHRun("sudo %s", *CPJoinCmd); err != nil {
		return errors.New(fmt.Sprintf("Failed to join control-plane. (node=%s)", self.Name))
	} else if strings.Contains(output, "This node has joined the cluster") {
		if _, err = self.SSHRun("sudo systemctl restart mcks-bootstrap"); err != nil {
			logger.Warnf("[%s] mcks-bootstrap restart error (command='sudo systemctl restart mcks-bootstrap' cause=%v)", self.Name, err)
		}
	} else {
		logger.Errorf("[%s] control-plane join failed (the output not contains 'This node has joined the cluster')", self.Name)
		return errors.New(fmt.Sprintf("Failed to join control-plane. (vm=%s)", self.Name))
	}

	logger.Infof("[%s] completed the 'control-plane join' process", self.Name)
	return nil
}

func (self *VM) WorkerJoin(workerJoinCmd *string) error {

	logger.Infof("[%s] start the process of 'worker-node join'", self.Name)

	if output, err := self.SSHRun("sudo %s", *workerJoinCmd); err != nil {
		return errors.New(fmt.Sprintf("Failed to join worker-node. (vm=%s)", self.Name))
	} else if strings.Contains(output, "This node has joined the cluster") {
		if _, err = self.SSHRun("sudo systemctl restart mcks-bootstrap"); err != nil {
			logger.Warnf("[%s] mcks-bootstrap restart error (command='sudo systemctl restart mcks-bootstrap', cause=%v)", self.Name, err)
		}
	} else {
		logger.Errorf("[%s] worker join failed (the output not contains 'This node has joined the cluster')", self.Name)
		return errors.New(fmt.Sprintf("Failed to execute 'kubeadm join' command. (vm=%s)", self.Name))
	}

	logger.Infof("[%s] completed the 'worker-node join' process", self.Name)
	return nil

}

func (self *VM) AddNodeLabels() error {

	configFile := "admin.conf"
	if self.Role == config.WORKER {
		configFile = "kubelet.conf"
	}

	infos := map[string]interface{}{
		config.LABEL_KEY_CSP:    self.Csp,
		config.LABEL_KEY_REGION: self.Region.Region,
	}
	if self.Region.Zone != "" {
		infos[config.LABEL_KEY_ZONE] = self.Region.Zone
	}

	labels := ""
	for key, value := range infos {
		labels += fmt.Sprintf("%s=%s ", key, value)
	}

	if _, err := self.SSHRun("sudo kubectl label nodes %s %s --kubeconfig=/etc/kubernetes/%s;", self.Name, labels, configFile); err != nil {
		return errors.New(fmt.Sprintf("Failed to set label. (node=%s, label=%s)", self.Name, labels))
	}

	logger.Infof("[%s] set node label (label=%s)", self.Name, labels)
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
		cniFiles = append(cniFiles, CNI_CANAL_FILE)
	} else {
		cniFiles = append(cniFiles, CNI_KILO_CRDS_FILE)
		cniFiles = append(cniFiles, CNI_KILO_KUBEADM_FILE)
		cniFiles = append(cniFiles, CNI_KILO_FLANNEL_FILE)
	}
	return
}
