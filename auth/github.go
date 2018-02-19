package auth

import (
	"context"
	"log"

	"github.com/f110/git-lfs-cloud/config"
	"github.com/f110/git-lfs-cloud/database"
	"github.com/gliderlabs/ssh"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	PermitPublicKeys = map[string][]ssh.PublicKey{}
	DefaultClient    *GitHub
)

type GitHub struct {
	client *github.Client
}

func NewGitHub(token string) *GitHub {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	return &GitHub{client: github.NewClient(tc)}
}

func (gh *GitHub) CrawlRepositories(repos map[string]*config.RepositoryConfig) error {
	for _, v := range repos {
		err := gh.crawlRepository(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gh *GitHub) GetMembers(owner, repo string) ([]*github.User, error) {
	teamsOpt := &github.ListOptions{PerPage: 100}
	teams, _, err := gh.client.Repositories.ListTeams(context.Background(), owner, repo, teamsOpt)
	if err != nil {
		return nil, err
	}

	allUsers := make([]*github.User, 0)
	for _, team := range teams {
		membersOpt := &github.OrganizationListTeamMembersOptions{}
		membersOpt.PerPage = 100
		users, _, err := gh.client.Organizations.ListTeamMembers(context.Background(), *team.ID, membersOpt)
		if err != nil {
			return nil, err
		}
		allUsers = append(allUsers, users...)
	}

	userMap := make(map[int64]*github.User)
	for _, user := range allUsers {
		if _, ok := userMap[*user.ID]; ok == false {
			userMap[*user.ID] = user
		}
	}

	repoUsers := make([]*github.User, 0)
	for _, v := range userMap {
		repoUsers = append(repoUsers, v)
	}

	return repoUsers, nil
}

func (gh *GitHub) InvalidateRepositoryCache(owner, repo string) error {
	users, err := database.ReadRepositoryUsers(owner + "/" + repo)
	if err != nil {
		return err
	}

	for _, u := range users {
		log.Printf("Delete %s's public keys", u)
		err := database.DeletePublicKeys(u)
		if err != nil {
			return err
		}
	}

	log.Printf("Delete %s/%s user list", owner, repo)
	return database.DeleteRepositoryUser(owner + "/" + repo)
}

func (gh *GitHub) GetPubkey(login string) ([]*github.Key, error) {
	opt := &github.ListOptions{PerPage: 100}
	keys, _, err := gh.client.Users.ListKeys(context.Background(), login, opt)
	return keys, err
}

func (gh *GitHub) crawlRepository(repo *config.RepositoryConfig) error {
	users, err := gh.readOrGetRepositoryMember(repo.Owner, repo.Repo)
	if err != nil {
		return err
	}

	for _, user := range users {
		keys, err := gh.readOrGetUserPublicKeys(user)
		if err != nil {
			return err
		}
		PermitPublicKeys[user] = keys
	}

	return nil
}

func (gh *GitHub) readOrGetRepositoryMember(owner, repo string) ([]string, error) {
	users, err := database.ReadRepositoryUsers(owner + "/" + repo)
	if err != nil || len(users) == 0 {
		githubUsers, err := gh.GetMembers(owner, repo)
		if err != nil {
			return nil, err
		}
		userNames := make([]string, 0, len(githubUsers))
		for _, u := range githubUsers {
			userNames = append(userNames, *u.Login)
		}
		log.Printf("Save %s/%s user list", owner, repo)
		database.SaveRepositoryUsers(owner+"/"+repo, userNames)
		users = userNames
	}

	return users, nil
}

func (gh *GitHub) readOrGetUserPublicKeys(login string) ([]ssh.PublicKey, error) {
	keys, err := database.ReadPublicKeys(login)
	if err != nil || len(keys) == 0 {
		githubKeys, err := gh.GetPubkey(login)
		if err != nil {
			return nil, err
		}
		opensshKeys := make([]ssh.PublicKey, 0, len(githubKeys))
		for _, key := range githubKeys {
			pubKey, _, _, _, _ := ssh.ParseAuthorizedKey([]byte(*key.Key))
			opensshKeys = append(opensshKeys, pubKey)
		}
		log.Printf("Save %s's public keys", login)
		err = database.SavePubKey(login, opensshKeys)
		if err != nil {
			log.Print(err)
			return nil, err
		}

		keys = opensshKeys
	}

	return keys, nil
}
