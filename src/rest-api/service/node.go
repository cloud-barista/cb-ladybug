package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/cloud-barista/cb-ladybug/src/core/common"
	"github.com/cloud-barista/cb-ladybug/src/core/model"
	"github.com/cloud-barista/cb-ladybug/src/core/model/tumblebug"
	"github.com/cloud-barista/cb-ladybug/src/utils/config"
	"github.com/cloud-barista/cb-ladybug/src/utils/lang"

	ssh "github.com/cloud-barista/cb-spider/cloud-control-manager/vm-ssh"
	logger "github.com/sirupsen/logrus"
)

func ListNode(namespace string, clusterName string) (*model.NodeList, error) {
	nodes := model.NewNodeList(namespace, clusterName)
	err := nodes.SelectList()
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func GetNode(namespace string, clusterName string, nodeName string) (*model.Node, error) {
	node := model.NewNode(namespace, clusterName, nodeName)
	err := node.Select()
	if err != nil {
		return nil, err
	}

	return node, nil
}

func AddNode(namespace string, clusterName string, req *model.NodeReq) (*model.NodeList, error) {

	//TODO [update/hard-coding] connection config
	csp := config.CSP_GCP
	if strings.Contains(namespace, "aws") {
		csp = config.CSP_AWS
	}
	//host user account
	account := GetUserAccount(csp)

	// get join command
	cpNode, err := getCPNode(namespace, clusterName)
	if err != nil {
		return nil, errors.New("control-plane node not found")
	}
	workerJoinCmd, err := getWorkerJoinCmdForAddNode(account, cpNode)
	if err != nil {
		return nil, errors.New("get join command error")
	}

	mcisName := clusterName
	mcis := tumblebug.NewMCIS(namespace, mcisName)

	exists, err := mcis.GET()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("MCIS not found")
	}

	vpcName := fmt.Sprintf("%s-vpc", clusterName)
	firewallName := fmt.Sprintf("%s-allow-external", clusterName)
	sshkeyName := fmt.Sprintf("%s-sshkey", clusterName)
	imageName := fmt.Sprintf("%s-Ubuntu1804", req.Config)
	specName := fmt.Sprintf("%s-spec", clusterName)

	// vpc
	logger.Infof("start create vpc (name=%s)", vpcName)
	vpc := tumblebug.NewVPC(namespace, vpcName, req.Config)
	exists, e := vpc.GET()
	if e != nil {
		return nil, e
	}
	if exists {
		logger.Infof("reuse vpc (name=%s, cause='already exists')", vpcName)
	} else {
		if e = vpc.POST(); e != nil {
			return nil, e
		}
		logger.Infof("create vpc OK.. (name=%s)", vpcName)
	}

	// firewall
	logger.Infof("start create firewall (name=%s)", firewallName)
	fw := tumblebug.NewFirewall(namespace, firewallName, req.Config)
	fw.VPCId = vpcName
	exists, e = fw.GET()
	if e != nil {
		return nil, e
	}
	if exists {
		logger.Infof("reuse firewall (name=%s, cause='already exists')", firewallName)
	} else {
		if e = fw.POST(); e != nil {
			return nil, e
		}
		logger.Infof("create firewall OK.. (name=%s)", firewallName)
	}

	// sshKey
	logger.Infof("start create ssh key (name=%s)", sshkeyName)
	sshKey := tumblebug.NewSSHKey(namespace, sshkeyName, req.Config)
	sshKey.Username = "cb-cluster"
	exists, e = sshKey.GET()
	if e != nil {
		return nil, e
	}
	if exists {
		logger.Infof("reuse ssh key (name=%s, cause='already exists')", sshkeyName)
	} else {
		if e = sshKey.POST(); e != nil {
			return nil, e
		}
		logger.Infof("create ssh key OK.. (name=%s)", sshkeyName)
	}

	// image
	logger.Infof("start create image (name=%s)", imageName)
	// get image id
	imageId, e := GetVmImageId(csp, req.Config)
	if e != nil {
		return nil, e
	}

	image := tumblebug.NewImage(namespace, imageName, req.Config)
	image.CspImageId = imageId
	exists, e = image.GET()
	if e != nil {
		return nil, e
	}
	if exists {
		logger.Infof("reuse image (name=%s, cause='already exists')", imageName)
	} else {
		if e = image.POST(); e != nil {
			return nil, e
		}
		logger.Infof("create image OK.. (name=%s)", imageName)
	}

	// spec
	logger.Infof("start create worker node spec (name=%s)", specName)
	spec := tumblebug.NewSpec(namespace, specName, req.Config)
	spec.CspSpecName = req.WorkerNodeSpec
	spec.Role = config.WORKER
	exists, e = spec.GET()
	if e != nil {
		return nil, e
	}
	if exists {
		logger.Infof("reuse worker node spec (name=%s, cause='already exists')", specName)
	} else {
		if e = spec.POST(); e != nil {
			return nil, e
		}
		logger.Infof("create worker node spec OK.. (name=%s)", specName)
	}

	// vm
	var VMs []model.VM
	for i := 0; i < req.WorkerNodeCount; i++ {
		vm := tumblebug.NewTVm(namespace, mcisName)
		vm.VM = model.VM{
			Name:         lang.GetNodeName(clusterName, spec.Role),
			Config:       req.Config,
			VPC:          vpc.Name,
			Subnet:       vpc.Subnets[0].Name,
			Firewall:     []string{fw.Name},
			SSHKey:       sshKey.Name,
			Image:        image.Name,
			Spec:         spec.Name,
			UserAccount:  account,
			UserPassword: "",
			Description:  "",
			Credential:   sshKey.PrivateKey,
			Role:         spec.Role,
		}

		// vm 생성
		logger.Infof("start create VM (mcisname=%s, nodename=%s)", mcisName, vm.VM.Name)
		err := vm.POST()
		if err != nil {
			logger.Warnf("create VM error (mcisname=%s, nodename=%s)", mcisName, vm.VM.Name)
			return nil, err
		}
		VMs = append(VMs, vm.VM)
		logger.Infof("create VM OK.. (mcisname=%s, nodename=%s)", mcisName, vm.VM.Name)
	}

	var wg sync.WaitGroup
	c := make(chan error)
	wg.Add(len(VMs))

	logger.Infoln("start connect VMs")
	for _, vm := range VMs {
		go func(vm model.VM) {
			defer wg.Done()
			sshInfo := ssh.SSHInfo{
				UserName:   account,
				PrivateKey: []byte(vm.Credential),
				ServerPort: fmt.Sprintf("%s:22", vm.PublicIP),
			}

			_ = vm.ConnectionTest(&sshInfo)
			err := vm.CopyScripts(&sshInfo)
			if err != nil {
				c <- err
			}
			bootstrapResult, err := vm.Bootstrap(&sshInfo)
			if err != nil {
				c <- err
			}
			if !bootstrapResult {
				c <- errors.New(vm.Name + " bootstrap failed")
			}
			result, err := vm.WorkerJoinForAddNode(&sshInfo, &workerJoinCmd)
			if err != nil {
				c <- err
			}
			if !result {
				c <- errors.New(vm.Name + " join failed")
			}
		}(vm)
	}

	go func() {
		wg.Wait()
		close(c)
		logger.Infoln("end connect VMs")
	}()

	for err := range c {
		if err != nil {
			logger.Warnf("worker join error (cause=%v)", err)
			return nil, err
		}
	}

	// insert store
	nodes := model.NewNodeList(namespace, clusterName)
	for _, vm := range VMs {
		node := model.NewNodeVM(namespace, clusterName, vm)
		err := node.Insert()
		if err != nil {
			return nil, err
		}
		nodes.Items = append(nodes.Items, *node)
	}

	return nodes, nil
}

func RemoveNode(namespace string, clusterName string, nodeName string) (*model.Status, error) {
	status := model.NewStatus()
	status.Code = model.STATUS_UNKNOWN

	cpNode, err := getCPNode(namespace, clusterName)
	if err != nil {
		status.Message = "control-plane node not found"
		return status, err
	}

	var userAccount string
	var hostName string
	if strings.Contains(namespace, "gcp") {
		userAccount = "cb-user"
		hostName = nodeName
	} else {
		userAccount = "ubuntu"

		hostName, err = getAWSHostName(namespace, clusterName, nodeName, userAccount)
		if err != nil {
			status.Message = "get aws node name error"
			return status, err
		}
	}

	// drain node
	sshInfo := ssh.SSHInfo{
		UserName:   userAccount,
		PrivateKey: []byte(cpNode.Credential),
		ServerPort: fmt.Sprintf("%s:22", cpNode.PublicIP),
	}
	cmd := fmt.Sprintf("sudo kubectl drain %s --kubeconfig=/etc/kubernetes/admin.conf --ignore-daemonsets", hostName)
	result, err := ssh.SSHRun(sshInfo, cmd)
	if err != nil {
		status.Message = "kubectl drain failed"
		return status, err
	}
	if strings.Contains(result, fmt.Sprintf("node/%s drained", hostName)) || strings.Contains(result, fmt.Sprintf("node/%s evicted", hostName)) {
		logger.Infoln("drain node success")
	} else {
		status.Message = "kubectl drain failed"
		return status, err
	}

	// delete node
	cmd = fmt.Sprintf("sudo kubectl delete node %s --kubeconfig=/etc/kubernetes/admin.conf", hostName)
	result, err = ssh.SSHRun(sshInfo, cmd)
	if err != nil {
		status.Message = "kubectl delete node failed"
		return status, err
	}
	if strings.Contains(result, "deleted") {
		logger.Infoln("delete node success")
	} else {
		status.Message = "kubectl delete node failed"
		return status, errors.New("kubectl delete node failed")
	}

	// delete vm
	vm := tumblebug.NewTVm(namespace, clusterName)
	vm.VM.Name = nodeName
	err = vm.DELETE()
	if err != nil {
		status.Message = "delete vm failed"
		return status, err
	}

	// delete node in store
	node := model.NewNode(namespace, clusterName, nodeName)
	if err := node.Delete(); err != nil {
		status.Message = err.Error()
		return status, nil
	}

	status.Code = model.STATUS_SUCCESS
	status.Message = "success"

	return status, nil
}

func getCPNode(namespace string, clusterName string) (*model.Node, error) {
	key := lang.GetStoreNodeKey(namespace, clusterName, "")
	keyValues, err := common.CBStore.GetList(key, true)
	if err != nil {
		return nil, err
	}
	if keyValues == nil {
		return nil, errors.New(fmt.Sprintf("%s not found", key))
	}
	cpNode := &model.Node{}
	for _, keyValue := range keyValues {
		node := &model.Node{}
		json.Unmarshal([]byte(keyValue.Value), &node)
		if node.Role == config.CONTROL_PLANE {
			cpNode = node
			break
		}
	}

	return cpNode, nil
}

func getAWSHostName(namespace string, clusterName string, nodeName string, userAccount string) (string, error) {
	key := lang.GetStoreNodeKey(namespace, clusterName, "")
	keyValues, err := common.CBStore.GetList(key, true)
	if err != nil {
		return "", err
	}
	if keyValues == nil {
		return "", errors.New(fmt.Sprintf("%s not found", key))
	}
	wNode := &model.Node{}
	for _, keyValue := range keyValues {
		node := &model.Node{}
		json.Unmarshal([]byte(keyValue.Value), &node)
		if node.Name == nodeName {
			wNode = node
			break
		}
	}

	sshInfo := ssh.SSHInfo{
		UserName:   userAccount,
		PrivateKey: []byte(wNode.Credential),
		ServerPort: fmt.Sprintf("%s:22", wNode.PublicIP),
	}
	cmd := "/bin/hostname"
	result, err := ssh.SSHRun(sshInfo, cmd)
	if err != nil {
		return "", err
	}

	return result, nil
}

func getWorkerJoinCmdForAddNode(userAccount string, cpNode *model.Node) (string, error) {
	sshInfo := ssh.SSHInfo{
		UserName:   userAccount,
		PrivateKey: []byte(cpNode.Credential),
		ServerPort: fmt.Sprintf("%s:22", cpNode.PublicIP),
	}
	cmd := "sudo kubeadm token create --print-join-command"
	result, err := ssh.SSHRun(sshInfo, cmd)
	if err != nil {
		return "", err
	}
	return result, nil
}
