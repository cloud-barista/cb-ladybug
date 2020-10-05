package router

import (
	"net/http"

	"github.com/cloud-barista/cb-ladybug/src/core/common"
	"github.com/cloud-barista/cb-ladybug/src/core/model"
	"github.com/cloud-barista/cb-ladybug/src/rest-api/service"
	"github.com/cloud-barista/cb-ladybug/src/utils/app"
	"github.com/labstack/echo/v4"
)

// ListCluster
// @Tags Cluster
// @Summary List Cluster
// @Description List Cluster
// @ID ListCluster
// @Accept json
// @Produce json
// @Param	namespace	path	string	true  "namespace"
// @Success 200 {object} model.ClusterList
// @Router /ns/{namespace}/clusters [get]
func ListCluster(c echo.Context) error {
	if err := app.Validate(c, []string{"namespace"}); err != nil {
		common.CBLog.Error(err)
		return app.SendMessage(c, http.StatusBadRequest, err.Error())
	}

	clusterList, err := service.ListCluster(c.Param("namespace"))
	if err != nil {
		common.CBLog.Error(err)
		return app.SendMessage(c, http.StatusBadRequest, err.Error())
	}

	return app.Send(c, http.StatusOK, clusterList)
}

// GetCluster
// @Tags Cluster
// @Summary Get Cluster
// @Description Get Cluster
// @ID GetCluster
// @Accept json
// @Produce json
// @Param	namespace	path	string	true  "namespace"
// @Param	cluster	path	string	true  "cluster"
// @Success 200 {object} model.Cluster
// @Router /ns/{namespace}/clusters/{cluster} [get]
func GetCluster(c echo.Context) error {
	if err := app.Validate(c, []string{"namespace", "cluster"}); err != nil {
		common.CBLog.Error(err)
		return app.SendMessage(c, http.StatusBadRequest, err.Error())
	}

	cluster, err := service.GetCluster(c.Param("namespace"), c.Param("cluster"))
	if err != nil {
		common.CBLog.Error(err)
		return app.SendMessage(c, http.StatusBadRequest, err.Error())
	}

	return app.Send(c, http.StatusOK, cluster)
}

// CreateCluster
// @Tags Cluster
// @Summary Create Cluster
// @Description Create Cluster
// @ID CreateCluster
// @Accept json
// @Produce json
// @Param	namespace	path	string	true  "namespace"
// @Param	cluster	path	string	true  "cluster"
// @Param json body model.ClusterReq true "Reuest json"
// @Success 200 {object} model.Cluster
// @Router /ns/{namespace}/clusters/{cluster} [post]
func CreateCluster(c echo.Context) error {
	if err := app.Validate(c, []string{"namespace", "cluster"}); err != nil {
		common.CBLog.Error(err)
		return app.SendMessage(c, http.StatusBadRequest, err.Error())
	}

	clusterReq := &model.ClusterReq{}
	if err := c.Bind(clusterReq); err != nil {
		common.CBLog.Error(err)
		return app.SendMessage(c, http.StatusBadRequest, err.Error())
	}

	cluster, err := service.CreateCluster(c.Param("namespace"), c.Param("cluster"), clusterReq)
	if err != nil {
		common.CBLog.Error(err)
		return app.SendMessage(c, http.StatusBadRequest, err.Error())
	}

	return app.Send(c, http.StatusOK, cluster)
}

// DestroyCluster
// @Tags Cluster
// @Summary Destroy Cluster
// @Description Destroy Cluster
// @ID DestroyCluster
// @Accept json
// @Produce json
// @Param	namespace	path	string	true  "namespace"
// @Param	cluster	path	string	true  "cluster"
// @Success 200 {object} model.Status
// @Router /ns/{namespace}/clusters/{cluster} [delete]
func DestroyCluster(c echo.Context) error {
	if err := app.Validate(c, []string{"namespace", "cluster"}); err != nil {
		common.CBLog.Error(err)
		return app.SendMessage(c, http.StatusBadRequest, err.Error())
	}

	status, err := service.DestroyCluster(c.Param("namespace"), c.Param("cluster"))
	if err != nil {
		common.CBLog.Error(err)
		return app.SendMessage(c, http.StatusBadRequest, err.Error())
	}

	return app.Send(c, http.StatusOK, status)
}
