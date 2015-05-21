package marathon

import (
	"os"
	"testing"
	"time"

	"github.com/Banno/go-marathon"
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

	c := marathon.NewClientForUrl(marathonUrl)

	appId := "test-app"

	c.AppDelete(appId, true)

	time.Sleep(time.Second * 1)

	appMutable := marathon.AppMutable{
		Id:   appId,
		Cpus: 0.01,
		Mem:  256,
		Container: &marathon.Container{
			Docker: &marathon.Docker{
				Image: "registry.banno-internal.com/small-deployable:latest",
			},
			Type: "DOCKER",
		},
	}

	_, err := c.AppCreate(appMutable)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 6)

	appRead, err := c.AppRead(appId)

	if err != nil {
		t.Fatal(err)
	}

	appVersion := appRead.Tasks[0].Version
	taskId := appRead.Tasks[0].TaskId

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
