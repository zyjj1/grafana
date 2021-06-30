package dashboardpanels

import (
	"encoding/json"

	"github.com/grafana/grafana/pkg/components/simplejson"
)

type MySQLPanelParser struct {
	Format string `json:"format"`
	RefID  string `json:"refId"`
	RawSQL string `json:"rawSql"`
}

func (m MySQLPanelParser) ParsePanel(datasourceID int64, panels []*simplejson.Json) ([]*simplejson.Json, error) {

	mappedData := make([]*simplejson.Json, len(panels))

	for i, p := range panels {
		panel, err := p.MarshalJSON()
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(panel, &m)
		if err != nil {
			return nil, err
		}

		mappedData[i] = simplejson.NewFromAny(map[string]interface{}{
			"format":        m.Format,
			"refId":         m.RefID,
			"rawSql":        m.RawSQL,
			"intervalMs":    60000,
			"maxDataPoints": 400,
			"datasourceId":  datasourceID,
		})
	}

	return mappedData, nil
}
