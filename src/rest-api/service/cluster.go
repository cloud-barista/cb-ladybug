package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/cloud-barista/cb-ladybug/src/core/model"
	"github.com/cloud-barista/cb-ladybug/src/core/model/tumblebug"
	"github.com/cloud-barista/cb-ladybug/src/utils/config"
	"github.com/cloud-barista/cb-ladybug/src/utils/lang"

	ssh "github.com/cloud-barista/cb-spider/cloud-control-manager/vm-ssh"
	logger "github.com/sirupsen/logrus"
)

func ListCluster(namespace string) (*model.ClusterList, error) {
	clusters := model.NewClusterList(namespace)

	err := clusters.SelectList()
	if err != nil {
		return nil, err
	}

	return clusters, nil
}

func GetCluster(namespace string, clusterName string) (*model.Cluster, error) {
	cluster := model.NewCluster(namespace, clusterName)
	err := cluster.Select()
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

func CreateCluster(namespace string, req *model.ClusterReq) (*model.Cluster, error) {

	clusterName := req.Name
	cluster := model.NewCluster(namespace, clusterName)
	cluster.UId = lang.GetUid()
	mcisName := clusterName

	// Namespace 존재여부 확인
	ns := tumblebug.NewNS(namespace)
	exists, err := ns.GET()
	if err != nil {
		return cluster, err
	}
	if !exists {
		return cluster, errors.New(fmt.Sprintf("namespace does not exist (name=%s)", namespace))
	}

	// MCIR 존재여부 확인
	mcis := tumblebug.NewMCIS(namespace, mcisName)
	exists, err = mcis.GET()
	if err != nil {
		return cluster, err
	}
	if exists {
		return cluster, errors.New("MCIS already exists")
	}

	// MCIR 명칭 정의 (존재하면 재활용 없다면 생성 기준)
	vpcName := fmt.Sprintf("%s-vpc", clusterName)
	firewallName := fmt.Sprintf("%s-allow-external", clusterName)
	sshkeyName := fmt.Sprintf("%s-sshkey", clusterName)
	specName := fmt.Sprintf("%s-spec", clusterName)

	//TODO [update/hard-coding] connection config
	csp := config.CSP_GCP
	if strings.Contains(namespace, "aws") {
		csp = config.CSP_AWS
	}
	connConfig := fmt.Sprintf("cb-%s-config", csp)

	//host user account
	account := GetUserAccount(csp)

	// get image id
	imageId, e := GetVmImageId(csp, connConfig)
	if e != nil {
		return cluster, e
	}

	// 1. create vpc
	logger.Infof("start create vpc (name=%s)", vpcName)
	vpc := tumblebug.NewVPC(namespace, vpcName, connConfig)
	exists, e = vpc.GET()
	if e != nil {
		return cluster, e
	}
	if exists {
		logger.Infof("reuse vpc (name=%s, cause='already exists')", vpcName)
	} else {
		if e = vpc.POST(); e != nil {
			return cluster, e
		}
		logger.Infof("create vpc OK.. (name=%s)", vpcName)
	}

	// 2. create firewall
	logger.Infof("start create firewall (name=%s)", firewallName)
	fw := tumblebug.NewFirewall(namespace, firewallName, connConfig)
	fw.VPCId = vpcName
	exists, e = fw.GET()
	if e != nil {
		return cluster, e
	}
	if exists {
		logger.Infof("reuse firewall (name=%s, cause='already exists')", firewallName)
	} else {
		if e = fw.POST(); e != nil {
			return cluster, e
		}
		logger.Infof("create firewall OK.. (name=%s)", firewallName)
	}

	// 3. create sshKey
	logger.Infof("start create ssh key (name=%s)", sshkeyName)
	sshKey := tumblebug.NewSSHKey(namespace, sshkeyName, connConfig)
	sshKey.Username = account
	exists, e = sshKey.GET()
	if e != nil {
		return cluster, e
	}
	if exists {
		logger.Infof("reuse ssh key (name=%s, cause='already exists')", sshkeyName)
	} else {
		if e = sshKey.POST(); e != nil {
			return cluster, e
		}
		logger.Infof("create ssh key OK.. (name=%s)", sshkeyName)
	}

	// 4. create image
	imageName := fmt.Sprintf("%s-Ubuntu1804", connConfig)
	logger.Infof("start create image (name=%s)", imageName)
	image := tumblebug.NewImage(namespace, imageName, connConfig)
	image.CspImageId = imageId
	exists, e = image.GET()
	if e != nil {
		return cluster, e
	}
	if exists {
		logger.Infof("reuse image (name=%s, cause='already exists')", imageName)
	} else {
		if e = image.POST(); e != nil {
			return cluster, e
		}
		logger.Infof("create image OK.. (name=%s)", imageName)
	}

	// control-plane
	// 5. create spec
	logger.Infof("start create control plane spec (name=%s)", specName)
	spec := tumblebug.NewSpec(namespace, specName, connConfig)
	spec.CspSpecName = req.ControlPlaneNodeSpec
	spec.Role = config.CONTROL_PLANE
	exists, e = spec.GET()
	if e != nil {
		return cluster, e
	}
	if exists {
		logger.Infof("reuse control plane spec (name=%s, cause='already exists')", specName)
	} else {
		if e = spec.POST(); e != nil {
			return cluster, e
		}
		logger.Infof("create control plane spec OK.. (name=%s)", specName)
	}

	// 6. vm
	for i := 0; i < req.ControlPlaneNodeCount; i++ {
		vm := model.VM{
			Name:         lang.GetNodeName(clusterName, spec.Role),
			Config:       connConfig,
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
		mcis.VMs = append(mcis.VMs, vm)
	}

	// worker node
	// 5. create spec
	logger.Infof("start create worker node spec (name=%s)", specName)
	spec = tumblebug.NewSpec(namespace, specName, connConfig)
	spec.CspSpecName = req.WorkerNodeSpec
	spec.Role = config.WORKER
	exists, e = spec.GET()
	if e != nil {
		return cluster, e
	}
	if exists {
		logger.Infof("reuse worker node spec (name=%s, cause='already exists')", specName)
	} else {
		if e = spec.POST(); e != nil {
			return cluster, e
		}
		logger.Infof("create worker node spec OK.. (name=%s)", specName)
	}

	// 6. vm
	for i := 0; i < req.WorkerNodeCount; i++ {
		vm := model.VM{
			Name:         lang.GetNodeName(clusterName, spec.Role),
			Config:       connConfig,
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
		mcis.VMs = append(mcis.VMs, vm)
	}

	// MCIS 생성
	logger.Infof("start create MCIS (name=%s)", mcisName)
	if err = mcis.POST(); err != nil {
		return cluster, err
	}
	logger.Infof("create MCIS OK.. (name=%s)", mcisName)

	// 결과값 저장
	var nodes []model.Node
	cluster.MCIS = mcisName
	for _, vm := range mcis.VMs {
		node := model.NewNodeVM(namespace, cluster.Name, vm)
		node.UId = lang.GetUid()

		// insert node in store
		nodes = append(nodes, *node)
		err := node.Insert()
		if err != nil {
			return nil, err
		}
	}
	err = cluster.Insert()
	if err != nil {
		return nil, err
	}

	var workerJoinCmd string
	var wg sync.WaitGroup
	c := make(chan error)
	wg.Add(len(mcis.VMs))

	// bootstrap
	logger.Infoln("start k8s bootstrap")
	for _, vm := range mcis.VMs {
		err = cluster.Update()
		if err != nil {
			return nil, err
		}

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
				cluster.Fail()
				c <- err
			}

			bootstrapResult, err := vm.Bootstrap(&sshInfo)
			if err != nil {
				cluster.Fail()
				c <- err
			}
			if !bootstrapResult {
				cluster.Fail()
				c <- errors.New(vm.Name + " bootstrap failed")
			}
		}(vm)
	}

	go func() {
		wg.Wait()
		close(c)
		logger.Infoln("end k8s bootstrap")
	}()

	for err := range c {
		if err != nil {
			return nil, err
		}
	}

	// init & join
	logger.Infoln("start k8s init & join")
	for _, vm := range mcis.VMs {
		sshInfo := ssh.SSHInfo{
			UserName:   account,
			PrivateKey: []byte(vm.Credential),
			ServerPort: fmt.Sprintf("%s:22", vm.PublicIP),
		}

		if vm.Role == config.CONTROL_PLANE {
			logger.Infoln("init")
			var clusterConfig string
			workerJoinCmd, clusterConfig, err = vm.ControlPlaneInit(&sshInfo, vm.PublicIP)
			if err != nil {
				logger.Warnf("%s init failed (cause=%v)", vm.Name, err)
				cluster.Fail()
				return nil, err
			}
			cluster.ClusterConfig = clusterConfig
		} else {
			logger.Infoln("join")
			joinResult, err := vm.WorkerJoin(&sshInfo, &workerJoinCmd)
			if err != nil {
				logger.Warnf("%s join error (cause=%v)", vm.Name, err)
				cluster.Fail()
				return nil, err
			}
			if !joinResult {
				logger.Warnf("%s join failed", vm.Name)
				cluster.Fail()
				return nil, errors.New(vm.Name + " join failed")
			}
		}

		logger.Infoln("kilo annotation")
		err := vm.KiloAnnotation(&sshInfo, vm.Role)
		if err != nil {
			logger.Warnf("%s kilo annotation failed", vm.Name)
		}
	}
	logger.Infoln("end k8s init & join")

	cluster.Complete()
	cluster.Nodes = nodes

	return cluster, nil
}

func DeleteCluster(namespace string, clusterName string) (*model.Status, error) {
	mcisName := clusterName //cluster 이름과 동일하게 (임시)

	status := model.NewStatus()
	status.Code = model.STATUS_UNKNOWN

	// 1. delete mcis
	logger.Infof("start delete MCIS (name=%s)", mcisName)
	mcis := tumblebug.NewMCIS(namespace, mcisName)
	exist, err := mcis.GET()
	if err != nil {
		return status, err
	}
	if exist {
		if err = mcis.DELETE(); err != nil {
			return status, err
		}
		// sleep 이후 확인하는 로직 추가 필요
		logger.Infof("delete MCIS OK.. (name=%s)", mcisName)
		status.Code = model.STATUS_SUCCESS
		status.Message = "success"

		cluster := model.NewCluster(namespace, clusterName)
		if err := cluster.Delete(); err != nil {
			status.Message = "delete success but cannot delete from the store"
			return status, nil
		}
	} else {
		status.Code = model.STATUS_NOT_EXIST
		logger.Infof("delete MCIS skip (cannot find).. (name=%s)", mcisName)
	}

	return status, nil
}
