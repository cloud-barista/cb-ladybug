package model

import (
	"github.com/cloud-barista/cb-mcks/src/core/app"
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
	InvalidMCIRReason                         = ClusterReason("InvalidMCIRReason")
	CreateMCISFailedReason                    = ClusterReason("CreateMCISFailedReason")
	GetControlPlaneConnectionInfoFailedReason = ClusterReason("GetControlPlaneConnectionInfoFailedReason")
	GetWorkerConnectionInfoFailedReason       = ClusterReason("GetWorkerConnectionInfoFailedReason")
	CreateVpcFailedReason                     = ClusterReason("CreateVpcFailedReason")
	CreateSecurityGroupFailedReason           = ClusterReason("CreateSecurityGroupFailedReason")
	CreateSSHKeyFailedReason                  = ClusterReason("CreateSSHKeyFailedReason")
	CreateVmImageFailedReason                 = ClusterReason("CreateVmImageFailedReason")
	CreateVmSpecFailedReason                  = ClusterReason("CreateVmSpecFailedReason")
	CreateNLBFailedReason                     = ClusterReason("CreateNLBFailedReason")
	AddNodeEntityFailedReason                 = ClusterReason("AddNodeEntityFailedReason")
	SetupBoostrapFailedReason                 = ClusterReason("SetupBoostrapFailedReason")
	SetupHaproxyFailedReason                  = ClusterReason("SetupHaproxyFailedReason")
	InitControlPlaneFailedReason              = ClusterReason("InitControlPlaneFailedReason")
	InitExternalEtcdFailedReason              = ClusterReason("InitExternalEtcdFailedReason")
	SetupNetworkCNIFailedReason               = ClusterReason("SetupNetworkCNIFailedReason")
	SetupStorageClassFailedReason             = ClusterReason("SetupStorageClassFailedReason")
	JoinControlPlaneFailedReason              = ClusterReason("JoinControlPlaneFailedReason")
	JoinWorkerFailedReason                    = ClusterReason("JoinWorkerFailedReason")
)

type Model struct {
	Name string   `json:"name"`
	Kind app.Kind `json:"kind"`
}
type ListModel struct {
	Kind app.Kind `json:"kind"`
}

type Cluster struct {
	Model
	Status          ClusterStatus    `json:"status"`
	MCIS            string           `json:"mcis"`
	Namespace       string           `json:"namespace"`
	Version         string           `json:"k8sVersion"`
	ClusterConfig   string           `json:"clusterConfig"`
	CpLeader        string           `json:"cpLeader"`
	CpGroup         string           `json:"cpGroup"`
	NetworkCni      app.NetworkCni   `json:"networkCni" enums:"canal,kilo"`
	Label           string           `json:"label"`
	InstallMonAgent string           `json:"installMonAgent" example:"no" default:"yes"`
	Loadbalancer    app.Loadbalancer `json:"loadbalancer" enums:"haproxy,nlb" example:"haproxy" default:"haproxy"`
	Etcd            app.Etcd         `json:"etcd" enums:"local,external" example:"local" default:"local"`
	Description     string           `json:"description"`
	CreatedTime     string           `json:"createdTime" example:"2022-01-02T12:00:00Z" default:""`
	Nodes           []*Node          `json:"nodes"`
}

type ClusterStatus struct {
	Phase   ClusterPhase  `json:"phase" enums:"Pending,Provisioning,Provisioned,Failed"`
	Reason  ClusterReason `json:"reason"`
	Message string        `json:"message"`
}

type ClusterList struct {
	ListModel
	namespace string
	Items     []Cluster `json:"items"`
}

type Node struct {
	Model
	namespace   string
	clusterName string
	Credential  string   `json:"credential"`
	PublicIP    string   `json:"publicIp"`
	Role        app.ROLE `json:"role" enums:"control-plane,worker"`
	Spec        string   `json:"spec"`
	Csp         app.CSP  `json:"csp" enums:"aws,gcp,azure,alibaba,tencent,openstack,ibm,cloudit,ncp,ncpvpc,nhncloud"`
	CreatedTime string   `json:"createdTime" example:"2022-01-02T12:00:00Z" default:""`
	CspLabel    string   `json:"cspLabel"`
	RegionLabel string   `json:"regionLabel"`
	ZoneLabel   string   `json:"zoneLabel"`
}

type NodeList struct {
	ListModel
	namespace   string
	clusterName string
	Items       []*Node `json:"items"`
}
