package mcar

import (
	"errors"
	"fmt"

	"github.com/beego/beego/v2/core/validation"
	"github.com/cloud-barista/cb-ladybug/src/core/model"
	"github.com/cloud-barista/cb-ladybug/src/utils/config"
	"github.com/cloud-barista/cb-ladybug/src/utils/lang"
)

// ===== [ Constants and Variables ] =====

// ===== [ Types ] =====

// MCARService - LADYBUG 서비스 구현
type MCARService struct {
}

// ===== [ Implementations ] =====

func (s *MCARService) Validate(params map[string]string) error {
	valid := validation.Validation{}

	for key, element := range params {
		valid.Required(element, key)
	}

	if valid.HasErrors() {
		for _, err := range valid.Errors {
			return errors.New(fmt.Sprintf("[%s]%s", err.Key, err.Error()))
		}
	}
	return nil
}

func (s *MCARService) ClusterReqDef(clusterReq *model.ClusterReq) error {
	clusterReq.Config.Kubernetes.NetworkCni = lang.NVL(clusterReq.Config.Kubernetes.NetworkCni, config.NETWORKCNI_KILO)
	clusterReq.Config.Kubernetes.PodCidr = lang.NVL(clusterReq.Config.Kubernetes.PodCidr, config.POD_CIDR)
	clusterReq.Config.Kubernetes.ServiceCidr = lang.NVL(clusterReq.Config.Kubernetes.ServiceCidr, config.SERVICE_CIDR)
	clusterReq.Config.Kubernetes.ServiceDnsDomain = lang.NVL(clusterReq.Config.Kubernetes.ServiceDnsDomain, config.SERVICE_DOMAIN)

	return nil
}

func (s *MCARService) ClusterReqValidate(req model.ClusterReq) error {
	if len(req.ControlPlane) == 0 {
		return errors.New("control plane node count must be one")
	}
	if len(req.Worker) == 0 {
		return errors.New("worker node count must be at least one")
	}
	if !(req.Config.Kubernetes.NetworkCni == config.NETWORKCNI_CANAL || req.Config.Kubernetes.NetworkCni == config.NETWORKCNI_KILO) {
		return errors.New("network cni allows only Kilo or Canal")
	}

	return nil
}

func (s *MCARService) NodeReqValidate(req model.NodeReq) error {
	if len(req.ControlPlane) > 0 {
		return errors.New("control plane node not supported")
	}
	if len(req.Worker) == 0 {
		return errors.New("worker node count must be at least one")
	}

	return nil
}

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====
