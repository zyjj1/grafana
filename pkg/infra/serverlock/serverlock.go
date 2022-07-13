package serverlock

import (
	"context"
	"time"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/sqlstore"
)

func ProvideService(sqlStore *sqlstore.SQLStore) *ServerLockService {
	return &ServerLockService{
		SQLStore: sqlStore,
		log:      log.New("infra.lockservice"),
	}
}

// ServerLockService allows servers in HA mode to claim a lock
// and execute an function if the server was granted the lock
type ServerLockService struct {
	SQLStore *sqlstore.SQLStore
	log      log.Logger
}

// LockAndExecute try to create a lock for this server and only executes the
// `fn` function when successful. This should not be used at low internal. But services
// that needs to be run once every ex 10m.
func (sl *ServerLockService) LockAndExecute(ctx context.Context, actionName string, maxInterval time.Duration, fn func(ctx context.Context)) error {
	// gets or creates a lockable row
	rowLock, err := sl.getOrCreate(ctx, actionName)
	if err != nil {
		return err
	}

	// avoid execution if last lock happened less than `maxInterval` ago
	if sl.isLockWithinInterval(rowLock, maxInterval) {
		return nil
	}

	// try to get lock based on rowLow version
	acquiredLock, _, err := sl.acquireLock(ctx, rowLock)
	if err != nil {
		return err
	}

	// everything is ok, so we execute the fn
	if acquiredLock {
		fn(ctx)
	}
	return nil
}

// LockExecuteAndRelease executes the same logic as LockAndExecute, but at the end releases the lock,
// so does not need to wait to it to timeout to execute again
func (sl *ServerLockService) LockExecuteAndRelease(ctx context.Context, actionName string, maxInterval time.Duration, fn func(ctx context.Context)) error {
	// gets or creates a lockable row
	rowLock, err := sl.getOrCreate(ctx, actionName)
	if err != nil {
		return err
	}

	// avoid execution if last lock happened less than `maxInterval` ago
	if sl.isLockWithinInterval(rowLock, maxInterval) {
		return nil
	}

	// try to get lock based on rowLow version
	acquiredLock, newVersion, err := sl.acquireLock(ctx, rowLock)
	if err != nil {
		return err
	}

	// everything is ok, so we execute the fn
	if acquiredLock {
		fn(ctx)
	}

	// finally we release the lock
	err = sl.releaseLock(ctx, newVersion, rowLock)
	if err != nil {
		// We will not return an error in this case, just log it
		sl.log.Error("Error releasing lock.", "err", err)
	}
	return nil
}

func (sl *ServerLockService) isLockWithinInterval(rowLock *serverLock, maxInterval time.Duration) bool {
	if rowLock.LastExecution != 0 {
		lastExecutionTime := time.Unix(rowLock.LastExecution, 0)
		if time.Since(lastExecutionTime) < maxInterval {
			return true
		}
	}
	return false
}

// acquireLock attempts to acquire a lock, and updates its version
func (sl *ServerLockService) acquireLock(ctx context.Context, serverLock *serverLock) (bool, int64, error) {
	var result bool
	newVersion := serverLock.Version + 1

	err := sl.SQLStore.WithDbSession(ctx, func(dbSession *sqlstore.DBSession) error {
		sql := `UPDATE server_lock SET
					version = ?,
					last_execution = ?
				WHERE
					id = ? AND version = ?`

		res, err := dbSession.Exec(sql, newVersion, time.Now().Unix(), serverLock.Id, serverLock.Version)
		if err != nil {
			return err
		}

		affected, err := res.RowsAffected()
		result = affected == 1

		return err
	})

	return result, newVersion, err
}

// releaseLock will delete the row at the database from
func (sl *ServerLockService) releaseLock(ctx context.Context, newVersion int64, serverLock *serverLock) error {
	err := sl.SQLStore.WithDbSession(ctx, func(dbSession *sqlstore.DBSession) error {
		sql := `DELETE FROM server_lock WHERE id=? AND version=?`

		res, err := dbSession.Exec(sql, serverLock.Id, newVersion)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if affected != 1 {
			sl.log.Debug("Error releasing lock ", "affected", affected)
		}
		return err
	})

	return err
}

func (sl *ServerLockService) getOrCreate(ctx context.Context, actionName string) (*serverLock, error) {
	var result *serverLock

	err := sl.SQLStore.WithTransactionalDbSession(ctx, func(dbSession *sqlstore.DBSession) error {
		lockRows := []*serverLock{}
		err := dbSession.Where("operation_uid = ?", actionName).Find(&lockRows)
		if err != nil {
			return err
		}

		if len(lockRows) > 0 {
			result = lockRows[0]
			return nil
		}

		lockRow := &serverLock{
			OperationUID:  actionName,
			LastExecution: 0,
		}

		_, err = dbSession.Insert(lockRow)
		if err != nil {
			return err
		}

		result = lockRow
		return nil
	})

	return result, err
}
