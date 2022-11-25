package provision

func awsGetMetadataLocalHostname(m *Machine) (string, error) {
	return m.executeSSH("curl http://169.254.169.254/latest/meta-data/local-hostname")
}

func ncpvpcGetMetadataServerName(m *Machine) (string, error) {
	return m.executeSSH("curl http://169.254.169.254/latest/meta-data/serverName")
}
