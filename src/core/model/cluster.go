package model

import (
	"encoding/json"
	"strings"

	"github.com/cloud-barista/cb-mcks/src/core/common"
	"github.com/cloud-barista/cb-mcks/src/utils/lang"
)

type ClusterPhase string
type ClusterReason string

const (
	ClusterPhasePending      = ClusterPhase("Pending")
	ClusterPhaseProvisioning = ClusterPhase("Provisioning")
	ClusterPhaseProvisioned  = ClusterPhase("Provisioned")
	ClusterPhaseFailed       = ClusterPhase("Failed")
	ClusterPhaseDeleting     = ClusterPhase("Deleting")

	GetMCISFailedReason                       = ClusterReason("GetMCISFailedReason")
	AlreadyExistMCISFailedReason              = ClusterReason("AlreadyExistMCISFailedReason")
	CreateMCISFailedReason                    = ClusterReason("CreateMCISFailedReason")
	GetControlPlaneConnectionInfoFailedReason = ClusterReason("GetControlPlaneConnectionInfoFailedReason")
	GetWorkerConnectionInfoFailedReason       = ClusterReason("GetWorkerConnectionInfoFailedReason")
	CreateVpcFailedReason                     = ClusterReason("CreateVpcFailedReason")
	CreateSecurityGroupFailedReason           = ClusterReason("CreateSecurityGroupFailedReason")
	CreateSSHKeyFailedReason                  = ClusterReason("CreateSSHKeyFailedReason")
	CreateVmImageFailedReason                 = ClusterReason("CreateVmImageFailedReason")
	CreateVmSpecFailedReason                  = ClusterReason("CreateVmSpecFailedReason")
	AddNodeEntityFailedReason                 = ClusterReason("AddNodeEntityFailedReason")
	SetupBoostrapFailedReason                 = ClusterReason("SetupBoostrapFailedReason")
	SetupHaproxyFailedReason                  = ClusterReason("SetupHaproxyFailedReason")
	InitControlPlaneFailedReason              = ClusterReason("InitControlPlaneFailedReason")
	SetupNetworkCNIFailedReason               = ClusterReason("SetupNetworkCNIFailedReason")
	JoinControlPlaneFailedReason              = ClusterReason("JoinControlPlaneFailedReason")
	JoinWorkerFailedReason                    = ClusterReason("JoinWorkerFailedReason")
)

type Cluster struct {
	Model
	Status struct {
		Phase   ClusterPhase  `json:"phase" enums:"Pending,Provisioning,Provisioned,Failed"`
		Reason  ClusterReason `json:"reason"`
		Message string        `json:"message"`
	} `json:"status"`
	MCIS          string `json:"mcis"`
	Namespace     string `json:"namespace"`
	ClusterConfig string `json:"clusterConfig"`
	CpLeader      string `json:"cpLeader"`
	NetworkCni    string `json:"networkCni" enums:"kilo,canal"`
	Nodes         []Node `json:"nodes"`
}

type ClusterList struct {
	ListModel
	namespace string
	Items     []Cluster `json:"items"`
}

func NewCluster(namespace string, name string) *Cluster {
	return &Cluster{
		Model:     Model{Kind: KIND_CLUSTER, Name: name},
		Namespace: namespace,
		Status: struct {
			Phase   ClusterPhase  "json:\"phase\" enums:\"Pending,Provisioning,Provisioned,Failed\""
			Reason  ClusterReason "json:\"reason\""
			Message string        "json:\"message\""
		}{Phase: ClusterPhasePending, Reason: "", Message: ""},
		Nodes: []Node{},
	}
}

func NewClusterList(namespace string) *ClusterList {
	return &ClusterList{
		ListModel: ListModel{Kind: KIND_CLUSTER_LIST},
		namespace: namespace,
		Items:     []Cluster{},
	}
}

func (self *Cluster) UpdatePhase(phase ClusterPhase) error {
	self.Status.Phase = phase
	if phase != ClusterPhaseFailed {
		self.Status.Reason = ""
		self.Status.Message = ""
	}
	return self.putStore()
}

func (self *Cluster) FailReason(reason ClusterReason, message string) error {
	self.Status.Phase = ClusterPhaseFailed
	self.Status.Reason = reason
	self.Status.Message = message
	return self.putStore()
}

func (self *Cluster) putStore() error {
	key := lang.GetStoreClusterKey(self.Namespace, self.Name)
	value, _ := json.Marshal(self)
	err := common.CBStore.Put(key, string(value))
	if err != nil {
		return err
	}
	return nil
}

func (self *Cluster) Select() (bool, error) {
	exists := false

	key := lang.GetStoreClusterKey(self.Namespace, self.Name)
	keyValue, err := common.CBStore.Get(key)
	if err != nil {
		return exists, err
	}
	exists = (keyValue != nil)
	if exists {
		json.Unmarshal([]byte(keyValue.Value), &self)
		err = getClusterNodes(self)
		if err != nil {
			return exists, err
		}
	}

	return exists, nil
}

func (self *Cluster) Delete() error {
	// delete node
	keyValues, err := common.CBStore.GetList(lang.GetStoreNodeKey(self.Namespace, self.Name, ""), true)
	if err != nil {
		return err
	}
	for _, keyValue := range keyValues {
		err = common.CBStore.Delete(keyValue.Key)
		if err != nil {
			return err
		}
	}

	// delete cluster
	key := lang.GetStoreClusterKey(self.Namespace, self.Name)
	err = common.CBStore.Delete(key)
	if err != nil {
		return err
	}

	return nil
}

func (self *ClusterList) SelectList() error {
	keyValues, err := common.CBStore.GetList(lang.GetStoreClusterKey(self.namespace, ""), true)
	if err != nil {
		return err
	}
	self.Items = []Cluster{}
	for _, keyValue := range keyValues {
		if !strings.Contains(keyValue.Key, "/nodes") {
			cluster := &Cluster{}
			json.Unmarshal([]byte(keyValue.Value), &cluster)

			err = getClusterNodes(cluster)
			if err != nil {
				return err
			}
			self.Items = append(self.Items, *cluster)
		}
	}

	return nil
}

func getClusterNodes(cluster *Cluster) error {
	nodeKeyValues, err := common.CBStore.GetList(lang.GetStoreNodeKey(cluster.Namespace, cluster.Name, ""), true)
	if err != nil {
		return err
	}
	for _, nodeKeyValue := range nodeKeyValues {
		node := &Node{}
		json.Unmarshal([]byte(nodeKeyValue.Value), &node)
		cluster.Nodes = append(cluster.Nodes, *node)
	}

	return nil
}
