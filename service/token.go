package service

import (
	"errors"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/ProjectLighthouseCAU/heimdall/config"
	"github.com/ProjectLighthouseCAU/heimdall/crypto"
	"github.com/ProjectLighthouseCAU/heimdall/model"
	"github.com/ProjectLighthouseCAU/heimdall/repository"
	"github.com/redis/go-redis/v9"
)

type TokenService struct {
	redis           *redis.Client
	tokenRepository repository.TokenRepository

	openAuthConnections map[string][]chan *model.AuthUpdateMessage // username -> update channel
	lock                *sync.Mutex
}

func NewTokenService(redis *redis.Client, tokenRepository repository.TokenRepository) TokenService {
	return TokenService{redis, tokenRepository, make(map[string][]chan *model.AuthUpdateMessage), &sync.Mutex{}}
}

func newRandomToken() (string, error) {
	s, err := crypto.NewRandomAlphaNumString(20)
	if err != nil {
		return "", err
	}
	s = "API-TOK_" + s[0:4] + "-" + s[4:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20]
	return s, nil
}

func (ts *TokenService) GetToken(user *model.User) (*model.Token, error) {
	if user.Roles == nil {
		log.Println("user.Roles is nil")
	}
	if user.ApiToken == nil {
		return nil, model.NotFoundError{
			Message: "This user does not have a valid API token",
			Err:     nil,
		}
	}
	return user.ApiToken, nil
}

// Generates a new API token for a user if the user does not have an API token (or expired)
// Returns true if the token was generated
func (ts *TokenService) GenerateApiTokenIfNotExists(user *model.User) (bool, error) {
	token := user.ApiToken
	if token != nil && token.ExpiresAt.After(time.Now()) {
		return false, nil
	}
	newToken, err := newRandomToken()
	if err != nil {
		return false, model.InternalServerError{Message: "Could not generate token", Err: err}
	}
	token = &model.Token{
		Token:     newToken,
		ExpiresAt: time.Now().Add(config.GetDuration("API_TOKEN_EXPIRATION_TIME", 3*24*time.Hour)),
		UserID:    user.ID,
	}
	err = ts.tokenRepository.Save(token)
	if err != nil {
		return false, model.InternalServerError{Message: "Error storing token", Err: err}
	}

	// notify subscribers
	ts.lock.Lock()
	defer ts.lock.Unlock()

	chans := ts.openAuthConnections[user.Username]
	if chans == nil {
		return true, nil
	}
	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}
	for _, c := range chans {
		c <- &model.AuthUpdateMessage{
			Username:  user.Username,
			Token:     token.Token,
			ExpiresAt: token.ExpiresAt,
			Roles:     roles,
		}
	}
	return true, nil
}

// Invalidates API token of user if it exists (not expired)
// and returns whether the token existed
func (ts *TokenService) InvalidateApiTokenIfExists(user *model.User) (bool, error) {
	return ts.invalidateApiTokenIfExists(user, true)
}

func (ts *TokenService) invalidateApiTokenIfExists(user *model.User, notify bool) (bool, error) {
	token := user.ApiToken
	if token == nil {
		return false, nil
	}
	err := ts.tokenRepository.DeleteByID(token.ID)
	if err != nil {
		return false, err
	}
	if !notify {
		return true, nil
	}

	// notify subscribers
	ts.lock.Lock()
	defer ts.lock.Unlock()

	chans := ts.openAuthConnections[user.Username]
	if chans == nil {
		return true, nil
	}
	for _, c := range chans {
		close(c) // closed channel (and therefore closed connection) indicates invalidated token
	}
	return true, nil
}

// Notify that the roles of a user have changed
// the given user must have its roles field pre-loaded from the database before calling
func (ts *TokenService) NotifyRoleUpdate(user *model.User) error {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	chans := ts.openAuthConnections[user.Username]
	if chans == nil {
		return nil
	}
	token := user.ApiToken
	if token == nil {
		return model.InternalServerError{Message: "API token was nil while notifying a role change for user " + user.Username}
	}
	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}
	for _, c := range chans {
		c <- &model.AuthUpdateMessage{
			Username:  user.Username,
			Token:     token.Token,
			ExpiresAt: token.ExpiresAt,
			Roles:     roles,
		}
	}
	return nil
}

// Invalidates and re-generates the users API token if it exists
// Does not re-generate the token if it didn't exist before
func (ts *TokenService) RegenerateApiToken(user *model.User) (bool, error) {
	tokenExisted, err := ts.invalidateApiTokenIfExists(user, false) // don't notify yet
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

func (ts *TokenService) SubscribeToChanges(username string) chan *model.AuthUpdateMessage {
	c := make(chan *model.AuthUpdateMessage, 1)

	ts.lock.Lock()
	defer ts.lock.Unlock()

	chans := ts.openAuthConnections[username]
	if chans == nil {
		ts.openAuthConnections[username] = []chan *model.AuthUpdateMessage{c}
	} else {
		ts.openAuthConnections[username] = append(ts.openAuthConnections[username], c)
	}
	return c
}

func (ts *TokenService) UnsubscribeFromChanges(username string, c chan *model.AuthUpdateMessage) error {
	if c == nil {
		return errors.New("cannot unsubscribe from nil channel")
	}

	ts.lock.Lock()
	defer ts.lock.Unlock()

	chans, ok := ts.openAuthConnections[username]
	if !ok {
		return errors.New("no subscription for user " + username)
	}
	if !slices.Contains(chans, c) {
		return errors.New("this channel has not subscribed to " + username)
	}

	// delete entire map key if this is the only subscription
	if len(chans) <= 1 {
		delete(ts.openAuthConnections, username)
		return nil
	}
	// delete channel from slice otherwise
	ts.openAuthConnections[username] = deleteElement(ts.openAuthConnections[username], c)
	return nil
}

func deleteElement[S ~[]E, E comparable](s S, e E) S {
	return slices.DeleteFunc(s, func(_e E) bool { return _e == e })
}
