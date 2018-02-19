package config

type Config struct {
	Host           string
	CertFile       string `toml:"cert_file"`
	KeyFile        string `toml:"key_file"`
	DisableHttps   bool   `toml:"disable_https"`
	Storage        string
	Repositories   map[string]*RepositoryConfig
	LocalCacheFile string `toml:"local_cache_file"`
	Organizations  []string
	GitHub         GitHubConfig
}

type GitHubConfig struct {
	Token string
}

type RepositoryConfig struct {
	Owner          string
	Repo           string
	Storage        string
	CredentialFile string `toml:"credential_file"`
	AccessID       string `toml:"access_id"`
	Bucket         string
	Region         string
}
