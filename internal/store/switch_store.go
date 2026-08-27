package store

import (
	"encoding/json"
	"fmt"
	"time"

	"task263-interlock/internal/model"
)

// SwitchStore 道岔持久化。
type SwitchStore struct {
	db *DB
}

// NewSwitchStore 创建道岔存储。
func NewSwitchStore(db *DB) *SwitchStore { return &SwitchStore{db: db} }

const swCols = "id,name,position,normal_to,reverse_to,line_name,created_at,updated_at"

func scanSwitch(sc interface{ Scan(...any) error }) (*model.Switch, error) {
	var s model.Switch
	var created, updated string
	if err := sc.Scan(&s.ID, &s.Name, &s.Position, &s.NormalTo, &s.ReverseTo, &s.LineName, &created, &updated); err != nil {
		return nil, err
	}
	s.CreatedAt, _ = time.Parse(time.RFC3339, created)
	s.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return &s, nil
}

// Create 插入道岔。
func (st *SwitchStore) Create(s *model.Switch) error {
	if err := s.Valid(); err != nil {
		return err
	}
	_, err := st.db.Exec(
		`INSERT INTO switches (id,name,position,normal_to,reverse_to,line_name,created_at,updated_at)
		 VALUES (?,?,?,?,?,?,?,?)`,
		s.ID, s.Name, s.Position, s.NormalTo, s.ReverseTo, s.LineName,
		s.CreatedAt.Format(time.RFC3339), s.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrAlreadyExists
		}
		return fmt.Errorf("insert switch: %w", err)
	}
	return nil
}

// Get 按 ID 读取道岔。
func (st *SwitchStore) Get(id string) (*model.Switch, error) {
	row := st.db.QueryRow(`SELECT `+swCols+` FROM switches WHERE id=?`, id)
	s, err := scanSwitch(row)
	if err != nil {
		if err == sqliteNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return s, nil
}

// List 列出全部道岔。
func (st *SwitchStore) List() ([]*model.Switch, error) {
	rows, err := st.db.Query(`SELECT ` + swCols + ` FROM switches ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Switch
	for rows.Next() {
		s, err := scanSwitch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// UpdatePosition 更新道岔位置。
func (st *SwitchStore) UpdatePosition(id string, pos model.SwitchPosition) error {
	if pos != model.SwitchNormal && pos != model.SwitchReverse {
		return model.ErrUnknownSwitchPos
	}
	res, err := st.db.Exec(`UPDATE switches SET position=?, updated_at=? WHERE id=?`,
		pos, time.Now().Format(time.RFC3339), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// Delete 删除道岔。
func (st *SwitchStore) Delete(id string) error {
	res, err := st.db.Exec(`DELETE FROM switches WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// decodeSwitchReqs 反序列化道岔要求列表。
func decodeSwitchReqs(raw string) ([]model.SwitchRequirement, error) {
	var out []model.SwitchRequirement
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return out, nil
}
