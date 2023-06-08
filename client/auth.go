package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
	"github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"

	"github.com/femnad/lyk/config"
	"github.com/femnad/mare"
)

const (
	tokenFilePath = "~/.local/share/lyk/token.json"
)

type authResult struct {
	err   error
	token *oauth2.Token
}

var (
	auth         *spotifyauth.Authenticator
	authResultCh = make(chan authResult)
	scopes       = []string{
		// Get current track: https://developer.spotify.com/documentation/web-api/reference/get-the-users-currently-playing-track
		spotifyauth.ScopeUserReadCurrentlyPlaying,
		// Save track: https://developer.spotify.com/documentation/web-api/reference/save-tracks-user
		spotifyauth.ScopeUserLibraryModify,
	}
	tokenFile = mare.ExpandUser(tokenFilePath)
)

func generateState() string {
	return uuid.New().String()
}

func closeAndCheck(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.Fatalf("error closing token file %s, %v", tokenFile, err)
	}
}

func saveToken(token oauth2.Token) error {
	dir, _ := path.Split(tokenFile)
	err := mare.EnsureDir(dir)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(tokenFile, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer closeAndCheck(f)

	encoder := json.NewEncoder(f)
	return encoder.Encode(token)
}

func hasSavedToken() (bool, error) {
	_, err := os.Stat(tokenFile)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func saveTokenFromClient(client *spotify.Client) error {
	token, err := client.Token()
	if err != nil {
		return fmt.Errorf("error getting token from client: %v", err)
	}

	return saveToken(*token)
}

func clientFromSavedToken(ctx context.Context) (*spotify.Client, error) {
	f, err := os.Open(tokenFile)
	if err != nil {
		return &spotify.Client{}, err
	}
	defer closeAndCheck(f)

	var token oauth2.Token
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&token)
	if err != nil {
		return &spotify.Client{}, err
	}

	client := spotify.New(auth.Client(ctx, &token))
	if token.Expiry.Before(time.Now()) {
		err = saveTokenFromClient(client)
		if err != nil {
			return &spotify.Client{}, err
		}
	}

	return client, nil
}

// Mostly derived from https://github.com/zmb3/spotify/blob/master/examples/authenticate/authcode/authenticate.go.
func authenticate(ctx context.Context, cfg config.Config) (*spotify.Client, error) {
	callbackPath, err := cfg.Path()
	if err != nil {
		return &spotify.Client{}, fmt.Errorf("error determining path of callback URI: %v", err)
	}

	state := generateState()
	completeAuth := func(w http.ResponseWriter, r *http.Request) {
		tok, rErr := auth.Token(r.Context(), state, r)
		if rErr != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			authResultCh <- authResult{err: rErr}
			return
		}

		if st := r.FormValue("state"); st != state {
			http.NotFound(w, r)
			rErr = fmt.Errorf("State mismatch: %s != %s\n", st, state)
			authResultCh <- authResult{err: rErr}
			return
		}

		authResultCh <- authResult{token: tok, err: nil}
	}

	port, err := cfg.Port()
	if err != nil {
		return &spotify.Client{}, fmt.Errorf("error determining port of callback URI")
	}

	http.HandleFunc(callbackPath, completeAuth)
	go func() {
		addr := fmt.Sprintf(":%s", port)
		hErr := http.ListenAndServe(addr, nil)
		if hErr != nil {
			log.Fatalf("error listening for redirect response: %v", hErr)
		}
	}()

	authUrl := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", authUrl)

	result := <-authResultCh
	resultErr := result.err
	if resultErr != nil {
		return &spotify.Client{}, resultErr
	}

	client := spotify.New(auth.Client(ctx, result.token))

	user, err := client.CurrentUser(ctx)
	if err != nil {
		return &spotify.Client{}, err
	}

	log.Printf("Logged as %s", user.ID)

	err = saveToken(*result.token)
	if err != nil {
		return &spotify.Client{}, err
	}

	return client, nil
}

func setAuth(cfg config.Config) error {
	authOptions := []spotifyauth.AuthenticatorOption{
		spotifyauth.WithRedirectURL(cfg.RedirectURI()),
		spotifyauth.WithScopes(scopes...),
	}

	if !cfg.ClientIdInEnv() {
		clientId, err := cfg.ClientId()
		if err != nil {
			return err
		}
		authOptions = append(authOptions, spotifyauth.WithClientID(clientId))
	}
	if !cfg.ClientSecretInEnv() {
		clientSecret, err := cfg.ClientSecret()
		if err != nil {
			return err
		}
		authOptions = append(authOptions, spotifyauth.WithClientSecret(clientSecret))
	}

	auth = spotifyauth.New(authOptions...)
	return nil
}

func Get(ctx context.Context, configFile string) (*spotify.Client, error) {
	cfg, err := config.Get(configFile)
	if err != nil {
		return &spotify.Client{}, err
	}

	err = setAuth(cfg)
	if err != nil {
		return &spotify.Client{}, err
	}

	hasToken, err := hasSavedToken()
	if err != nil {
		return &spotify.Client{}, fmt.Errorf("error checking for saved auth token: %v", err)
	}

	var client *spotify.Client
	if hasToken {
		client, err = clientFromSavedToken(ctx)
		if err != nil {
			return &spotify.Client{}, fmt.Errorf("error authenticating with saved token: %v", err)
		}
		return client, nil
	} else {
		client, err = authenticate(ctx, cfg)
		if err != nil {
			return client, fmt.Errorf("error authenticating for the first time: %v", err)
		}

	}

	return client, err
}
