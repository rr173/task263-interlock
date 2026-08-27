package store

import (
	"encoding/json"
	"fmt"
	"time"

	"task263-interlock/internal/model"
)

// RouteStore 进路持久化。
type RouteStore struct {
	db *DB
}

// NewRouteStore 创建进路存储。
func NewRouteStore(db *DB) *RouteStore { return &RouteStore{db: db} }

const routeCols = "id,name,origin_seg,dest_seg,path_segs,switches,release,state,version_id,created_at,updated_at"

func scanRoute(sc interface{ Scan(...any) error }) (*model.Route, error) {
	var r model.Route
	var pathRaw, swRaw, relRaw, created, updated string
	if err := sc.Scan(&r.ID, &r.Name, &r.OriginSeg, &r.DestSeg, &pathRaw, &swRaw, &relRaw,
		&r.State, &r.VersionID, &created, &updated); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(pathRaw), &r.PathSegs); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(swRaw), &r.Switches); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(relRaw), &r.Release); err != nil {
		return nil, err
	}
	r.CreatedAt, _ = time.Parse(time.RFC3339, created)
	r.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return &r, nil
}

// Create 插入进路。
func (st *RouteStore) Create(r *model.Route) error {
	if err := r.Valid(); err != nil {
		return err
	}
	_, err := st.db.Exec(
		`INSERT INTO routes (id,name,origin_seg,dest_seg,path_segs,switches,release,state,version_id,created_at,updated_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		r.ID, r.Name, r.OriginSeg, r.DestSeg,
		jsonEncode(r.PathSegs), jsonEncode(r.Switches), jsonEncode(r.Release),
		r.State, r.VersionID, r.CreatedAt.Format(time.RFC3339), r.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrAlreadyExists
		}
		return fmt.Errorf("insert route: %w", err)
	}
	return nil
}

// Get 按 ID 读取进路。
func (st *RouteStore) Get(id string) (*model.Route, error) {
	row := st.db.QueryRow(`SELECT `+routeCols+` FROM routes WHERE id=?`, id)
	r, err := scanRoute(row)
	if err != nil {
		if err == sqliteNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return r, nil
}

// List 列出全部进路。
func (st *RouteStore) List() ([]*model.Route, error) {
	rows, err := st.db.Query(`SELECT ` + routeCols + ` FROM routes ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Route
	for rows.Next() {
		r, err := scanRoute(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListByVersion 按版本列出进路。
func (st *RouteStore) ListByVersion(versionID string) ([]*model.Route, error) {
	rows, err := st.db.Query(`SELECT `+routeCols+` FROM routes WHERE version_id=? ORDER BY id`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Route
	for rows.Next() {
		r, err := scanRoute(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// UpdateState 更新进路状态。
func (st *RouteStore) UpdateState(id string, state model.RouteState) error {
	res, err := st.db.Exec(`UPDATE routes SET state=?, updated_at=? WHERE id=?`,
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

// UpdateVersion 更新进路归属的联锁版本（幂等）。
func (st *RouteStore) UpdateVersion(id, versionID string) error {
	res, err := st.db.Exec(`UPDATE routes SET version_id=?, updated_at=? WHERE id=?`,
		versionID, time.Now().Format(time.RFC3339), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}
