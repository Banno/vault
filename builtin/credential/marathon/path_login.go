package marathon

import (
	"errors"
	"github.com/banno/go-marathon"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"strings"
	"time"
)

const (
	StartupThresholdSeconds = time.Second * 5
)

func pathLogin(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "login",
		Fields: map[string]*framework.FieldSchema{
			"marathon_app_id": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "MARATHON_APP_ID env var from a Marathon task",
			},
			"marathon_app_version": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "MARATHON_APP_VERSION env var from a Marathon task",
			},
			"mesos_task_id": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "MESOS_TASK env var from a Marathon task",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.WriteOperation: b.pathLogin,
		},
	}
}

func (b *backend) pathLogin(
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	// Get all our stored state
	config, err := b.Config(req.Storage)
	if err != nil {
		return nil, err
	}

	if config.MarathonUrl == "" {
		return logical.ErrorResponse(
			"configure the marathon credential backend first"), nil
	}

	appId := data.Get("marathon_app_id").(string)
	appVersion := data.Get("marathon_app_version").(string)
	taskId := data.Get("mesos_task_id").(string)

	client, err := b.Client(config.MarathonUrl)
	if err != nil {
		return nil, err
	}

	// Get marathon task data
	app, err := client.AppRead(appId)
	if err != nil {
		return nil, err
	}

	found := false
	for _, appTask := range app.Tasks {
		if appTask.Version == appVersion {
			found = true

			_, err := appTaskStartedWithinThreshold(&appTask)
			if err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("App version not found")
	}

	policyName := strings.TrimPrefix(appId, "/")

	return &logical.Response{
		Auth: &logical.Auth{
			Policies: []string{policyName},
			Metadata: map[string]string{
				"marathon_app_id":      appId,
				"marathon_app_version": appVersion,
				"mesos_task_id":        taskId,
			},
			DisplayName: appId,
		},
	}, nil
}

func appTaskStartedWithinThreshold(appTask *marathon.AppTask) (bool, error) {
	delta := time.Now().Sub(appTask.StartedAt)
	if delta > StartupThresholdSeconds {
		return false, errors.New("App did not startup within threshold")
	}
	return true, nil
}
