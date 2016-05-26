package marathon

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	marathon "github.com/gambol99/go-marathon"
	"github.com/hashicorp/vault/logical"
	logicaltest "github.com/hashicorp/vault/logical/testing"
)

const AppID = "test-app"

func buildBackend(t *testing.T) logical.Backend {
	defaultLeaseTTLVal := time.Hour * 24
	maxLeaseTTLVal := time.Hour * 24 * 30
	b, err := Factory(&logical.BackendConfig{
		Logger: nil,
		System: &logical.StaticSystemView{
			DefaultLeaseTTLVal: defaultLeaseTTLVal,
			MaxLeaseTTLVal:     maxLeaseTTLVal,
		},
	})
	if err != nil {
		t.Fatalf("Unable to create backend: %s", err)
	}

	return b
}

func TestBackend_login(t *testing.T) {
	logicaltest.Test(t, logicaltest.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Backend:  buildBackend(t),
		Steps: []logicaltest.TestStep{
			testAccStepConfig(t, nil),
			testAccLogin(t, AppID),
		},
		Teardown: func() error {
			return tearDown(AppID)
		},
	})
}

func TestBackend_invalid(t *testing.T) {
	logicaltest.Test(t, logicaltest.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Backend:  buildBackend(t),
		Steps: []logicaltest.TestStep{
			testAccStepConfig(t, nil),
			testAccLoginInvalid(t, AppID),
		},
		Teardown: func() error {
			return tearDown(AppID)
		},
	})
}

func TestBackend_multiple_marathon(t *testing.T) {

	// setup http mock server
	fakeMarathon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"version":"0.28.0","git_sha":"961edbd82e691a619a4c171a7aadc9c32957fa73","git_tag":"0.28.0","build_date":"2016-03-17 17:45:11","build_time":1458236711.0,"build_user":"root","start_time":1462460944.44985,"id":"0b1472ec-7dad-4bcf-9214-683760553cea","pid":"master@10.10.3.91:5050","hostname":"mesos-master1-aws.uat.banno-internal.com","activated_slaves":0.0,"deactivated_slaves":0.0,"cluster":"uat","leader":"master@10.10.1.226:5050","log_dir":"\/var\/log\/mesos-master","flags":{"allocation_interval":"500ms","allocator":"HierarchicalDRF","authenticate":"false","authenticate_http":"false","authenticate_slaves":"false","authenticators":"crammd5","authorizers":"local","cluster":"uat","framework_sorter":"drf","help":"false","hostname":"mesos-master1-aws.uat.banno-internal.com","hostname_lookup":"true","http_authenticators":"basic","initialize_driver_logging":"true","ip":"10.10.3.91","log_auto_initialize":"true","log_dir":"\/var\/log\/mesos-master","logbufsecs":"0","logging_level":"INFO","max_completed_frameworks":"50","max_completed_tasks_per_framework":"1000","max_slave_ping_timeouts":"5","offer_timeout":"1secs","port":"5050","quiet":"false","quorum":"1","recovery_slave_removal_limit":"100%","registry":"replicated_log","registry_fetch_timeout":"1mins","registry_store_timeout":"20secs","registry_strict":"false","roles":"daemon","root_submissions":"true","slave_ping_timeout":"15secs","slave_reregister_timeout":"10mins","user_sorter":"drf","version":"false","webui_dir":"\/usr\/share\/mesos\/webui","weights":"daemon=2","work_dir":"\/tmp\/mesos-master","zk":"zk:\/\/zookeeper0-aws.uat.banno-internal.com:2181,zookeeper1-aws.uat.banno-internal.com:2181,zookeeper2-aws.uat.banno-internal.com:2181\/mesos","zk_session_timeout":"10secs"},"slaves":[],"frameworks":[],"completed_frameworks":[],"orphan_tasks":[],"unregistered_frameworks":[]}`)
		ioutil.WriteFile("/tmp/marathon.test", []byte("done"), 0644)
	}))
	defer fakeMarathon.Close()

	deadMarathonURL := "http://127.0.0.1:9090"

	logicaltest.Test(t, logicaltest.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Backend:  buildBackend(t),
		Steps: []logicaltest.TestStep{
			testAccStepConfig(t, &deadMarathonURL),
			testAccLogin(t, AppID),
		},
		Teardown: func() error {
			return tearDown(AppID)
		},
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("MARATHON_URL"); v == "" {
		t.Fatal("MARATHON_URL must be set for acceptance tests")
	}
	if v := os.Getenv("MESOS_URL"); v == "" {
		t.Fatal("MESOS_URL must be set for acceptance tests")
	}
}

func testAccStepConfig(t *testing.T, nonLeaderURL *string) logicaltest.TestStep {
	mesosURLs := []string{os.Getenv("MARATHON_URL")}

	if nonLeaderURL != nil {
		mesosURLs = append([]string{*nonLeaderURL}, mesosURLs...)
	}

	marathonURL := strings.Join(mesosURLs, ",")
	mesosURL := os.Getenv("MESOS_URL")

	return logicaltest.TestStep{
		Operation: logical.UpdateOperation,
		Path:      "config",
		Data: map[string]interface{}{
			"marathon_url": marathonURL,
			"mesos_url":    mesosURL,
		},
	}
}

func testAccLogin(t *testing.T, appID string) logicaltest.TestStep {
	marathonURL := os.Getenv("MARATHON_URL")

	c, err := marathonClient(marathonURL)
	if err != nil {
		t.Fatal(err)
	}

	task, err := startTestTask(c, appID)

	if err != nil {
		t.Fatal(err)
	}

	appVersion := task.Version
	taskID := task.ID

	return logicaltest.TestStep{
		Operation: logical.UpdateOperation,
		Path:      "login",
		Data: map[string]interface{}{
			"marathon_app_id":      appID,
			"marathon_app_version": appVersion,
			"mesos_task_id":        taskID,
		},
		Unauthenticated: true,

		Check: logicaltest.TestCheckAuth([]string{"default", appID}),
	}
}

func testAccLoginInvalid(t *testing.T, appID string) logicaltest.TestStep {
	marathonURL := os.Getenv("MARATHON_URL")

	c, err := marathonClient(marathonURL)
	if err != nil {
		t.Fatal(err)
	}

	task, err := startTestTask(c, appID)

	if err != nil {
		t.Fatal(err)
	}

	appVersion := task.Version
	taskID := task.ID

	time.Sleep(time.Second * 5)

	stopTestTask(c, appID)

	return logicaltest.TestStep{
		Operation: logical.UpdateOperation,
		Path:      "login",
		Data: map[string]interface{}{
			"marathon_app_id":      appID,
			"marathon_app_version": appVersion,
			"mesos_task_id":        taskID,
		},
		ErrorOk:         true,
		Unauthenticated: true,

		Check: logicaltest.TestCheckError(),
	}
}

func tearDown(appID string) error {
	marathonURL := os.Getenv("MARATHON_URL")

	c, err := marathonClient(marathonURL)
	if err != nil {
		return err
	}

	return stopTestTask(c, appID)
}

func marathonClient(marathonURL string) (marathon.Marathon, error) {
	config := marathon.NewDefaultConfig()
	config.URL = marathonURL
	config.LogOutput = os.Stdout
	c, err := marathon.NewClient(config)

	return c, err
}

func startTestTask(c marathon.Marathon, appID string) (*marathon.Task, error) {
	stopTestTask(c, appID)

	time.Sleep(time.Second * 2)

	application := marathon.NewDockerApplication()
	application.Name(appID)
	application.AddArgs("sleep").AddArgs("10000")
	application.CPU(0.1).Memory(256).Count(1)
	application.Container.Docker.Container("alpine")

	app, err := c.CreateApplication(application)
	if err != nil {
		return nil, err
	}

	err = c.WaitOnApplication(app.ID, time.Second*30)
	if err != nil {
		return nil, err
	}

	appRead, err := c.Application(app.ID)
	if err != nil {
		return nil, err
	}

	if len(appRead.Tasks) == 0 {
		return nil, errors.New("Test App failed to start")
	}

	return appRead.Tasks[0], nil
}

func stopTestTask(c marathon.Marathon, appID string) error {

	deploymentID, err := c.DeleteApplication(appID)
	if err != nil {
		// app is already deleted
		return nil
	}

	err = c.WaitOnDeployment(deploymentID.DeploymentID, time.Second*30)

	if err != nil {
		return err
	}

	return nil
}
