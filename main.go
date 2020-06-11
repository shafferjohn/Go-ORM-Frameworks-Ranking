package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

type Repo struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Stars         int    `json:"stargazers_count"`
	Forks         int    `json:"forks_count"`
	OpenIssues    int    `json:"open_issues_count"`
	URL           string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
	LastCommit    Commit `json:"-"`
}

type Commit struct {
	Commit struct {
		Author struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

var (
	Repos []Repo
	AccessToken = os.Getenv("ACCESS_TOKEN")
	GithubAPI = "https://api.github.com"
)

func main() {
	content, err := ioutil.ReadFile("repos.txt")
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(content), "\n")

	Repos = make([]Repo, 0)
	for _, line := range lines {
		var repo Repo
		var commit Commit

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		u, err := url.Parse(line)
		if err != nil {
			log.Fatal(err)
		}
		if u.Host != "github.com" {
			continue
		}
		paths := strings.Split(u.Path, "/")
		if len(paths) < 3 {
			continue
		}

		resp, err := http.Get(fmt.Sprintf("%s/repos/%s/%s?access_token=%s", GithubAPI, paths[1], paths[2], AccessToken))
		if err != nil || resp == nil {
			log.Fatal(err)
		}
		if resp.StatusCode != 200 {
			log.Fatal(resp.StatusCode)
		}

		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&repo)
		if err != nil {
			log.Fatal(err)
		}

		resp, err = http.Get(fmt.Sprintf("%s/repos/%s/%s/commits/%s?access_token=%s", GithubAPI, paths[1], paths[2], repo.DefaultBranch, AccessToken))
		if err != nil || resp == nil {
			log.Fatal(err)
		}
		if resp.StatusCode != 200 {
			log.Fatal(resp.StatusCode)
		}

		decoder = json.NewDecoder(resp.Body)
		err = decoder.Decode(&commit)
		if err != nil {
			log.Fatal(err)
		}
		repo.LastCommit = commit
		Repos = append(Repos, repo)
	}
	sort.Slice(Repos, func(i int, j int) bool {
		return Repos[i].Stars > Repos[j].Stars
	})
	save()
}

func save() {
	head := `Go ORM Frameworks Ranking List
==========

**a list of the most github stars repositories related to Go ORM**

*ranked by stars*

| Project Name | Stars | Forks | Open Issues | Description | Last Commit |
| ------------ | ----- | ----- | ----------- | ----------- | ----------- |
`
	readme, err := os.OpenFile("README.md", os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer readme.Close()
	readme.WriteString(head)
	for _, repo := range Repos {
		line := fmt.Sprintf("| [%s](%s) | %d | %d | %d | %s | %s |\n", repo.Name, repo.URL, repo.Stars, repo.Forks, repo.OpenIssues, repo.Description, repo.LastCommit.Commit.Author.Date.Format("2006-01-02 15:04:05"))
		readme.WriteString(line)
	}
	readme.WriteString(fmt.Sprintf("\n*Last Automatic Update Time: %v*", time.Now().Format("2006-01-02 15:04:05")))
}
