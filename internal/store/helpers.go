package store

import (
	"database/sql"
	"errors"
	"strings"
)

// sqliteNoRows 别名 SQL 标准无行错误。
var sqliteNoRows = sql.ErrNoRows

// isUniqueViolation 判断是否为 UNIQUE/PK 约束冲突。
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "primary key")
}

// isForeignKeyViolation 判断是否为外键约束冲突。
func isForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "foreign key")
}

// wrapErr 将 sql 错误统一映射为领域错误。
func wrapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return sqliteNoRows
	}
	if isUniqueViolation(err) {
		return errUniqueViolation
	}
	return err
}

// errUniqueViolation 唯一约束冲突的领域错误。
var errUniqueViolation = errors.New("唯一约束冲突")
