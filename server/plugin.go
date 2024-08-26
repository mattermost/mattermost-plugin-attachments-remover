package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"

	"github.com/mattermost/mattermost-plugin-attachments-remover/server/sqlstore"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	// api is the instance of the API struct that handles the plugin's API endpoints.
	api *API

	// sqlStore is the instance of the SQLStore struct that handles the plugin's database interactions.
	SQLStore *sqlstore.SQLStore
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.api.ServeHTTP(w, r)
}

func (p *Plugin) OnActivate() error {
	var err error
	p.api, err = setupAPI(p)
	if err != nil {
		return fmt.Errorf("error setting up the API: %w", err)
	}

	// Setup direct SQL Store access via the plugin api
	papi := pluginapi.NewClient(p.API, p.Driver)
	SQLStore, err := sqlstore.New(papi.Store, &papi.Log)
	if err != nil {
		p.API.LogError("cannot create SQLStore", "err", err)
		return err
	}
	p.SQLStore = SQLStore

	// Register command
	if err := p.API.RegisterCommand(p.getCommand()); err != nil {
		p.API.LogError("failed to register command", "err", err)
		return fmt.Errorf("failed to register command: %w", err)
	}

	return nil
}
