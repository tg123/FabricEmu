//go:build windows
// +build windows

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/tg123/FabricEmu"
)

func main() {

	packagerootdir := os.Args[1]
	configdir := os.Args[2]
	binarypath := os.Args[3]
	sfruntimedir := os.Args[4]
	// if len(os.Args) < 2 {
	// 	fmt.Printf("%v <path/to/app>", os.Args[0])
	// 	return
	// }

	h, err := FabricEmu.NewReplicaAgent(FabricEmu.ReplicaAgentConfig{
		OnNewReplicaOpened: func(replica *FabricEmu.Replica) {
			fmt.Println("new replica opened")
			time.Sleep(1 * time.Second) // TODO not fully opened, seems by design
			if err := replica.ChangeRole(FabricEmu.ReplicaRolePrimary); err != nil {
				panic(err)
			}
		},
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

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	go func() {
		ch <- cmd.Wait()
	}()

	panic(<-ch)
}
