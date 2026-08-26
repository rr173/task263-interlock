package store

import (
	"fmt"
	"time"

	"task263-interlock/internal/model"
)

// SnapshotStore 验证快照持久化。
type SnapshotStore struct {
	db *DB
}

// NewSnapshotStore 创建快照存储。
func NewSnapshotStore(db *DB) *SnapshotStore { return &SnapshotStore{db: db} }

const snapCols = "id,version_id,name,state,topology_hash,conflict_total,exception_count,published_at,superseded_by,created_at,updated_at"

func scanSnapshot(sc interface{ Scan(...any) error }) (*model.ValidationSnapshot, error) {
	var s model.ValidationSnapshot
	var pubRaw, created, updated string
	if err := sc.Scan(&s.ID, &s.VersionID, &s.Name, &s.State, &s.TopologyHash,
		&s.ConflictTotal, &s.ExceptionCount, &pubRaw, &s.SupersededBy, &created, &updated); err != nil {
		return nil, err
	}
	if pubRaw != "" {
		t, err := time.Parse(time.RFC3339, pubRaw)
		if err == nil {
			s.PublishedAt = &t
		}
	}
	s.CreatedAt, _ = time.Parse(time.RFC3339, created)
	s.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return &s, nil
}

// Create 插入快照。
func (st *SnapshotStore) Create(s *model.ValidationSnapshot) error {
	if err := s.Valid(); err != nil {
		return err
	}
	pubRaw := ""
	if s.PublishedAt != nil {
		pubRaw = s.PublishedAt.Format(time.RFC3339)
	}
	_, err := st.db.Exec(
		`INSERT INTO snapshots (id,version_id,name,state,topology_hash,conflict_total,exception_count,published_at,superseded_by,created_at,updated_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		s.ID, s.VersionID, s.Name, s.State, s.TopologyHash, s.ConflictTotal,
		s.ExceptionCount, pubRaw, s.SupersededBy,
		s.CreatedAt.Format(time.RFC3339), s.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrAlreadyExists
		}
		return fmt.Errorf("insert snapshot: %w", err)
	}
	return nil
}

// Get 按 ID 读取快照。
func (st *SnapshotStore) Get(id string) (*model.ValidationSnapshot, error) {
	row := st.db.QueryRow(`SELECT `+snapCols+` FROM snapshots WHERE id=?`, id)
	s, err := scanSnapshot(row)
	if err != nil {
		if err == sqliteNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return s, nil
}

// List 列出全部快照。
func (st *SnapshotStore) List() ([]*model.ValidationSnapshot, error) {
	rows, err := st.db.Query(`SELECT ` + snapCols + ` FROM snapshots ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.ValidationSnapshot
	for rows.Next() {
		s, err := scanSnapshot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ListByVersion 按版本列出快照。
func (st *SnapshotStore) ListByVersion(versionID string) ([]*model.ValidationSnapshot, error) {
	rows, err := st.db.Query(`SELECT `+snapCols+` FROM snapshots WHERE version_id=? ORDER BY created_at`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.ValidationSnapshot
	for rows.Next() {
		s, err := scanSnapshot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// UpdateState 更新快照状态（发布/替代）。
func (st *SnapshotStore) UpdateState(s *model.ValidationSnapshot) error {
	pubRaw := ""
	if s.PublishedAt != nil {
		pubRaw = s.PublishedAt.Format(time.RFC3339)
	}
	res, err := st.db.Exec(
		`UPDATE snapshots SET state=?, published_at=?, superseded_by=?, updated_at=? WHERE id=?`,
		s.State, pubRaw, s.SupersededBy, time.Now().Format(time.RFC3339), s.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}
