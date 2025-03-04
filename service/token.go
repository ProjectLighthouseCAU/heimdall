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
)

// TODO: implement permanent API tokens!
// TODO: notify Beacon about username changed and user deleted (DONE)
// DON'T: let Beacon delete resources based on the AuthUpdateMessages alone - Beacon is not always subscribed to every user's AuthUpdates
// TODO: let Beacon delete resource when username_invalid is true
// TODO: add endpoint for Beacon to query the list of usernames (for creating and deleting resources / user directories)

type TokenService struct {
	tokenRepository repository.TokenRepository
	userRepository  repository.UserRepository

	openAuthConnections map[string][]chan *model.AuthUpdateMessage // username -> update channel
	lock                *sync.Mutex
}

func NewTokenService(tokenRepository repository.TokenRepository, userRepository repository.UserRepository) TokenService {
	return TokenService{tokenRepository, userRepository, make(map[string][]chan *model.AuthUpdateMessage), &sync.Mutex{}}
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
// the given user must have its roles and api token field pre-loaded from the database before calling
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
		ExpiresAt: time.Now().Add(config.ApiTokenExpirationTime),
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
			Username:        user.Username,
			Token:           token.Token,
			ExpiresAt:       token.ExpiresAt,
			Roles:           roles,
			UsernameInvalid: false,
		}
	}
	return true, nil
}

func (ts *TokenService) NotifyUsernameInvalid(user *model.User) {
	token := user.ApiToken
	if token != nil {
		err := ts.tokenRepository.DeleteByID(token.ID)
		if err != nil {
			log.Println("NotifyUsernameInvalid: could not delete token with id", token.ID, ":", err)
		}
	}

	// notify subscribers
	ts.lock.Lock()
	defer ts.lock.Unlock()

	chans := ts.openAuthConnections[user.Username]
	if chans == nil {
		return
	}
	for _, c := range chans {
		c <- &model.AuthUpdateMessage{ // TODO: maybe not needed? close of channel (and connection) signals that auth data is invalid -> ensure connection closes!
			Username:        user.Username,
			Token:           "",
			ExpiresAt:       time.Now(),
			Roles:           []string{},
			UsernameInvalid: true,
		}
		close(c) // closed channel (and therefore closed connection) indicates invalidated token
	}
}

// Notify that the roles of a user have changed
// the given user must have its roles and api token field pre-loaded from the database before calling
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
			Username:        user.Username,
			Token:           token.Token,
			ExpiresAt:       token.ExpiresAt,
			Roles:           roles,
			UsernameInvalid: false,
		}
	}
	return nil
}

// Invalidates an existing API token of a user and re-generates a new one
func (ts *TokenService) RegenerateApiToken(user *model.User) error {
	user.ApiToken = nil // forces re-generation
	generated, err := ts.GenerateApiTokenIfNotExists(user)
	if err != nil {
		return err
	}
	if !generated {
		return model.InternalServerError{Message: "API Token could not be regenerated!"}
	}
	return nil
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
	log.Println("Subscribed to", username)
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
	log.Println("Unsubscribed from", username)
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
