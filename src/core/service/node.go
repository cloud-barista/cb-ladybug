package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cloud-barista/cb-mcks/src/core/model"
	"github.com/cloud-barista/cb-mcks/src/core/model/tumblebug"
	"github.com/cloud-barista/cb-mcks/src/utils/config"
	"github.com/cloud-barista/cb-mcks/src/utils/lang"
	"golang.org/x/sync/errgroup"

	ssh "github.com/cloud-barista/cb-spider/cloud-control-manager/vm-ssh"
	logger "github.com/sirupsen/logrus"
)

func ListNode(namespace string, clusterName string) (*model.NodeList, error) {
	err := CheckNamespace(namespace)
	if err != nil {
		return nil, err
	}

	err = CheckMcis(namespace, clusterName)
	if err != nil {
		return nil, err
	}

	nodes := model.NewNodeList(namespace, clusterName)
	err = nodes.SelectList()
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func GetNode(namespace string, clusterName string, nodeName string) (*model.Node, error) {
	err := CheckNamespace(namespace)
	if err != nil {
		return nil, err
	}

	err = CheckMcis(namespace, clusterName)
	if err != nil {
		return nil, err
	}

	node := model.NewNode(namespace, clusterName, nodeName)
	err = node.Select()
	if err != nil {
		return nil, err
	}

	return node, nil
}

func AddNode(namespace string, clusterName string, req *model.NodeReq) (*model.NodeList, error) {
	err := CheckNamespace(namespace)
	if err != nil {
		return nil, err
	}

	err = CheckMcis(namespace, clusterName)
	if err != nil {
		return nil, err
	}

	err = CheckClusterStatus(namespace, clusterName)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to get cluster status (cause=%v)", err))
	}

	mcisName := clusterName

	// get join command & network cni
	workerJoinCmd, err := getWorkerJoinCmdForAddNode(namespace, clusterName)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to get join command (cause=%v)", err))
	}
	networkCni, err := getClusterNetworkCNI(namespace, clusterName)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to get network cni (cause=%v)", err))
	}

	var nodeConfigInfos []NodeConfigInfo
	// worker
	wk, err := SetNodeConfigInfos(req.Worker, config.WORKER)
	if err != nil {
		return nil, err
	}
	nodeConfigInfos = append(nodeConfigInfos, wk...)

	cIdx := 0
	wIdx := 0
	maxCIdx, maxWIdx := getMaxIdx(namespace, clusterName)
	var TVMs []tumblebug.TVM
	var sTVMs []tumblebug.TVM

	for _, nodeConfigInfo := range nodeConfigInfos {
		// MCIR - 존재하면 재활용 없다면 생성 기준
		// 1. create vpc
		vpc, err := nodeConfigInfo.CreateVPC(namespace)
		if err != nil {
			return nil, err
		}

		// 2. create firewall
		fw, err := nodeConfigInfo.CreateFirewall(namespace)
		if err != nil {
			return nil, err
		}

		// 3. create sshKey
		sshKey, err := nodeConfigInfo.CreateSshKey(namespace)
		if err != nil {
			return nil, err
		}

		// 4. create image
		image, err := nodeConfigInfo.CreateImage(namespace)
		if err != nil {
			return nil, err
		}

		// 5. create spec
		spec, err := nodeConfigInfo.CreateSpec(namespace)
		if err != nil {
			return nil, err
		}

		// 6. vm
		for i := 0; i < nodeConfigInfo.Count; i++ {
			if nodeConfigInfo.Role == config.CONTROL_PLANE {
				cIdx++
			} else {
				wIdx++
			}
			tvm := tumblebug.NewTVm(namespace, mcisName)
			tvm.VM = model.VM{
				Config:       nodeConfigInfo.Connection,
				VPC:          vpc.Name,
				Subnet:       vpc.Subnets[0].Name,
				Firewall:     []string{fw.Name},
				SSHKey:       sshKey.Name,
				Image:        image.Name,
				Spec:         spec.Name,
				UserAccount:  nodeConfigInfo.Account,
				UserPassword: "",
				Description:  "",
				Credential:   sshKey.PrivateKey,
				Role:         nodeConfigInfo.Role,
				Csp:          nodeConfigInfo.Csp,
			}

			if nodeConfigInfo.Role == config.CONTROL_PLANE {
				tvm.VM.Name = lang.GetNodeName(config.CONTROL_PLANE, maxCIdx+cIdx)
			} else {
				tvm.VM.Name = lang.GetNodeName(config.WORKER, maxWIdx+wIdx)
			}

			// vm 생성
			logger.Infof("start create VM (mcisname=%s, nodename=%s)", mcisName, tvm.VM.Name)
			err := tvm.POST()
			if err != nil {
				logger.Warnf("create VM error (mcisname=%s, nodename=%s)", mcisName, tvm.VM.Name)
				deleteVMs(namespace, clusterName, sTVMs)
				return nil, err
			}
			logger.Infof("create VM OK.. (mcisname=%s, nodename=%s)", mcisName, tvm.VM.Name)

			TVMs = append(TVMs, *tvm)
			sTVMs = append(sTVMs, *tvm)
		}
	}

	logger.Infoln("start connect VMs")
	eg, _ := errgroup.WithContext(context.Background())

	for _, tvm := range TVMs {
		vm := tvm.VM
		eg.Go(func() error {
			sshInfo := ssh.SSHInfo{
				UserName:   GetUserAccount(vm.Csp),
				PrivateKey: []byte(vm.Credential),
				ServerPort: fmt.Sprintf("%s:22", vm.PublicIP),
			}

			if vm.Status != config.Running || vm.PublicIP == "" {
				return errors.New(fmt.Sprintf("Cannot do ssh, VM IP is not Running (name=%s, ip=%s, systemMessage=%s)", vm.Name, vm.PublicIP, vm.SystemMessage))
			}

			err := vm.ConnectionTest(&sshInfo)
			if err != nil {
				return err
			}

			err = vm.CopyScripts(&sshInfo, networkCni)
			if err != nil {
				return err
			}

			logger.Infof("set systemd service (vm=%s)", vm.Name)
			err = vm.SetSystemd(&sshInfo, networkCni)
			if err != nil {
				return err
			}

			logger.Infof("bootstrap (vm=%s)", vm.Name)
			err = vm.Bootstrap(&sshInfo)
			if err != nil {
				return err
			}

			logger.Infof("join (vm=%s)", vm.Name)
			err = vm.WorkerJoin(&sshInfo, &workerJoinCmd)
			if err != nil {
				return err
			}

			logger.Infof("add labels (vm=%s)", vm.Name)
			err = vm.AddNodeLabels(&sshInfo)
			if err != nil {
				logger.Warnf("failed to add node labels (vm=%s, cause= %s)", vm.Name, err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		logger.Warnf("worker join error (cause=%v)", err)
		deleteVMs(namespace, clusterName, TVMs)
		return nil, err
	}

	// insert store
	nodes := model.NewNodeList(namespace, clusterName)
	for _, vm := range TVMs {
		node := model.NewNodeVM(namespace, clusterName, vm.VM)
		err := node.Insert()
		if err != nil {
			return nil, err
		}
		nodes.Items = append(nodes.Items, *node)
	}

	return nodes, nil
}

func RemoveNode(namespace string, clusterName string, nodeName string) (*model.Status, error) {
	err := CheckNamespace(namespace)
	if err != nil {
		return nil, err
	}

	err = CheckMcis(namespace, clusterName)
	if err != nil {
		return nil, err
	}

	node := model.NewNode(namespace, clusterName, nodeName)
	err = node.Select()
	if err != nil {
		return nil, err
	}

	status := model.NewStatus()
	status.Code = model.STATUS_UNKNOWN

	cpNode, err := getCPLeaderNode(namespace, clusterName)
	if err != nil {
		status.Message = "failed to find control-plane node"
		return status, errors.New(fmt.Sprintf("%s (cause=%v)", status.Message, err))
	}

	hostName, err := getHostName(node)
	if err != nil {
		status.Message = "failed to get hostname"
		return status, errors.New(fmt.Sprintf("%s (cause=%v)", status.Message, err))
	}

	// drain node
	userAccount := GetUserAccount(cpNode.Csp)
	sshInfo := ssh.SSHInfo{
		UserName:   userAccount,
		PrivateKey: []byte(cpNode.Credential),
		ServerPort: fmt.Sprintf("%s:22", cpNode.PublicIP),
	}
	cmd := fmt.Sprintf("sudo kubectl drain %s --kubeconfig=/etc/kubernetes/admin.conf --ignore-daemonsets --force --delete-local-data", hostName)
	logger.Infof("[RemoveNode] %s $ %s", cpNode.Name, cmd)
	result, err := ssh.SSHRun(sshInfo, cmd)
	if err != nil {
		status.Message = "kubectl drain failed"
		return status, errors.New(fmt.Sprintf("%s (cause=%v)", status.Message, err))
	}
	if strings.Contains(result, fmt.Sprintf("node/%s drained", hostName)) || strings.Contains(result, fmt.Sprintf("node/%s evicted", hostName)) {
		logger.Infoln("drain node success")
	} else {
		status.Message = "kubectl drain failed"
		return status, errors.New(fmt.Sprintf("%s (cause=%v)", status.Message, err))
	}

	// delete node
	cmd = fmt.Sprintf("sudo kubectl delete node %s --kubeconfig=/etc/kubernetes/admin.conf", hostName)
	logger.Infof("[RemoveNode] %s $ %s", cpNode.Name, cmd)
	result, err = ssh.SSHRun(sshInfo, cmd)
	if err != nil {
		status.Message = "kubectl delete node failed"
		return status, errors.New(fmt.Sprintf("%s (cause=%v)", status.Message, err))
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
		return status, errors.New(fmt.Sprintf("%s (cause=%v)", status.Message, err))
	}

	// delete node in store
	if err := node.Delete(); err != nil {
		status.Message = err.Error()
		return status, nil
	}

	status.Code = model.STATUS_SUCCESS
	status.Message = fmt.Sprintf("node '%s' has been deleted", nodeName)

	return status, nil
}

func getCPLeaderNode(namespace string, clusterName string) (*model.Node, error) {
	cluster := model.NewCluster(namespace, clusterName)
	err := cluster.Select()
	if err != nil {
		return nil, err
	}
	cpLeaderName := cluster.CpLeader
	if cpLeaderName == "" {
		return nil, errors.New("control-place node is empty")
	}

	cpNode := model.NewNode(namespace, clusterName, cpLeaderName)
	err = cpNode.Select()
	if err != nil {
		return nil, err
	}

	return cpNode, nil
}

func getClusterNetworkCNI(namespace string, clusterName string) (string, error) {
	cluster := model.NewCluster(namespace, clusterName)
	err := cluster.Select()
	if err != nil {
		return "", err
	}

	networkCni := cluster.NetworkCni
	if networkCni == "" {
		return "", errors.New("network cni is empty")
	}

	return networkCni, nil
}

func getHostName(node *model.Node) (string, error) {
	userAccount := GetUserAccount(node.Csp)
	sshInfo := ssh.SSHInfo{
		UserName:   userAccount,
		PrivateKey: []byte(node.Credential),
		ServerPort: fmt.Sprintf("%s:22", node.PublicIP),
	}
	cmd := "/bin/hostname"
	hostName, err := ssh.SSHRun(sshInfo, cmd)
	if err != nil {
		return "", err
	}
	hostName = strings.ToLower(hostName)
	return hostName, nil
}

func getWorkerJoinCmdForAddNode(namespace string, clusterName string) (string, error) {
	cpNode, err := getCPLeaderNode(namespace, clusterName)
	if err != nil {
		return "", err
	}
	userAccount := GetUserAccount(cpNode.Csp)
	sshInfo := ssh.SSHInfo{
		UserName:   userAccount,
		PrivateKey: []byte(cpNode.Credential),
		ServerPort: fmt.Sprintf("%s:22", cpNode.PublicIP),
	}
	cmd := "sudo kubeadm token create --print-join-command"
	logger.Infof("[getWorkerJoinCmdForAddNode] %s $ %s", cpNode.Name, cmd)
	joinCommand, err := ssh.SSHRun(sshInfo, cmd)
	if err != nil {
		return "", err
	}
	if joinCommand == "" {
		return "", errors.New("join command is empty")
	}

	return joinCommand, nil
}

func getMaxIdx(namespace string, clusterName string) (maxCpIdx int, maxWkIdx int) {
	maxCpIdx = 0
	maxWkIdx = 0

	nodes := model.NewNodeList(namespace, clusterName)
	err := nodes.SelectList()
	if err != nil {
		return
	}

	var arrCp, arrWk []int
	for _, node := range nodes.Items {
		slice := strings.Split(node.Name, "-")
		role := len(slice) - 3
		idx := len(slice) - 2

		if slice[role] == "c" {
			arrCp = append(arrCp, lang.GetIdxToInt(slice[idx]))
		} else if slice[role] == "w" {
			arrWk = append(arrWk, lang.GetIdxToInt(slice[idx]))
		}
	}
	fmt.Println(maxCpIdx, maxWkIdx)
	maxCpIdx = lang.GetMaxNumber(arrCp)
	maxWkIdx = lang.GetMaxNumber(arrWk)
	return
}

func deleteVMs(namespace string, clusterName string, TVMs []tumblebug.TVM) error {
	logger.Infof("delete VMs")
	for _, tvm := range TVMs {
		vm := tumblebug.NewTVm(namespace, clusterName)
		vm.VM.Name = tvm.VM.Name
		if err := vm.DELETE(); err != nil {
			logger.Errorf("failed to delete vm (nodeName=%s, cause=%v)", tvm.VM.Name, err)
			continue
		}
	}
	return nil
}
