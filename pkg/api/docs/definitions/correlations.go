package definitions

import (
	"github.com/grafana/grafana/pkg/services/correlations"
)

// swagger:route POST /datasources/uid/{uid}/correlations correlations createCorrelation
//
// Add correlation.
//
// Responses:
// 200: createCorrelationResponse
// 400: badRequestError
// 401: unauthorisedError
// 403: forbiddenError
// 404: notFoundError
// 500: internalServerError

// swagger:parameters createCorrelation
type CreateCorrelationParams struct {
	// in:body
	// required:true
	Body correlations.CreateCorrelationCommand `json:"body"`
	// in:path
	// required:true
	SourceUID string `json:"uid"`
}

//swagger:response createCorrelationResponse
type CreateCorrelationResponse struct {
	// in: body
	Body correlations.CreateCorrelationResponse `json:"body"`
}

// swagger:route DELETE /datasources/uid/{uid}/correlations/{correlationUid} correlations deleteCorrelation
//
// Delete a correlation.
//
// Responses:
// 200: deleteCorrelationResponse
// 401: unauthorisedError
// 403: forbiddenError
// 404: notFoundError
// 500: internalServerError

// swagger:parameters deleteCorrelation
type DeleteCorrelationParams struct {
	// in:path
	// required:true
	DatasourceUID string `json:"uid"`
	// in:path
	// required:true
	CorrelationUID string `json:"correlationUid"`
}

//swagger:response deleteCorrelationResponse
type DeleteCorrelationResponse struct {
	// in: body
	Body correlations.DeleteCorrelationResponse `json:"body"`
}

// swagger:route PUT /datasources/uid/{uid}/correlations/{correlationUid} correlations updateCorrelation
//
// Updates a correlation.
//
// Responses:
// 200: updateCorrelationResponse
// 401: unauthorisedError
// 403: forbiddenError
// 404: notFoundError
// 500: internalServerError

// swagger:parameters updateCorrelation
type UpdateCorrelationParams struct {
	// in:path
	// required:true
	DatasourceUID string `json:"uid"`
	// in:path
	// required:true
	CorrelationUID string `json:"correlationUid"`
	// in: body
	Body correlations.UpdateCorrelationCommand `json:"body"`
}

//swagger:response updateCorrelationResponse
type UpdateCorrelationResponse struct {
	// in: body
	Body correlations.UpdateCorrelationResponse `json:"body"`
}

// swagger:route GET /datasources/uid/{uid}/correlations/{correlationUid} correlations getCorrelation
//
// Gets a correlation.
//
// Responses:
// 200: getCorrelationResponse
// 401: unauthorisedError
// 404: notFoundError
// 500: internalServerError

// swagger:parameters getCorrelation
type GetCorrelationParams struct {
	// in:path
	// required:true
	DatasourceUID string `json:"uid"`
	// in:path
	// required:true
	CorrelationUID string `json:"correlationUid"`
}

//swagger:response getCorrelationResponse
type GetCorrelationResponse struct {
	// in: body
	Body correlations.CorrelationDTO `json:"body"`
}

// swagger:route GET /datasources/uid/{uid}/correlations correlations getCorrelationsBySourceUID
//
// Gets all correlations originating from the given data source.
//
// Responses:
// 200: getCorrelationsBySourceUIDResponse
// 401: unauthorisedError
// 404: notFoundError
// 500: internalServerError

// swagger:parameters getCorrelationsBySourceUID
type GetCorrelationsBySourceUIDParams struct {
	// in:path
	// required:true
	DatasourceUID string `json:"uid"`
}

// swagger:route GET /datasources/uid/correlations correlations getCorrelations
//
// Gets all correlations.
//
// Responses:
// 200: getCorrelationsResponse
// 401: unauthorisedError
// 404: notFoundError
// 500: internalServerError

//swagger:response getCorrelationsResponse
type GetCorrelations struct {
	// in: body
	Body []correlations.CorrelationDTO `json:"body"`
}
