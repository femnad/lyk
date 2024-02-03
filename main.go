package main

import (
	"context"
	"fmt"
	"log"

	"github.com/alexflint/go-arg"

	"github.com/femnad/lyk/cmd"
)

const (
	version = "v0.2.0"
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
