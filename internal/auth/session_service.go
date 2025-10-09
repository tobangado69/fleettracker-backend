package auth

import (
	"context"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/tobangado69/fleettracker-pro/backend/pkg/errors"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// SessionResponse represents a session response
type SessionResponse struct {
	ID        string    `json:"id"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	IsActive  bool      `json:"is_active"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	IsCurrent bool      `json:"is_current"` // Is this the current session
}

// GetActiveSessions retrieves all active sessions for a user
func (s *Service) GetActiveSessions(ctx context.Context, userID string, currentToken string) ([]SessionResponse, *apperrors.AppError) {
	var sessions []models.Session
	
	// Query active sessions for the user
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ? AND expires_at > ?", userID, true, time.Now()).
		Order("created_at DESC").
		Find(&sessions).Error
	
	if err != nil {
		return nil, apperrors.NewInternalError("Failed to retrieve sessions").WithInternal(err)
	}
	
	// Convert to response format
	responses := make([]SessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = SessionResponse{
			ID:        session.ID,
			UserAgent: session.UserAgent,
			IPAddress: session.IPAddress,
			IsActive:  session.IsActive,
			ExpiresAt: session.ExpiresAt,
			CreatedAt: session.CreatedAt,
			IsCurrent: session.Token == currentToken, // Mark current session
		}
	}
	
	return responses, nil
}

// RevokeSession revokes a specific session
func (s *Service) RevokeSession(ctx context.Context, userID, sessionID string) *apperrors.AppError {
	// Find the session
	var session models.Session
	err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", sessionID, userID).
		First(&session).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.NewNotFoundError("Session not found")
		}
		return apperrors.NewInternalError("Failed to find session").WithInternal(err)
	}
	
	// Deactivate the session
	updates := map[string]interface{}{
		"is_active": false,
	}
	
	if err := s.db.Model(&session).Updates(updates).Error; err != nil {
		return apperrors.NewInternalError("Failed to revoke session").WithInternal(err)
	}
	
	// Invalidate session in Redis cache
	cacheKey := "session:" + sessionID
	if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
		// Log but don't fail - cache invalidation is best-effort
		// The session is already deactivated in the database
	}
	
	return nil
}

// RevokeAllSessions revokes all sessions for a user (except current)
func (s *Service) RevokeAllSessions(ctx context.Context, userID string, exceptSessionID string) *apperrors.AppError {
	// Build query
	query := s.db.Model(&models.Session{}).
		Where("user_id = ? AND is_active = ?", userID, true)
	
	// Exclude current session if provided
	if exceptSessionID != "" {
		query = query.Where("id != ?", exceptSessionID)
	}
	
	// Deactivate all matching sessions
	updates := map[string]interface{}{
		"is_active": false,
	}
	
	if err := query.Updates(updates).Error; err != nil {
		return apperrors.NewInternalError("Failed to revoke sessions").WithInternal(err)
	}
	
	// Invalidate cache for all user sessions
	pattern := "session:user:" + userID + ":*"
	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err == nil && len(keys) > 0 {
		s.redis.Del(ctx, keys...)
	}
	
	return nil
}

// CleanupExpiredSessions removes expired sessions from database
func (s *Service) CleanupExpiredSessions(ctx context.Context) error {
	// Delete expired sessions
	result := s.db.WithContext(ctx).
		Where("expires_at < ? OR is_active = ?", time.Now(), false).
		Delete(&models.Session{})
	
	if result.Error != nil {
		return result.Error
	}
	
	// Log cleanup
	if result.RowsAffected > 0 {
		// Session cleanup successful
	}
	
	return nil
}

