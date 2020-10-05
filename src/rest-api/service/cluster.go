package service

import (
	"github.com/cloud-barista/cb-ladybug/src/core/model"
)

func ListCluster(namespace string) (*model.ClusterList, error) {
	clusters := model.NewClusterList()

	return clusters, nil
}

func GetCluster(namespace string, clusterName string) (*model.Cluster, error) {
	cluster := model.NewCluster(clusterName)

	return cluster, nil
}

func CreateCluster(namespace string, clusterName string, clusterReq *model.ClusterReq) (*model.Cluster, error) {
	cluster := model.NewCluster(clusterName)

	return cluster, nil
}

func DestroyCluster(namespace string, clusterName string) (*model.Status, error) {
	status := model.NewStatus(model.STATUS_SUCCESS)
	status.Message = "success"

	return status, nil
}
