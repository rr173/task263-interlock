package store

import (
	"fmt"
	"time"

	"task263-interlock/internal/model"
)

// ExceptionStore 例外裁决持久化。
type ExceptionStore struct {
	db *DB
}

// NewExceptionStore 创建例外存储。
func NewExceptionStore(db *DB) *ExceptionStore { return &ExceptionStore{db: db} }

const excCols = "id,version_id,conflict_id,state,reason,owner,created_at,updated_at"

func scanException(sc interface{ Scan(...any) error }) (*model.Exception, error) {
	var e model.Exception
	var created, updated string
	if err := sc.Scan(&e.ID, &e.VersionID, &e.ConflictID, &e.State, &e.Reason, &e.Owner,
		&created, &updated); err != nil {
		return nil, err
	}
	e.CreatedAt, _ = time.Parse(time.RFC3339, created)
	e.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return &e, nil
}

// Create 插入例外。
func (st *ExceptionStore) Create(e *model.Exception) error {
	if err := e.Valid(); err != nil {
		return err
	}
	_, err := st.db.Exec(
		`INSERT INTO exceptions (id,version_id,conflict_id,state,reason,owner,created_at,updated_at)
		 VALUES (?,?,?,?,?,?,?,?)`,
		e.ID, e.VersionID, e.ConflictID, e.State, e.Reason, e.Owner,
		e.CreatedAt.Format(time.RFC3339), e.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrAlreadyExists
		}
		return fmt.Errorf("insert exception: %w", err)
	}
	return nil
}

// Get 按 ID 读取例外。
func (st *ExceptionStore) Get(id string) (*model.Exception, error) {
	row := st.db.QueryRow(`SELECT `+excCols+` FROM exceptions WHERE id=?`, id)
	e, err := scanException(row)
	if err != nil {
		if err == sqliteNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return e, nil
}

// ListByVersion 按版本列出例外。
func (st *ExceptionStore) ListByVersion(versionID string) ([]*model.Exception, error) {
	rows, err := st.db.Query(`SELECT `+excCols+` FROM exceptions WHERE version_id=? ORDER BY created_at`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Exception
	for rows.Next() {
		e, err := scanException(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// UpdateState 更新例外状态。
func (st *ExceptionStore) UpdateState(id string, state model.ExceptionState) error {
	res, err := st.db.Exec(`UPDATE exceptions SET state=?, updated_at=? WHERE id=?`,
		state, time.Now().Format(time.RFC3339), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}
