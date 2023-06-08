package cmd

import (
	"context"
	"fmt"

	"github.com/femnad/lyk/client"
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
		return fmt.Errorf("no playback in progress")
	}

	currentId := current.Item.ID
	err = spotify.AddTracksToLibrary(ctx, currentId)
	if err != nil {
		return fmt.Errorf("error adding track %s to library: %v", currentId, err)
	}

	return nil
}
