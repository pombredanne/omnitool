package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/cli"
)

// ParseCommonFlags does the work for setting up the common environment
func ParseCommonFlags(c *cli.Context) (*string, *string, *HostGroup, *error) {
	u := c.GlobalString("username")
	k := c.GlobalString("keyfile")

	// Parse groups into list
	ml := c.GlobalString("list")
	mg := c.GlobalString("group")

	// Load machine list file
	mf, err := LoadFile(ml)
	if err != nil {
		return nil, nil, nil, &err
	}

	// Lost machine addresses by group name
	machineList := mf.Get(mg)

	return &u, &k, &machineList, nil
}

// GenerateFlags takes a list of flags and appends them to the flags that are
// common to all commands, for a complete list
func GenerateFlags(subFlags []cli.Flag) []cli.Flag {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:   "list, l",
			Value:  "machines.list",
			Usage:  "Path to machine list file",
			EnvVar: "OMNI_MACHINE_LIST",
		},

		cli.StringFlag{
			Name:   "username, u",
			Value:  os.Getenv("USER"),
			Usage:  "Username for machine group",
			EnvVar: "OMNI_USERNAME",
		},

		cli.StringFlag{
			Name:   "keyfile, k",
			Value:  os.Getenv("HOME") + "/.ssh/id_rsa.pub",
			Usage:  "Path to auth key",
			EnvVar: "OMNI_KEYFILE",
		},

		cli.StringFlag{
			Name:   "group, g",
			Value:  "vagrants",
			Usage:  "Machine group to perform task on",
			EnvVar: "OMNI_MACHINE_GROUP",
		},
	}

	for i := 0; i < len(subFlags); i++ {
		flag := subFlags[i]
		flags = append(flags, flag)
	}

	return flags
}

func cmdRun(c *cli.Context) {
	if len(c.Args()) != 1 {
		cli.ShowCommandHelp(c, "command")
		return
	}

	cmd := strings.Join(c.Args(), " ")

	username, key, machineList, err := ParseCommonFlags(c)
	if err != nil {
		log.Fatal(err)
		return
	}

	results := make(chan SSHResponse)
	timeout := time.After(60 * time.Second)
	MapCmd(*machineList, *username, *key, cmd, results)

	for i := 0; i < len(*machineList); i++ {
		select {
		case r := <-results:
			fmt.Printf("Hostname: %s\n", r.Hostname)
			fmt.Printf("Result:\n%s\n", r.Result)
		case <-timeout:
			fmt.Println("Timed out!")
		}
	}

	fmt.Println("CMD: ", c.Args())
}

func cmdScp(c *cli.Context) {
	if len(c.Args()) != 2 {
		cli.ShowCommandHelp(c, "scp")
		return
	}

	localPath := c.Args()[0]
	remotePath := c.Args()[1]

	username, key, machineList, err := ParseCommonFlags(c)
	if err != nil {
		log.Fatal(err)
		return
	}

	results := make(chan SSHResponse)
	timeout := time.After(60 * time.Second)
	MapScp(*machineList, *username, *key, localPath, remotePath, results)

	for i := 0; i < len(*machineList); i++ {
		select {
		case r := <-results:
			fmt.Printf("Hostname: %s\n", r.Hostname)
			fmt.Printf("Result:\n%s\n", r.Result)
		case <-timeout:
			fmt.Println("Timed out!")
		}
	}

	fmt.Println("SCP: ", c.Args())
}

func main() {
	// App setup
	app := cli.NewApp()
	app.Name = "omnitool"
	app.Usage = "Simple SSH pools, backed by machine lists"
	app.Version = "0.1"
	app.Flags = GenerateFlags([]cli.Flag{})

	// Subcommands
	app.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "Runs command on machine group",
			Action: cmdRun,
			Flags:  GenerateFlags([]cli.Flag{}),
		},
		{
			Name:   "scp",
			Usage:  "Copies file to machine group",
			Action: cmdScp,
			Flags:  GenerateFlags([]cli.Flag{}),
		},
	}

	// Do it
	app.Run(os.Args)
}
