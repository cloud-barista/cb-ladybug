package service

import (
	"fmt"
	"strings"

	"github.com/cloud-barista/cb-ladybug/src/core/model/tumblebug"
	logger "github.com/sirupsen/logrus"
)

func (nodeConfigInfo *NodeConfigInfo) CreateVPC(namespace string, clusterName string) (*tumblebug.VPC, error) {
	vpcName := fmt.Sprintf("%s-%s-vpc", clusterName, nodeConfigInfo.Csp)
	logger.Infof("start create vpc (name=%s)", vpcName)
	vpc := tumblebug.NewVPC(namespace, vpcName, nodeConfigInfo.Connection)
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
	return vpc, nil
}

func (nodeConfigInfo *NodeConfigInfo) CreateFirewall(namespace string, clusterName string) (*tumblebug.Firewall, error) {
	firewallName := fmt.Sprintf("%s-%s-allow-external", clusterName, nodeConfigInfo.Csp)
	vpcName := fmt.Sprintf("%s-%s-vpc", clusterName, nodeConfigInfo.Csp)
	logger.Infof("start create firewall (name=%s)", firewallName)
	fw := tumblebug.NewFirewall(namespace, firewallName, nodeConfigInfo.Connection)
	fw.VPCId = vpcName
	exists, e := fw.GET()
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
	return fw, nil
}

func (nodeConfigInfo *NodeConfigInfo) CreateSshKey(namespace string, clusterName string) (*tumblebug.SSHKey, error) {
	sshkeyName := fmt.Sprintf("%s-%s-sshkey", clusterName, nodeConfigInfo.Csp)
	logger.Infof("start create ssh key (name=%s)", sshkeyName)
	sshKey := tumblebug.NewSSHKey(namespace, sshkeyName, nodeConfigInfo.Connection)
	sshKey.Username = nodeConfigInfo.Account
	exists, e := sshKey.GET()
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
	return sshKey, nil
}

func (nodeConfigInfo *NodeConfigInfo) CreateImage(namespace string, clusterName string) (*tumblebug.Image, error) {
	imageId, e := GetVmImageId(nodeConfigInfo.Csp, nodeConfigInfo.Connection)
	if e != nil {
		return nil, e
	}

	imageName := fmt.Sprintf("%s-%s-Ubuntu1804", nodeConfigInfo.Connection, nodeConfigInfo.Region)
	logger.Infof("start create image (name=%s)", imageName)
	image := tumblebug.NewImage(namespace, imageName, nodeConfigInfo.Connection)
	image.CspImageId = imageId
	exists, e := image.GET()
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
	return image, nil
}

func (nodeConfigInfo *NodeConfigInfo) CreateSpec(namespace string, clusterName string) (*tumblebug.Spec, error) {
	specName := fmt.Sprintf("%s-spec", strings.ReplaceAll(nodeConfigInfo.Spec, ".", "-"))
	logger.Infof("start create spec (name=%s)", specName)
	spec := tumblebug.NewSpec(namespace, specName, nodeConfigInfo.Connection)
	spec.CspSpecName = nodeConfigInfo.Spec
	spec.Role = nodeConfigInfo.Role
	exists, e := spec.GET()
	if e != nil {
		return nil, e
	}
	if exists {
		logger.Infof("reuse spec (name=%s, cause='already exists')", specName)
	} else {
		if e = spec.POST(); e != nil {
			return nil, e
		}
		logger.Infof("create spec OK.. (name=%s)", specName)
	}
	return spec, nil
}
