package config

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestRead(t *testing.T) {
	f, err := ioutil.TempFile("", "test_config")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(`host = "git-lfs.localdomain"
cache_dir = "./.cert_cache"
storage = "s3"
local_cache_file = "test.db"
organizations = ["monsterstrike"]

[repositories]
  [repositories."f110/test1"]
  [repositories."f110/test2"]

[github]
token = "hogefuga"`)

	config, err := Read(f.Name())
	if err != nil {
		t.Error(err)
	}
	if config.Host != "git-lfs.localdomain" {
		t.Error("Failed parse host")
	}
	if config.CacheDir != "./.cert_cache" {
		t.Error("Failed parse cache_dir")
	}
	if config.Storage != "s3" {
		t.Error("Storage is not s3")
	}
	if len(config.Repositories) != 2 {
		t.Error("failed parse Repositories")
	}
	if config.Repositories["f110/test1"].Owner != "f110" || config.Repositories["f110/test1"].Repo != "test1" {
		t.Errorf("failed parse Repositories value: %v", config.Repositories)
	}
	if config.LocalCacheFile != "test.db" {
		t.Errorf("failed parse LocalCacheFile: %s", config.LocalCacheFile)
	}
	if len(config.Organizations) != 1 {
		t.Errorf("failed parse Organizations: %v", config.Organizations)
	}
	if config.GitHub.Token != "hogefuga" {
		t.Errorf("failed parse github.token: %s", config.GitHub.Token)
	}
}
