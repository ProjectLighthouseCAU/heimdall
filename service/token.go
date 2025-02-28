package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ProjectLighthouseCAU/heimdall/crypto"
	"github.com/ProjectLighthouseCAU/heimdall/model"
	"github.com/redis/go-redis/v9"
)

type TokenService struct {
	redis *redis.Client
}

func NewTokenService(redis *redis.Client) TokenService {
	return TokenService{redis}
}

// Get all information about a users API token (username, token, roles, expiration)
func (ts *TokenService) GetToken(user *model.User) (*model.APIToken, error) {
	hash, err := ts.redis.HGetAll(context.TODO(), user.Username).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, model.NotFoundError{Err: err}
		}
		return nil, model.InternalServerError{Err: err}
	}
	now := time.Now()

	// workaround: ExpireTime should return time instead of duration, see https://github.com/redis/go-redis/issues/2657
	res, err := ts.redis.Do(context.TODO(), "EXPIRETIME", user.Username).Int64()
	if err != nil {
		return nil, model.InternalServerError{Message: "Redis could not get expire time"}
	}

	expiresAt := time.Unix(int64(res), 0)
	if res == -2 || res != -1 && expiresAt.Before(now) { // -2: key does not exist, -1: no expiration
		return nil, model.NotFoundError{}
	}
	token, ok := hash["token"]
	if !ok {
		return nil, model.NotFoundError{}
	}
	rolesJson, ok := hash["roles"]
	if !ok {
		return nil, model.NotFoundError{}
	}
	var roles []string
	err = json.Unmarshal([]byte(rolesJson), &roles)
	if err != nil {
		return nil, model.InternalServerError{Message: "Error unmarshaling json roles"}
	}
	return &model.APIToken{
		Username:  user.Username,
		Token:     token,
		Roles:     roles,
		ExpiresAt: expiresAt,
	}, nil
}

func newRandomToken() (string, error) {
	s, err := crypto.NewRandomAlphaNumString(20)
	if err != nil {
		return "", err
	}
	s = "API-TOK_" + s[0:4] + "-" + s[4:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20]
	return s, nil
}

// Generates a new API token for a user if the user does not have an API token (or expired)
// Returns true if the token was generated
func (ts *TokenService) GenerateApiTokenIfNotExists(user *model.User) (bool, error) {
	res, err := ts.redis.Exists(context.TODO(), user.Username).Result()
	if err != nil {
		return false, model.InternalServerError{Message: "Error retrieving API token from redis", Err: err}
	}
	if res != 0 { // exists
		return false, nil
	}
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Name
	}
	rolesJson, err := json.Marshal(roles)
	if err != nil {
		return false, model.InternalServerError{Message: "Error marshaling roles to json", Err: err}
	}
	newToken, err := newRandomToken()
	if err != nil {
		return false, model.InternalServerError{Message: "Could not generate token", Err: err}
	}
	_, err = ts.redis.Pipelined(context.TODO(), func(p redis.Pipeliner) error {
		p.HSet(context.TODO(), user.Username, "token", newToken, "roles", string(rolesJson))
		if !user.PermanentAPIToken {
			p.Expire(context.TODO(), user.Username, 3*24*time.Hour)
		}
		return nil
	})
	if err != nil {
		err = ts.redis.Del(context.TODO(), user.Username).Err()
		if err != nil {
			return false, model.InternalServerError{Message: "Error storing token and roles in redis or setting expiration date and could not delete key afterwards", Err: err}
		}
		return false, model.InternalServerError{Message: "Error storing token and roles in redis or setting expiration date", Err: err}
	}
	return true, nil
}

// Invalidates API token of user if it exists (not expired)
// and returns whether the token existed
func (ts *TokenService) InvalidateApiTokenIfExists(user *model.User) (bool, error) {
	existed, err := ts.redis.Del(context.TODO(), user.Username).Result()
	if err != nil {
		if err == redis.Nil {
			return false, model.NotFoundError{Err: err}
		}
		return false, model.InternalServerError{Message: "Error deleting API token in redis", Err: err}
	}
	return existed != 0, nil
}

// Invalidates and re-generates the users API token if it exists
// Does not re-generate the token if it didn't exist before
func (ts *TokenService) RegenerateApiToken(user *model.User) (bool, error) {
	tokenExisted, err := ts.InvalidateApiTokenIfExists(user)
	if err != nil {
		return false, err
	}
	if tokenExisted {
		_, err := ts.GenerateApiTokenIfNotExists(user)
		if err != nil {
			return true, err
		}
	}
	return tokenExisted, err
}

// Update the roles in redis without re-generating the token
// Returns true if key exists and roles were updated
func (ts *TokenService) UpdateRolesIfExists(user *model.User) (bool, error) {
	res, err := ts.redis.Exists(context.TODO(), user.Username).Result()
	if err != nil {
		return false, model.InternalServerError{Message: "Error retrieving API token from redis", Err: err}
	}
	if res == 0 { // does not exists
		return false, nil
	}
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Name
	}
	rolesJson, err := json.Marshal(roles)
	if err != nil {
		return false, model.InternalServerError{Message: "Error marshaling roles to json", Err: err}
	}
	err = ts.redis.HSet(context.TODO(), user.Username, "roles", string(rolesJson)).Err()
	if err != nil {
		return false, model.InternalServerError{Message: "Error storing roles in redis", Err: err}
	}
	return true, nil
}
