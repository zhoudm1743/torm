package migration

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/zhoudm1743/torm/db"
)

// SafeMigrator 安全的表结构迁移器
type SafeMigrator struct {
	conn         db.ConnectionInterface
	dryRun       bool
	backupTables bool
	logger       *log.Logger
}

// NewSafeMigrator 创建安全迁移器
func NewSafeMigrator(conn db.ConnectionInterface) *SafeMigrator {
	return &SafeMigrator{
		conn:         conn,
		dryRun:       false,
		backupTables: true,
		logger:       log.Default(),
	}
}

// SetDryRun 设置是否为预演模式
func (sm *SafeMigrator) SetDryRun(dryRun bool) *SafeMigrator {
	sm.dryRun = dryRun
	return sm
}

// SetBackupTables 设置是否备份表
func (sm *SafeMigrator) SetBackupTables(backup bool) *SafeMigrator {
	sm.backupTables = backup
	return sm
}

// SetLogger 设置日志器
func (sm *SafeMigrator) SetLogger(logger *log.Logger) *SafeMigrator {
	sm.logger = logger
	return sm
}

// SafeAlterTable 安全地执行表结构变更
func (sm *SafeMigrator) SafeAlterTable(tableName string, differences []ColumnDifference) (*MigrationResult, error) {
	result := &MigrationResult{
		TableName: tableName,
		StartTime: time.Now(),
		Changes:   differences,
		Success:   false,
	}

	if len(differences) == 0 {
		result.Success = true
		result.Message = "No changes needed"
		return result, nil
	}

	// 生成变更SQL
	alterGenerator := NewAlterGenerator(sm.conn)
	alterStatements, err := alterGenerator.GenerateAlterSQL(tableName, differences)
	if err != nil {
		result.Error = err
		return result, fmt.Errorf("failed to generate ALTER statements: %w", err)
	}

	result.Statements = alterStatements

	// 预演模式：只打印SQL，不执行
	if sm.dryRun {
		sm.logger.Println("🔍 DRY RUN MODE - SQL statements to be executed:")
		for i, stmt := range alterStatements {
			sm.logger.Printf("  %d. %s", i+1, stmt)
		}
		result.Success = true
		result.Message = "Dry run completed - no changes were applied"
		return result, nil
	}

	// 备份表（如果启用）
	var backupTableName string
	if sm.backupTables {
		backupTableName, err = sm.createTableBackup(tableName)
		if err != nil {
			result.Error = err
			return result, fmt.Errorf("failed to create table backup: %w", err)
		}
		result.BackupTable = backupTableName
		sm.logger.Printf("✅ Table backup created: %s", backupTableName)
	}

	// 开始事务
	tx, err := sm.conn.Begin()
	if err != nil {
		result.Error = err
		return result, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// 执行变更
	for i, statement := range alterStatements {
		// 跳过注释
		if strings.HasPrefix(strings.TrimSpace(statement), "--") {
			continue
		}

		sm.logger.Printf("Executing (%d/%d): %s", i+1, len(alterStatements), statement)

		_, err := tx.Exec(statement)
		if err != nil {
			// 回滚事务
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				sm.logger.Printf("❌ Failed to rollback transaction: %v", rollbackErr)
			}

			result.Error = err
			result.FailedStatement = statement

			// 如果有备份，提供恢复指令
			if backupTableName != "" {
				result.RecoveryInstructions = fmt.Sprintf(
					"To restore the table, run: RENAME TABLE %s TO temp_table, %s TO %s, temp_table TO %s_failed",
					tableName, backupTableName, tableName, tableName,
				)
			}

			return result, fmt.Errorf("failed to execute statement '%s': %w", statement, err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		result.Error = err
		return result, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result.Success = true
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Message = fmt.Sprintf("Successfully applied %d changes", len(differences))

	sm.logger.Printf("✅ Table structure updated successfully in %v", result.Duration)

	return result, nil
}

// createTableBackup 创建表备份
func (sm *SafeMigrator) createTableBackup(tableName string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	backupTableName := fmt.Sprintf("%s_backup_%s", tableName, timestamp)

	driver := sm.conn.GetDriver()
	var backupSQL string

	switch driver {
	case "mysql":
		backupSQL = fmt.Sprintf("CREATE TABLE %s LIKE %s", backupTableName, tableName)
	case "postgres", "postgresql":
		backupSQL = fmt.Sprintf("CREATE TABLE %s AS TABLE %s", backupTableName, tableName)
	case "sqlite", "sqlite3":
		backupSQL = fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM %s", backupTableName, tableName)
	default:
		return "", fmt.Errorf("unsupported database driver for backup: %s", driver)
	}

	// 创建表结构
	_, err := sm.conn.Exec(backupSQL)
	if err != nil {
		return "", fmt.Errorf("failed to create backup table structure: %w", err)
	}

	// 复制数据（MySQL需要单独执行）
	if driver == "mysql" {
		copySQL := fmt.Sprintf("INSERT INTO %s SELECT * FROM %s", backupTableName, tableName)
		_, err = sm.conn.Exec(copySQL)
		if err != nil {
			// 清理创建的备份表
			sm.conn.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", backupTableName))
			return "", fmt.Errorf("failed to copy data to backup table: %w", err)
		}
	}

	return backupTableName, nil
}

// RestoreFromBackup 从备份恢复表
func (sm *SafeMigrator) RestoreFromBackup(tableName, backupTableName string) error {
	driver := sm.conn.GetDriver()

	// 开始事务
	tx, err := sm.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// 删除当前表
	_, err = tx.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to drop current table: %w", err)
	}

	// 恢复备份
	var restoreSQL string
	switch driver {
	case "mysql", "postgres", "postgresql", "sqlite", "sqlite3":
		restoreSQL = fmt.Sprintf("RENAME TABLE %s TO %s", backupTableName, tableName)
		if driver == "postgres" || driver == "postgresql" {
			restoreSQL = fmt.Sprintf("ALTER TABLE %s RENAME TO %s", backupTableName, tableName)
		}
	default:
		tx.Rollback()
		return fmt.Errorf("unsupported database driver for restore: %s", driver)
	}

	_, err = tx.Exec(restoreSQL)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to restore table: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit restore transaction: %w", err)
	}

	sm.logger.Printf("✅ Table %s restored from backup %s", tableName, backupTableName)
	return nil
}

// CleanupBackups 清理旧的备份表
func (sm *SafeMigrator) CleanupBackups(tablePrefix string, keepDays int) error {
	driver := sm.conn.GetDriver()
	var query string

	switch driver {
	case "mysql":
		query = `
			SELECT table_name 
			FROM information_schema.tables 
			WHERE table_schema = DATABASE() 
			AND table_name LIKE ?
			AND table_name REGEXP '_backup_[0-9]{8}_[0-9]{6}$'
		`
	case "postgres", "postgresql":
		query = `
			SELECT tablename 
			FROM pg_tables 
			WHERE schemaname = 'public' 
			AND tablename LIKE $1
			AND tablename ~ '_backup_[0-9]{8}_[0-9]{6}$'
		`
	case "sqlite", "sqlite3":
		query = `
			SELECT name 
			FROM sqlite_master 
			WHERE type='table' 
			AND name LIKE ?
			AND name GLOB '*_backup_[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]_[0-9][0-9][0-9][0-9][0-9][0-9]'
		`
	default:
		return fmt.Errorf("unsupported database driver for cleanup: %s", driver)
	}

	pattern := tablePrefix + "_backup_%"
	rows, err := sm.conn.Query(query, pattern)
	if err != nil {
		return fmt.Errorf("failed to query backup tables: %w", err)
	}
	defer rows.Close()

	cutoffTime := time.Now().AddDate(0, 0, -keepDays)
	var tablesToDrop []string

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}

		// 解析时间戳
		if tableTime := sm.extractTableTimestamp(tableName); tableTime != nil {
			if tableTime.Before(cutoffTime) {
				tablesToDrop = append(tablesToDrop, tableName)
			}
		}
	}

	// 删除过期的备份表
	for _, tableName := range tablesToDrop {
		dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
		_, err := sm.conn.Exec(dropSQL)
		if err != nil {
			sm.logger.Printf("⚠️ Failed to drop backup table %s: %v", tableName, err)
		} else {
			sm.logger.Printf("🗑️ Cleaned up backup table: %s", tableName)
		}
	}

	if len(tablesToDrop) > 0 {
		sm.logger.Printf("✅ Cleaned up %d old backup tables", len(tablesToDrop))
	}

	return nil
}

// extractTableTimestamp 从表名中提取时间戳
func (sm *SafeMigrator) extractTableTimestamp(tableName string) *time.Time {
	parts := strings.Split(tableName, "_backup_")
	if len(parts) != 2 {
		return nil
	}

	timestamp := parts[1]
	if len(timestamp) != 15 { // YYYYMMDD_HHMMSS = 15 characters
		return nil
	}

	timeStr := timestamp[:8] + timestamp[9:] // Remove underscore
	if t, err := time.Parse("20060102150405", timeStr); err == nil {
		return &t
	}

	return nil
}

// MigrationResult 迁移结果
type MigrationResult struct {
	TableName            string
	StartTime            time.Time
	EndTime              time.Time
	Duration             time.Duration
	Changes              []ColumnDifference
	Statements           []string
	Success              bool
	Error                error
	Message              string
	BackupTable          string
	FailedStatement      string
	RecoveryInstructions string
}

// PrintSummary 打印迁移摘要
func (mr *MigrationResult) PrintSummary() {
	fmt.Printf("\n🎯 Migration Summary for table: %s\n", mr.TableName)
	fmt.Printf("⏱️ Duration: %v\n", mr.Duration)
	fmt.Printf("📝 Changes: %d\n", len(mr.Changes))

	if mr.Success {
		fmt.Printf("✅ Status: SUCCESS\n")
		fmt.Printf("💬 Message: %s\n", mr.Message)
	} else {
		fmt.Printf("❌ Status: FAILED\n")
		if mr.Error != nil {
			fmt.Printf("💬 Error: %s\n", mr.Error.Error())
		}
		if mr.FailedStatement != "" {
			fmt.Printf("💔 Failed Statement: %s\n", mr.FailedStatement)
		}
	}

	if mr.BackupTable != "" {
		fmt.Printf("💾 Backup Table: %s\n", mr.BackupTable)
	}

	if mr.RecoveryInstructions != "" {
		fmt.Printf("🔧 Recovery: %s\n", mr.RecoveryInstructions)
	}

	fmt.Println()
}
