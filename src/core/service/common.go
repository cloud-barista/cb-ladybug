package service

import (
	"errors"
	"fmt"

	"github.com/cloud-barista/cb-mcks/src/core/model"
	"github.com/cloud-barista/cb-mcks/src/core/model/tumblebug"
	"github.com/cloud-barista/cb-mcks/src/utils/config"
)

type NodeConfigInfo struct {
	model.NodeConfig
	Csp     config.CSP `json:"csp"`
	Role    string     `json:"role"`
	Account string     `json:"account"`
}

func SetNodeConfigInfos(nodeConfigs []model.NodeConfig, role string) ([]NodeConfigInfo, error) {
	var nodeConfigInfos []NodeConfigInfo

	for _, nodeConfig := range nodeConfigs {
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
			return nil, errors.New(fmt.Sprintf("%s get region error (connectionName=%s)", role, nodeConfig.Connection))
		}
		if !exists {
			return nil, errors.New(fmt.Sprintf("%s region does not exist (connectionName=%s)", role, nodeConfig.Connection))
		}

		var nodeConfigInfo NodeConfigInfo
		nodeConfigInfo.Connection = nodeConfig.Connection
		nodeConfigInfo.Count = nodeConfig.Count
		nodeConfigInfo.Spec = nodeConfig.Spec
		nodeConfigInfo.Csp = csp
		nodeConfigInfo.Role = role
		nodeConfigInfo.Account = GetUserAccount(nodeConfigInfo.Csp)

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
