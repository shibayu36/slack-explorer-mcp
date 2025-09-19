package main

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/slack-go/slack"
)

const (
	userRepositoryTTL      = 30 * time.Minute
	userCacheSweepInterval = 5 * time.Minute
)

// SessionCache holds cached users for a specific session
type SessionCache struct {
	users     []slack.User
	fetchedAt time.Time
}

// UserRepository manages user information with session-based caching
type UserRepository struct {
	sessionCaches map[SessionID]*SessionCache
	mu            sync.RWMutex
	stopCh        chan struct{}
	wg            sync.WaitGroup
	closeOnce     sync.Once

	// for test
	now func() time.Time
}

// NewUserRepository creates a new UserRepository
func NewUserRepository() *UserRepository {
	r := &UserRepository{
		sessionCaches: make(map[SessionID]*SessionCache),
		stopCh:        make(chan struct{}),

		now: time.Now,
	}

	r.wg.Add(1)
	go r.sweeper()

	return r
}

// FindByDisplayName searches for users by display name
func (r *UserRepository) FindByDisplayName(
	ctx context.Context,
	client SlackClient,
	displayName string,
	exact bool,
) ([]slack.User, error) {
	sessionID := SessionIDFromContext(ctx)

	r.mu.RLock()
	cache, exists := r.sessionCaches[sessionID]
	r.mu.RUnlock()

	if exists && !r.isExpired(cache) {
		return r.searchInUsers(cache.users, displayName, exact), nil
	}

	users, err := client.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.sessionCaches[sessionID] = &SessionCache{
		users:     users,
		fetchedAt: r.now(),
	}
	r.mu.Unlock()

	return r.searchInUsers(users, displayName, exact), nil
}

func (r *UserRepository) searchInUsers(users []slack.User, displayName string, exact bool) []slack.User {
	var matches []slack.User
	for _, user := range users {
		if exact {
			if user.Profile.DisplayName == displayName {
				matches = append(matches, user)
			}
		} else {
			if strings.Contains(user.Profile.DisplayName, displayName) {
				matches = append(matches, user)
			}
		}
	}
	return matches
}

func (r *UserRepository) isExpired(cache *SessionCache) bool {
	if cache == nil {
		return true
	}
	return r.now().Sub(cache.fetchedAt) > userRepositoryTTL
}

func (r *UserRepository) sweeper() {
	defer r.wg.Done()

	ticker := time.NewTicker(userCacheSweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.sweepExpiredCaches()
		case <-r.stopCh:
			return
		}
	}
}

func (r *UserRepository) sweepExpiredCaches() {
	now := r.now()

	r.mu.Lock()
	for sessionID, cache := range r.sessionCaches {
		if now.Sub(cache.fetchedAt) > userRepositoryTTL {
			delete(r.sessionCaches, sessionID)
		}
	}
	r.mu.Unlock()
}

func (r *UserRepository) Close() {
	r.closeOnce.Do(func() {
		close(r.stopCh)
	})
	r.wg.Wait()
}
