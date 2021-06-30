package dashboardpanels

import (
	"fmt"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
)

type DatasourceQuery interface {
	ParsePanel(int64, []*simplejson.Json) ([]*simplejson.Json, error)
}

type DashboardPanel struct {
	Panels []Panel `json:"panels"`
}

type Panel struct {
	ID         int64              `json:"id"`
	Datasource string             `json:"datasource"`
	Targets    []*simplejson.Json `json:"targets"`
	Type       string             `json:"type"`
}

var metricQueries = map[string]DatasourceQuery{
	"MySQL": MySQLPanelParser{},
}

func (panels DashboardPanel) GetDatasourcePanelQueries(datasource *models.DataSource) ([]*simplejson.Json, error) {
	for _, p := range panels.Panels {
		if datasource.Name == p.Datasource {
			if dsPanel, ok := metricQueries[p.Datasource]; ok {
				return dsPanel.ParsePanel(datasource.Id, p.Targets)
			} else {
				return nil, ErrorPanelNotImplemented{datasource: datasource.Name}
			}
		}
	}

	return nil, ErrorPanelNotFound{datasource: datasource.Name}
}

type ErrorPanelNotFound struct {
	datasource string
}

func (e ErrorPanelNotFound) Error() string {
	return fmt.Sprintf("Panel not found for datasource: %s", e.datasource)
}

type ErrorPanelNotImplemented struct {
	datasource string
}

func (e ErrorPanelNotImplemented) Error() string {
	return fmt.Sprintf("Panel for %s is not implemented", e.datasource)
}
