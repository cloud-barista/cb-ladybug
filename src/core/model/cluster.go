package model

type Cluster struct {
	Model
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	MCIS          string `json:"mcis"`
	ClusterConfig string `json:"clusterConfig"`
	Nodes         []Node `json:"nodes"`
}

type ClusterList struct {
	Kind     string    `json:"kind"`
	Clusters []Cluster `json:"clusters"`
}

type ClusterReq struct {
	ClusterConfig         string `json:"clusterConfig"`
	Name                  string `json:"name"`
	ControlPlaneSpec      string `json:"controlPlaneSpec"`
	ControlPlaneNodeCount int    `json:"controlPlaneNodeCount"`
	WorkerNodeSpec        string `json:"workerNodeSpec"`
	WorkerNodeCount       int    `json:"workerNodeCount"`
}

func NewCluster(name string) *Cluster {
	return &Cluster{
		Model: Model{Kind: KIND_CLUSTER},
		Name:  name,
		Nodes: []Node{},
	}
}

func NewClusterList() *ClusterList {
	return &ClusterList{
		Kind:     KIND_CLUSTER_LIST,
		Clusters: []Cluster{},
	}
}
