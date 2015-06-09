package marathon

import (
	"errors"
	"os"
	"testing"
	"time"

	marathon "github.com/gambol99/go-marathon"
	"github.com/hashicorp/vault/logical"
	logicaltest "github.com/hashicorp/vault/logical/testing"
)

func TestBackend_login(t *testing.T) {
	logicaltest.Test(t, logicaltest.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Backend:  Backend(),
		Steps: []logicaltest.TestStep{
			testAccStepConfig(t),
			testAccLogin(t),
		},
	})
}

func TestBackend_invalid(t *testing.T) {
	logicaltest.Test(t, logicaltest.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Backend:  Backend(),
		Steps: []logicaltest.TestStep{
			testAccStepConfig(t),
			testAccLoginInvalid(t),
		},
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("MARATHON_URL"); v == "" {
		t.Fatal("MARATHON_URL must be set for acceptance tests")
	}
	if v := os.Getenv("MESOS_URL"); v == "" {
		t.Fatal("MARATHON_URL must be set for acceptance tests")
	}
}

func testAccStepConfig(t *testing.T) logicaltest.TestStep {
	marathonUrl := os.Getenv("MARATHON_URL")
	mesosUrl := os.Getenv("MESOS_URL")

	return logicaltest.TestStep{
		Operation: logical.WriteOperation,
		Path:      "config",
		Data: map[string]interface{}{
			"marathon_url": marathonUrl,
			"mesos_url":    mesosUrl,
		},
	}
}

func testAccLogin(t *testing.T) logicaltest.TestStep {
	marathonUrl := os.Getenv("MARATHON_URL")

	c, err := marathonClient(marathonUrl)
	if err != nil {
		t.Fatal(err)
	}

	appId := "test-app"
	task, err := startTestTask(c, appId)

	if err != nil {
		t.Fatal(err)
	}

	appVersion := task.Version
	taskId := task.ID

	return logicaltest.TestStep{
		Operation: logical.WriteOperation,
		Path:      "login",
		Data: map[string]interface{}{
			"marathon_app_id":      appId,
			"marathon_app_version": appVersion,
			"mesos_task_id":        taskId,
		},
		Unauthenticated: true,

		Check: logicaltest.TestCheckAuth([]string{appId}),
	}
}

func testAccLoginInvalid(t *testing.T) logicaltest.TestStep {
	marathonUrl := os.Getenv("MARATHON_URL")

	c, err := marathonClient(marathonUrl)
	if err != nil {
		t.Fatal(err)
	}

	appId := "test-app"
	task, err := startTestTask(c, appId)

	if err != nil {
		t.Fatal(err)
	}

	appVersion := task.Version
	taskId := task.ID

	time.Sleep(time.Second * 1)

	c.DeleteApplication(appId)

	time.Sleep(time.Second * 1)

	return logicaltest.TestStep{
		Operation: logical.WriteOperation,
		Path:      "login",
		Data: map[string]interface{}{
			"marathon_app_id":      appId,
			"marathon_app_version": appVersion,
			"mesos_task_id":        taskId,
		},
		ErrorOk:         true,
		Unauthenticated: true,

		Check: logicaltest.TestCheckError(),
	}
}

func marathonClient(marathonUrl string) (marathon.Marathon, error) {
	config := marathon.NewDefaultConfig()
	config.URL = marathonUrl
	config.LogOutput = os.Stdout
	c, err := marathon.NewClient(config)

	return c, err
}

func startTestTask(c marathon.Marathon, appId string) (*marathon.Task, error) {
	c.DeleteApplication(appId)

	time.Sleep(time.Second * 1)

	application := marathon.NewDockerApplication()
	application.Name(appId)
	application.Arg("sleep").Arg("10000")
	application.CPU(0.1).Memory(256).Count(1)
	application.Container.Docker.Container("alpine")

	err := c.CreateApplication(application, true)

	if err != nil {
		return nil, err
	}

	time.Sleep(time.Second * 6)

	appRead, err := c.Application(appId)

	if err != nil {
		return nil, err
	}

	if len(appRead.Tasks) == 0 {
		return nil, errors.New("Test App failed to start")
	}

	return appRead.Tasks[0], nil
}
