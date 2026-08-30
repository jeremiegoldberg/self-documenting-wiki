// wikigen generates a documentation index from what a forge already knows
// about its repositories.
//
// The idea is deliberately narrow: a wiki written by hand rots, because
// nothing forces it to follow the code. A wiki generated from the repository
// descriptions and topics cannot drift, because those live next to the code
// and are edited by the people who change it.
//
// Run it from CI, push the result, and put a banner on the page saying manual
// edits will be lost. That banner is the whole point.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/jeremiegoldberg/self-documenting-wiki/internal/forge"
)

type group struct {
	Topic string
	Repos []forge.Repo
}

func main() {
	var (
		provider = flag.String("provider", "gitlab", "gitlab or github")
		host     = flag.String("host", "gitlab.com", "gitlab host")
		group_   = flag.String("group", "", "gitlab group id or path")
		owner    = flag.String("owner", "", "github user or organisation")
		tmplPath = flag.String("template", "templates/wiki.md.tmpl", "template file")
		out      = flag.String("out", "-", "output file, or - for stdout")
		skip     = flag.String("skip", "archived,prototype", "comma separated topics to leave out")
		untitled = flag.String("untitled", "Uncategorised", "heading for repositories with no topic")
	)
	flag.Parse()

	if err := run(*provider, *host, *group_, *owner, *tmplPath, *out, *skip, *untitled); err != nil {
		fmt.Fprintln(os.Stderr, "wikigen:", err)
		os.Exit(1)
	}
}

func run(provider, host, groupID, owner, tmplPath, out, skip, untitled string) error {
	var client forge.Client
	switch provider {
	case "gitlab":
		if groupID == "" {
			return fmt.Errorf("-group is required for gitlab")
		}
		client = forge.GitLab{Host: host, GroupID: groupID, Token: os.Getenv("FORGE_TOKEN")}
	case "github":
		if owner == "" {
			return fmt.Errorf("-owner is required for github")
		}
		client = forge.GitHub{Owner: owner, Token: os.Getenv("FORGE_TOKEN")}
	default:
		return fmt.Errorf("unknown provider %q", provider)
	}

	repos, err := client.Repos()
	if err != nil {
		return err
	}

	groups := regroup(repos, splitSet(skip), untitled)

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return err
	}

	w := os.Stdout
	if out != "-" {
		if w, err = os.Create(out); err != nil {
			return err
		}
		defer w.Close()
	}
	return tmpl.Execute(w, groups)
}

// regroup sorts repositories under one heading per topic. A repository with
// several topics appears under each of them, which is usually what a reader
// wants: they look for "aws", not for the one true taxonomy.
func regroup(repos []forge.Repo, skip map[string]bool, untitled string) []group {
	byTopic := map[string][]forge.Repo{}
	for _, r := range repos {
		if r.Archived {
			continue
		}
		topics := r.Topics
		if len(topics) == 0 {
			topics = []string{untitled}
		}
		for _, t := range topics {
			if skip[strings.ToLower(t)] {
				continue
			}
			byTopic[t] = append(byTopic[t], r)
		}
	}

	out := make([]group, 0, len(byTopic))
	for topic, rs := range byTopic {
		sort.Slice(rs, func(i, j int) bool { return rs[i].Name < rs[j].Name })
		out = append(out, group{Topic: topic, Repos: rs})
	}
	// Untitled last: the reader wants the organised part first.
	sort.Slice(out, func(i, j int) bool {
		if (out[i].Topic == untitled) != (out[j].Topic == untitled) {
			return out[j].Topic == untitled
		}
		return out[i].Topic < out[j].Topic
	})
	return out
}

func splitSet(csv string) map[string]bool {
	set := map[string]bool{}
	for _, s := range strings.Split(csv, ",") {
		if s = strings.ToLower(strings.TrimSpace(s)); s != "" {
			set[s] = true
		}
	}
	return set
}
