package store

import (
	"encoding/json"
	"fmt"
	"time"

	"task263-interlock/internal/model"
)

// SegmentStore 区段持久化。
type SegmentStore struct {
	db *DB
}

// NewSegmentStore 创建区段存储。
func NewSegmentStore(db *DB) *SegmentStore { return &SegmentStore{db: db} }

const segCols = "id,name,kind,length_m,state,line_name,created_at,updated_at"

func scanSegment(sc interface{ Scan(...any) error }) (*model.Segment, error) {
	var s model.Segment
	var created, updated string
	if err := sc.Scan(&s.ID, &s.Name, &s.Kind, &s.LengthM, &s.State, &s.LineName, &created, &updated); err != nil {
		return nil, err
	}
	s.CreatedAt, _ = time.Parse(time.RFC3339, created)
	s.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return &s, nil
}

// Create 插入区段。
func (st *SegmentStore) Create(s *model.Segment) error {
	if err := s.Valid(); err != nil {
		return err
	}
	_, err := st.db.Exec(
		`INSERT INTO segments (id,name,kind,length_m,state,line_name,created_at,updated_at)
		 VALUES (?,?,?,?,?,?,?,?)`,
		s.ID, s.Name, s.Kind, s.LengthM, s.State, s.LineName,
		s.CreatedAt.Format(time.RFC3339), s.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrAlreadyExists
		}
		return fmt.Errorf("insert segment: %w", err)
	}
	return nil
}

// Get 按 ID 读取区段。
func (st *SegmentStore) Get(id string) (*model.Segment, error) {
	row := st.db.QueryRow(`SELECT `+segCols+` FROM segments WHERE id=?`, id)
	s, err := scanSegment(row)
	if err != nil {
		if err == sqliteNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return s, nil
}

// List 列出全部区段。
func (st *SegmentStore) List() ([]*model.Segment, error) {
	rows, err := st.db.Query(`SELECT ` + segCols + ` FROM segments ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Segment
	for rows.Next() {
		s, err := scanSegment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// UpdateState 更新区段占用状态（幂等）。
func (st *SegmentStore) UpdateState(id string, state model.SegmentState) error {
	if _, ok := map[model.SegmentState]bool{
		model.SegmentClear: true, model.SegmentOccupied: true,
		model.SegmentReserved: true, model.SegmentUnknown: true,
	}[state]; !ok {
		return model.ErrBadSegmentState
	}
	res, err := st.db.Exec(`UPDATE segments SET state=?, updated_at=? WHERE id=?`,
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

// Delete 删除区段。
func (st *SegmentStore) Delete(id string) error {
	res, err := st.db.Exec(`DELETE FROM segments WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// jsonEncode 编码 JSON 字段。
func jsonEncode(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
