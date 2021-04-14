package config

type CSP string

const (
	CONTROL_PLANE = "control-plane"
	WORKER        = "worker"

	BOOTSTRAP_FILE               = "bootstrap.sh"
	INIT_FILE                    = "k8s-init.sh"
	LADYBUG_BOOTSTRAP_CANAL_FILE = "ladybug-bootstrap-canal.sh"
	LADYBUG_BOOTSTRAP_KILO_FILE  = "ladybug-bootstrap-kilo.sh"
	SYSTEMD_SERVICE_FILE         = "systemd-service.sh"

	GCP_IMAGE_ID = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-1804-bionic-v20201014"

	CSP_AWS CSP = "aws"
	CSP_GCP CSP = "gcp"

	NETWORKCNI_KILO  = "kilo"
	NETWORKCNI_CANAL = "canal"

	POD_CIDR       = "10.244.0.0/16"
	SERVICE_CIDR   = "10.96.0.0/12"
	SERVICE_DOMAIN = "cluster.local"
)
