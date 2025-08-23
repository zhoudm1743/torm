package db

import (
	"context"
	"database/sql"
	"fmt"
)

// DBTransaction 事务实现
type DBTransaction struct {
	tx  *sql.Tx
	ctx context.Context
}

// NewTransaction 创建新事务
func NewTransaction(conn ConnectionInterface) (*DBTransaction, error) {
	if conn == nil {
		return nil, fmt.Errorf("连接不能为空")
	}

	db := conn.GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}

	return &DBTransaction{
		tx:  tx,
		ctx: context.Background(),
	}, nil
}

// Query 执行查询
func (t *DBTransaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(t.ctx, query, args...)
}

// QueryRow 执行查询单行
func (t *DBTransaction) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(t.ctx, query, args...)
}

// Exec 执行语句
func (t *DBTransaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(t.ctx, query, args...)
}

// Commit 提交事务
func (t *DBTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback 回滚事务
func (t *DBTransaction) Rollback() error {
	return t.tx.Rollback()
}

// Transaction 便捷的事务执行函数
func Transaction(fn func(tx TransactionInterface) error, connectionName ...string) error {
	connName := "default"
	if len(connectionName) > 0 {
		connName = connectionName[0]
	}

	conn, err := DefaultManager().Connection(connName)
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	tx, err := NewTransaction(conn)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// BeginTransaction 开始事务
func BeginTransaction(connectionName ...string) (*DBTransaction, error) {
	connName := "default"
	if len(connectionName) > 0 {
		connName = connectionName[0]
	}

	conn, err := DefaultManager().Connection(connName)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	return NewTransaction(conn)
}
