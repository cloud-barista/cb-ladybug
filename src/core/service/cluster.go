package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloud-barista/cb-mcks/src/core/app"
	"github.com/cloud-barista/cb-mcks/src/core/model"
	"github.com/cloud-barista/cb-mcks/src/core/provision"
	"github.com/cloud-barista/cb-mcks/src/core/tumblebug"
	"github.com/cloud-barista/cb-mcks/src/utils/lang"

	logger "github.com/sirupsen/logrus"
)

/* get clusters */
func ListCluster(namespace string) (*model.ClusterList, error) {

	// validate namespace
	if err := verifyNamespace(namespace); err != nil {
		return nil, err
	}

	clusters := model.NewClusterList(namespace)
	if err := clusters.SelectList(); err != nil {
		return nil, err
	}
	return clusters, nil
}

/* get a cluster */
func GetCluster(namespace string, clusterName string) (*model.Cluster, error) {

	// validate namespace
	if err := verifyNamespace(namespace); err != nil {
		return nil, err
	}

	// get
	cluster := model.NewCluster(namespace, clusterName)
	if exists, err := cluster.Select(); err != nil {
		return nil, err
	} else if !exists {
		return nil, errors.New(fmt.Sprintf("Could not be found a cluster '%s' (namespace=%s)", clusterName, namespace))
	}
	return cluster, nil
}

/* create a cluster */
func CreateCluster(namespace string, req *app.ClusterReq) (*model.Cluster, error) {

	// validate a namespace
	if err := verifyNamespace(namespace); err != nil {
		return nil, err
	}
	// ibm, cloudit 일 경우에는 현재 haproxy만 사용하도록 함. 추후 지원 예정
	if req.Config.Kubernetes.Loadbalancer != app.LB_HAPROXY {
		connection := tumblebug.NewConnection(req.ControlPlane[0].Connection)
		exists, _ := connection.GET()
		if exists {
			if strings.ToLower(connection.ProviderName) == string(app.CSP_IBM) || strings.ToLower(connection.ProviderName) == string(app.CSP_NCP) || strings.ToLower(connection.ProviderName) == string(app.CSP_NCPVPC) {
				return nil, errors.New(fmt.Sprintf("%s does not yet supported nlb loadbalancer.", strings.ToLower(connection.ProviderName)))
			}
		}
	}

	k8sVersion := fmt.Sprintf("%s-00", req.Config.Kubernetes.Version)

	// validate prameters
	if req.ControlPlane[0].Count < 1 {
		return nil, errors.New("Control-Plane count must be at least one.")
	}
	if len(req.Worker) < 1 {
		return nil, errors.New("Worker must be at least one.")
	} else {
		for _, worker := range req.Worker {
			if worker.Count < 1 {
				return nil, errors.New(fmt.Sprintf("Worker count must be at least one. (connection=%s)", worker.Connection))
			}
		}
	}

	clusterName := req.Name
	mcisName := clusterName

	// validate exists & clean-up cluster
	cluster := model.NewCluster(namespace, clusterName)
	if exists, err := cluster.Select(); err != nil {
		return nil, err
	} else if exists == true {
		// clean-up if "exists" & "failed-status"
		if cluster.Status.Phase == model.ClusterPhaseFailed {
			logger.Infof("[%s.%s] Clean up a cluster (phase=%s, reason=%s, cause='cluster is already exists') ", namespace, clusterName, cluster.Status.Phase, cluster.Status.Reason)
			_, err = DeleteCluster(namespace, clusterName)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New(fmt.Sprintf("The cluster '%s' already exists. (namespace=%s)", clusterName, namespace))
		}
	}
	logger.Infof("[%s.%s] Validation & clean-up has been completed.", namespace, clusterName)

	// set cluster paramaters
	cluster.Version = k8sVersion
	cluster.NetworkCni = req.Config.Kubernetes.NetworkCni
	cluster.Label = req.Label
	cluster.InstallMonAgent = req.Config.InstallMonAgent
	cluster.Loadbalancer = req.Config.Kubernetes.Loadbalancer
	cluster.Etcd = req.Config.Kubernetes.Etcd
	cluster.Description = req.Description
	provisioner := provision.NewProvisioner(cluster)

	//update phase(provisioning)
	if err := cluster.UpdatePhase(model.ClusterPhaseProvisioning); err != nil {
		return nil, err
	}
	logger.Infof("[%s.%s] The phase update has been completed.", namespace, clusterName)

	// validate exists a MCIS
	mcis := tumblebug.NewMCIS(namespace, mcisName)
	if exists, err := mcis.GET(); err != nil {
		cluster.FailReason(model.GetMCISFailedReason, err.Error())
		return nil, errors.New(cluster.Status.Message)
	} else if exists {
		cluster.FailReason(model.AlreadyExistMCISFailedReason, fmt.Sprintf("MCIS already exists. (namespace=%s, mcis=%s)", namespace, mcisName))
		return nil, errors.New(cluster.Status.Message)
	}
	logger.Infof("[%s.%s] MCIS validation has been completed. (mcis=%s)", namespace, clusterName, mcisName)

	// create a MCIR - "vpc, f/w, sshkey, image, spec" - with vlidations
	mcir := NewMCIR(namespace, app.CONTROL_PLANE, *req.ControlPlane[0])

	reason, msg := mcir.CreateIfNotExist()
	if reason != "" {
		cluster.FailReason(reason, msg)
		return nil, errors.New(msg)
	} else {
		// make mics reuqest & provisioner data
		name := lang.GenerateNewNodeName(string(app.CONTROL_PLANE), 1)
		cluster.CpGroup = name
		mcis.VMs = append(mcis.VMs, mcir.NewVM(namespace, name, mcisName, strconv.Itoa(req.ControlPlane[0].Count), req.ControlPlane[0].RootDisk.Type, req.ControlPlane[0].RootDisk.Size))
	}
	logger.Infof("[%s.%s] MCIR(control-plane) creation has been completed.", namespace, clusterName)

	idx := 0
	for _, worker := range req.Worker {
		mcir := NewMCIR(namespace, app.WORKER, *worker)
		reason, msg := mcir.CreateIfNotExist()
		if reason != "" {
			cluster.FailReason(reason, msg)
			return nil, errors.New(msg)
		} else {
			// make mics reuqest & provisioner data
			for i := 0; i < mcir.vmCount; i++ {
				name := lang.GenerateNewNodeName(string(app.WORKER), idx+1)
				mcis.VMs = append(mcis.VMs, mcir.NewVM(namespace, name, mcisName, "", worker.RootDisk.Type, worker.RootDisk.Size))
				provisioner.AppendWorkerNodeMachine(name+"-1", mcir.csp, mcir.region, mcir.zone, mcir.credential)
				idx = idx + 1
			}
		}
	}
	logger.Infof("[%s.%s] MCIR(worker nodes) creation has been completed.", namespace, clusterName)

	// create a MCIS (contains vm)
	mcis.Label = app.MCIS_LABEL
	mcis.InstallMonAgent = cluster.InstallMonAgent
	mcis.SystemLabel = app.MCIS_SYSTEMLABEL
	if err := mcis.POST(); err != nil {
		cluster.FailReason(model.CreateMCISFailedReason, fmt.Sprintf("Failed to create a MCIS. (cause='%v')", err))
		return nil, errors.New(cluster.Status.Message)
	} else {
		logger.Debugf("[%s.%s] MCIS status is '%s' & vms='%v'", namespace, clusterName, mcis.Status, mcis.VMs)
	}
	cluster.MCIS = mcisName
	logger.Infof("[%s.%s] MCIS creation has been completed.", namespace, clusterName)
	cluster.CpLeader = mcis.VMs[0].Name

	for _, vms := range mcis.VMs {
		if cluster.CpGroup == vms.VmGroupId {
			provisioner.AppendControlPlaneMachine(vms.Name, mcir.csp, mcir.region, mcir.zone, mcir.credential)
		}
	}
	//create a NLB (contains control-plane)
	if cluster.Loadbalancer != app.LB_HAPROXY {
		NLB := mcir.NewNLB(namespace, mcisName, cluster.CpGroup)
		if exists, err := NLB.GET(); err != nil {
			cluster.FailReason(model.CreateNLBFailedReason, err.Error())
			return nil, errors.New(cluster.Status.Message)
		} else if !exists {
			if err := NLB.POST(); err != nil {
				cluster.FailReason(model.CreateNLBFailedReason, fmt.Sprintf("Failed to create a NLB. (cause='%v')", NLB))
				return nil, errors.New(cluster.Status.Message)
			}
			logger.Infof("[%s] NLB creation has been completed. (%s)", req.ControlPlane[0].Connection, NLB.TargetGroup.VmGroupId)
		}
	}

	// update received data & save nodes metadata
	if nodes, err := provisioner.BindVM(mcis.VMs); err != nil {
		cluster.FailReason(model.AddNodeEntityFailedReason, err.Error())
		cleanUpCluster(*cluster, mcis)
		return nil, errors.New(cluster.Status.Message)
	} else {
		cluster.Nodes = nodes
		if err := cluster.PutStore(); err != nil {
			cluster.FailReason(model.AddNodeEntityFailedReason, fmt.Sprintf("Failed to add node entity. (cause='%v')", err))
			return nil, errors.New(cluster.Status.Message)
		}
	}

	// kubernetes provisioning : bootstrap
	time.Sleep(2 * time.Second)
	if err := provisioner.Bootstrap(); err != nil {
		cluster.FailReason(model.SetupBoostrapFailedReason, fmt.Sprintf("Bootstrap failed. (cause='%v')", err))
		cleanUpCluster(*cluster, mcis)
		return nil, errors.New(cluster.Status.Message)
	}
	logger.Infof("[%s.%s] Bootstrap has been completed.", namespace, clusterName)

	if cluster.Loadbalancer == app.LB_HAPROXY {
		// kubernetes provisioning : haproxy
		if err := provisioner.InstallHAProxy(); err != nil {
			cluster.FailReason(model.SetupHaproxyFailedReason, fmt.Sprintf("Failed to install haproxy. (cause='%v')", err))
			cleanUpCluster(*cluster, mcis)
			return nil, errors.New(cluster.Status.Message)
		}
		logger.Infof("[%s.%s] HAProxy installation has been completed.", namespace, clusterName)
	}

	if cluster.Etcd == app.ETCD_EXTERNAL {
		time.Sleep(2 * time.Second)
		if err := provisioner.InitExternalEtcd(); err != nil {
			cluster.FailReason(model.InitExternalEtcdFailedReason, fmt.Sprintf("Failed to initialize External etcd. (cause='%v')", err))
			cleanUpCluster(*cluster, mcis)
			return nil, errors.New(cluster.Status.Message)
		}
		logger.Infof("[%s.%s] External etcd initialize has been completed.", namespace, clusterName)
	}

	// kubernetes provisioning :control-plane init
	var joinCmds []string
	joinCmds, kubeconfig, err := provisioner.InitControlPlane(req.Config.Kubernetes)
	if err != nil {
		cluster.FailReason(model.InitControlPlaneFailedReason, fmt.Sprintf("Fail to initialize Control-plane. (cause='%v')", err))
		cleanUpCluster(*cluster, mcis)
		return nil, errors.New(cluster.Status.Message)
	}
	cluster.ClusterConfig = kubeconfig
	logger.Infof("[%s.%s] Control-Plane initialize has been completed.", namespace, clusterName)

	// kubernetes provisioning : control-plane join
	for _, machine := range provisioner.ControlPlaneMachines {
		if provisioner.Cluster.CpLeader != machine.Name {
			if err := machine.JoinControlPlane(&joinCmds[0]); err != nil {
				cluster.FailReason(model.JoinControlPlaneFailedReason, fmt.Sprintf("Fail to control-plane join. (node=%s)", machine.Name))
				cleanUpCluster(*cluster, mcis)
				return nil, errors.New(cluster.Status.Message)
			}
		}
	}
	logger.Infof("[%s.%s] Control-Plane join has been completed.", namespace, clusterName)

	// kubernetes provisioning : worker node join
	for _, machine := range provisioner.WorkerNodeMachines {
		if err := machine.JoinWorker(&joinCmds[1]); err != nil {
			cluster.FailReason(model.JoinWorkerFailedReason, fmt.Sprintf("Fail to worker-node join. (node=%s)", machine.Name))
			cleanUpCluster(*cluster, mcis)
			return nil, errors.New(cluster.Status.Message)
		}
	}
	logger.Infof("[%s.%s] Woker-nodes join has been completed.", namespace, clusterName)

	// assign node labels (topology.cloud-barista.github.io/csp , topology.kubernetes.io/region, topology.kubernetes.io/zone)
	if err = provisioner.AssignNodeLabelAnnotation(); err != nil {
		logger.Warnf("[%s.%s] Failed to assign node labels (cause='%v')", namespace, clusterName, err)
	} else {
		logger.Infof("[%s.%s] Node label assignment has been completed.", namespace, clusterName)
	}

	// kubernetes provisioning : deploy network-cni
	if err = provisioner.InstallNetworkCni(); err != nil {
		cluster.FailReason(model.SetupNetworkCNIFailedReason, fmt.Sprintf("Failed to install network-cni. (cni=%s)", req.Config.Kubernetes.NetworkCni))
		cleanUpCluster(*cluster, mcis)
		return nil, errors.New(cluster.Status.Message)
	}
	logger.Infof("[%s.%s] CNI installation has been completed.", namespace, clusterName)

	// kubernetes provisioning : setting storageclass
	if req.Config.Kubernetes.StorageClass.Nfs.Server != "" && req.Config.Kubernetes.StorageClass.Nfs.Path != "" {
		if err = provisioner.InstallStorageClassNFS(req.Config.Kubernetes.StorageClass.Nfs); err != nil {
			cluster.FailReason(model.SetupStorageClassFailedReason, fmt.Sprintf("Failed to install storageclass. (cause='%v')", err))
			cleanUpCluster(*cluster, mcis)
			return nil, errors.New(cluster.Status.Message)
		}
		logger.Infof("[%s.%s] Storageclass installation has been completed.", namespace, clusterName)
	}

	// save nodes metadata & update status
	for _, node := range cluster.Nodes {
		node.CreatedTime = lang.GetNowUTC()
	}
	cluster.UpdatePhase(model.ClusterPhaseProvisioned)
	logger.Infof("[%s.%s] Cluster creation has been completed.", namespace, clusterName)
	return cluster, nil
}

/* delete a cluster */
func DeleteCluster(namespace string, clusterName string) (*app.Status, error) {

	// validate namespace
	if err := verifyNamespace(namespace); err != nil {
		return nil, err
	}
	// validate exists
	cluster := model.NewCluster(namespace, clusterName)
	if exists, err := cluster.Select(); err != nil {
		return nil, err
	} else if !exists {
		return app.NewStatus(app.STATUS_NOTFOUND, fmt.Sprintf("Could not be found cluster '%s'. (namespace=%s)", clusterName, namespace)), nil
	}

	// set a stauts
	cluster.UpdatePhase(model.ClusterPhaseDeleting)

	// delete a MCIS
	if cluster.MCIS != "" {
		logger.Infof("[%s.%s] MCIS deletion start.", namespace, clusterName)
		mcis := tumblebug.NewMCIS(namespace, cluster.MCIS)
		if exist, err := mcis.GET(); err != nil {
			return nil, err
		} else if exist {
			if err = cleanUpMCIS(clusterName, mcis); err != nil {
				return nil, err
			} else {
				logger.Infof("[%s.%s] Clean-up MCIS has been completed.", namespace, clusterName)
			}
		}
		logger.Infof("[%s.%s] MCIS deletion has been completed.", namespace, clusterName)
	}

	// delete a cluster-entity
	if err := cluster.Delete(); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to delete a cluster-entity. (namespace=%s, cluster=%s)", namespace, clusterName))
	}

	logger.Infof("[%s.%s] Cluster deletion has been completed.", namespace, clusterName)
	return app.NewStatus(app.STATUS_SUCCESS, fmt.Sprintf("Cluster '%s' has been deleted", clusterName)), nil
}

/* clean-up a Cluster(with MCIS) & update a cluster-entity */
func cleanUpCluster(cluster model.Cluster, mcis *tumblebug.MCIS) {
	for _, node := range cluster.Nodes {
		node.Credential = ""
		node.PublicIP = ""
	}
	if err := cluster.PutStore(); err != nil {
		logger.Warnf("[%s.%s] Failed to update a cluster-entity. (cause='%v')", cluster.Namespace, cluster.Name, err)
	}

	err := cleanUpMCIS(cluster.Name, mcis)
	if err != nil {
		logger.Warnf("[%s.%s] Failed to clean up a MCIS. (cause='%v')", cluster.Namespace, cluster.Name, err)
	} else {
		logger.Infof("[%s.%s] Garbage data has been cleaned.", cluster.Namespace, cluster.Name)
	}
}

/* clean-up a MCIS  */
func cleanUpMCIS(clusterName string, mcis *tumblebug.MCIS) error {

	if err := mcis.TERMINATE(); err != nil {
		return errors.New(fmt.Sprintf("Failed to terminate a MCIS (mcis=%s, cause='%v')", mcis.Name, err))
	}
	time.Sleep(5 * time.Second)

	if _, err := mcis.DELETE(); err != nil {
		if err = mcis.REFINE(); err != nil {
			logger.Warnf("[%s.%s] Failed to refine a MCIS. (mcis=%s, cause='%v')", mcis.Namespace, clusterName, mcis.Name, err)
		}
		if _, err = mcis.DELETE(); err != nil {
			return errors.New(fmt.Sprintf("Failed to delete a MCIS (cause='%v')", err))
		}
	}

	return nil

}

/* verify namespace  */
func verifyNamespace(namespace string) error {
	ns := tumblebug.NewNS(namespace)
	if exists, err := ns.GET(); err != nil {
		return err
	} else if !exists {
		return errors.New(fmt.Sprintf("Could not be found a namespace. (namespace=%s)", namespace))
	}
	return nil
}
