package provision

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/cloud-barista/cb-mcks/src/core/app"
	"github.com/cloud-barista/cb-mcks/src/core/model"
	"github.com/cloud-barista/cb-mcks/src/core/tumblebug"
	"github.com/cloud-barista/cb-mcks/src/utils/lang"

	"golang.org/x/sync/errgroup"
)

/* new a instance of provider */
func NewProvisioner(cluster *model.Cluster) *Provisioner {
	provisioner := &Provisioner{
		Cluster:              cluster,
		WorkerNodeMachines:   make(map[string]*WorkerNodeMachine),
		ControlPlaneMachines: make(map[string]*ControlPlaneMachine),
	}
	if cluster.CpLeader != "" {
		for _, node := range cluster.Nodes {
			if node.Name == cluster.CpLeader {
				provisioner.leader = &ControlPlaneMachine{Machine: &Machine{
					Name:       node.Name,
					PublicIP:   node.PublicIP,
					Username:   tumblebug.VM_USER_ACCOUNT,
					Credential: node.Credential,
				}}
			}
		}
	}
	return provisioner
}

/* append a control-plane-machine */
func (self *Provisioner) AppendControlPlaneMachine(name string, csp app.CSP, region string, zone string, credential string) {

	machine := &ControlPlaneMachine{
		Machine: &Machine{
			Name:       name,
			CSP:        csp,
			Role:       app.CONTROL_PLANE,
			Region:     region,
			Zone:       zone,
			Credential: credential,
		},
	}
	self.ControlPlaneMachines[name] = machine
	if len(self.ControlPlaneMachines) == 1 {
		self.leader = machine
	}

}

/* append a worker-node-machine */
func (self *Provisioner) AppendWorkerNodeMachine(name string, csp app.CSP, region string, zone string, credential string) {
	self.WorkerNodeMachines[name] = &WorkerNodeMachine{
		Machine: &Machine{
			Name:       name,
			CSP:        csp,
			Role:       app.WORKER,
			Region:     region,
			Zone:       zone,
			Credential: credential,
		},
	}
}

/* set fileds each machines (public-ip, region, zone, spec, username) */
func (self *Provisioner) BindVM(vms []tumblebug.VM) ([]*model.Node, error) {

	nodes := []*model.Node{}
	for _, vm := range vms {

		// validate created vm
		if vm.Status == tumblebug.VMSTATUS_FAILED {
			status := app.Status{}
			if err := json.Unmarshal([]byte(vm.SystemMessage), &status); err != nil {
				status.Message = vm.SystemMessage
			}
			return nil, errors.New(fmt.Sprintf("Failed to create a vm (status=%s, cause='%s')", vm.Status, status.Message))
		} else if vm.PublicIP == "" {
			return nil, errors.New(fmt.Sprintf("Failed to create a vm (status=%s, cause='unbounded public-ip')", vm.Status))
		}

		var machine *Machine

		if self.leader.Name == vm.Name {
			machine = self.leader.Machine
		} else {
			_, exists := self.ControlPlaneMachines[vm.Name]
			if exists {
				machine = self.ControlPlaneMachines[vm.Name].Machine
			} else {
				_, exists = self.WorkerNodeMachines[vm.Name]
				if exists {
					machine = self.WorkerNodeMachines[vm.Name].Machine
				}
			}
		}
		if machine != nil {
			machine.PublicIP = vm.PublicIP
			machine.PrivateIP = vm.PrivateIP
			machine.Username = lang.NVL(vm.UserAccount, tumblebug.VM_USER_ACCOUNT)
			machine.Region = lang.NVL(vm.Region.Region, machine.Region) // region, zone 공백인 경우가 간혹 있음
			machine.Zone = lang.NVL(vm.Region.Zone, machine.Zone)
			machine.Spec = vm.CspViewVmDetail.VMSpecName
			nodes = append(nodes, machine.NewNode())
		} else {
			return nil, errors.New(fmt.Sprintf("Can't be found node by name '%s'", vm.Name))
		}
	}

	return nodes, nil
}

/* bootstrap */
func (self *Provisioner) Bootstrap() error {

	// bootstrap
	eg, _ := errgroup.WithContext(context.Background())

	for _, m := range self.GetMachinesAll() {
		machine := m
		eg.Go(func() error {
			if err := machine.ConnectionTest(); err != nil {
				return err
			}
			if err := machine.bootstrap(self.Cluster); err != nil {
				return err
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

/* setup haproxy */
func (self *Provisioner) InstallHAProxy() error {

	var servers string
	for _, machine := range self.ControlPlaneMachines {
		servers += fmt.Sprintf("  server  %s  %s:6443  check\\n", machine.Name, machine.PrivateIP)
	}
	if output, err := self.leader.executeSSH("sudo sed 's/^{{SERVERS}}/%s/g' %s/%s", servers, REMOTE_TARGET_PATH, "haproxy.sh"); err != nil {
		return err
	} else {
		if _, err = self.leader.executeSSH(output); err != nil {
			return err
		}
	}

	return nil
}
func (self *Provisioner) InitExternalEtcd() error {
	var ips string
	var hosts string

	for _, machine := range self.ControlPlaneMachines {
		ips += fmt.Sprintf("%s ", machine.PrivateIP)
		hosts += fmt.Sprintf("%s %s ", machine.Name, machine.PrivateIP)
	}
	if _, err := self.leader.executeSSH("sudo echo '%s'>$HOME/id_rsa; sudo mv $HOME/id_rsa $HOME/.ssh/id_rsa; sudo chmod 600 $HOME/.ssh/id_rsa", self.leader.Credential); err != nil {
		return errors.New(fmt.Sprintf("Failed to create private-key."))
	}
	if _, err := self.leader.executeSSH(REMOTE_TARGET_PATH+"/etcd-ca.sh %s", ips); err != nil {
		return errors.New(fmt.Sprintf("Failed to create etcd certificates. (etcd-ca.sh)"))
	}
	for _, machine := range self.ControlPlaneMachines {
		if _, err := self.leader.executeSSH("scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r /tmp/ca/* cb-user@%s:", machine.PublicIP); err != nil {
			return errors.New(fmt.Sprintf("[%s] Failed to copy certificate.", machine.Name))
		}

		if _, err := machine.executeSSH(REMOTE_TARGET_PATH+"/etcd-conf.sh %s", hosts); err != nil {
			return errors.New(fmt.Sprintf("[%s] Failed to configure etcd cluster. (etcd-conf.sh)", machine.Name))
		}
	}
	return nil
}

// coantrol-plane init
func (self *Provisioner) InitControlPlane(kubernetesConfigReq app.ClusterConfigKubernetesReq) ([]string, string, error) {

	var joinCmd []string
	var port string
	if self.Cluster.Loadbalancer == app.LB_HAPROXY {
		port = "9998"
	} else {
		port = "6443"
	}
	if self.Cluster.Etcd == app.ETCD_EXTERNAL {
		var etcdIp string
		for _, machine := range self.ControlPlaneMachines {
			etcdIp += fmt.Sprintf("%s ", machine.PrivateIP)
		}
		if output, err := self.leader.executeSSH("cd %s;./%s %s %s %s %s %s %s", REMOTE_TARGET_PATH, "k8s-init-etcd.sh", kubernetesConfigReq.PodCidr, kubernetesConfigReq.ServiceCidr, kubernetesConfigReq.ServiceDnsDomain, self.leader.PublicIP, port, etcdIp); err != nil {
			return nil, "", errors.New("Failed to initialize control-plane. (k8s-init-etcd.sh)")
		} else if strings.Contains(output, "Your Kubernetes control-plane has initialized successfully") {
			joinCmd = getJoinCmd(output)
		} else {
			return nil, "", errors.New("to initialize control-plane (the output not contains 'Your Kubernetes control-plane has initialized successfully')")
		}
	} else {
		if output, err := self.leader.executeSSH("cd %s;./%s %s %s %s %s %s", REMOTE_TARGET_PATH, "k8s-init.sh", kubernetesConfigReq.PodCidr, kubernetesConfigReq.ServiceCidr, kubernetesConfigReq.ServiceDnsDomain, self.leader.PublicIP, port); err != nil {
			return nil, "", errors.New("Failed to initialize control-plane. (k8s-init.sh)")
		} else if strings.Contains(output, "Your Kubernetes control-plane has initialized successfully") {
			joinCmd = getJoinCmd(output)
		} else {
			return nil, "", errors.New("to initialize control-plane (the output not contains 'Your Kubernetes control-plane has initialized successfully')")
		}
	}

	ouput, _ := self.leader.executeSSH("sudo cat /etc/kubernetes/admin.conf")

	return joinCmd, ouput, nil
}

/* install network-cni */
func (self *Provisioner) InstallNetworkCni() error {

	cniYamls := []string{}
	if self.Cluster.NetworkCni == app.NETWORKCNI_CANAL {
		cniYamls = append(cniYamls, CNI_CANAL_FILE)
	} else {
		cniYamls = append(cniYamls, CNI_KILO_FLANNEL_FILE)
		cniYamls = append(cniYamls, CNI_KILO_CRDS_FILE)
		cniYamls = append(cniYamls, CNI_KILO_KUBEADM_FILE)
	}

	for _, file := range cniYamls {
		if _, err := self.Kubectl("apply -f %s/%s", REMOTE_TARGET_PATH, file); err != nil {
			return err
		}
	}

	return nil
}

func (self *Provisioner) InstallStorageClassNFS(storageReq app.ClusterStorageClassNfsReq) error {

	storageYamls := []string{}
	if storageReq.Server != "" {
		storageYamls = append(storageYamls, SC_NFS_RBAC_FILE)
		storageYamls = append(storageYamls, SC_NFS_CLASS_FILE)
	}

	for _, file := range storageYamls {
		if _, err := self.Kubectl("apply -f %s/%s", REMOTE_TARGET_PATH, file); err != nil {
			return err
		}
	}
	if storageReq.Server != "" {
		if _, err := self.leader.executeSSH("cd %s;./%s %s %s ", REMOTE_TARGET_PATH, "addons/nfs/deploy_v4.0.16.sh", storageReq.Path, storageReq.Server); err != nil {
			return errors.New("Failed to setup storageCalss controla-plane.")
		}
	}

	return nil
}

/* assign node labels */
func (self *Provisioner) AssignNodeLabelAnnotation() error {

	// commons labels
	for _, machine := range self.GetMachinesAll() {
		if _, err := self.Kubectl("label nodes %s %s=%s", machine.Name, app.LABEL_KEY_CSP, machine.CSP); err != nil {
			return err
		}
		if _, err := self.Kubectl("label nodes %s %s=%s", machine.Name, app.LABEL_KEY_REGION, machine.Region); err != nil {
			return err
		}
		if _, err := self.Kubectl("label nodes %s %s=%s", machine.Name, app.LABEL_KEY_ZONE, machine.Zone); err != nil {
			return err
		}
	}

	// network-cni annotations
	if self.Cluster.NetworkCni == app.NETWORKCNI_KILO {
		for _, machine := range self.GetMachinesAll() {
			// use a full mesh network
			if _, err := self.Kubectl("annotate nodes %s kilo.squat.ai/location=%s", machine.Name, machine.Name); err != nil {
				return err
			}
			if _, err := self.Kubectl("annotate nodes %s kilo.squat.ai/persistent-keepalive=25", machine.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

/* new generate worker-node join command */
func (self *Provisioner) NewWorkerJoinCommand() (string, error) {

	if joinCommand, err := self.leader.executeSSH("sudo kubeadm token create --print-join-command"); err != nil {
		return "", err
	} else if joinCommand == "" {
		return "", errors.New("join command is empty")
	} else {
		return joinCommand, nil
	}
}

/* execute kubectl */
func (self *Provisioner) Kubectl(format string, a ...interface{}) (string, error) {

	command := fmt.Sprintf(format, a...)
	command = fmt.Sprintf("sudo kubectl %s --kubeconfig=/etc/kubernetes/admin.conf", command)
	if output, err := self.leader.executeSSH(command); err != nil {
		return "", errors.New(fmt.Sprintf("Failed to kubectl. (command='%s')", command))
	} else {
		return output, nil
	}

}

/* get machines */
func (self *Provisioner) GetMachinesAll() []*Machine {

	machines := []*Machine{}
	for _, m := range self.ControlPlaneMachines {
		machines = append(machines, m.Machine)
	}
	for _, m := range self.WorkerNodeMachines {
		machines = append(machines, m.Machine)
	}
	return machines
}

/* drain a node + delete node + delete a VM */
func (self *Provisioner) DrainAndDeleteNode(nodeName string) error {

	if output, err := self.Kubectl("drain %s --ignore-daemonsets --force --delete-local-data", nodeName); err != nil {
		return errors.New(fmt.Sprintf("Failed to drain a node (node=%s, output='%s')", nodeName, output))
	}
	if output, err := self.Kubectl("delete node %s", nodeName); err != nil {
		return errors.New(fmt.Sprintf("Failed to delete a node (node=%s, output='%s')", nodeName, output))
	}
	vm := tumblebug.NewVM(self.Cluster.Namespace, nodeName, self.Cluster.MCIS)
	if exists, err := vm.DELETE(); err != nil {
		return errors.New(fmt.Sprintf("Failed to remove a VM (%s)", vm.Name))
	} else if !exists {
		return errors.New(fmt.Sprintf("Failed to remove a VM (vm=%s, cause='Colud not be found a VM')", vm.Name))
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
