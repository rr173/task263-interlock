package store

import (
	"encoding/json"
	"fmt"
	"time"

	"task263-interlock/internal/model"
)

// VersionStore 联锁版本持久化。
type VersionStore struct {
	db *DB
}

// NewVersionStore 创建版本存储。
func NewVersionStore(db *DB) *VersionStore { return &VersionStore{db: db} }

const verCols = "id,name,state,segment_ids,switch_ids,route_ids,exception_ids,conflict_count,last_validated_at,created_at,updated_at"

func scanVersion(sc interface{ Scan(...any) error }) (*model.InterlockingVersion, error) {
	var v model.InterlockingVersion
	var segRaw, swRaw, rtRaw, excRaw, lastRaw, created, updated string
	var last *time.Time
	if err := sc.Scan(&v.ID, &v.Name, &v.State, &segRaw, &swRaw, &rtRaw, &excRaw,
		&v.ConflictCount, &lastRaw, &created, &updated); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(segRaw), &v.SegmentIDs)
	_ = json.Unmarshal([]byte(swRaw), &v.SwitchIDs)
	_ = json.Unmarshal([]byte(rtRaw), &v.RouteIDs)
	_ = json.Unmarshal([]byte(excRaw), &v.ExceptionIDs)
	if lastRaw != "" {
		t, err := time.Parse(time.RFC3339, lastRaw)
		if err == nil {
			last = &t
		}
	}
	v.LastValidatedAt = last
	v.CreatedAt, _ = time.Parse(time.RFC3339, created)
	v.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return &v, nil
}

// Create 插入版本。
func (st *VersionStore) Create(v *model.InterlockingVersion) error {
	if err := v.Valid(); err != nil {
		return err
	}
	lastRaw := ""
	if v.LastValidatedAt != nil {
		lastRaw = v.LastValidatedAt.Format(time.RFC3339)
	}
	_, err := st.db.Exec(
		`INSERT INTO versions (id,name,state,segment_ids,switch_ids,route_ids,exception_ids,conflict_count,last_validated_at,created_at,updated_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		v.ID, v.Name, v.State, jsonEncode(v.SegmentIDs), jsonEncode(v.SwitchIDs),
		jsonEncode(v.RouteIDs), jsonEncode(v.ExceptionIDs), v.ConflictCount, lastRaw,
		v.CreatedAt.Format(time.RFC3339), v.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrAlreadyExists
		}
		return fmt.Errorf("insert version: %w", err)
	}
	return nil
}

// Get 按 ID 读取版本。
func (st *VersionStore) Get(id string) (*model.InterlockingVersion, error) {
	row := st.db.QueryRow(`SELECT `+verCols+` FROM versions WHERE id=?`, id)
	v, err := scanVersion(row)
	if err != nil {
		if err == sqliteNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return v, nil
}

// List 列出全部版本。
func (st *VersionStore) List() ([]*model.InterlockingVersion, error) {
	rows, err := st.db.Query(`SELECT ` + verCols + ` FROM versions ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.InterlockingVersion
	for rows.Next() {
		v, err := scanVersion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// UpdateState 更新版本状态并维护冲突数/验证时间。
func (st *VersionStore) UpdateState(v *model.InterlockingVersion) error {
	lastRaw := ""
	if v.LastValidatedAt != nil {
		lastRaw = v.LastValidatedAt.Format(time.RFC3339)
	}
	res, err := st.db.Exec(
		`UPDATE versions SET state=?, route_ids=?, conflict_count=?, last_validated_at=?, exception_ids=?, updated_at=? WHERE id=?`,
		v.State, jsonEncode(v.RouteIDs), v.ConflictCount, lastRaw, jsonEncode(v.ExceptionIDs),
		time.Now().Format(time.RFC3339), v.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// AppendExceptionID 在版本例外列表追加 ID（不触发冲突压制）。
func (st *VersionStore) AppendExceptionID(versionID, excID string) error {
	v, err := st.Get(versionID)
	if err != nil {
		return err
	}
	for _, e := range v.ExceptionIDs {
		if e == excID {
			return nil // 幂等
		}
	}
	v.ExceptionIDs = append(v.ExceptionIDs, excID)
	return st.UpdateState(v)
}
