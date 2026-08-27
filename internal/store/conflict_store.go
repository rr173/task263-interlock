package store

import (
	"encoding/json"
	"fmt"
	"time"

	"task263-interlock/internal/model"
)

// ConflictStore 冲突持久化。
type ConflictStore struct {
	db *DB
}

// NewConflictStore 创建冲突存储。
func NewConflictStore(db *DB) *ConflictStore { return &ConflictStore{db: db} }

const conflictCols = "id,version_id,kind,state,route_a,route_b,object_id,detail,steps,created_at,updated_at"

func scanConflict(sc interface{ Scan(...any) error }) (*model.Conflict, error) {
	var c model.Conflict
	var stepsRaw, created, updated string
	if err := sc.Scan(&c.ID, &c.VersionID, &c.Kind, &c.State, &c.RouteA, &c.RouteB,
		&c.ObjectID, &c.Detail, &stepsRaw, &created, &updated); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(stepsRaw), &c.Steps)
	c.CreatedAt, _ = time.Parse(time.RFC3339, created)
	c.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return &c, nil
}

// Create 插入冲突。
func (st *ConflictStore) Create(c *model.Conflict) error {
	if err := c.Valid(); err != nil {
		return err
	}
	_, err := st.db.Exec(
		`INSERT INTO conflicts (id,version_id,kind,state,route_a,route_b,object_id,detail,steps,created_at,updated_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		c.ID, c.VersionID, c.Kind, c.State, c.RouteA, c.RouteB, c.ObjectID, c.Detail,
		jsonEncode(c.Steps), c.CreatedAt.Format(time.RFC3339), c.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrAlreadyExists
		}
		return fmt.Errorf("insert conflict: %w", err)
	}
	return nil
}

// Get 按 ID 读取冲突。
func (st *ConflictStore) Get(id string) (*model.Conflict, error) {
	row := st.db.QueryRow(`SELECT `+conflictCols+` FROM conflicts WHERE id=?`, id)
	c, err := scanConflict(row)
	if err != nil {
		if err == sqliteNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return c, nil
}

// ListByVersion 按版本列出冲突（shared_segment 由服务层决定是否落库）。
func (st *ConflictStore) ListByVersion(versionID string) ([]*model.Conflict, error) {
	rows, err := st.db.Query(`SELECT `+conflictCols+` FROM conflicts WHERE version_id=? ORDER BY created_at`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Conflict
	for rows.Next() {
		c, err := scanConflict(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// UpdateState 更新冲突状态。
func (st *ConflictStore) UpdateState(id string, state model.ConflictState) error {
	res, err := st.db.Exec(`UPDATE conflicts SET state=?, updated_at=? WHERE id=?`,
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

// DeleteByVersion 清理某版本的全部冲突（重新验证前调用）。
func (st *ConflictStore) DeleteByVersion(versionID string) error {
	_, err := st.db.Exec(`DELETE FROM conflicts WHERE version_id=?`, versionID)
	return err
}
