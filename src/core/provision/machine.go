package provision

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/cloud-barista/cb-mcks/src/core/app"
	"github.com/cloud-barista/cb-mcks/src/core/model"
	ssh "github.com/cloud-barista/cb-spider/cloud-control-manager/vm-ssh"

	logger "github.com/sirupsen/logrus"
)

/* ssh execution */
func (self *Machine) executeSSH(format string, a ...interface{}) (string, error) {

	address := fmt.Sprintf("%s:22", self.PublicIP)
	command := fmt.Sprintf(format, a...)

	logger.Infof("[%s] SSH executing. (server=%s, command='%s')", self.Name, address, command)
	output, err := ssh.SSHRun(
		ssh.SSHInfo{
			UserName:   self.Username,
			PrivateKey: []byte(self.Credential),
			ServerPort: address,
		}, command)
	if err != nil {
		if strings.Contains(err.Error(), "handshake failed") {
			if self.Username == "" {
				logger.Warnf("[%s] Failed to run SSH command - username is empty (server=%s, key=%s, command='%s', cause='%v')", self.Name, address, len(self.Credential), command, err)
			} else if self.Credential == "" {
				logger.Warnf("[%s] Failed to run SSH command - private-key is empty (server=%s, command='%s', cause='%v')", self.Name, address, command, err)
			} else {
				logger.Warnf("[%s] Failed to run SSH command. (server=%s, command='%s', cause='%v')", self.Name, address, command, err)
			}
		} else {
			logger.Warnf("[%s] Failed to run SSH command. (server=%s, username=%s, key=%s, command='%s', cause='%v')", self.Name, address, self.Username, len(self.Credential), command, err)
		}
	}
	return output, err
}

/* scp execution */
func (self *Machine) executeSCP(source string, destination string) error {

	//validate files exist
	if _, err := os.Stat(source); err != nil {
		return errors.New(fmt.Sprintf("SCP source file does not exist (%s)", source))
	}

	address := fmt.Sprintf("%s:22", self.PublicIP)

	err := ssh.SSHCopy(
		ssh.SSHInfo{
			UserName:   self.Username,
			PrivateKey: []byte(self.Credential),
			ServerPort: address,
		}, source, destination)
	if err != nil {
		logger.Warnf("[%s] Failed to copy files. (server=%s, destination='%s', cause='%v')", self.Name, address, destination, err)
	} else {
		logger.Infof("[%s] File copy has been completed. (server=%s, destination='%s')", self.Name, address, destination)
	}
	return err
}

/* ssh onnectivity test */
func (self *Machine) checkConnectivity() error {

	address := fmt.Sprintf("%s:22", self.PublicIP)
	timeout := time.Second * time.Duration(10)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	if conn != nil {
		defer conn.Close()
		return nil
	}

	return errors.New("Failed to validate connectivity.")
}

/* ssh connect test */
func (self *Machine) ConnectionTest() error {

	retryCheck := 15
	for i := 0; i < retryCheck; i++ {
		err := self.checkConnectivity()
		if err == nil {
			// verify SSH connect
			if _, err := self.executeSSH("/bin/hostname"); err == nil {
				break
			} else {
				logger.Infof("[%s] Failed to validate SSH connection. (ip=%s, retry=%d)", self.Name, self.PublicIP, i)
			}
		} else {
			logger.Infof("[%s] Dial timeout. (dial=tcp://%s:22, retry=%d)", self.Name, self.PublicIP, i)
		}
		if i == retryCheck-1 {
			return errors.New(fmt.Sprintf("SSH connection retry count has exceeded. (node=%s, ip=%s)", self.Name, self.PublicIP))
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}

func (self *Machine) GetHostname() (string, error) {
	if self.CSP == app.CSP_AWS {
		return awsGetMetadataLocalHostname(self)
	} else if self.CSP == app.CSP_OPENSTACK {
		return self.NameInCsp, nil
	} else if self.CSP == app.CSP_NCPVPC {
		return ncpvpcGetMetadataServerName(self)
	} else {
		return "", errors.New(fmt.Sprintf("Failed to get the fullname: no CSP (node=%s)", self.Name))
	}
}

/* bootstrap */
func (self *Machine) bootstrap(clusterInfo *model.Cluster) error {

	//verfiy
	if self.CSP == "" || self.Region == "" || self.Name == "" || self.PublicIP == "" || clusterInfo.ServiceType == "" {
		return errors.New(fmt.Sprintf("There are mandatory fields. (node=%s, role=%s, csp=%s, region=%s, publicip=%s, servicetype=%s)", self.Name, self.Role, self.CSP, self.Region, self.PublicIP, clusterInfo.ServiceType))
	}

	// 1. copy files
	//  - list-up copy bootstrap files
	sourcePath := fmt.Sprintf("%s/src/scripts", *app.Config.AppRootPath)
	sourceFiles := []string{"bootstrap.sh"}

	//  - list-up for control-plane
	if self.Role == app.CONTROL_PLANE {
		sourceFiles = append(sourceFiles, "haproxy.sh")
		if _, err := self.executeSSH("mkdir -p %s/addons/cni/%s", REMOTE_TARGET_PATH, clusterInfo.NetworkCni); err != nil {
			return errors.New(fmt.Sprintf("Failed to create a addon directory. (node=%s, path='%s')", self.Name, "addons/cni/"+clusterInfo.NetworkCni))
		}

		if clusterInfo.ServiceType != app.ST_SINGLE {
			if clusterInfo.NetworkCni == app.NETWORKCNI_CANAL {
				sourceFiles = append(sourceFiles, CNI_CANAL_FILE)
			} else {
				sourceFiles = append(sourceFiles, CNI_KILO_CRDS_FILE, CNI_KILO_KUBEADM_FILE, CNI_KILO_FLANNEL_FILE)
			}
		} else { // clusterInfo.ServiceType == app.ST_SINGLE
			if clusterInfo.NetworkCni == app.NETWORKCNI_FLANNEL {
				sourceFiles = append(sourceFiles, CNI_FLANNEL_FILE)
			} else if clusterInfo.NetworkCni == app.NETWORKCNI_CALICO {
				sourceFiles = append(sourceFiles, CNI_CALICO_FILE)
			}

			sourceFiles = append(sourceFiles, "gen-cloud-config.sh")
			if _, err := self.executeSSH("mkdir -p %s/addons/ccm/%s", REMOTE_TARGET_PATH, self.CSP); err != nil {
				return errors.New(fmt.Sprintf("Failed to create a addon directory. (node=%s, path='%s')", self.Name, "addons/ccm/"+self.CSP))
			}

			if self.CSP == app.CSP_AWS {
				sourceFiles = append(sourceFiles,
					CCM_AWS_ROLE_SA_FILE,
					CCM_AWS_DS_FILE)
			} else if self.CSP == app.CSP_OPENSTACK {
				sourceFiles = append(sourceFiles,
					CCM_OPENSTACK_ROLE_BINDINGS_FILE,
					CCM_OPENSTACK_ROLES_FILE,
					CCM_OPENSTACK_DS_FILE)
			} else if self.CSP == app.CSP_NCPVPC {
				sourceFiles = append(sourceFiles,
					CCM_NCPVPC_ROLE_SA_FILE,
					CCM_NCPVPC_DS_FILE)
			}
		}

		if clusterInfo.Etcd == app.ETCD_EXTERNAL {
			sourceFiles = append(sourceFiles, "etcd-conf.sh", "k8s-init-etcd.sh")
			if clusterInfo.CpLeader == self.Name {
				sourceFiles = append(sourceFiles, "etcd-ca.sh")
			}
		} else {
			sourceFiles = append(sourceFiles, "k8s-init.sh")
		}

		if _, err := self.executeSSH("mkdir -p %s/addons/%s", REMOTE_TARGET_PATH, "nfs"); err != nil {
			return errors.New(fmt.Sprintf("Failed to create a addon directory. (node=%s, path='%s')", self.Name, "addons/"+"nfs"))
		}
		sourceFiles = append(sourceFiles, SC_NFS_RBAC_FILE, SC_NFS_CLASS_FILE, "addons/nfs/deploy_v4.0.16.sh")
	}

	//  - copy list-up files
	for _, f := range sourceFiles {
		src := fmt.Sprintf("%s/%s", sourcePath, f)
		dest := fmt.Sprintf("%s/%s", REMOTE_TARGET_PATH, f)
		if err := self.executeSCP(src, dest); err != nil {
			return errors.New(fmt.Sprintf("Failed to copy bootstrap files. (node=%s, destination='%s', cause='%v')", self.Name, dest, err))
		}
	}

	// 2. execute bootstrap.sh
	var hostname string = self.Name
	if clusterInfo.ServiceType == app.ST_SINGLE {
		var err error
		if hostname, err = self.GetHostname(); err != nil {
			return err
		}
	}

	if _, err := self.executeSSH(REMOTE_TARGET_PATH+"/bootstrap.sh %s %s %s %s %s %s", clusterInfo.Version, self.CSP, hostname, self.PublicIP, clusterInfo.NetworkCni, clusterInfo.ServiceType); err != nil {
		return errors.New(fmt.Sprintf("Failed to execute bootstrap.sh (node=%s)", self.Name))
	}

	return nil

}

/* control-plane join */
func (self *ControlPlaneMachine) JoinControlPlane(CPJoinCmd *string) error {

	if *CPJoinCmd == "" {
		return errors.New("Control-plane-join-command is a mandatory parameter.")
	}

	if output, err := self.executeSSH("sudo %s", *CPJoinCmd); err != nil {
		return errors.New(fmt.Sprintf("Failed to join control-plane. (node=%s)", self.Name))
	} else if strings.Contains(output, "This node has joined the cluster") {
		if _, err = self.executeSSH("sudo systemctl restart mcks-bootstrap"); err != nil {
			logger.Warnf("[%s] mcks-bootstrap restart error (command='sudo systemctl restart mcks-bootstrap' cause='%v')", self.Name, err)
		}
	} else {
		return errors.New(fmt.Sprintf("Failed to join control-plane. (node=%s)", self.Name))
	}

	return nil
}

/* woker node join */
func (self *WorkerNodeMachine) JoinWorker(workerJoinCmd *string) error {

	if *workerJoinCmd == "" {
		return errors.New("Worker-join-command is a mandatory parameter.")
	}

	if output, err := self.executeSSH("sudo %s", *workerJoinCmd); err != nil {
		return errors.New(fmt.Sprintf("Failed to join worker-node. (node=%s)", self.Name))
	} else if strings.Contains(output, "This node has joined the cluster") {
		if _, err = self.executeSSH("sudo systemctl restart mcks-bootstrap"); err != nil {
			logger.Warnf("[%s] mcks-bootstrap restart error (command='sudo systemctl restart mcks-bootstrap', cause='%v')", self.Name, err)
		}
	} else {
		return errors.New(fmt.Sprintf("Failed to execute 'kubeadm join' command. (node=%s)", self.Name))
	}

	return nil

}

/* new instance of node-entity */
func (self *Machine) NewNode() *model.Node {

	return &model.Node{
		Model:       model.Model{Kind: app.KIND_NODE, Name: self.Name},
		Credential:  self.Credential,
		Role:        self.Role,
		Spec:        self.Spec,
		Csp:         self.CSP,
		PublicIP:    self.PublicIP,
		PrivateIP:   self.PrivateIP,
		CspLabel:    fmt.Sprintf("%s=%s", app.LABEL_KEY_CSP, string(self.CSP)),
		RegionLabel: fmt.Sprintf("%s=%s", app.LABEL_KEY_REGION, self.Region),
		ZoneLabel:   fmt.Sprintf("%s=%s", app.LABEL_KEY_ZONE, self.Zone),
	}
}
