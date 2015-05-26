package marathon

import (
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"

	"github.com/banno/go-marathon"
)

func Factory(map[string]string) (logical.Backend, error) {
	return Backend(), nil
}

func Backend() *framework.Backend {
	var b backend
	b.Backend = &framework.Backend{
		Help: backendHelp,

		Paths: append([]*framework.Path{
			pathConfig(),
			pathLogin(&b),
		}),

		PathsSpecial: &logical.Paths{
			Root: []string{
				"config",
			},
			Unauthenticated: []string{
				"login",
			},
		},

		AuthRenew: b.pathLoginRenew,
	}

	return b.Backend
}

type backend struct {
	*framework.Backend
}

// Client returns the Marathon client to communicate to Marathon
func (b *backend) Client(marathonUrl string) (*marathon.Client, error) {
	client := marathon.NewClientForUrl(marathonUrl)

	return client, nil
}

const backendHelp = `
The Marathon credential provider allows task authentication via Marathon.

Tasks provide a marathon_app_id, marathon_app_version and mesos_task_id
and the credential provider can authenticate the task with the Marathon
API.
`
