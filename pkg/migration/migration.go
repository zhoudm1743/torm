package migration

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"torm/pkg/db"
)

// MigrationInterface 迁移接口
type MigrationInterface interface {
	// Up 执行迁移
	Up(conn db.ConnectionInterface) error
	// Down 回滚迁移
	Down(conn db.ConnectionInterface) error
	// Version 获取迁移版本号
	Version() string
	// Description 获取迁移描述
	Description() string
}

// Migration 基础迁移结构
type Migration struct {
	version     string
	description string
	upFunc      func(conn db.ConnectionInterface) error
	downFunc    func(conn db.ConnectionInterface) error
}

// NewMigration 创建新迁移
func NewMigration(version, description string,
	upFunc, downFunc func(conn db.ConnectionInterface) error) *Migration {
	return &Migration{
		version:     version,
		description: description,
		upFunc:      upFunc,
		downFunc:    downFunc,
	}
}

// Up 执行迁移
func (m *Migration) Up(conn db.ConnectionInterface) error {
	if m.upFunc == nil {
		return fmt.Errorf("up function not defined for migration %s", m.version)
	}
	return m.upFunc(conn)
}

// Down 回滚迁移
func (m *Migration) Down(conn db.ConnectionInterface) error {
	if m.downFunc == nil {
		return fmt.Errorf("down function not defined for migration %s", m.version)
	}
	return m.downFunc(conn)
}

// Version 获取迁移版本号
func (m *Migration) Version() string {
	return m.version
}

// Description 获取迁移描述
func (m *Migration) Description() string {
	return m.description
}

// MigrationRecord 迁移记录
type MigrationRecord struct {
	ID          int64     `json:"id"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	AppliedAt   time.Time `json:"applied_at"`
	Batch       int       `json:"batch"`
}

// Migrator 迁移器
type Migrator struct {
	conn       db.ConnectionInterface
	migrations []MigrationInterface
	tableName  string
	logger     db.LoggerInterface
	autoCreate bool
}

// NewMigrator 创建新的迁移器
func NewMigrator(conn db.ConnectionInterface, logger db.LoggerInterface) *Migrator {
	return &Migrator{
		conn:       conn,
		migrations: make([]MigrationInterface, 0),
		tableName:  "migrations",
		logger:     logger,
		autoCreate: true,
	}
}

// SetTableName 设置迁移表名
func (m *Migrator) SetTableName(tableName string) *Migrator {
	m.tableName = tableName
	return m
}

// SetAutoCreate 设置是否自动创建迁移表
func (m *Migrator) SetAutoCreate(autoCreate bool) *Migrator {
	m.autoCreate = autoCreate
	return m
}

// Register 注册迁移
func (m *Migrator) Register(migration MigrationInterface) *Migrator {
	m.migrations = append(m.migrations, migration)
	return m
}

// RegisterFunc 注册函数式迁移
func (m *Migrator) RegisterFunc(version, description string,
	upFunc, downFunc func(conn db.ConnectionInterface) error) *Migrator {
	migration := NewMigration(version, description, upFunc, downFunc)
	return m.Register(migration)
}

// ensureMigrationTable 确保迁移表存在
func (m *Migrator) ensureMigrationTable() error {
	if !m.autoCreate {
		return nil
	}

	var createTableSQL string
	switch m.conn.GetDriver() {
	case "mysql":
		createTableSQL = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			version VARCHAR(255) NOT NULL UNIQUE,
			description TEXT,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			batch INT NOT NULL,
			KEY idx_version (version),
			KEY idx_batch (batch)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`, m.tableName)

	case "postgres", "postgresql":
		createTableSQL = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGSERIAL PRIMARY KEY,
			version VARCHAR(255) NOT NULL UNIQUE,
			description TEXT,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			batch INTEGER NOT NULL
		)`, m.tableName)

		// 创建索引
		indexSQL := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS idx_%s_version ON %s(version);
		CREATE INDEX IF NOT EXISTS idx_%s_batch ON %s(batch);
		`, m.tableName, m.tableName, m.tableName, m.tableName)

		if _, err := m.conn.Exec(indexSQL); err != nil {
			m.logError("Failed to create indexes", err)
		}

	case "sqlite", "sqlite3":
		createTableSQL = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version TEXT NOT NULL UNIQUE,
			description TEXT,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			batch INTEGER NOT NULL
		)`, m.tableName)

		// 创建索引
		indexSQL := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS idx_%s_version ON %s(version);
		CREATE INDEX IF NOT EXISTS idx_%s_batch ON %s(batch);
		`, m.tableName, m.tableName, m.tableName, m.tableName)

		if _, err := m.conn.Exec(indexSQL); err != nil {
			m.logError("Failed to create indexes", err)
		}

	case "sqlserver", "mssql":
		createTableSQL = fmt.Sprintf(`
		IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='%s' AND xtype='U')
		CREATE TABLE %s (
			id BIGINT IDENTITY(1,1) PRIMARY KEY,
			version NVARCHAR(255) NOT NULL UNIQUE,
			description NTEXT,
			applied_at DATETIME2 DEFAULT GETDATE(),
			batch INT NOT NULL
		)`, m.tableName, m.tableName)

	default:
		return fmt.Errorf("unsupported database driver: %s", m.conn.GetDriver())
	}

	_, err := m.conn.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	m.logInfo("Migration table ensured", "table", m.tableName)
	return nil
}

// getAppliedMigrations 获取已应用的迁移
func (m *Migrator) getAppliedMigrations() (map[string]*MigrationRecord, error) {
	applied := make(map[string]*MigrationRecord)

	query := fmt.Sprintf("SELECT id, version, description, applied_at, batch FROM %s ORDER BY id", m.tableName)
	rows, err := m.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		record := &MigrationRecord{}
		err := rows.Scan(&record.ID, &record.Version, &record.Description, &record.AppliedAt, &record.Batch)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration record: %w", err)
		}
		applied[record.Version] = record
	}

	return applied, nil
}

// getNextBatch 获取下一个批次号
func (m *Migrator) getNextBatch() (int, error) {
	query := fmt.Sprintf("SELECT COALESCE(MAX(batch), 0) + 1 FROM %s", m.tableName)
	row := m.conn.QueryRow(query)

	var nextBatch int
	err := row.Scan(&nextBatch)
	if err != nil {
		return 0, fmt.Errorf("failed to get next batch: %w", err)
	}

	return nextBatch, nil
}

// recordMigration 记录迁移
func (m *Migrator) recordMigration(migration MigrationInterface, batch int) error {
	query := fmt.Sprintf("INSERT INTO %s (version, description, batch) VALUES (?, ?, ?)", m.tableName)
	_, err := m.conn.Exec(query, migration.Version(), migration.Description(), batch)
	if err != nil {
		return fmt.Errorf("failed to record migration %s: %w", migration.Version(), err)
	}
	return nil
}

// removeMigrationRecord 移除迁移记录
func (m *Migrator) removeMigrationRecord(version string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE version = ?", m.tableName)
	_, err := m.conn.Exec(query, version)
	if err != nil {
		return fmt.Errorf("failed to remove migration record %s: %w", version, err)
	}
	return nil
}

// sortMigrations 排序迁移
func (m *Migrator) sortMigrations(migrations []MigrationInterface) {
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version() < migrations[j].Version()
	})
}

// Up 执行所有待执行的迁移
func (m *Migrator) Up() error {
	if err := m.ensureMigrationTable(); err != nil {
		return err
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	// 获取待执行的迁移
	pending := make([]MigrationInterface, 0)
	for _, migration := range m.migrations {
		if _, exists := applied[migration.Version()]; !exists {
			pending = append(pending, migration)
		}
	}

	if len(pending) == 0 {
		m.logInfo("No pending migrations")
		return nil
	}

	// 排序待执行的迁移
	m.sortMigrations(pending)

	// 获取下一个批次号
	batch, err := m.getNextBatch()
	if err != nil {
		return err
	}

	m.logInfo("Starting migration batch", "batch", batch, "count", len(pending))

	// 执行迁移
	for _, migration := range pending {
		m.logInfo("Applying migration", "version", migration.Version(), "description", migration.Description())

		start := time.Now()
		err := migration.Up(m.conn)
		duration := time.Since(start)

		if err != nil {
			m.logError("Migration failed", err, "version", migration.Version())
			return fmt.Errorf("migration %s failed: %w", migration.Version(), err)
		}

		// 记录迁移
		if err := m.recordMigration(migration, batch); err != nil {
			return err
		}

		m.logInfo("Migration applied successfully", "version", migration.Version(), "duration", duration)
	}

	m.logInfo("Migration batch completed", "batch", batch, "applied", len(pending))
	return nil
}

// Down 回滚指定数量的迁移
func (m *Migrator) Down(steps int) error {
	if err := m.ensureMigrationTable(); err != nil {
		return err
	}

	if steps <= 0 {
		return fmt.Errorf("steps must be greater than 0")
	}

	// 获取已应用的迁移，按批次倒序
	query := fmt.Sprintf("SELECT version FROM %s ORDER BY batch DESC, id DESC LIMIT ?", m.tableName)
	rows, err := m.conn.Query(query, steps)
	if err != nil {
		return fmt.Errorf("failed to query migrations to rollback: %w", err)
	}
	defer rows.Close()

	versions := make([]string, 0)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, version)
	}

	if len(versions) == 0 {
		m.logInfo("No migrations to rollback")
		return nil
	}

	m.logInfo("Starting rollback", "count", len(versions))

	// 按相反顺序回滚
	for _, version := range versions {
		// 找到对应的迁移
		var migration MigrationInterface
		for _, m := range m.migrations {
			if m.Version() == version {
				migration = m
				break
			}
		}

		if migration == nil {
			m.logError("Migration not found for rollback", fmt.Errorf("migration %s not found", version))
			return fmt.Errorf("migration %s not found", version)
		}

		m.logInfo("Rolling back migration", "version", version, "description", migration.Description())

		start := time.Now()
		err := migration.Down(m.conn)
		duration := time.Since(start)

		if err != nil {
			m.logError("Rollback failed", err, "version", version)
			return fmt.Errorf("rollback %s failed: %w", version, err)
		}

		// 移除迁移记录
		if err := m.removeMigrationRecord(version); err != nil {
			return err
		}

		m.logInfo("Migration rolled back successfully", "version", version, "duration", duration)
	}

	m.logInfo("Rollback completed", "rolled_back", len(versions))
	return nil
}

// Reset 重置所有迁移
func (m *Migrator) Reset() error {
	if err := m.ensureMigrationTable(); err != nil {
		return err
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		m.logInfo("No migrations to reset")
		return nil
	}

	return m.Down(len(applied))
}

// Status 获取迁移状态
func (m *Migrator) Status() ([]*MigrationStatus, error) {
	if err := m.ensureMigrationTable(); err != nil {
		return nil, err
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return nil, err
	}

	status := make([]*MigrationStatus, 0, len(m.migrations))
	for _, migration := range m.migrations {
		s := &MigrationStatus{
			Version:     migration.Version(),
			Description: migration.Description(),
		}

		if record, exists := applied[migration.Version()]; exists {
			s.Applied = true
			s.AppliedAt = &record.AppliedAt
			s.Batch = record.Batch
		}

		status = append(status, s)
	}

	// 按版本排序
	sort.Slice(status, func(i, j int) bool {
		return status[i].Version < status[j].Version
	})

	return status, nil
}

// MigrationStatus 迁移状态
type MigrationStatus struct {
	Version     string     `json:"version"`
	Description string     `json:"description"`
	Applied     bool       `json:"applied"`
	AppliedAt   *time.Time `json:"applied_at,omitempty"`
	Batch       int        `json:"batch,omitempty"`
}

// String 返回状态的字符串表示
func (s *MigrationStatus) String() string {
	status := "pending"
	if s.Applied {
		status = fmt.Sprintf("applied (batch %d)", s.Batch)
	}
	return fmt.Sprintf("%-20s %-10s %s", s.Version, status, s.Description)
}

// Fresh 清空数据库并重新执行所有迁移
func (m *Migrator) Fresh() error {
	m.logInfo("Starting fresh migration")

	// 重置所有迁移
	if err := m.Reset(); err != nil {
		return fmt.Errorf("failed to reset migrations: %w", err)
	}

	// 执行所有迁移
	if err := m.Up(); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	m.logInfo("Fresh migration completed")
	return nil
}

// logInfo 记录信息日志
func (m *Migrator) logInfo(message string, args ...interface{}) {
	if m.logger != nil {
		m.logger.Info(message, args...)
	}
}

// logError 记录错误日志
func (m *Migrator) logError(message string, err error, args ...interface{}) {
	if m.logger != nil {
		allArgs := append([]interface{}{"error", err}, args...)
		m.logger.Error(message, allArgs...)
	}
}

// PrintStatus 打印迁移状态
func (m *Migrator) PrintStatus() error {
	status, err := m.Status()
	if err != nil {
		return err
	}

	fmt.Println("Migration Status:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-20s %-15s %s\n", "VERSION", "STATUS", "DESCRIPTION")
	fmt.Println(strings.Repeat("-", 80))

	for _, s := range status {
		fmt.Println(s.String())
	}

	fmt.Println(strings.Repeat("-", 80))
	return nil
}
