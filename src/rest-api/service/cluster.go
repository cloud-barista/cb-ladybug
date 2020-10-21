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
)

type CSP struct {
	Name                  string
	Config                string
	ControlPlaneNodeSpec  string
	ControlPlaneNodeCount int
	WorkerNodeSpec        string
	WorkerNodeCount       int
}

func ListCluster(namespace string) (*model.ClusterList, error) {
	clusters := model.NewClusterList()

	result, err := clusters.SelectList(namespace, clusters)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetCluster(namespace string, clusterName string) (*model.Cluster, error) {
	cluster := model.NewCluster(namespace, clusterName)

	result, err := cluster.Select(cluster)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func CreateCluster(namespace string, req *model.ClusterReq) (*model.Cluster, error) {
	var csps []CSP

	var cspPrefix, imageId, userAccount string
	if strings.Contains(namespace, "gcp") {
		userAccount = "cb-user"
		cspPrefix = "cb-gcp"
		imageId = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-1804-bionic-v20201014"
	} else {
		userAccount = "ubuntu"
		cspPrefix = "cb-aws"
		imageId = "ami-02b658ac34935766f"
	}

	csps = append(csps, CSP{Name: cspPrefix, Config: cspPrefix + "-config", ControlPlaneNodeSpec: req.ControlPlaneNodeSpec, ControlPlaneNodeCount: req.ControlPlaneNodeCount, WorkerNodeSpec: req.WorkerNodeSpec, WorkerNodeCount: req.WorkerNodeCount})

	clusterName := req.Name
	cluster := model.NewCluster(namespace, clusterName)
	cluster.UId = lang.GetUid()
	mcisName := clusterName

	mcis := tumblebug.NewMCIS(namespace, mcisName)
	exists, err := mcis.GET()
	if err != nil {
		return cluster, err
	}
	if exists {
		return cluster, errors.New("MCIS already exists")
	}

	// MCIR 생성 (존재하면 재활용 없다면 생성 기준)
	vpcName := fmt.Sprintf("%s-vpc", clusterName)
	firewallName := fmt.Sprintf("%s-allow-external", clusterName)
	sshkeyName := fmt.Sprintf("%s-sshkey", clusterName)
	specName := fmt.Sprintf("%s-spec", clusterName)

	for _, csp := range csps {
		// 1. create vpc
		fmt.Println(fmt.Sprintf("start create vpc (name=%s)", vpcName))
		vpc := tumblebug.NewVPC(namespace, vpcName, csp.Config)
		exists, e := vpc.GET()
		if e != nil {
			return cluster, e
		}
		if exists {
			fmt.Println(fmt.Sprintf("reuse vpc (name=%s, cause='already exists')", vpcName))
		} else {
			if e = vpc.POST(); e != nil {
				return cluster, e
			}
			fmt.Println(fmt.Sprintf("create vpc OK.. (name=%s)", vpcName))
		}

		// 2. create firewall
		fmt.Println(fmt.Sprintf("start create firewall (name=%s)", firewallName))
		fw := tumblebug.NewFirewall(namespace, firewallName, csp.Config)
		fw.VPCId = vpcName
		exists, e = fw.GET()
		if e != nil {
			return cluster, e
		}
		if exists {
			fmt.Println(fmt.Sprintf("reuse firewall (name=%s, cause='already exists')", firewallName))
		} else {
			if e = fw.POST(); e != nil {
				return cluster, e
			}
			fmt.Println(fmt.Sprintf("create firewall OK.. (name=%s)", firewallName))
		}

		// 3. create sshKey
		fmt.Println(fmt.Sprintf("start create ssh key (name=%s)", sshkeyName))
		sshKey := tumblebug.NewSSHKey(namespace, sshkeyName, csp.Config)
		sshKey.Username = "cb-cluster"
		exists, e = sshKey.GET()
		if e != nil {
			return cluster, e
		}
		if exists {
			fmt.Println(fmt.Sprintf("reuse ssh key (name=%s, cause='already exists')", sshkeyName))
		} else {
			if e = sshKey.POST(); e != nil {
				return cluster, e
			}
			fmt.Println(fmt.Sprintf("create ssh key OK.. (name=%s)", sshkeyName))
		}

		// 4. create image
		imageName := fmt.Sprintf("%s-Ubuntu1804", csp.Config)
		fmt.Println(fmt.Sprintf("start create image (name=%s)", imageName))
		image := tumblebug.NewImage(namespace, imageName, csp.Config)
		image.CspImageId = imageId
		exists, e = image.GET()
		if e != nil {
			return cluster, e
		}
		if exists {
			fmt.Println(fmt.Sprintf("reuse image (name=%s, cause='already exists')", imageName))
		} else {
			if e = image.POST(); e != nil {
				return cluster, e
			}
			fmt.Println(fmt.Sprintf("create image OK.. (name=%s)", imageName))
		}

		// control-plane
		// 5. create spec
		fmt.Println(fmt.Sprintf("start create control plane spec (name=%s)", specName))
		spec := tumblebug.NewSpec(namespace, specName, csp.Config)
		spec.CspSpecName = csp.ControlPlaneNodeSpec
		spec.Role = config.CONTROL_PLANE
		exists, e = spec.GET()
		if e != nil {
			return cluster, e
		}
		if exists {
			fmt.Println(fmt.Sprintf("reuse control plane spec (name=%s, cause='already exists')", specName))
		} else {
			if e = spec.POST(); e != nil {
				return cluster, e
			}
			fmt.Println(fmt.Sprintf("create control plane spec OK.. (name=%s)", specName))
		}

		nodeCnt := 1
		// 6. vm
		for i := 0; i < csp.ControlPlaneNodeCount; i++ {
			vm := model.VM{
				Name:         fmt.Sprintf("%s-%s-%d", csp.Name, clusterName, nodeCnt),
				Config:       csp.Config,
				VPC:          vpc.Name,
				Subnet:       vpc.Subnets[0].Name,
				Firewall:     []string{fw.Name},
				SSHKey:       sshKey.Name,
				Image:        image.Name,
				Spec:         spec.Name,
				UserAccount:  userAccount,
				UserPassword: "",
				Description:  "",
				Credential:   sshKey.PrivateKey,
				Role:         spec.Role,
			}
			mcis.VMs = append(mcis.VMs, vm)
			nodeCnt++
		}

		// worker node
		// 5. create spec
		fmt.Println(fmt.Sprintf("start create worker node spec (name=%s)", specName))
		spec = tumblebug.NewSpec(namespace, specName, csp.Config)
		spec.CspSpecName = csp.WorkerNodeSpec
		spec.Role = config.WORKER
		exists, e = spec.GET()
		if e != nil {
			return cluster, e
		}
		if exists {
			fmt.Println(fmt.Sprintf("reuse worker node spec (name=%s, cause='already exists')", specName))
		} else {
			if e = spec.POST(); e != nil {
				return cluster, e
			}
			fmt.Println(fmt.Sprintf("create worker node spec OK.. (name=%s)", specName))
		}

		// 6. vm
		for i := 0; i < csp.WorkerNodeCount; i++ {
			vm := model.VM{
				Name:         fmt.Sprintf("%s-%s-%d", csp.Name, clusterName, nodeCnt),
				Config:       csp.Config,
				VPC:          vpc.Name,
				Subnet:       vpc.Subnets[0].Name,
				Firewall:     []string{fw.Name},
				SSHKey:       sshKey.Name,
				Image:        image.Name,
				Spec:         spec.Name,
				UserAccount:  userAccount,
				UserPassword: "",
				Description:  "",
				Credential:   sshKey.PrivateKey,
				Role:         spec.Role,
			}
			mcis.VMs = append(mcis.VMs, vm)
			nodeCnt++
		}
	}

	// MCIS 생성
	fmt.Println(fmt.Sprintf("start create MCIS (name=%s)", mcisName))
	if err = mcis.POST(); err != nil {
		return cluster, err
	}
	fmt.Println(fmt.Sprintf("create MCIS OK.. (name=%s)", mcisName))

	// 결과값 저장
	cluster.MCIS = mcisName
	for _, vm := range mcis.VMs {
		node := model.NewNode(vm)
		node.UId = lang.GetUid()
		cluster.Nodes = append(cluster.Nodes, *node)
	}
	cluster.Insert(cluster)

	var workerJoinCmd string
	var wg sync.WaitGroup
	c := make(chan error)
	wg.Add(len(mcis.VMs))

	// bootstrap
	fmt.Println("start k8s bootstrap")
	for _, vm := range mcis.VMs {
		cluster.Update(cluster)

		go func(vm model.VM) {
			defer wg.Done()
			sshInfo := ssh.SSHInfo{
				UserName:   userAccount,
				PrivateKey: []byte(vm.Credential),
				ServerPort: fmt.Sprintf("%s:22", vm.PublicIP),
			}

			_ = vm.ConnectionTest(&sshInfo, &vm)
			err := vm.CopyScripts(&sshInfo, &vm)
			if err != nil {
				cluster.Fail(cluster)
				c <- err
			}

			bootstrapResult, err := vm.Bootstrap(&sshInfo)
			if err != nil {
				cluster.Fail(cluster)
				c <- err
			}
			if !bootstrapResult {
				cluster.Fail(cluster)
				c <- errors.New(vm.Name + " bootstrap failed")
			}
		}(vm)
	}

	go func() {
		wg.Wait()
		close(c)
		fmt.Println("end k8s bootstrap")
	}()

	for err := range c {
		if err != nil {
			return nil, err
		}
	}

	// init & join
	fmt.Println("start k8s init & join")
	for _, vm := range mcis.VMs {
		sshInfo := ssh.SSHInfo{
			UserName:   userAccount,
			PrivateKey: []byte(vm.Credential),
			ServerPort: fmt.Sprintf("%s:22", vm.PublicIP),
		}

		if vm.Role == config.CONTROL_PLANE {
			var clusterConfig string
			workerJoinCmd, clusterConfig, err = vm.ControlPlaneInit(&sshInfo, vm.PublicIP)
			if err != nil {
				fmt.Println(vm.Name+" init failed", err)
				cluster.Fail(cluster)
				return nil, err
			}
			cluster.ClusterConfig = clusterConfig
		} else {
			joinResult, err := vm.WorkerJoin(&sshInfo, &workerJoinCmd)
			if err != nil {
				fmt.Println(vm.Name+" join error", err)
				cluster.Fail(cluster)
				return nil, err
			}
			if !joinResult {
				cluster.Fail(cluster)
				return nil, errors.New(vm.Name + " join failed")
			}
		}
	}
	fmt.Println("end k8s init & join")

	cluster.Complete(cluster)

	return cluster, nil
}

func DestroyCluster(namespace string, clusterName string) (*model.Status, error) {
	mcisName := clusterName //cluster 이름과 동일하게 (임시)

	status := model.NewStatus()
	status.Code = model.STATUS_UNKNOWN

	// 1. delete mcis
	fmt.Println(fmt.Sprintf("start delete MCIS (name=%s)", mcisName))
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
		fmt.Println(fmt.Sprintf("delete MCIS OK.. (name=%s)", mcisName))
		status.Code = model.STATUS_SUCCESS
		status.Message = "success"

		cluster := model.NewCluster(namespace, clusterName)
		if err := cluster.Delete(cluster); err != nil {
			status.Message = "delete success but cannot delete from the store"
			return status, nil
		}
	} else {
		status.Code = model.STATUS_NOT_EXIST
		fmt.Println(fmt.Sprintf("delete MCIS skip (cannot find).. (name=%s)", mcisName))
	}

	return status, nil
}
