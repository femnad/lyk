package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/femnad/lyk/client"
	"github.com/femnad/lyk/notify"
)

func LikeCurrentSong(ctx context.Context, configFile string) error {
	spotify, err := client.Get(ctx, configFile)
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

	msg := fmt.Sprintf("%s by %s on %s", track.Name, artistStr, track.Album.Name)
	return notify.Send("Liked Song", msg)
}
