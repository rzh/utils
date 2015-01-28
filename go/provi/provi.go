package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/codegangsta/cli"

	"code.google.com/p/go.crypto/ssh"
)

var (
	config string
	wt     sync.WaitGroup
)

const CLR_0 = "\x1b[30;1m"
const CLR_R = "\x1b[31;1m"
const CLR_G = "\x1b[32;1m"
const CLR_Y = "\x1b[33;1m"
const CLR_B = "\x1b[34;1m"
const CLR_M = "\x1b[35;1m"
const CLR_C = "\x1b[36;1m"
const CLR_W = "\x1b[37;1m"
const CLR_N = "\x1b[0m"

type Jobs struct {
	Servers  []string `json:"servers"`
	Pem_file string   `json:"pem"`
	Tasks    []string `json:"tasks"`
	User     string   `json:"user"`
}

// ssh key support
type keychain struct {
	keys []ssh.Signer
}

func (k *keychain) Key(i int) (ssh.PublicKey, error) {
	if i < 0 || i >= len(k.keys) {
		return nil, nil
	}
	return k.keys[i].PublicKey(), nil
}

/*
func (k *keychain) Sign(i int, rand io.Reader, data []byte) (sig []byte, err error) {
	return k.keys[i].Sign(rand, data)
}*/

func (k *keychain) add(key ssh.Signer) {
	k.keys = append(k.keys, key)
}

func (k *keychain) loadPEM(file string) (ssh.Signer, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	key, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, err
	}
	k.add(key)
	return key, nil
}

// end of ssh key support

func (j *Jobs) runTasks() {

	// load ssh key
	k := new(keychain)
	key, err := k.loadPEM(j.Pem_file)

	if err != nil {
		panic("Cannot load key [" + j.Pem_file + "]: " + err.Error())
	}

	// not ready
	ts := time.Now().Format(time.RFC3339)
	for i := 0; i < len(j.Servers); i++ {
		wt.Add(1)
		// create log file
		log_file := "logs/provi_log__" + j.Servers[i] + "--" + ts + ".log"

		outfile, err := os.Create(log_file)
		if err != nil {
			log.Fatal(err)
		}
		defer outfile.Close()

		// create ssh session
		config := &ssh.ClientConfig{
			User: j.User,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(key),
			},
		}

		client, err := ssh.Dial("tcp", j.Servers[i]+":22", config)
		if err != nil {
			log.Fatal("Failed to connect to [" + j.Servers[i] + "] with error: " + err.Error())
		}

		go func(x int, s *ssh.Client, o *os.File) {
			defer wt.Done()
			j.runTask(x, s, o)
		}(i, client, outfile)
	}
	wt.Wait()
}

func (p *Jobs) runTask(i int, s *ssh.Client, outfile *os.File) {
	for j := 0; j < len(p.Tasks); j++ {
		p.runCmd(i, p.Tasks[j], s, outfile)
	}
}

func (p *Jobs) runCmd(i int, cmd string, client *ssh.Client, outfile *os.File) {
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create SSH session to server [" + p.Servers[i] + "] with error: " + err.Error())
	}
	defer session.Close()

	session.Stdout = outfile
	session.Stderr = outfile

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Fatalf("request for pseudo terminal failed: %s", err)
	}

	err = session.Start(cmd)
	if err != nil {
		log.Fatal("run command ["+cmd+"] on server ("+p.Servers[i]+") failed with -> ", err)
	}
	err = session.Wait()

	// TODO: make cmd not longer than 80
	if err == nil {
		log.Printf("[%15s] %-80s %s%s%s", p.Servers[i], cmd, CLR_G, "√", CLR_N)
	} else {
		log.Printf("[%15s] %-80s %s%s%s", p.Servers[i], cmd, CLR_R, "✗", CLR_N)
	}
}

func init() {
	//	flag.StringVar(&config, "config", "", "Config JSON for the run")
}

func run(config string) error {
	var t Jobs

	// read json file
	jsonBlob, e := ioutil.ReadFile(config)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	err := json.Unmarshal(jsonBlob, &t)
	if err != nil {
		fmt.Println("error:", err)
		return err
	}

	t.runTasks()

	return nil
}

func main() {
	app := cli.NewApp()
	app.Usage = "cluster management tool"
	app.Name = "provi"
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{

		// init a cluster from template
		{
			Name:      "init",
			ShortName: "i",
			Usage:     "initiate a cloud from template",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "cluster, c",
					Value: "config.json",
					Usage: "name of the cluster",
				},
				cli.StringFlag{
					Name:  "template, t",
					Value: "shard_template.json",
					Usage: "name of the template",
				},
			},
			Action: func(c *cli.Context) {
				log.Printf("initialize cluster %s%s%s with template %s%s%s\n",
					CLR_G, c.String("cluster"), CLR_N,
					CLR_Y, c.String("template"), CLR_N)
			},
		},

		// provision the cluster based on roles
		{
			Name:      "provision",
			ShortName: "p",
			Usage:     "provision a cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "cluster, c",
					Value: "config.json",
					Usage: "name of the cluster",
				},
			},
			Action: func(c *cli.Context) {
				println("providion cluster: ", c.Args().First())
				run(c.String("cluster"))
			},
		},

		// manuplate collection of templates
		{
			Name:      "template",
			ShortName: "r",
			Usage:     "dealing with templates",
			Subcommands: []cli.Command{
				{
					Name:  "add",
					Usage: "add a new template",
					Action: func(c *cli.Context) {
						println("new task template: ", c.Args().First())
					},
				},
				{
					Name:  "remove",
					Usage: "remove an existing template",
					Action: func(c *cli.Context) {
						println("removed task template: ", c.Args().First())
					},
				},
			},
		},
	}

	app.Run(os.Args)
}