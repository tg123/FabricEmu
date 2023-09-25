//go:build windows
// +build windows

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/tg123/FabricEmu"
	"github.com/tg123/jobobject"
	"github.com/urfave/cli/v2"
)

var config struct {
	packgerootdir string
	configdir     string
	binarypath    string
	sfruntimedir  string
	stateful      bool
}

func main() {
	app := &cli.App{
		Name:  "simplelauncher",
		Usage: "simple launcher for sf app",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "packagerootdir",
				Usage:       "path to package root dir",
				Required:    true,
				Destination: &config.packgerootdir,
			},
			&cli.StringFlag{
				Name:        "configdir",
				Usage:       "path to config dir",
				Required:    true,
				Destination: &config.configdir,
			},
			&cli.StringFlag{
				Name:        "binarypath",
				Usage:       "path to binary",
				Required:    true,
				Destination: &config.binarypath,
			},
			&cli.StringFlag{
				Name:        "sfruntimedir",
				Usage:       "path to sfruntime dir",
				Required:    true,
				Destination: &config.sfruntimedir,
			},
			&cli.BoolFlag{
				Name:        "stateful",
				Usage:       "stateful service",
				Destination: &config.stateful,
			},
		},
		Action: func(ctx *cli.Context) error {
			run()
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run() {

	packagerootdir := config.packgerootdir
	configdir := config.configdir
	binarypath := config.binarypath
	sfruntimedir := config.sfruntimedir
	stateful := config.stateful

	// if len(os.Args) < 2 {
	// 	fmt.Printf("%v <path/to/app>", os.Args[0])
	// 	return
	// }

	h, err := FabricEmu.NewReplicaAgent(FabricEmu.ReplicaAgentConfig{
		OnNewReplicaOpened: func(replica *FabricEmu.Replica) {
			fmt.Println("new replica opened")

			if (!stateful) {
				return
			}

			time.Sleep(1 * time.Second) // TODO not fully opened, seems by design

			fmt.Println("change role to primary")
			if err := replica.ChangeRole(FabricEmu.ReplicaRolePrimary); err != nil {
				panic(err)
			}
		},
		Stateful: stateful,
	})
	if err != nil {
		panic(err)
	}

	ch := make(chan error, 2)
	go func() {
		ch <- h.Wait()
	}()

	cmd, err := h.ExecServicePkg(
		packagerootdir,
		binarypath,
		sfruntimedir,
		configdir,
	)
	if err != nil {
		panic(err)
	}

	r, w := io.Pipe()

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			color.HiBlue(scanner.Text())
		}
	}()

	cmd.Stdout = w
	cmd.Stderr = w

	jobObject, err := jobobject.Create()
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	if err := jobObject.AddProcess(cmd.Process); err != nil {
		panic(err)
	}

	go func() {
		ch <- cmd.Wait()
	}()

	panic(<-ch)
}
