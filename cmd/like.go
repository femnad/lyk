package cmd

import (
	"context"
	"fmt"
	"strings"

	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"github.com/femnad/lyk/notify"
	"github.com/femnad/spoaut"
)

const (
	tokenFile = "~/.local/share/lyk/token.json"
)

var (
	scopes = []string{
		// Get current track: https://developer.spotify.com/documentation/web-api/reference/get-the-users-currently-playing-track
		spotifyauth.ScopeUserReadCurrentlyPlaying,
		// Save track: https://developer.spotify.com/documentation/web-api/reference/save-tracks-user
		spotifyauth.ScopeUserLibraryModify,
	}
)

func LikeCurrentSong(ctx context.Context, configFile string) error {
	cfg := spoaut.Config{
		ConfigFile: configFile,
		Scopes:     scopes,
		TokenFile:  tokenFile,
	}

	spotify, err := spoaut.Client(ctx, cfg)
	if err != nil {
		return err
	}

	current, err := spotify.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return fmt.Errorf("error getting currently playing song: %v", err)
	}

	if !current.Playing {
		return sendErrorNotification("No track is playing at the moment")
	}

	track := current.Item
	err = spotify.AddTracksToLibrary(ctx, track.ID)
	if err != nil {
		return fmt.Errorf("error adding track %s to library: %v", track.ID, err)
	}

	var artists []string
	for _, artist := range track.Artists {
		artists = append(artists, artist.Name)
	}
	artistStr := strings.Join(artists, ",")

	msg := fmt.Sprintf("%s <i>by</i> %s <i>on</i> %s", track.Name, artistStr, track.Album.Name)
	return notify.Send("Liked Song", msg)
}
