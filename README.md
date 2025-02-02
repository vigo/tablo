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
  -fdc string
    	field delimiter char to split the line input
  -ldc string
    	line delimiter char to split the input (default "\n")
  -output string
    	where to send output (default "stdout")
  -version
    	display version information
```

---

## Usage

```bash
echo "${PATH}" | tablo -ldc ":" # ":" as line delimeter

┌───────────────────────────────────────────────────────────────────────────────────┐
│ /opt/homebrew/opt/postgresql@16/bin                                               │
├───────────────────────────────────────────────────────────────────────────────────┤
│ /Users/vigo/.cargo/bin                                                            │
├───────────────────────────────────────────────────────────────────────────────────┤
│ /Users/vigo/.orbstack/bin                                                         │
└───────────────────────────────────────────────────────────────────────────────────┘
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
cat /etc/passwd | tablo -fdc=":"
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
┌────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ REPOSITORY                                              TAG       IMAGE ID       CREATED        SIZE   │
├────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ vigo/basichttpdebugger                                  latest    911f45e85b68   19 hours ago   12.7MB │
├────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ ghcr.io/vbyazilim/basichttpdebugger/basichttpdebugger   latest    b72784c93710   19 hours ago   12.7MB │
└────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

Or;

```bash
 docker images | tablo . -fdc " "
┌───────────────────────────────────────────────────────┬────────┬──────────────┬────┬─────────┬──────┬────────┐
│ REPOSITORY                                            │ TAG    │ IMAGE        │ ID │ CREATED │ SIZE │        │
├───────────────────────────────────────────────────────┼────────┼──────────────┼────┼─────────┼──────┼────────┤
│ vigo/basichttpdebugger                                │ latest │ 911f45e85b68 │ 19 │ hours   │ ago  │ 12.7MB │
├───────────────────────────────────────────────────────┼────────┼──────────────┼────┼─────────┼──────┼────────┤
│ ghcr.io/vbyazilim/basichttpdebugger/basichttpdebugger │ latest │ b72784c93710 │ 19 │ hours   │ ago  │ 12.7MB │
└───────────────────────────────────────────────────────┴────────┴──────────────┴────┴─────────┴──────┴────────┘
```

---

## Contributor(s)

* [Uğur Özyılmazel](https://github.com/vigo) - Creator, maintainer

---

## Contribute

All PR’s are welcome!

1. `fork` (https://github.com/vigo/tablo/fork)
1. Create your `branch` (`git checkout -b my-feature`)
1. `commit` yours (`git commit -am 'add some functionality'`)
1. `push` your `branch` (`git push origin my-feature`)
1. Than create a new **Pull Request**!

---

## License

This project is licensed under MIT (MIT)

---

This project is intended to be a safe, welcoming space for collaboration, and
contributors are expected to adhere to the [code of conduct][coc].

[coc]: https://github.com/vigo/tablo/blob/main/CODE_OF_CONDUCT.md
[001]: https://www.nushell.sh/