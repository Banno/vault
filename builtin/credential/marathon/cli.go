package marathon

import (
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

type CLIHandler struct{}

func (h *CLIHandler) Auth(c *api.Client, m map[string]string) (string, error) {
	mount, ok := m["mount"]
	if !ok {
		mount = "marathon"
	}

	appId, ok := m["marathon_app_id"]
	if !ok {
		return "", fmt.Errorf("'marathon_app_id' var must be set")
	}
	appVersion, ok := m["marathon_app_version"]
	if !ok {
		return "", fmt.Errorf("'marathon_app_version' var must be set")
	}
	taskId, ok := m["mesos_task_id"]
	if !ok {
		return "", fmt.Errorf("'mesos_task_id' var must be set")
	}

	path := fmt.Sprintf("auth/%s/login", mount)
	secret, err := c.Logical().Write(path, map[string]interface{}{
		"marathon_app_id":      appId,
		"marathon_app_version": appVersion,
		"mesos_task_id":        taskId,
	})
	if err != nil {
		return "", err
	}
	if secret == nil {
		return "", fmt.Errorf("empty response from credential provider")
	}

	return secret.Auth.ClientToken, nil
}

func (h *CLIHandler) Help() string {
	help := `
The Marathon credential provider allows you to authenticate via Marathon tasks.
To use it, specify "marathon_app_id", "marathon_app_version" and "mesos_task_id.

    Example: vault auth -method=marathon marathon_app_id=<app_id> marathon_app_version=<app_version> mesos_task_id=<task_id>

	`

	return strings.TrimSpace(help)
}
