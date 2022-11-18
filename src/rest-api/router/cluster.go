package router

import (
	"errors"
	"net/http"
	"time"

	"github.com/cloud-barista/cb-mcks/src/core/app"
	"github.com/cloud-barista/cb-mcks/src/core/service"
	"github.com/cloud-barista/cb-mcks/src/utils/lang"
	"github.com/labstack/echo/v4"

	logger "github.com/sirupsen/logrus"
)

// ListCluster godoc
// @Tags Cluster
// @Summary List all Clusters
// @Description List all Clusters
// @ID ListCluster
// @Accept json
// @Produce json
// @Param	namespace	path	string	true  "Namespace ID"
// @Success 200 {object} model.ClusterList
// @Failure 400 {object} app.Status
// @Router /ns/{namespace}/clusters [get]
func ListCluster(c echo.Context) error {
	clusterList, err := service.ListCluster(c.Param("namespace"))
	if err != nil {
		logger.Warnf("(ListCluster) %s'", err.Error())
		return app.SendMessage(c, http.StatusBadRequest, err.Error())
	}

	return app.Send(c, http.StatusOK, clusterList)
}

// GetCluster godoc
// @Tags Cluster
// @Summary Get Cluster
// @Description Get Cluster
// @ID GetCluster
// @Accept json
// @Produce json
// @Param	namespace	path	string	true  "Namespace ID"
// @Param	cluster	path	string	true  "Cluster Name"
// @Success 200 {object} model.Cluster
// @Failure 400 {object} app.Status
// @Failure 404 {object} app.Status
// @Router /ns/{namespace}/clusters/{cluster} [get]
func GetCluster(c echo.Context) error {
	if err := app.Validate(c, []string{"namespace", "cluster"}); err != nil {
		logger.Warnf("(CreateCluster) %s", err.Error())
		return app.SendMessage(c, http.StatusBadRequest, err.Error())
	}

	cluster, err := service.GetCluster(c.Param("namespace"), c.Param("cluster"))
	if err != nil {
		logger.Warnf("(GetCluster) %s", err.Error())
		return app.SendMessage(c, http.StatusNotFound, err.Error())
	}

	return app.Send(c, http.StatusOK, cluster)
}

// CreateCluster godoc
// @Tags Cluster
// @Summary Create Cluster
// @Description Create Cluster
// @ID CreateCluster
// @Accept json
// @Produce json
// @Param	namespace	path	string	true  "Namespace ID"
// @Param ClusterReq body app.ClusterReq true "Request Body to create cluster"
// @Success 200 {object} model.Cluster
// @Failure 400 {object} app.Status
// @Failure 500 {object} app.Status
// @Router /ns/{namespace}/clusters [post]
func CreateCluster(c echo.Context) error {
	start := time.Now()
	clusterReq := &app.ClusterReq{}
	if err := c.Bind(clusterReq); err != nil {
		logger.Warnf("(CreateCluster) %s", err.Error())
		return app.SendMessage(c, http.StatusBadRequest, err.Error())
	}

	err := validateCreateClusterReq(clusterReq)
	if err != nil {
		logger.Warnf("(CreateCluster) %s", err.Error())
		return app.SendMessage(c, http.StatusBadRequest, err.Error())
	}
	cluster, err := service.CreateCluster(c.Param("namespace"), clusterReq)
	if err != nil {
		logger.Warnf("(CreateCluster) %s", err.Error())
		return app.SendMessage(c, http.StatusInternalServerError, err.Error())
	}

	logger.Info("(CreateCluster) Duration = ", time.Since(start))
	return app.Send(c, http.StatusOK, cluster)
}

// DeleteCluster godoc
// @Tags Cluster
// @Summary Delete Cluster
// @Description Delete Cluster
// @ID DeleteCluster
// @Accept json
// @Produce json
// @Param	namespace	path	string	true  "Namespace ID"
// @Param	cluster	path	string	true  "Cluster Name"
// @Success 200 {object} app.Status
// @Failure 400 {object} app.Status
// @Failure 500 {object} app.Status
// @Router /ns/{namespace}/clusters/{cluster} [delete]
func DeleteCluster(c echo.Context) error {
	start := time.Now()

	if err := app.Validate(c, []string{"namespace", "cluster"}); err != nil {
		logger.Warnf("(DeleteCluster) %s", err.Error())
		return app.SendMessage(c, http.StatusBadRequest, err.Error())
	}

	status, err := service.DeleteCluster(c.Param("namespace"), c.Param("cluster"))
	if err != nil {
		logger.Warnf("(DeleteCluster) %s", err.Error())
		return app.SendMessage(c, http.StatusInternalServerError, err.Error())
	}

	logger.Info("(DeleteCluster) Duration = ", time.Since(start))
	return app.Send(c, http.StatusOK, status)
}

func validateCreateClusterReq(clusterReq *app.ClusterReq) error {

	clusterReq.Config.Kubernetes.Version = lang.NVL(clusterReq.Config.Kubernetes.Version, "1.23.13")
	clusterReq.Config.Kubernetes.PodCidr = lang.NVL(clusterReq.Config.Kubernetes.PodCidr, app.POD_CIDR)
	clusterReq.Config.Kubernetes.ServiceCidr = lang.NVL(clusterReq.Config.Kubernetes.ServiceCidr, app.SERVICE_CIDR)
	clusterReq.Config.Kubernetes.ServiceDnsDomain = lang.NVL(clusterReq.Config.Kubernetes.ServiceDnsDomain, app.SERVICE_DOMAIN)
	if len(clusterReq.Config.Kubernetes.Loadbalancer) == 0 {
		clusterReq.Config.Kubernetes.Loadbalancer = app.LB_HAPROXY
	}
	if len(clusterReq.Config.Kubernetes.Loadbalancer) == 0 {
		clusterReq.Config.Kubernetes.Etcd = app.ETCD_LOCAL
	}
	if len(clusterReq.Config.Kubernetes.NetworkCni) == 0 {
		clusterReq.Config.Kubernetes.NetworkCni = app.NETWORKCNI_KILO
	}
	clusterReq.Config.InstallMonAgent = lang.NVL(clusterReq.Config.InstallMonAgent, "no")

	if len(clusterReq.ControlPlane) == 0 {
		return errors.New("Control plane node must be at least one")
	}
	if len(clusterReq.ControlPlane) > 1 {
		return errors.New("Only one control plane node is supported")
	}
	if len(clusterReq.Worker) == 0 {
		return errors.New("Worker node must be at least one")
	}
	if !(clusterReq.Config.Kubernetes.NetworkCni == app.NETWORKCNI_CANAL || clusterReq.Config.Kubernetes.NetworkCni == app.NETWORKCNI_KILO) {
		return errors.New("Network-cni allows only canal or kilo")
	}
	if len(clusterReq.Config.Kubernetes.Loadbalancer) != 0 && !(clusterReq.Config.Kubernetes.Loadbalancer == app.LB_HAPROXY || clusterReq.Config.Kubernetes.Loadbalancer == app.LB_NLB) {
		return errors.New("loadbalancer allows only haproxy or nlb")
	}
	if !(clusterReq.Config.Kubernetes.Etcd == app.ETCD_LOCAL || clusterReq.Config.Kubernetes.Etcd == app.ETCD_EXTERNAL) {
		return errors.New("etcd allows only local or external")
	}
	if clusterReq.Config.Kubernetes.Etcd == app.ETCD_EXTERNAL && (clusterReq.ControlPlane[0].Count != 3 && clusterReq.ControlPlane[0].Count != 5 && clusterReq.ControlPlane[0].Count != 7) {
		return errors.New("External etcd must have 3,5,7 controlPlane count")
	}
	if len(clusterReq.Name) == 0 {
		return errors.New("Cluster name is empty")
	} else {
		err := lang.VerifyClusterName(clusterReq.Name)
		if err != nil {
			return err
		}
	}
	if len(clusterReq.Config.Kubernetes.PodCidr) > 0 {
		err := lang.VerifyCIDR("podCidr", clusterReq.Config.Kubernetes.PodCidr)
		if err != nil {
			return err
		}
	}
	if len(clusterReq.Config.Kubernetes.ServiceCidr) > 0 {
		err := lang.VerifyCIDR("serviceCidr", clusterReq.Config.Kubernetes.ServiceCidr)
		if err != nil {
			return err
		}
	}

	// control plane nodes
	for _, set := range clusterReq.ControlPlane {
		if err := validateNodeSetReq(set); err != nil {
			return err
		}
	}
	// worker nodes
	for _, set := range clusterReq.Worker {
		if err := validateNodeSetReq(set); err != nil {
			return err
		}
	}

	return nil
}
