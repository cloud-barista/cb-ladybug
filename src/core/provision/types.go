package provision

import (
	"github.com/cloud-barista/cb-mcks/src/core/app"
	"github.com/cloud-barista/cb-mcks/src/core/model"
)

const (
	REMOTE_TARGET_PATH = "/tmp"

	CNI_CANAL_FILE        = "addons/cni/canal/canal_v3.20.0.yaml"
	CNI_KILO_CRDS_FILE    = "addons/cni/kilo/crds_v0.3.0.yaml"
	CNI_KILO_KUBEADM_FILE = "addons/cni/kilo/kilo-kubeadm-flannel_v0.3.0.yaml"
	CNI_KILO_FLANNEL_FILE = "addons/cni/kilo/kube-flannel_v0.14.0.yaml"
	CNI_FLANNEL_FILE      = "addons/cni/flannel/kube-flannel_v0.19.0.yml"
	CNI_CALICO_FILE       = "addons/cni/calico/calico-v0.3.1.yaml"

	SC_NFS_RBAC_FILE  = "addons/nfs/rbac_v4.0.16.yaml"
	SC_NFS_CLASS_FILE = "addons/nfs/class_v4.0.16.yaml"

	CCM_CLOUD_CONFIG_FILE            = "cloud-config"
	CCM_AWS_ROLE_SA_FILE             = "addons/ccm/aws/clusterrole-service-account.yaml"
	CCM_AWS_DS_FILE                  = "addons/ccm/aws/aws-cloud-controller-manager-daemonset.yaml"
	CCM_OPENSTACK_ROLE_BINDINGS_FILE = "addons/ccm/openstack/cloud-controller-manager-role-bindings.yaml"
	CCM_OPENSTACK_ROLES_FILE         = "addons/ccm/openstack/cloud-controller-manager-roles.yaml"
	CCM_OPENSTACK_DS_FILE            = "addons/ccm/openstack/openstack-cloud-controller-manager-ds.yaml"
	CCM_NCPVPC_ROLE_SA_FILE          = "addons/ccm/ncpvpc/clusterrole-service-account.yaml"
	CCM_NCPVPC_DS_FILE               = "addons/ccm/ncpvpc/ncp-cloud-controller-manager-daemonset.yaml"
)

type Machine struct {
	Name       string
	NameInCsp  string
	PublicIP   string
	PrivateIP  string
	Username   string
	CSP        app.CSP
	Role       app.ROLE
	Region     string
	Zone       string
	Spec       string
	Credential string
}
type ControlPlaneMachine struct {
	*Machine
}
type WorkerNodeMachine struct {
	*Machine
}

type Provisioner struct {
	Cluster              *model.Cluster
	leader               *ControlPlaneMachine
	ControlPlaneMachines map[string]*ControlPlaneMachine
	WorkerNodeMachines   map[string]*WorkerNodeMachine
}
