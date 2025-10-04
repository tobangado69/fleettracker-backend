package repository

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BaseRepository implements the base repository interface using GORM
type BaseRepository[T any] struct {
	db    *gorm.DB
	model *T
}

// NewBaseRepository creates a new base repository instance
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	var model T
	return &BaseRepository[T]{
		db:    db,
		model: &model,
	}
}

// Create creates a new entity
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}
	return nil
}

// GetByID retrieves an entity by its ID
func (r *BaseRepository[T]) GetByID(ctx context.Context, id string) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("entity not found with id: %s", id)
		}
		return nil, fmt.Errorf("failed to get entity by id: %w", err)
	}
	return &entity, nil
}

// Update updates an existing entity
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}
	return nil
}

// Delete soft deletes an entity (if model has DeletedAt field) or hard deletes
func (r *BaseRepository[T]) Delete(ctx context.Context, id string) error {
	var entity T
	
	// Check if the model has DeletedAt field for soft delete
	if r.hasDeletedAtField() {
		if err := r.db.WithContext(ctx).Delete(&entity, "id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to delete entity: %w", err)
		}
	} else {
		// Hard delete
		if err := r.db.WithContext(ctx).Unscoped().Delete(&entity, "id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to delete entity: %w", err)
		}
	}
	return nil
}

// List retrieves entities with filtering and pagination
func (r *BaseRepository[T]) List(ctx context.Context, filters FilterOptions, pagination Pagination) ([]*T, error) {
	var entities []*T
	query := r.db.WithContext(ctx)
	
	// Apply filters
	query = r.applyFilters(query, filters)
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	// Execute query
	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to list entities: %w", err)
	}
	
	return entities, nil
}

// Count counts entities with filtering
func (r *BaseRepository[T]) Count(ctx context.Context, filters FilterOptions) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(r.model)
	
	// Apply filters
	query = r.applyFilters(query, filters)
	
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count entities: %w", err)
	}
	
	return count, nil
}

// WithTransaction executes a function within a database transaction
func (r *BaseRepository[T]) WithTransaction(ctx context.Context, fn func(Repository[T]) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &BaseRepository[T]{
			db:    tx,
			model: r.model,
		}
		return fn(txRepo)
	})
}

// applyFilters applies filter options to a GORM query
func (r *BaseRepository[T]) applyFilters(query *gorm.DB, filters FilterOptions) *gorm.DB {
	// Apply basic where conditions
	for field, value := range filters.Where {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}
	
	// Apply where in conditions
	for field, values := range filters.WhereIn {
		query = query.Where(fmt.Sprintf("%s IN ?", field), values)
	}
	
	// Apply where not conditions
	for field, value := range filters.WhereNot {
		query = query.Where(fmt.Sprintf("%s != ?", field), value)
	}
	
	// Apply like conditions
	for field, pattern := range filters.WhereLike {
		query = query.Where(fmt.Sprintf("%s LIKE ?", field), "%"+pattern+"%")
	}
	
	// Apply date range filters
	for field, dateRange := range filters.DateRange {
		if dateRange.Start != "" {
			query = query.Where(fmt.Sprintf("%s >= ?", field), dateRange.Start)
		}
		if dateRange.End != "" {
			query = query.Where(fmt.Sprintf("%s <= ?", field), dateRange.End)
		}
	}
	
	// Apply custom conditions
	for _, condition := range filters.Conditions {
		query = r.applyCondition(query, condition)
	}
	
	// Apply company isolation if company_id is provided
	if filters.CompanyID != "" {
		query = query.Where("company_id = ?", filters.CompanyID)
	}
	
	// Apply text search
	if filters.Search != "" && len(filters.SearchIn) > 0 {
		var searchConditions []string
		var searchArgs []interface{}
		
		for _, field := range filters.SearchIn {
			searchConditions = append(searchConditions, fmt.Sprintf("%s ILIKE ?", field))
			searchArgs = append(searchArgs, "%"+filters.Search+"%")
		}
		
		if len(searchConditions) > 0 {
			query = query.Where(strings.Join(searchConditions, " OR "), searchArgs...)
		}
	}
	
	return query
}

// applyCondition applies a custom condition to a GORM query
func (r *BaseRepository[T]) applyCondition(query *gorm.DB, condition Condition) *gorm.DB {
	switch strings.ToUpper(condition.Operator) {
	case "=":
		return query.Where(fmt.Sprintf("%s = ?", condition.Field), condition.Value)
	case "!=":
		return query.Where(fmt.Sprintf("%s != ?", condition.Field), condition.Value)
	case ">":
		return query.Where(fmt.Sprintf("%s > ?", condition.Field), condition.Value)
	case ">=":
		return query.Where(fmt.Sprintf("%s >= ?", condition.Field), condition.Value)
	case "<":
		return query.Where(fmt.Sprintf("%s < ?", condition.Field), condition.Value)
	case "<=":
		return query.Where(fmt.Sprintf("%s <= ?", condition.Field), condition.Value)
	case "IN":
		return query.Where(fmt.Sprintf("%s IN ?", condition.Field), condition.Value)
	case "NOT IN":
		return query.Where(fmt.Sprintf("%s NOT IN ?", condition.Field), condition.Value)
	case "LIKE":
		return query.Where(fmt.Sprintf("%s LIKE ?", condition.Field), condition.Value)
	case "ILIKE":
		return query.Where(fmt.Sprintf("%s ILIKE ?", condition.Field), condition.Value)
	case "IS NULL":
		return query.Where(fmt.Sprintf("%s IS NULL", condition.Field))
	case "IS NOT NULL":
		return query.Where(fmt.Sprintf("%s IS NOT NULL", condition.Field))
	default:
		// Default to equality
		return query.Where(fmt.Sprintf("%s = ?", condition.Field), condition.Value)
	}
}

// applyPagination applies pagination to a GORM query
func (r *BaseRepository[T]) applyPagination(query *gorm.DB, pagination Pagination) *gorm.DB {
	// Calculate offset and limit
	offset := pagination.Offset
	limit := pagination.Limit
	
	// Use page and page_size if offset and limit are not provided
	if offset == 0 && limit == 0 {
		if pagination.Page > 0 && pagination.PageSize > 0 {
			offset = (pagination.Page - 1) * pagination.PageSize
			limit = pagination.PageSize
		}
	}
	
	// Apply default limit if none specified
	if limit == 0 {
		limit = 20 // Default page size
	}
	
	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	return query
}

// hasDeletedAtField checks if the model has a DeletedAt field for soft delete
func (r *BaseRepository[T]) hasDeletedAtField() bool {
	t := reflect.TypeOf(*r.model)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "DeletedAt" {
			return true
		}
	}
	return false
}

// TransactionManager implements transaction management using GORM
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// Begin starts a new database transaction
func (tm *TransactionManager) Begin(ctx context.Context) (Transaction, error) {
	tx := tm.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	
	return &TransactionImpl{tx: tx}, nil
}

// WithTransaction executes a function within a database transaction
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(Transaction) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txImpl := &TransactionImpl{tx: tx}
		return fn(txImpl)
	})
}

// TransactionImpl implements the Transaction interface
type TransactionImpl struct {
	tx *gorm.DB
}

// Commit commits the transaction
func (t *TransactionImpl) Commit() error {
	return t.tx.Commit().Error
}

// Rollback rolls back the transaction
func (t *TransactionImpl) Rollback() error {
	return t.tx.Rollback().Error
}

// Repository returns a repository instance within the transaction
func (t *TransactionImpl) Repository[T any]() Repository[T] {
	return &BaseRepository[T]{
		db:    t.tx,
		model: new(T),
	}
}

// QueryBuilder provides a fluent interface for building complex queries
type QueryBuilder struct {
	db    *gorm.DB
	model interface{}
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(db *gorm.DB, model interface{}) *QueryBuilder {
	return &QueryBuilder{
		db:    db,
		model: model,
	}
}

// Where adds a WHERE condition
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.db = qb.db.Where(condition, args...)
	return qb
}

// And adds an AND condition
func (qb *QueryBuilder) And(condition string, args ...interface{}) *QueryBuilder {
	qb.db = qb.db.Where(condition, args...)
	return qb
}

// Or adds an OR condition
func (qb *QueryBuilder) Or(condition string, args ...interface{}) *QueryBuilder {
	qb.db = qb.db.Or(condition, args...)
	return qb
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder) OrderBy(field string, direction string) *QueryBuilder {
	if direction == "" {
		direction = "ASC"
	}
	qb.db = qb.db.Order(fmt.Sprintf("%s %s", field, strings.ToUpper(direction)))
	return qb
}

// Limit adds a LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.db = qb.db.Limit(limit)
	return qb
}

// Offset adds an OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.db = qb.db.Offset(offset)
	return qb
}

// Preload adds preloading for associations
func (qb *QueryBuilder) Preload(associations ...string) *QueryBuilder {
	for _, association := range associations {
		qb.db = qb.db.Preload(association)
	}
	return qb
}

// Select specifies fields to select
func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	qb.db = qb.db.Select(strings.Join(fields, ", "))
	return qb
}

// Group adds a GROUP BY clause
func (qb *QueryBuilder) Group(fields ...string) *QueryBuilder {
	qb.db = qb.db.Group(strings.Join(fields, ", "))
	return qb
}

// Having adds a HAVING clause
func (qb *QueryBuilder) Having(condition string, args ...interface{}) *QueryBuilder {
	qb.db = qb.db.Having(condition, args...)
	return qb
}

// Joins adds JOIN clauses
func (qb *QueryBuilder) Joins(query string, args ...interface{}) *QueryBuilder {
	qb.db = qb.db.Joins(query, args...)
	return qb
}

// Build returns the underlying GORM DB instance
func (qb *QueryBuilder) Build() *gorm.DB {
	return qb.db
}

// RepositoryManager manages all repositories
type RepositoryManager struct {
	db                    *gorm.DB
	transactionManager    *TransactionManager
	users                 UserRepository
	vehicles              VehicleRepository
	drivers               DriverRepository
	gpsTracks             GPSTrackRepository
	trips                 TripRepository
	geofences             GeofenceRepository
	companies             CompanyRepository
	auditLogs             AuditLogRepository
	sessions              SessionRepository
	passwordResetTokens   PasswordResetTokenRepository
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(db *gorm.DB) *RepositoryManager {
	tm := NewTransactionManager(db)
	
	return &RepositoryManager{
		db:                  db,
		transactionManager: tm,
		users:               NewUserRepository(db),
		vehicles:            NewVehicleRepository(db),
		drivers:             NewDriverRepository(db),
		gpsTracks:           NewGPSTrackRepository(db),
		trips:               NewTripRepository(db),
		geofences:           NewGeofenceRepository(db),
		companies:           NewCompanyRepository(db),
		auditLogs:           NewAuditLogRepository(db),
		sessions:            NewSessionRepository(db),
		passwordResetTokens: NewPasswordResetTokenRepository(db),
	}
}

// GetTransactionManager returns the transaction manager
func (rm *RepositoryManager) GetTransactionManager() TransactionManager {
	return *rm.transactionManager
}

// GetUsers returns the user repository
func (rm *RepositoryManager) GetUsers() UserRepository {
	return rm.users
}

// GetVehicles returns the vehicle repository
func (rm *RepositoryManager) GetVehicles() VehicleRepository {
	return rm.vehicles
}

// GetDrivers returns the driver repository
func (rm *RepositoryManager) GetDrivers() DriverRepository {
	return rm.drivers
}

// GetGPSTracks returns the GPS track repository
func (rm *RepositoryManager) GetGPSTracks() GPSTrackRepository {
	return rm.gpsTracks
}

// GetTrips returns the trip repository
func (rm *RepositoryManager) GetTrips() TripRepository {
	return rm.trips
}

// GetGeofences returns the geofence repository
func (rm *RepositoryManager) GetGeofences() GeofenceRepository {
	return rm.geofences
}

// GetCompanies returns the company repository
func (rm *RepositoryManager) GetCompanies() CompanyRepository {
	return rm.companies
}

// GetAuditLogs returns the audit log repository
func (rm *RepositoryManager) GetAuditLogs() AuditLogRepository {
	return rm.auditLogs
}

// GetSessions returns the session repository
func (rm *RepositoryManager) GetSessions() SessionRepository {
	return rm.sessions
}

// GetPasswordResetTokens returns the password reset token repository
func (rm *RepositoryManager) GetPasswordResetTokens() PasswordResetTokenRepository {
	return rm.passwordResetTokens
}

// HealthCheck performs a health check on the database connection
func (rm *RepositoryManager) HealthCheck(ctx context.Context) error {
	sqlDB, err := rm.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	
	return nil
}

// GetStats returns database connection statistics
func (rm *RepositoryManager) GetStats() (map[string]interface{}, error) {
	sqlDB, err := rm.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                   stats.InUse,
		"idle":                     stats.Idle,
		"wait_count":               stats.WaitCount,
		"wait_duration":            stats.WaitDuration.String(),
		"max_idle_closed":          stats.MaxIdleClosed,
		"max_idle_time_closed":     stats.MaxIdleTimeClosed,
		"max_lifetime_closed":      stats.MaxLifetimeClosed,
	}, nil
}
