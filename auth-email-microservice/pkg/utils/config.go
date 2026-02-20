package utils

func GetConfigPath(configPath string) string {
	if configPath == "docker" {
		return "./config/config-docker.yaml"
	}

	// Default to docker config if no config path is specified
	if configPath == "" {
		return "./config/config-docker.yaml"
	}

	return "./config/config-local.yaml"
}
