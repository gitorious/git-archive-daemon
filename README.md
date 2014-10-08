# git-archive-daemon

[![Build Status](https://travis-ci.org/gitorious/git-archive-daemon.svg?branch=master)](https://travis-ci.org/gitorious/git-archive-daemon)

git-archive-daemon is a scalable, high-performance HTTP API for serving
archives of git repositories.

It utilizes `git archive` tool for actual archive generation.

## Features

* *Laziness* - Archives are generated on demand, when they're requested.
* *Caching* - Requests for the same combination of tree, prefix and format are
  served from a disk cache.
* *Work pooling* - Archiving is done by workers from a configurable, fixed size
  pool. This allows for putting predictable, limited load on the machine.
* *Request grouping* - When the archive is not cached then all requests for it
  are grouped together, waiting for the single archiving job to complete.
  This avoids duplicate work and allows git-archive-daemon to handle high
  volume of requests.

## Installation

Currently you need Go development environment to build git-archive-daemon.

The following command will fetch the package and build the binary at
`$GOPATH/bin/git-archive-daemon`:

    go get gitorious.org/gitorious/git-archive-daemon

## Usage

### Starting

Usage:

    git-archive-daemon [options]

Options:

* `-r <repos-dir>` - Directory containing git repositories, defaults to "."
* `-c <cache-dir>` - Cache dir for storing archives, defaults to "."
* `-t <tmp-dir>` - Tmp dir for archive generation, defaults to system tmp dir
* `-l <[addr]:port>` -  Address/port to listen on, defaults to 127.0.0.1:5000
* `-w <workers>` - Number of workers, defaults to 10

Example:

    git-archive-daemon -r /var/git/repositories -c /var/cache/archives -l :80

### API

    GET /<repo-path>[?params]

Params:

* `ref` - branch/tag name or commit sha
* `format` - tar.gz or zip
* `prefix` - (optional) prepended to each filename in the archive (passed to `git
  archive` via `--prefix` option)
* `filename` - (optional) filename for the response, returned in
  `Content-Disposition: attachment` HTTP header.

Example:

    GET /my-project/repo?ref=master&format=tar.gz&prefix=my-project/

This will generate and send tar.gz archive of master branch of repository at
`<repos-dir>/my-project/repo`.

## License

git-archive-daemon is free software licensed under the
[GNU Affero General Public License](http://www.gnu.org/licenses/agpl-3.0.html).
git-archive-daemon is developed as part of the Gitorious project.
