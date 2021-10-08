package orgs

import (
	"context"

	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/sqlstore"
)

type Service interface {
	GetSignedInUserOrgList(c *models.ReqContext) response.Response
	GetUserOrgList(c *models.ReqContext) response.Response
}

type OSSService struct {
	bus      bus.Bus
	sqlStore *sqlstore.SQLStore
}

func ProvideOrgService(sqlStore *sqlstore.SQLStore) *OSSService {
	return &OSSService{sqlStore: sqlStore}
}

func (s *OSSService) GetSignedInUserOrgList(c *models.ReqContext) response.Response {
	return s.getUserOrgsList(c.Req.Context(), c.UserId)
}

func (s *OSSService) GetUserOrgList(c *models.ReqContext) response.Response {
	return s.getUserOrgsList(c.Req.Context(), c.ParamsInt64(":id"))
}

func (s *OSSService) getUserOrgsList(ctx context.Context, userID int64) response.Response {
	query := models.GetUserOrgListQuery{UserId: userID}
	if err := s.sqlStore.GetUserOrgList(ctx, &query); err != nil {
		return response.Error(500, "Failed to get user organizations", err)
	}

	return response.JSON(200, query.Result)
}
