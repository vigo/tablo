![Version](https://img.shields.io/badge/version-0.0.0-orange.svg)

# tablo

`tablo` is a golang data pipeline tool inspired by [Nushell][001], where every
element is data and everything flows seamlessly through pipes.

---

## Installation

```bash
go install github.com/vigo/tablo@latest
```

Command line args:

```bash
tablo -h

Usage of tablo:
  -f string
    	field delimiter char to split the line input (short) (default " ")
  -field-delimeter-char string
    	field delimiter char to split the line input (default " ")
  -l string
    	line delimiter char to split the input (short) (default "\n")
  -line-delimeter-char string
    	line delimiter char to split the input (default "\n")
  -no-separate-rows
    	draw separation line under rows
  -nsr
    	draw separation line under rows (short)
  -o string
    	where to send output (short) (default "stdout")
  -output string
    	where to send output (default "stdout")
  -version
    	display version information
```

---

## Usage

```bash
echo "${PATH}" | tablo -l ":" # ":" as line delimeter
┌───────────────────────────────────────────────────────────────────────────────────┐
│ /opt/homebrew/opt/postgresql@16/bin                                               │
├───────────────────────────────────────────────────────────────────────────────────┤
│ /Users/vigo/.cargo/bin                                                            │
├───────────────────────────────────────────────────────────────────────────────────┤
│ /Users/vigo/.orbstack/bin                                                         │
└───────────────────────────────────────────────────────────────────────────────────┘
# output is trimmed...
```

You can disable row separation line with `-nsr` flag:

```bash
echo "${PATH}" | tablo -l ":" -nsr
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
```

Or check your `/etc/passwd`:

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

docker images | tablo -nsr REPOSITORY "IMAGE ID"
┌───────────────────────────────────────────────────────┬──────────────┐
│ REPOSITORY                                            │ IMAGE ID     │
├───────────────────────────────────────────────────────┼──────────────┤
│ vigo/basichttpdebugger                                │ 911f45e85b68 │
│ ghcr.io/vbyazilim/basichttpdebugger/basichttpdebugger │ b72784c93710 │
└───────────────────────────────────────────────────────┴──────────────┘
```

You can set output for save:

```bash
docker images | tablo -o /tmp/docker-images.txt REPOSITORY "IMAGE ID"
result saved to: /tmp/docker-images.txt

# verify
cat /tmp/docker-images.txt
```

---

## TODO

**2025-02-02**

- add `json`, `csv` support
- add `json`, `csv` filtering

---

## Change Log

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