# self-documenting-wiki

> Documentation that cannot drift, because none of it is written by hand: the
> index is generated from what your forge already knows about its repositories.

![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)
![Dependencies](https://img.shields.io/badge/dependencies-none-brightgreen)
![Forges](https://img.shields.io/badge/forges-GitLab%20%C2%B7%20GitHub-FC6D26?logo=gitlab&logoColor=white)
![Licence](https://img.shields.io/badge/licence-MIT-blue)

A small Go tool that generates a documentation index from what a code forge
already knows about its repositories: their description and their topics.

No database, no wiki editor, no dependency outside the standard library.

## Why

I have watched the same thing happen in every team I have worked with.

Documentation goes through three filters, and very little comes out the other
end. Most of the time it is never written, because it is planned for the end of
a project and the end of a project does not exist. When it is written, nobody
can find it: an old wiki, the space of a team that no longer exists, a Markdown
file in a tooling repository nobody opens. And when it is found, it lies,
because it describes a platform that changed eighteen months ago.

The third case is the dangerous one. Nobody distrusts a document that looks
tidy.

The way out is not to write more documentation. It is to stop writing the part
that can be derived. A repository already carries a description and a set of
topics, they live next to the code, and they are edited by the people who
change it. That is a source of truth. This tool turns it into a page.

The generated page carries a banner saying that manual edits will be lost.
**That banner is the point.** It is what stops the page from slowly becoming a
hand-maintained document again.

## Usage

```sh
# a GitLab group, including its subgroups
wikigen -provider gitlab -group 1234567 -out wiki/home.md

# a self-hosted GitLab
wikigen -provider gitlab -host gitlab.example.com -group platform -out wiki/home.md

# a GitHub user or organisation
wikigen -provider github -owner jeremiegoldberg -out wiki/home.md
```

For private repositories, set `FORGE_TOKEN`. Public groups need no token at all.

| Flag | Default | What it does |
|---|---|---|
| `-provider` | `gitlab` | `gitlab` or `github` |
| `-host` | `gitlab.com` | GitLab host, for self-hosted instances |
| `-group` | | GitLab group id or path, subgroups included |
| `-owner` | | GitHub user or organisation |
| `-template` | `templates/wiki.md.tmpl` | Go template used to render the page |
| `-out` | `-` | Output file, `-` for stdout |
| `-skip` | `archived,prototype` | Topics to leave out of the page |
| `-untitled` | `Uncategorised` | Heading for repositories with no topic |

Archived repositories are always excluded.

A repository with several topics appears under each of them. That is deliberate:
a reader looks for "aws", not for the one true taxonomy.

## Running it from CI

The point is that nobody runs this by hand. Regenerate on a schedule, and on
every push to the default branch.

```yaml
# .gitlab-ci.yml
generate-wiki:
  image: golang:1.22
  rules:
    - if: $CI_PIPELINE_SOURCE == "schedule"
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
  script:
    - go run ./cmd/wikigen -provider gitlab -group "$GITLAB_GROUP_ID" -out home.md
    - git clone "https://git:${WIKI_TOKEN}@${CI_SERVER_HOST}/${CI_PROJECT_PATH}.wiki.git" wiki
    - cp home.md wiki/
    - cd wiki
    - git add home.md
    - git diff --cached --quiet || git commit -m "regenerate wiki"
    - git push
```

`git diff --cached --quiet ||` matters: without it you get a commit on every
run, and a history nobody can read.

## Making it useful

The tool is worth exactly as much as the descriptions it reads. Two habits make
the difference:

- **A description on every repository.** One sentence, what it is for. If a
  repository cannot be described in one sentence, that is a finding in itself.
- **Topics as a real taxonomy.** Pick a small, closed list — cloud provider,
  layer, lifecycle — and hold it. Twenty ad-hoc topics produce a page nobody
  reads.

Neither is enforced by the tool, on purpose. It reports what exists. An empty
page means the repositories have nothing to say about themselves.

## Licence

MIT. See [LICENSE](LICENSE).
