package correlations

import (
	"errors"
)

var (
	ErrSourceDataSourceReadOnly           = errors.New("source data source is read only")
	ErrSourceDataSourceDoesNotExists      = errors.New("source data source does not exist")
	ErrTargetDataSourceDoesNotExists      = errors.New("target data source does not exist")
	ErrCorrelationFailedGenerateUniqueUid = errors.New("failed to generate unique correlation UID")
	ErrCorrelationIdentifierNotSet        = errors.New("source identifier and org id are needed to be able to edit correlations")
	ErrCorrelationNotFound                = errors.New("correlation not found")
)

// Correlation is the model for correlations definitions
type Correlation struct {
	ID          int64  `xorm:"pk autoincr 'id'"`
	UID         string `xorm:"uid"`
	SourceUID   string `xorm:"source_uid"`
	TargetUID   string `xorm:"target_uid"`
	Label       string `xorm:"label"`
	Description string `xorm:"description"`
}

type CorrelationDTO struct {
	// Unique identifier of the correlation
	// example: 50xhMlg9k
	UID string `json:"uid" xorm:"uid"`
	// UID of the data source the correlation originates from
	// example:d0oxYRg4z
	SourceUID string `json:"sourceUid"`
	// UID of the data source the correlation points to
	// example:PE1C5CBDA0504A6A3
	TargetUID string `json:"targetUid"`
	// Label identifying the correlation
	// example: My Label
	Label string `json:"label"`
	// Description of the correlation
	// example: Logs to Traces
	Description string `json:"description"`
}

// CreateCorrelationResponse is a response struct for CorrelationDTO
// swagger:model
type CreateCorrelationResponse struct {
	Result CorrelationDTO `json:"result"`
	// example: Correlation created
	Message string `json:"message"`
}

// CreateCorrelationCommand is the command for creating a correlation
// swagger:model
type CreateCorrelationCommand struct {
	// UID of the data source for which correlation is created.
	SourceUID         string `json:"-"`
	OrgId             int64  `json:"-"`
	SkipReadOnlyCheck bool   `json:"-"`
	// Target data source UID to which the correlation is created
	// example:PE1C5CBDA0504A6A3
	TargetUID string `json:"targetUid" binding:"Required"`
	// Optional label identifying the correlation
	// example: My label
	Label string `json:"label"`
	// Optional description of the correlation
	// example: Logs to Traces
	Description string `json:"description"`
}

// swagger:model
type DeleteCorrelationResponse struct {
	// example: Correlation deleted
	Message string `json:"message"`
}

// DeleteCorrelationCommand is the command for deleting a correlation
type DeleteCorrelationCommand struct {
	// UID of the correlation to be deleted.
	UID       string
	SourceUID string
	OrgId     int64
}

// swagger:model
type UpdateCorrelationResponse struct {
	Result CorrelationDTO `json:"result"`
	// example: Correlation updated
	Message string `json:"message"`
}

// UpdateCorrelationCommand is the command for updating a correlation
type UpdateCorrelationCommand struct {
	// UID of the correlation to be deleted.
	UID       string `json:"-"`
	SourceUID string `json:"-"`
	OrgId     int64  `json:"-"`

	// Optional label identifying the correlation
	// example: My label
	Label string `json:"label"`
	// Optional description of the correlation
	// example: Logs to Traces
	Description string `json:"description"`
}

// GetCorrelationQuery is the query to retrieve a single correlation
type GetCorrelationQuery struct {
	// UID of the correlation
	UID string `json:"-"`
	// UID of the source data source
	SourceUID string `json:"-"`
	OrgId     int64  `json:"-"`
}

// GetCorrelationsBySourceUIDQuery is the query to retrieve all correlations originating by the given Data Source
type GetCorrelationsBySourceUIDQuery struct {
	SourceUID string `json:"-"`
	OrgId     int64  `json:"-"`
}

type DeleteCorrelationsBySourceUIDCommand struct {
	SourceUID string
}

type DeleteCorrelationsByTargetUIDCommand struct {
	TargetUID string
}
