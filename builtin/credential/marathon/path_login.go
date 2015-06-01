package marathon

import (
	"errors"
	"fmt"
	marathon "github.com/gambol99/go-marathon"
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

	client, err := getMarathonClientFromConfig(b, req)

	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}

	appId := data.Get("marathon_app_id").(string)
	appVersion := data.Get("marathon_app_version").(string)
	taskId := data.Get("mesos_task_id").(string)

	if err != nil {
		return nil, err
	}

	appTask, err := getAppTaskFromValues(client, appId, appVersion)

	if err != nil {
		return nil, err
	}

	_, err = appTaskStartedWithinThreshold(appTask)

	if err != nil {
		return nil, err
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
			LeaseOptions: logical.LeaseOptions{
				Renewable: true,
				Lease:     time.Minute * 5,
			},
		},
	}, nil
}

func getMarathonClientFromConfig(b *backend, req *logical.Request) (*marathon.Client, error) {
	// Get all our stored state
	config, err := b.Config(req.Storage)
	if err != nil {
		return nil, err
	}

	if config.MarathonUrl == "" {
		return nil, errors.New("configure the marathon credential backend first")
	}
	return b.Client(config.MarathonUrl)
}

func getAppTaskFromValues(client *marathon.Client, appId string, appVersion string) (*marathon.Task, error) {
	// Get marathon task data
	app, err := client.Application(appId)
	if err != nil {
		return nil, err
	}

	for _, task := range app.Tasks {
		if task.Version == appVersion {
			return task, nil
		}
	}

	return nil, errors.New("App version not found")
}

func appTaskStartedWithinThreshold(appTask *marathon.Task) (bool, error) {
	startedAt, e := time.Parse(
		time.RFC3339,
		appTask.StartedAt)

	if e != nil {
		return false, errors.New(fmt.Sprintf("Failed to validate app startup time: %s", e.Error()))
	}

	delta := time.Now().Sub(startedAt)
	if delta > StartupThresholdSeconds {
		return false, errors.New("App did not startup within threshold")
	}
	return true, nil
}

func (b *backend) pathLoginRenew(
	req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	appId := req.Auth.Metadata["marathon_app_id"]
	appVersion := req.Auth.Metadata["marathon_app_version"]

	client, err := getMarathonClientFromConfig(b, req)

	if err != nil {
		return nil, err
	}

	appTask, err := getAppTaskFromValues(client, appId, appVersion)

	if err != nil {
		return nil, err
	}

	if appTask == nil {
		// not sure if this is necessary, but if appTask is nil,
		// do not renew
		return nil, nil
	}

	return framework.LeaseExtend(1*time.Hour, 0)(req, d)
}
