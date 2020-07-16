package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/websentry/websentry/server"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "config.json",
				Usage:   "Load configuration from `FILE`",
			},
		},
		Action: start,
		Usage:  "master server for websentry",
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func start(c *cli.Context) error {
	// TODO: Calling Init of each module manually is not a good idea.
	//       It's very easy to miss something or initializing in a wrong order.
	//       It's better to avoid using singleton and pass things around.
	err := server.Init(c.String("config"))
	return err
}
