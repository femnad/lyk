// This example demonstrates how to authenticate with Spotify using the authorization code flow.
// In order to run this example yourself, you'll need to:
//
//  1. Register an application at: https://developer.spotify.com/my-applications/
//     - Use "http://localhost:8080/callback" as the redirect URI
//  2. Set the SPOTIFY_ID environment variable to the client ID you got in step 1.
//  3. Set the SPOTIFY_SECRET environment variable to the client secret from step 1.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/alexflint/go-arg"

	"github.com/femnad/lyk/cmd"
)

const (
	version = "v0.1.0"
)

type args struct {
	File string `arg:"-f,--file" default:"~/.config/lyk/lyk.yml" help:"Config file path"`
}

func (args) Version() string {
	return fmt.Sprintf("%s %s", cmd.Name, version)
}

func main() {
	var parsed args
	arg.MustParse(&parsed)

	err := cmd.LikeCurrentSong(context.Background(), parsed.File)
	if err != nil {
		log.Fatalf("error liking current song: %v", err)
	}
}
