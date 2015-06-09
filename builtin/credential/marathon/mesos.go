package marathon

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type Task struct {
	Id      string `json:"id"`
	State   string `json:"state"`
	SlaveId string `json:"slave_id"`
}

type Framework struct {
	Tasks []Task `json:"tasks"`
}

type MesosState struct {
	Frameworks []Framework `json:"frameworks"`
}

func SlaveTaskIdIsValid(mesosUrl string, slaveTaskId string) (bool, error) {
	res, _ := http.Get(mesosUrl + "/state.json")
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var data MesosState
	_ = json.Unmarshal(body, &data)

	var taskSlaveId string
	for _, framework := range data.Frameworks {
		for _, task := range framework.Tasks {
			if task.Id == slaveTaskId && task.State == "TASK_RUNNING" {
				taskSlaveId = task.SlaveId
				break
			}
		}
		if taskSlaveId != "" {
			break
		}
	}

	if taskSlaveId == "" {
		return false, errors.New("Slave Task ID not found!")
	}

	return true, nil
}

// func main() {
// 	ip, _ := GetHostNameForSlaveTaskId("http://dev.banno.com:5050", "smaller-deployable.79b291e7-0bb5-11e5-a6dd-0242ac1100c2")

// 	fmt.Println(ip)
// }
