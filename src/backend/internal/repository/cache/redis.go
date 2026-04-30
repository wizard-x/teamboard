package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	rdb       *redis.Client
	authTTL   time.Duration
	boardTTL  time.Duration
	listTTL   time.Duration
}

type AuthContext struct {
	MemberID string `json:"member_id"`
	TeamID   string `json:"team_id"`
	Role     string `json:"role"`
}

func New(rdb *redis.Client, authTTLMinutes int) *Cache {
	return &Cache{
		rdb:      rdb,
		authTTL:  time.Duration(authTTLMinutes) * time.Minute,
		boardTTL: 5 * time.Minute,
		listTTL:  2 * time.Minute,
	}
}

// --- Auth Cache ---

func (c *Cache) GetAuthContext(ctx context.Context, keyPrefix, keyHash string) (*AuthContext, error) {
	key := fmt.Sprintf("auth:api_key:%s:%s", keyPrefix, keyHash)
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting auth cache: %w", err)
	}
	var authCtx AuthContext
	if err := json.Unmarshal([]byte(val), &authCtx); err != nil {
		return nil, fmt.Errorf("unmarshaling auth cache: %w", err)
	}
	return &authCtx, nil
}

func (c *Cache) SetAuthContext(ctx context.Context, keyPrefix, keyHash string, authCtx *AuthContext) error {
	key := fmt.Sprintf("auth:api_key:%s:%s", keyPrefix, keyHash)
	data, err := json.Marshal(authCtx)
	if err != nil {
		return fmt.Errorf("marshaling auth cache: %w", err)
	}
	return c.rdb.Set(ctx, key, data, c.authTTL).Err()
}

func (c *Cache) InvalidateAuthContext(ctx context.Context, keyPrefix, keyHash string) error {
	key := fmt.Sprintf("auth:api_key:%s:%s", keyPrefix, keyHash)
	return c.rdb.Del(ctx, key).Err()
}

// --- Board Cache ---

func (c *Cache) GetBoard(ctx context.Context, boardID string) ([]byte, error) {
	key := fmt.Sprintf("board:%s", boardID)
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting board cache: %w", err)
	}
	return val, nil
}

func (c *Cache) SetBoard(ctx context.Context, boardID string, data []byte) error {
	key := fmt.Sprintf("board:%s", boardID)
	return c.rdb.Set(ctx, key, data, c.boardTTL).Err()
}

func (c *Cache) InvalidateBoard(ctx context.Context, boardID string) error {
	key := fmt.Sprintf("board:%s", boardID)
	return c.rdb.Del(ctx, key).Err()
}

// --- Board List Cache ---

func (c *Cache) GetBoardList(ctx context.Context, teamID string) ([]byte, error) {
	key := fmt.Sprintf("board:list:%s", teamID)
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting board list cache: %w", err)
	}
	return val, nil
}

func (c *Cache) SetBoardList(ctx context.Context, teamID string, data []byte) error {
	key := fmt.Sprintf("board:list:%s", teamID)
	return c.rdb.Set(ctx, key, data, c.listTTL).Err()
}

func (c *Cache) InvalidateBoardList(ctx context.Context, teamID string) error {
	key := fmt.Sprintf("board:list:%s", teamID)
	return c.rdb.Del(ctx, key).Err()
}

// --- Rate Limiting ---

func (c *Cache) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int, error) {
	count, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("incrementing rate limit: %w", err)
	}
	if count == 1 {
		c.rdb.Expire(ctx, key, window)
	}
	return int(count), nil
}

func (c *Cache) GetRateLimitTTL(ctx context.Context, key string) (time.Duration, error) {
	return c.rdb.TTL(ctx, key).Result()
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}
