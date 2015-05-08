package marathon

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type App struct {
	AppMutable
	Deployments     []Deployment    `json:"deployments,omitempty"`
	Executor        string          `json:"executor,omitempty"`
	LastTaskFailure *AppTaskFailure `json:"lastTaskFailure,omitempty"`
	StoreUrls       []string        `json:"storeUrls,omitempty"`
	Tasks           []AppTask       `json:"tasks,omitempty"`
	TasksRunning    int             `json:"tasksRunning,omitempty"`
	TasksStaged     int             `json:"tasksStaged,omitempty"`
	User            string          `json:"user,omitempty"`
	Version         string          `json:"version,omitempty"`
}

type AppMutable struct {
	Id              string            `json:"id,omitempty"`
	Args            []string          `json:"args,omitempty"`
	BackoffSeconds  float64           `json:"backoffSeconds,omitempty"`
	BackoffFactor   float64           `json:"backoffFactor,omitempty"`
	Cmd             string            `json:"cmd,omitempty"`
	Constraints     [][]string        `json:"constraints,omitempty"`
	Container       *Container        `json:"container,omitempty"`
	Cpus            float64           `json:"cpus,omitempty"`
	Dependencies    []string          `json:"dependencies,omitempty"`
	Disk            float64           `json:"disk,omitempty"`
	Env             map[string]string `json:"env,omitempty"`
	HealthChecks    []HealthCheck     `json:"healthChecks,omitempty"`
	Instances       int               `json:"instances,omitempty"`
	Mem             float64           `json:"mem,omitempty"`
	Ports           []int             `json:"ports"`
	RequirePorts    bool              `json:"requirePorts,omitempty"`
	UpgradeStrategy *UpgradeStrategy  `json:"upgradeStrategy,omitempty"`
	Uris            []string          `json:"uris,omitempty"`
}

type AppTask struct {
	AppId     string    `json:"appId,omitempty"`
	TaskId    string    `json:"id,omitempty"`
	Host      string    `json:"host,omitempty"`
	Ports     []int     `json:"ports,omitempty"`
	StagedAt  string    `json:"stagedAt,omitempty"`
	StartedAt time.Time `json:"startedAt,omitempty"`
	Version   string    `json:"version,omitempty"`
}

type AppTaskFailure struct {
	AppId     string
	Host      string
	Message   string
	State     string
	TaskId    string
	Timestamp string
	Version   string
}

type AppUpdateResponse struct {
	DeploymentId string `json:"deploymentId,omitempty"`
	Version      string `json:"version,omitempty"`

	Deployments []Deployment `json:"deployments,omitempty"`
	Message     string       `json:"message,omitempty"`
}

type Container struct {
	Docker  *Docker  `json:"docker,omitempty"`
	Volumes []Volume `json:"volumes,omitempty"`
	Type    string   `json:"type,omitempty"`
}

type Deployment struct {
	Id             string             `json:"id,omitempty"`
	AffectedApps   []string           `json:affectedApps,omitempty`
	Steps          [][]DeploymentStep `json:steps,omitempty`
	CurrentActions []DeploymentStep   `json:currentActions,omitempty`
	Version        string             `json:version,omitempty`
	CurrentStep    int                `json:currentStep,omitempty`
	TotalSteps     int                `json:totalSteps,omitempty`
}

type DeploymentStep struct {
	Action string `json:"action,omitempty"`
	App    string `json:"app,omitempty"`
}

type Docker struct {
	Image        string        `json:"image,omitempty"`
	Network      string        `json:"network,omitempty"`
	Privileged   bool          `json:"privileged,omitempty"`
	PortMappings []PortMapping `json:"portMappings,omitempty"`
}

type HealthCheck struct {
	Protocol               string              `json:"protocol,omitempty"`
	Path                   string              `json:"path,omitempty"`
	GracePeriodSeconds     int                 `json:"gracePeriodSeconds,omitempty"`
	IntervalSeconds        int                 `json:"intervalSeconds,omitempty"`
	PortIndex              int                 `json:"portIndex,omitempty"`
	TimeoutSeconds         int                 `json:"timeoutSeconds,omitempty"`
	MaxConsecutiveFailures int                 `json:"maxConsecutiveFailures,omitempty"`
	Command                map[string]string   `json:"command,omitempty"`
	HealthCheckResults     []HealthCheckResult `json:"healthCheckResults,omitempty"`
}

type HealthCheckResult struct {
	Alive               bool
	ConsecutiveFailures float64
	FirstSuccess        string
	LastFailure         string
	LastSuccess         string
	TaskId              string
}

type PortMapping struct {
	ContainerPort int    `json:"containerPort,omitempty"`
	HostPort      int    `json:"hostPort,omitempty"`
	ServicePort   int    `json:"servicePort,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
}

type Volume struct {
	ContainerPath string `json:"containerPath,omitempty"`
	HostPath      string `json:"hostPath,omitempty"`
	Mode          string `json:"mode,omitempty"`
}

type UpgradeStrategy struct {
	MinimumHealthCapacity float64 `json:"minimumHealthCapacity,omitempty"`
	MaximumOverCapicity   float64 `json:"maximumOverhCapacity,omitempty"`
}

var (
	ErrAppNotFound = errors.New("marathon: App not found")
)

func (c *Client) AppCreate(appCreate AppMutable) (*App, error) {
	//log.Println(appCreate.Id)
	log.Printf("== CREATE REQUEST ==\n%#v\n", appCreate)

	appRead, err := c.AppRead(appCreate.Id)
	if err == nil {
		log.Printf("== App READ ==\n%#v\n", appRead)

		if appRead.Id == appCreate.Id {
			return nil, fmt.Errorf("App Already Exists: %s", appCreate.Id)
		}
	}

	b, _ := json.MarshalIndent(appCreate, "", "  ")

	log.Printf("== CREATE REQUEST JSON ==\n%#v\n", string(b))

	resp, err := c.postJson("/v2/apps", b)
	if err != nil {
		return nil, err
	}

	// debug log.
	log.Printf("== CREATE RESPONSE JSON ==\n%#v\n", string(resp))

	var app App
	if err := json.Unmarshal(resp, &app); err != nil {
		return nil, err
	}

	log.Printf("== CREATE RESPONSE ==\n%#v\n", app)

	return &app, nil
}

/*
// implement later
func (c *Client) Apps() ([]App, error) {

}
*/

func (c *Client) AppRead(id string) (*App, error) {
	resp, err := c.getJson("/v2/apps/" + id)
	if err != nil {
		if ce, ok := err.(ClientError); ok && ce.HttpStatusCode == 404 {
			return nil, ErrAppNotFound
		}
		return nil, err
	}

	// debug log
	//log.Println(string(resp))

	var a struct {
		App     App    `json:"app,omitempty"`
		Message string `json:"message,omitempty"`
	}

	if err := json.Unmarshal(resp, &a); err != nil {
		return nil, err
	}

	//log.Printf("unmarshall: %#v\nmessage: %#v", a, a.Message)

	if a.Message != "" {
		return nil, fmt.Errorf("AppRead: %#v\n", a.Message)
	}

	return &a.App, nil
}

func (c *Client) AppReadVersions(id string) ([]string, error) {
	resp, err := c.getJson("/v2/apps/" + id + "/versions")
	if err != nil {
		return nil, err
	}

	// debug log
	// log.Println(string(resp))

	var a struct {
		Versions []string `json:"versions"`
	}

	if err := json.Unmarshal(resp, &a); err != nil {
		return nil, err
	}

	return a.Versions, nil
}

func (c *Client) AppReadByVersion(id, version string) (*App, error) {
	resp, err := c.getJson("/v2/apps/" + id + "/versions/" + version)
	if err != nil {
		return nil, err
	}

	// debug log
	//fmt.Println(string(resp))

	var app App
	if err := json.Unmarshal(resp, &app); err != nil {
		return nil, err
	}

	return &app, nil
}

func (c *Client) AppUpdate(appUpdate AppMutable, force bool) (*AppUpdateResponse, error) {
	if appRead, _ := c.AppRead(appUpdate.Id); appRead.Id != appUpdate.Id {
		return nil, fmt.Errorf("App Doesn't Exist: %s", appUpdate.Id)
	}

	url := "/v2/apps/" + appUpdate.Id

	if force != false {
		url = fmt.Sprintf("%s?force=%t", url, force)
	}

	// debug log
	fmt.Printf("%+v\n", appUpdate)

	b, _ := json.MarshalIndent(appUpdate, "", "  ")

	//fmt.Println(string(b))

	resp, err := c.putJson(url, b)
	if err != nil {
		return nil, err
	}

	// debug log.
	//fmt.Println(string(resp))

	var appUpdateResponse AppUpdateResponse
	if err := json.Unmarshal(resp, &appUpdateResponse); err != nil {
		return nil, err
	}

	// error out if the app is locked by a deployment

	return &appUpdateResponse, nil
}

func (c *Client) AppRollback(appId string, version string, force bool) (*AppUpdateResponse, error) {
	if appRead, _ := c.AppRead(appId); appRead.Id != appId {
		return nil, fmt.Errorf("App Doesn't Exist: %s", appId)
	}

	appVersions, _ := c.AppReadVersions(appId)

	validVersion := false
	for _, a := range appVersions {
		if a == version {
			validVersion = true
			break
		}
	}
	if validVersion == false {
		return nil, fmt.Errorf("Version Doesn't Exist: %s", version)
	}

	url := "/v2/apps/" + appId
	if force != false {
		url = fmt.Sprintf("%s?force=%t", url, force)
	}

	app := &App{Version: version}

	b, _ := json.MarshalIndent(app, "", "  ")

	//fmt.Println(string(b))

	resp, err := c.putJson(url, b)
	if err != nil {
		return nil, err
	}

	// debug log.
	//fmt.Println(string(resp))

	var appUpdateResponse AppUpdateResponse
	if err := json.Unmarshal(resp, &appUpdateResponse); err != nil {
		return nil, err
	}

	// also check to see if the deployment is locked and err properly

	return &appUpdateResponse, nil
}

func (c *Client) AppDelete(id string, force bool) error {
	url := c.getFullUrl("/v2/apps/" + id)

	if force != false {
		url = fmt.Sprintf("%s?force=%t", url, force)
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("App Deletion Error: %s", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) AppReadTasks(id string) ([]AppTask, error) {
	resp, err := c.getJson("/v2/apps/" + id + "/tasks")
	if err != nil {
		return nil, err
	}

	// debug log
	//fmt.Println(string(resp))

	var a struct {
		Tasks []AppTask `json:"tasks"`
	}

	if err := json.Unmarshal(resp, &a); err != nil {
		return nil, err
	}

	return a.Tasks, nil
}

func (c *Client) AppDeleteTask(appId string, taskId string, scale bool) (AppTask, error) {
	url := c.getFullUrl("/v2/apps/" + appId + "/tasks/" + taskId)

	if scale != false {
		url = fmt.Sprintf("%s?scale=%t", url, scale)
	}

	//fmt.Println(url)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return AppTask{}, fmt.Errorf("AppTask Delete error: %s | %s | %s", appId, taskId, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var a struct {
		Task AppTask `json:"task"`
	}

	if err := json.Unmarshal(body, &a); err != nil {
		return AppTask{}, err
	}

	return a.Task, nil
}

func (c *Client) AppDeleteTasks(id string, host string, scale bool) ([]AppTask, error) {

	// consider switching to the url package
	url := c.getFullUrl("/v2/apps/" + id + "/tasks")
	switch {
	case host != "" && scale != false:
		url = fmt.Sprintf("%s?host=%s&scale=%t", url, host, scale)
	case host == "" && scale != false:
		url = fmt.Sprintf("%s?scale=%t", url, scale)
	case host != "" && scale == false:
		url = fmt.Sprintf("%s?host=%s", url, host)
	}

	// consider refactoring to a delete function in client.go
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("AppTasks Deletion Error: %s", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var a struct {
		Tasks []AppTask `json:"tasks"`
	}

	if err := json.Unmarshal(body, &a); err != nil {
		return nil, err
	}

	return a.Tasks, nil
}

func (c *Client) Deployments() ([]Deployment, error) {
	resp, err := c.getJson("/v2/deployments")
	if err != nil {
		return nil, err
	}

	var deployments []Deployment

	if err := json.Unmarshal(resp, &deployments); err != nil {
		return nil, err
	}

	return deployments, nil
}
