# :robot: Sonarqube PR Issues Review
[![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![Go Report Card][report-card-img]][report-card]

Simple Webhook for Sonarqube which publishes the issues found in the PR as review requesting changes.

The only SCM that is available for now is GitHub, feel free to open a PR if you want to add others.

## Getting started
First of all you need to setup the environment variables:

```
SONAR_API_KEY=SONAR_API_TOKEN
GH_TOKEN=GITHUB_API_TOKEN
SONAR_ROOT_URL=https://sonar-url-without-trailing-slash
WEBHOOK_SECRET=my-hook-secret # Not necessary if CLI
PORT=8080 # Not necessary if CLI
```
There are currently two ways to use this project:

### Webhook
To use the webhook you need to start the server by running:
```shell
$ sqpr
```

### CLI
This option is mostly to test, ideally you should use the webhook.

To use in the command line you can see the available flags by running `sqpr --help` (or `go run cmd/cli/main.go`):
```
$ sqpr -server --help

Usage of sqpr:
  -branch string
    	SCM Branch name (default "my-branch")
  -markaspublished
    	Mark the issue as published to avoid sending it again (default false)
  -project string
    	Sonarqube project name (default "my-project")
  -publish
    	Publish review
```

If you specify the `-branch` and `-project` it will print all the issues for the given branch:
```
$ sqpr -branch feat/newtest -project reynencourt_v3-vendor-onboarding

[OPEN] BUG: pkg/newwrong.go L5:
	- :bug: MAJOR: Refactor this piece of code to not have any dead code after this "return". ([go:S1763](https://sonar-url-without-trailing-slash/coding_rules?open=go:S1763&rule_key=go:S1763))
[OPEN] BUG: pkg/newwrong.go L12:
	- :bug: MAJOR: Refactor this piece of code to not have any dead code after this "return". ([go:S1763](https://sonar-url-without-trailing-slash/coding_rules?open=go:S1763&rule_key=go:S1763))

```

If you also specify the `-publish` it will print the issues and also create add the comments in the respective
PR for the given branch.

```
$ sqpr -branch feat/newtest -project reynencourt_v3-vendor-onboarding -publish

[OPEN] BUG: pkg/newwrong.go L5:
	- :bug: MAJOR: Refactor this piece of code to not have any dead code after this "return". ([go:S1763](https://sonar-url-without-trailing-slash/coding_rules?open=go:S1763&rule_key=go:S1763))
[OPEN] BUG: pkg/newwrong.go L12:
	- :bug: MAJOR: Refactor this piece of code to not have any dead code after this "return". ([go:S1763](https://sonar-url-without-trailing-slash/coding_rules?open=go:S1763&rule_key=go:S1763))
Issues review published!
```

The flag `-markaspublished` it will update the issue in the Sonarqube side, then the next time you run it
those marked issues will be filtered out. This is useful to avoid commenting about the same issues many
times in the PR.

```
$ sqpr -branch feat/newtest -project reynencourt_v3-vendor-onboarding -publish -markaspublished

[OPEN] BUG: pkg/newwrong.go L5:
	- :bug: MAJOR: Refactor this piece of code to not have any dead code after this "return". ([go:S1763](https://sonar-url-without-trailing-slash/coding_rules?open=go:S1763&rule_key=go:S1763))
[OPEN] BUG: pkg/newwrong.go L12:
	- :bug: MAJOR: Refactor this piece of code to not have any dead code after this "return". ([go:S1763](https://sonar-url-without-trailing-slash/coding_rules?open=go:S1763&rule_key=go:S1763))
Issues review published!
INFO[0000] --------------------------                   
INFO[0000] Mark as published result:                    
INFO[0000] 2 issues marked                              
INFO[0000] 0 issues ignored                             
INFO[0000] 0 issues failed                              
INFO[0000] --------------------------  
```

[doc-img]: http://img.shields.io/badge/GoDoc-Reference-blue.svg
[doc]: https://godoc.org/go.uber.org/fx

[ci-img]: https://github.com/uber-go/fx/actions/workflows/go.yml/badge.svg
[ci]: https://github.com/uber-go/fx/actions/workflows/go.yml

[cov-img]: https://codecov.io/gh/uber-go/fx/branch/master/graph/badge.svg
[cov]: https://codecov.io/gh/uber-go/fx/branch/master

[report-card-img]: https://goreportcard.com/badge/github.com/uber-go/fx
[report-card]: https://goreportcard.com/report/github.com/uber-go/fx
