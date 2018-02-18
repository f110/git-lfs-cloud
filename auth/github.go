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
		users, err := database.ReadRepositoryUsers(v.Owner + "/" + v.Repo)
		if err != nil || len(users) == 0 {
			githubUsers, err := gh.GetMembers(v.Owner, v.Repo)
			if err != nil {
				return err
			}
			userNames := make([]string, 0, len(githubUsers))
			for _, u := range githubUsers {
				userNames = append(userNames, *u.Login)
			}
			database.SaveRepositoryUsers(v.Owner+"/"+v.Repo, userNames)
			users = userNames
		}

		for _, user := range users {
			if _, ok := PermitPublicKeys[user]; ok == false {
				PermitPublicKeys[user] = make([]ssh.PublicKey, 0)
			}
			keys, err := database.ReadPublicKeys(user)
			if err != nil || len(keys) == 0 {
				githubKeys, err := gh.GetPubkey(user)
				if err != nil {
					return err
				}
				opensshKeys := make([]ssh.PublicKey, 0, len(githubKeys))
				for _, key := range githubKeys {
					pubKey, _, _, _, _ := ssh.ParseAuthorizedKey([]byte(*key.Key))
					opensshKeys = append(opensshKeys, pubKey)
				}
				err = database.SavePubKey(user, opensshKeys)
				if err != nil {
					log.Print(err)
					continue
				}
			} else {
				PermitPublicKeys[user] = append(PermitPublicKeys[user], keys...)
			}
		}
	}
	return nil
}

func (gh *GitHub) GetMembers(owner string, repo string) ([]*github.User, error) {
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

func (gh *GitHub) GetPubkey(login string) ([]*github.Key, error) {
	opt := &github.ListOptions{PerPage: 100}
	keys, _, err := gh.client.Users.ListKeys(context.Background(), login, opt)
	return keys, err
}
