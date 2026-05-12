![Version](https://img.shields.io/badge/version-0.3.0-orange.svg)
[![Documentation](https://godoc.org/github.com/vigo/tablo?status.svg)](https://pkg.go.dev/github.com/vigo/tablo)
[![Run go tests](https://github.com/vigo/tablo/actions/workflows/go-test.yml/badge.svg)](https://github.com/vigo/tablo/actions/workflows/go-test.yml)
[![Run golangci-lint](https://github.com/vigo/tablo/actions/workflows/go-lint.yml/badge.svg)](https://github.com/vigo/tablo/actions/workflows/go-lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vigo/tablo)](https://goreportcard.com/report/github.com/vigo/tablo)
[![codecov](https://codecov.io/github/vigo/tablo/graph/badge.svg?token=Q8ACC1DLGK)](https://codecov.io/github/vigo/tablo)
![Powered by Rake](https://img.shields.io/badge/powered_by-rake-blue?logo=ruby)

# tablo

`tablo` is a golang data pipeline tool inspired by [Nushell][001], where every
element is data and everything flows seamlessly through pipes.

---

## Installation

### Via Homebrew

```bash
brew tap vigo/tablo
brew install tablo
```

### Via Go

```bash
go install github.com/vigo/tablo@latest
```

Command line args:

```bash
usage: tablo [-flags] [COLUMN] [COLUMN] [COLUMN]

  flags:

  -version                          display version information (X.X.X)
  --bash-completion                 print bash completion script
  -f, -field-delimiter-char         field delimiter char to split the line input
                                    (omit for smart split: 2+ whitespace)
  -l, -line-delimiter-char          line delimiter char to split the input
                                    (default: "\n")
  -n, -no-separate-rows             do not draw separation line under rows
  -nb, -no-borders                  do not draw borders
  -nh, -no-headers                  hide the selected or detected header row
  -fi, -filter-indexes              filter columns by index
  -j, -json                         render output as json
  -o, -output                       where to send output, can be file path or stdout
                                    (default "stdout")

  examples:

  $ tablo                                         # interactive mode
  $ echo "${PATH}" | tablo -l ":"
  $ echo "${PATH}" | tablo -l ":" -n
  $ echo "foo bar" | tablo -f " "                 # exact split: 2 cells
  $ echo "foo  bar" | tablo                       # smart split: 2 cells (no -f)
  $ cat /path/to/file | tablo
  $ cat /path/to/file | tablo -n
  $ cat /etc/passwd | tablo -f ":"
  $ cat /etc/passwd | tablo -f ":" -n
  $ cat /etc/passwd | tablo -n -f ":" -fi "1,5"   # show columns 1 and 5 only
  $ cat /etc/passwd | tablo -n -f ":" -nb nobody  # list users only (macos)
  $ cat /etc/passwd | tablo -n -f ":" -nb root    # list users only (linux)
  $ docker images | tablo
  $ docker images | tablo REPOSITORY              # show only REPOSITORY colum
  $ docker images | tablo REPOSITORY "IMAGE ID"   # show REPOSITORY and IMAGE ID colums
  $ docker images | tablo -j                       # render rows as json

  # save output to a file
  $ docker images | tablo -o /path/to/docker-images.txt REPOSITORY "IMAGE ID"

  # use default file redirection
  $ docker images | tablo REPOSITORY "IMAGE ID" > /path/to/docker-images.txt

  # csv files
  $ cat /path/to/file.csv | tablo -f ";"
  $ cat /path/to/file.csv | tablo -f ";" -n
  $ cat /path/to/file.csv | tablo -f ";" -n -nb
  $ cat /path/to/file.csv | tablo -f ";" -n -nb -nh
  $ cat /path/to/file.csv | tablo -f ";" -n -nb -nh <HEADER>

```

---

## Usage

Use `-l` or `-line-delimiter-char` for line delimiter.

```bash
echo "${PATH}" | tablo -l ":"       # ":" as line delimiter
┌───────────────────────────────────────────────────────────────────────────────────┐
│ /opt/homebrew/opt/postgresql@16/bin                                               │
├───────────────────────────────────────────────────────────────────────────────────┤
│ /Users/vigo/.cargo/bin                                                            │
├───────────────────────────────────────────────────────────────────────────────────┤
│ /Users/vigo/.orbstack/bin                                                         │
└───────────────────────────────────────────────────────────────────────────────────┘
# output is trimmed...
```

You can disable row separation line with `-n` or `-no-separate-rows` flag:

```bash
echo "${PATH}" | tablo -l ":" -n    # ":" as line delimiter, remove row separation
┌───────────────────────────────────────────────────────────────────────────────────┐
│ /opt/homebrew/opt/postgresql@16/bin                                               │
│ /Users/vigo/.cargo/bin                                                            │
└───────────────────────────────────────────────────────────────────────────────────┘
# output is trimmed...
```

Let’s say you have a text file under `/tmp/foo`:

```bash
cat /tmp/foo | tablo
┌────────────────┐
│ this is line 1 │
├────────────────┤
│ this is line 2 │
├────────────────┤
│ this is line   │
└────────────────┘

cat /tmp/foo | tablo -n
┌────────────────┐
│ this is line 1 │
│ this is line 2 │
│ this is line   │
└────────────────┘

cat /tmp/foo | go run . -n -nb
this is line 1    
this is line 2    
this is line
```

### Field Delimiter Modes

`tablo` has two field-splitting modes:

- **Smart mode** (when `-f` is omitted): splits on **2 or more consecutive
  whitespace characters**. This is ideal for aligned column output from tools
  like `docker images`, `ps aux`, `ls -l`, where cell values may contain a
  single space (e.g. `"2 weeks ago"`, `"Up 22 hours"`). When the input looks
  like delimited data (CSV/TSV or similar), `tablo` auto-detects common
  delimiters such as `,`, `;`, tab, and `|`.
- **Exact mode** (when `-f <char>` is given): splits on **each occurrence** of
  the given character. Consecutive delimiters preserve empty cells, matching
  CSV semantics.

Example contrasting both modes:

```bash
# smart mode (no -f)
echo "foo  bar" | tablo
┌─────┬─────┐
│ foo │ bar │
└─────┴─────┘

# exact mode (-f " " forces single-space split)
echo "foo bar" | tablo -f " "
┌─────┬─────┐
│ foo │ bar │
└─────┴─────┘

# exact mode preserves empty cells from consecutive delimiters
echo "a:b::c" | tablo -f ":"
┌───┬───┬──┬───┐
│ a │ b │  │ c │
└───┴───┴──┴───┘
```

Or check your `/etc/passwd`, use `-f` or `-field-delimiter-char` flag for
custom field delimiter:

```bash
cat /etc/passwd | tablo -f ":"
┌────────────────────────┬───┬─────┬─────┬─────────────────────────────────────────────────┬───────────────────────────────┬──────────────────┐
│ nobody                 │ * │ -2  │ -2  │ Unprivileged User                               │ /var/empty                    │ /usr/bin/false   │
├────────────────────────┼───┼─────┼─────┼─────────────────────────────────────────────────┼───────────────────────────────┼──────────────────┤
│ root                   │ * │ 0   │ 0   │ System Administrator                            │ /var/root                     │ /bin/sh          │
├────────────────────────┼───┼─────┼─────┼─────────────────────────────────────────────────┼───────────────────────────────┼──────────────────┤
│ daemon                 │ * │ 1   │ 1   │ System Services                                 │ /var/root                     │ /usr/bin/false   │
├────────────────────────┼───┼─────┼─────┼─────────────────────────────────────────────────┼───────────────────────────────┼──────────────────┤
│ _uucp                  │ * │ 4   │ 4   │ Unix to Unix Copy Protocol                      │ /var/spool/uucp               │ /usr/sbin/uucico │
├────────────────────────┼───┼─────┼─────┼─────────────────────────────────────────────────┼───────────────────────────────┼──────────────────┤
│ _taskgated             │ * │ 13  │ 13  │ Task Gate Daemon                                │ /var/empty                    │ /usr/bin/false   │
├────────────────────────┼───┼─────┼─────┼─────────────────────────────────────────────────┼───────────────────────────────┼──────────────────┤
│ _networkd              │ * │ 24  │ 24  │ Network Services                                │ /var/networkd                 │ /usr/bin/false   │
├────────────────────────┼───┼─────┼─────┼─────────────────────────────────────────────────┼───────────────────────────────┼──────────────────┤
│ _oahd                  │ * │ 441 │ 441 │ OAH Daemon                                      │ /var/empty                    │ /usr/bin/false   │
└────────────────────────┴───┴─────┴─────┴─────────────────────────────────────────────────┴───────────────────────────────┴──────────────────┘
# output is trimmed...

cat /etc/passwd | tablo -f ":" -n
┌────────────────────────┬───┬─────┬─────┬─────────────────────────────────────────────────┬───────────────────────────────┬──────────────────┐
│ nobody                 │ * │ -2  │ -2  │ Unprivileged User                               │ /var/empty                    │ /usr/bin/false   │
│ root                   │ * │ 0   │ 0   │ System Administrator                            │ /var/root                     │ /bin/sh          │
└────────────────────────┴───┴─────┴─────┴─────────────────────────────────────────────────┴───────────────────────────────┴──────────────────┘
# output is trimmed...
```

### Bash Completion

Load completion into your current shell:

```bash
source <(tablo --bash-completion)
```

Or save it and load it from your shell profile:

```bash
tablo --bash-completion > ~/.tablo-completion.bash
source ~/.tablo-completion.bash
```

The completion suggests flags, common delimiter values for `-f` / `-l`, file
paths for the input/output arguments, and column names after you pass an input
file:

```bash
tablo -f ";" ./users.csv <TAB>
tablo -f ";" ./users.csv "First name" <TAB>
```

If your input doesn’t have a kind of header, you can use `-fi` or `-filter-indexes`
flag to filter by column index. Index values are not **zero-based**, example
illustrates how to display first (1) and fifth (5) columns only:

```bash
cat /etc/passwd | tablo -f ":" -n -fi "1,5"
┌────────────────────────┬─────────────────────────────────────────────────┐
│ nobody                 │ Unprivileged User                               │
│ root                   │ System Administrator                            │
└────────────────────────┴─────────────────────────────────────────────────┘
```

Or;

```bash
docker images | tablo
┌───────────────────────────────────────────────────────┬────────┬──────────────┬──────────────┬────────┐
│ REPOSITORY                                            │ TAG    │ IMAGE ID     │ CREATED      │ SIZE   │
├───────────────────────────────────────────────────────┼────────┼──────────────┼──────────────┼────────┤
│ vigo/basichttpdebugger                                │ latest │ 911f45e85b68 │ 22 hours ago │ 12.7MB │
├───────────────────────────────────────────────────────┼────────┼──────────────┼──────────────┼────────┤
│ ghcr.io/vbyazilim/basichttpdebugger/basichttpdebugger │ latest │ b72784c93710 │ 22 hours ago │ 12.7MB │
└───────────────────────────────────────────────────────┴────────┴──────────────┴──────────────┴────────┘

docker images | tablo -j
[
  {
    "REPOSITORY": "vigo/basichttpdebugger",
    "TAG": "latest",
    "IMAGE ID": "911f45e85b68",
    "CREATED": "22 hours ago",
    "SIZE": "12.7MB"
  },
  {
    "REPOSITORY": "ghcr.io/vbyazilim/basichttpdebugger/basichttpdebugger",
    "TAG": "latest",
    "IMAGE ID": "b72784c93710",
    "CREATED": "22 hours ago",
    "SIZE": "12.7MB"
  }
]
```

You can also filter if your input has a kind of header row:

```bash
docker images | tablo REPOSITORY
┌───────────────────────────────────────────────────────┐
│ REPOSITORY                                            │
├───────────────────────────────────────────────────────┤
│ vigo/basichttpdebugger                                │
├───────────────────────────────────────────────────────┤
│ ghcr.io/vbyazilim/basichttpdebugger/basichttpdebugger │
└───────────────────────────────────────────────────────┘

docker images | tablo REPOSITORY "IMAGE ID"
┌───────────────────────────────────────────────────────┬──────────────┐
│ REPOSITORY                                            │ IMAGE ID     │
├───────────────────────────────────────────────────────┼──────────────┤
│ vigo/basichttpdebugger                                │ 911f45e85b68 │
├───────────────────────────────────────────────────────┼──────────────┤
│ ghcr.io/vbyazilim/basichttpdebugger/basichttpdebugger │ b72784c93710 │
└───────────────────────────────────────────────────────┴──────────────┘

docker images | tablo -n REPOSITORY "IMAGE ID"
┌───────────────────────────────────────────────────────┬──────────────┐
│ REPOSITORY                                            │ IMAGE ID     │
├───────────────────────────────────────────────────────┼──────────────┤
│ vigo/basichttpdebugger                                │ 911f45e85b68 │
│ ghcr.io/vbyazilim/basichttpdebugger/basichttpdebugger │ b72784c93710 │
└───────────────────────────────────────────────────────┴──────────────┘

docker images | tablo -n -nb REPOSITORY "IMAGE ID"
 REPOSITORY                                            │ IMAGE ID     
 vigo/basichttpdebugger                                │ 911f45e85b68 
 ghcr.io/vbyazilim/basichttpdebugger/basichttpdebugger │ b72784c93710 

docker images | tablo -n -nb -nh REPOSITORY "IMAGE ID"
 vigo/basichttpdebugger                                │ 911f45e85b68 
 ghcr.io/vbyazilim/basichttpdebugger/basichttpdebugger │ b72784c93710 
```

You have a `users.csv` file:

    Username;Identifier;First name;Last name
    booker12;9012;Rachel;Booker
    grey07;2070;Laura;Grey
    johnson81;4081;Craig;Johnson
    jenkins46;9346;Mary;Jenkins
    smith79;5079;Jamie;Smith

and you can;

```bash
cat /path/to/username.csv | tablo -f ";"
┌───────────┬────────────┬────────────┬───────────┐
│ Username  │ Identifier │ First name │ Last name │
├───────────┼────────────┼────────────┼───────────┤
│ booker12  │ 9012       │ Rachel     │ Booker    │
├───────────┼────────────┼────────────┼───────────┤
│ grey07    │ 2070       │ Laura      │ Grey      │
├───────────┼────────────┼────────────┼───────────┤
│ johnson81 │ 4081       │ Craig      │ Johnson   │
├───────────┼────────────┼────────────┼───────────┤
│ jenkins46 │ 9346       │ Mary       │ Jenkins   │
├───────────┼────────────┼────────────┼───────────┤
│ smith79   │ 5079       │ Jamie      │ Smith     │
└───────────┴────────────┴────────────┴───────────┘

cat /path/to/username.csv | tablo -f ";" -n
┌───────────┬────────────┬────────────┬───────────┐
│ Username  │ Identifier │ First name │ Last name │
│ booker12  │ 9012       │ Rachel     │ Booker    │
│ grey07    │ 2070       │ Laura      │ Grey      │
│ johnson81 │ 4081       │ Craig      │ Johnson   │
│ jenkins46 │ 9346       │ Mary       │ Jenkins   │
│ smith79   │ 5079       │ Jamie      │ Smith     │
└───────────┴────────────┴────────────┴───────────┘

cat /path/to/username.csv | tablo -f ";" -n -nh
┌───────────┬──────┬────────┬─────────┐
│ booker12  │ 9012 │ Rachel │ Booker  │
│ grey07    │ 2070 │ Laura  │ Grey    │
│ johnson81 │ 4081 │ Craig  │ Johnson │
│ jenkins46 │ 9346 │ Mary   │ Jenkins │
│ smith79   │ 5079 │ Jamie  │ Smith   │
└───────────┴──────┴────────┴─────────┘

cat /path/to/username.csv | tablo -f ";" -n username
┌───────────┐
│ Username  │
├───────────┤
│ booker12  │
│ grey07    │
│ johnson81 │
│ jenkins46 │
│ smith79   │
└───────────┘

cat /path/to/username.csv | tablo -f ";" -n -nh username
┌───────────┐
│ booker12  │
│ grey07    │
│ johnson81 │
│ jenkins46 │
│ smith79   │
└───────────┘

cat /path/to/username.csv | tablo -f ";" -n -nb -nh username
booker12  
grey07    
johnson81 
jenkins46 
smith79 
```

You can set output for save:

```bash
docker images | tablo -o /tmp/docker-images.txt REPOSITORY "IMAGE ID"
result saved to: /tmp/docker-images.txt

# verify
cat /tmp/docker-images.txt
```

---

## Rake Tasks

```bash
rake -T

rake bump[revision]  # bump version, default: patch, available: major,minor,patch
rake coverage        # show test coverage
rake test            # run test
```

---

## TODO

**2025-02-02**

- add sorting such as `-sort <FIELD>`
- add `ls` support, such as `-where -size > 10mb`, `-sort ...`

---

## Change Log

**2026-05-10**

- add `-j` / `-json` output mode
- emit arrays of objects when input has a header row, otherwise arrays of arrays

**2026-05-02**

- **breaking:** `-f <char>` now splits on each occurrence of the delimiter
  (exact mode); consecutive delimiters preserve empty cells. Previously, `-f`
  used a "longest-run" rule that collapsed runs of the same character into a
  single delimiter and required `-f " "` to behave like 2+ whitespace.
- smart split (2+ whitespace) is now triggered by **omitting** `-f` instead of
  passing `-f " "`. Pipe-from-aligned-output workflows
  (`docker images | tablo`, `ps aux | tablo`) keep working without any `-f`
  flag.
- remove `maxConsecutiveRepeats` heuristic; clarify field-delimiter help text;
  document field-delimiter modes in README.

**2025-02-13**

- better line delimiter handling
- improve test coverage
- add `-nb`, `-nh` flags

**2025-02-05**

- reduce complexity in code
- add column filtering by index
- improve test coverage
- bug fixes

**2025-02-02**

- initial release

---

## License

This project is licensed under MIT (MIT)

---

This project is intended to be a safe, welcoming space for collaboration, and
contributors are expected to adhere to the [code of conduct][coc].

[coc]: https://github.com/vigo/tablo/blob/main/CODE_OF_CONDUCT.md
[001]: https://www.nushell.sh/
