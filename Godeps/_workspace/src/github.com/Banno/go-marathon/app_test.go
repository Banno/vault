package marathon

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestAppMethods(t *testing.T) {
	registryUrl := "registry.banno-internal.com"
	marathonHostname := os.Getenv("MARATHON_HOSTNAME")

	if marathonHostname == "" {
		t.Fatal("MARATHON_HOSTNAME must be set for the acceptance tests to work.")
	}

	imageName := "small-deployable"
	imageTag := "latest"
	fullImageName := registryUrl + "/" + imageName + ":" + imageTag

	appName := "/" + imageName

	c := NewClient(marathonHostname, 8080)

	fmt.Println("REQUEST")
	fmt.Println("=======")

	appMutable := AppMutable{
		Id:   appName,
		Cpus: 0.01,
		Container: &Container{
			Docker: &Docker{
				Image: fullImageName,
			},
			Type: "DOCKER",
		},
	}

	b, _ := json.MarshalIndent(appMutable, "", "  ")
	fmt.Println(string(b))

	fmt.Println("\nRESPONSE")
	fmt.Println("========")

	// actually create the app
	app, err := c.AppCreate(appMutable)
	if err != nil {
		fmt.Println(err)
	}

	b, _ = json.MarshalIndent(app, "", "  ")

	fmt.Println(string(b))

	time.Sleep(time.Second * 5)
	fmt.Println("\nREAD")
	fmt.Println("========")

	appRead, err := c.AppRead(appName)
	if err != nil {
		t.Fatal(err)
	}

	b, _ = json.MarshalIndent(appRead, "", "  ")

	fmt.Println(string(b))

	for tries := 0; ; tries++ {
		fmt.Println("\nDEPLOYMENTS")
		fmt.Println("========")

		deployments, err := c.Deployments()
		if err != nil {
			t.Fatal(err)
		}

		b, _ = json.MarshalIndent(deployments, "", "  ")

		fmt.Println(string(b))

		if len(deployments) == 0 {
			break
		}

		// Some threshold to not wait forever
		if tries == 20 {
			break
		}

		time.Sleep(time.Second * 3)
	}

	time.Sleep(time.Second * 5)
	fmt.Println("\nUPDATE")
	fmt.Println("========")

	// actually create the app
	appMutable.Instances = 5

	appUpdate, err := c.AppUpdate(appMutable, false)
	if err != nil {
		t.Fatal(err)
	}

	b, _ = json.MarshalIndent(appUpdate, "", "  ")

	fmt.Println(string(b))

	time.Sleep(time.Second * 5)

	fmt.Println("\nREAD")
	fmt.Println("========")

	appRead, err = c.AppRead(appName)
	if err != nil {
		t.Fatal(err)
	}

	b, _ = json.MarshalIndent(appRead, "", "  ")

	fmt.Println(string(b))

	time.Sleep(time.Second)
	fmt.Println("\nVERSIONS")
	fmt.Println("========")

	appReadVersions, err := c.AppReadVersions(appName)
	if err != nil {
		t.Fatal(err)
	}

	b, _ = json.MarshalIndent(appReadVersions, "", "  ")

	fmt.Println(string(b))

	time.Sleep(time.Second)
	fmt.Println("\nAPP by VERSION")
	fmt.Println("========")

	appReadByVersion, err := c.AppReadByVersion(appName, appRead.Version)
	if err != nil {
		t.Fatal(err)
	}

	b, _ = json.MarshalIndent(appReadByVersion, "", "  ")

	fmt.Println(string(b))

	time.Sleep(time.Second)
	fmt.Println("\nAPP TASKS")
	fmt.Println("========")

	appTasks, err := c.AppReadTasks(appName)
	if err != nil {
		t.Fatal(err)
	}

	b, _ = json.MarshalIndent(appTasks, "", "  ")

	fmt.Println(string(b))

	time.Sleep(time.Second)
	fmt.Println("\nDELETE APP TASK")
	fmt.Println("========")

	appDeleteTask, err := c.AppDeleteTask(appName, appRead.Tasks[0].TaskId, false)
	if err != nil {
		t.Fatal(err)
	}

	b, _ = json.MarshalIndent(appDeleteTask, "", "  ")

	fmt.Println(string(b))

	time.Sleep(time.Second * 5)
	fmt.Println("\nAPP TASKS")
	fmt.Println("========")

	appTasks, err = c.AppReadTasks(appName)
	if err != nil {
		t.Fatal(err)
	}

	b, _ = json.MarshalIndent(appTasks, "", "  ")

	fmt.Println(string(b))

	time.Sleep(time.Second)
	fmt.Println("\nDELETE APP TASKS")
	fmt.Println("========")

	time.Sleep(time.Second * 4)
	appDeleteTasks, err := c.AppDeleteTasks(appName, "", true)
	if err != nil {
		t.Fatal(err)
	}

	b, _ = json.MarshalIndent(appDeleteTasks, "", "  ")

	fmt.Println(string(b))

	time.Sleep(time.Second * 5)
	fmt.Println("\nAPP TASKS")
	fmt.Println("========")

	appTasks, err = c.AppReadTasks(appName)
	if err != nil {
		t.Fatal(err)
	}

	b, _ = json.MarshalIndent(appTasks, "", "  ")

	fmt.Println(string(b))

	time.Sleep(time.Second)
	fmt.Println("\nREVERT VERSION")
	fmt.Println("========")

	resp, err := c.AppRollback(appName, appReadVersions[1], false)
	if err != nil {
		t.Fatal(err)
	}

	b, _ = json.MarshalIndent(resp, "", " ")

	fmt.Println(string(b))

	time.Sleep(time.Second * 15)
	fmt.Println("\nDELETE APP")
	fmt.Println("========")

	err = c.AppDelete(appName, true)
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.AppRead(appName)
	if err == nil || err != ErrAppNotFound {
		t.Fatal(fmt.Errorf("Expected app to not be found."))
	}

	fmt.Println(string(b))
}
