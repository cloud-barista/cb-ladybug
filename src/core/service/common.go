package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloud-barista/cb-mcks/src/core/model"
	"github.com/cloud-barista/cb-mcks/src/core/model/tumblebug"
	"github.com/cloud-barista/cb-mcks/src/utils/config"
	"github.com/cloud-barista/cb-mcks/src/utils/lang"
)

type NodeConfigInfo struct {
	model.NodeConfig
	Csp     config.CSP `json:"csp"`
	Role    string     `json:"role"`
	ImageId string     `json:"imageId"`
}

func SetNodeConfigInfos(nodeConfigs []model.NodeConfig, role string) ([]NodeConfigInfo, error) {
	var nodeConfigInfos []NodeConfigInfo

	for _, nodeConfig := range nodeConfigs {
		if nodeConfig.Count < 1 {
			return nil, errors.New(fmt.Sprintf("%s count must be at least one (connectionName=%s)", role, nodeConfig.Connection))
		}

		conn := tumblebug.NewConnection(nodeConfig.Connection)
		exists, err := conn.GET()
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%s Connection connect error (connectionName=%s)", role, nodeConfig.Connection))
		}
		if !exists {
			return nil, errors.New(fmt.Sprintf("%s Connection does not exist (connectionName=%s)", role, nodeConfig.Connection))
		}
		csp, err := GetCSPName(conn.ProviderName)
		if err != nil {
			return nil, err
		}

		region := tumblebug.NewRegion(conn.RegionName)
		exists, err = region.GET()
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%s Region connect error (connectionName=%s)", role, nodeConfig.Connection))
		}
		if !exists {
			return nil, errors.New(fmt.Sprintf("%s Region does not exist (connectionName=%s)", role, nodeConfig.Connection))
		}

		imageId, err := GetVmImageId(csp, nodeConfig.Connection, region)
		if err != nil {
			return nil, err
		}

		err = CheckSpec(csp, nodeConfig.Connection, nodeConfig.Spec, role)
		if err != nil {
			return nil, err
		}

		var nodeConfigInfo NodeConfigInfo
		nodeConfigInfo.Connection = nodeConfig.Connection
		nodeConfigInfo.Count = nodeConfig.Count
		nodeConfigInfo.Spec = nodeConfig.Spec
		nodeConfigInfo.Csp = csp
		nodeConfigInfo.Role = role
		nodeConfigInfo.ImageId = imageId

		nodeConfigInfos = append(nodeConfigInfos, nodeConfigInfo)
	}

	return nodeConfigInfos, nil
}

func GetControlPlaneIPs(VMs []model.VM) []string {
	var IPs []string
	for _, vm := range VMs {
		if vm.Role == config.CONTROL_PLANE {
			IPs = append(IPs, vm.PrivateIP)
		}
	}
	return IPs
}

func GetVmImageName(name string) string {
	tmp := lang.GetOnlyLettersAndNumbers(name)

	return strings.ToLower(tmp)
}

func CheckNamespace(namespace string) error {
	ns := tumblebug.NewNS(namespace)
	exists, err := ns.GET()
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(fmt.Sprintf("namespace does not exist (name=%s)", namespace))
	}
	return nil
}

func CheckMcis(namespace string, mcisName string) error {
	mcis := tumblebug.NewMCIS(namespace, mcisName)
	exists, err := mcis.GET()
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(fmt.Sprintf("MCIS does not exist (name=%s)", mcisName))
	}
	return nil
}

func CheckClusterStatus(namespace string, clusterName string) error {
	cluster := model.NewCluster(namespace, clusterName)
	exists, err := cluster.Select()
	if err != nil {
		return err
	} else if exists == false {
		return errors.New(fmt.Sprintf("Cluster not found (namespace=%s, cluster=%s)", namespace, clusterName))
	} else if cluster.Status.Phase != model.ClusterPhaseProvisioned {
		return errors.New(fmt.Sprintf("cannot add node. status is '%s'", cluster.Status.Phase))
	}
	return nil
}

func CheckSpec(csp config.CSP, configName string, specName string, role string) error {
	lookupSpec := tumblebug.NewLookupSpec(configName, specName)
	err := lookupSpec.LookupSpec()
	if err != nil {
		return errors.New(fmt.Sprintf("failed to lookup spec (connection='%s', cause=%v)", configName, err))
	}

	if lookupSpec.SpiderSpecInfo.Name == "" {
		return errors.New(fmt.Sprintf("failed to find spec (connection='%s', specName='%s')", configName, specName))
	}

	if role == config.CONTROL_PLANE {
		vCpuCount, err := strconv.Atoi(lookupSpec.SpiderSpecInfo.VCpu.Count)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to convert vCpu count (connection='%s', specName='%s', vCpu.Count=%s)", configName, specName, lookupSpec.SpiderSpecInfo.VCpu.Count))
		}
		if vCpuCount < 2 {
			return errors.New(fmt.Sprintf("kubernetes control plane node needs 2 vCPU at least (connection='%s', specName='%s', vCpu.Count=%d)", configName, specName, vCpuCount))
		}
	}

	mem, err := strconv.Atoi(lookupSpec.SpiderSpecInfo.Mem)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to convert memory (connection='%s', specName='%s', Mem=%s)", configName, specName, lookupSpec.SpiderSpecInfo.Mem))
	}

	gbMem := mem
	if csp != config.CSP_TENCENT {
		gbMem = mem / 1024
	}
	if gbMem < 2 {
		return errors.New(fmt.Sprintf("kubernetes node needs 2 GiB or more of RAM (connection='%s', specName='%s', mem=%dGB)", configName, specName, gbMem))
	}

	return nil
}
