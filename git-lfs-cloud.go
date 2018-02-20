package main

import (
	"fmt"
	"os"
	"sync"

	_ "cloud.google.com/go/storage"
	_ "github.com/aws/aws-sdk-go/aws"
	"github.com/boltdb/bolt"
	"github.com/f110/git-lfs-cloud/auth"
	"github.com/f110/git-lfs-cloud/config"
	"github.com/f110/git-lfs-cloud/database"
	"github.com/f110/git-lfs-cloud/lfs"
)

var (
	globalConfig config.Config
)

func run() int {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: git-lfs-cloud [config file]")
		return 1
	}

	// Read config file
	conf, err := config.Read(os.Args[1])
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return 1
	}
	globalConfig = conf

	github := auth.NewGitHub(globalConfig.GitHub.Token)
	auth.DefaultClient = github

	// Open database file
	db, err := bolt.Open(globalConfig.LocalCacheFile, 0644, nil)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return 1
	}
	defer db.Close()
	database.Conn = db

	github.CrawlRepositories(globalConfig.Repositories)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		auth.SSHServer()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		lfs.ObjectServer(globalConfig.DisableHttps, globalConfig.CertFile, globalConfig.KeyFile, globalConfig.Repositories)
	}()
	wg.Wait()

	return 0
}

func main() {
	os.Exit(run())
}
