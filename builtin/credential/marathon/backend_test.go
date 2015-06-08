package marathon

import (
	"os"
	"testing"
	"time"

	marathon "github.com/gambol99/go-marathon"
	"github.com/hashicorp/vault/logical"
	logicaltest "github.com/hashicorp/vault/logical/testing"
)

func TestBackend_basic(t *testing.T) {
	logicaltest.Test(t, logicaltest.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Backend:  Backend(),
		Steps: []logicaltest.TestStep{
			testAccStepConfig(t),
			testAccLogin(t),
		},
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("MARATHON_URL"); v == "" {
		t.Fatal("MARATHON_URL must be set for acceptance tests")
	}
}

func testAccStepConfig(t *testing.T) logicaltest.TestStep {
	marathonUrl := os.Getenv("MARATHON_URL")

	return logicaltest.TestStep{
		Operation: logical.WriteOperation,
		Path:      "config",
		Data: map[string]interface{}{
			"marathon_url": marathonUrl,
		},
	}
}

func testAccLogin(t *testing.T) logicaltest.TestStep {
	marathonUrl := os.Getenv("MARATHON_URL")

	config := marathon.NewDefaultConfig()
	config.URL = marathonUrl
	config.LogOutput = os.Stdout
	c, err := marathon.NewClient(config)

	if err != nil {
		t.Fatal(err)
	}

	appId := "test-app"

	c.DeleteApplication(appId)

	time.Sleep(time.Second * 1)

	application := marathon.NewDockerApplication()
	application.Name(appId)
	application.Arg("sleep").Arg("10000")
	application.CPU(0.1).Memory(256).Count(1)
	application.Container.Docker.Container("alpine")

	if err := c.CreateApplication(application, true); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 6)

	appRead, err := c.Application(appId)

	if err != nil {
		t.Fatal(err)
	}

	appVersion := appRead.Tasks[0].Version
	taskId := appRead.Tasks[0].ID

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
