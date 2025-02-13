package tablo

import (
	"flag"
	"fmt"
	"os"
)

const usage = `usage: %[1]s [-flags] [COLUMN] [COLUMN] [COLUMN]

  flags:

  -version                          display version information (%s)
  -f, -field-delimiter-char         %s
                                    (default: "%s")
  -l, -line-delimiter-char          %s
                                    (default: "\n")
  -n, -no-separate-rows             %s
  -nb, -no-borders                  %s
  -nh, -no-headers                  %s
  -fi, -filter-indexes              %s
  -o, -output                       %s
                                    (default "stdout")

  examples:

  $ %[1]s                                         # interactive mode
  $ echo "${PATH}" | %[1]s -l ":"
  $ echo "${PATH}" | %[1]s -l ":" -n
  $ cat /path/to/file | %[1]s
  $ cat /path/to/file | %[1]s -n
  $ cat /etc/passwd | %[1]s -f ":"
  $ cat /etc/passwd | %[1]s -f ":" -n
  $ cat /etc/passwd | %[1]s -n -f ":" -fi "1,5"   # show columns 1 and 5 only
  $ cat /etc/passwd | %[1]s -n -f ":" -nb nobody  # list users only (macos)
  $ cat /etc/passwd | %[1]s -n -f ":" -nb root    # list users only (linux)
  $ docker images | %[1]s
  $ docker images | %[1]s REPOSITORY              # show only REPOSITORY colum
  $ docker images | %[1]s REPOSITORY "IMAGE ID"   # show REPOSITORY and IMAGE ID colums

  # save output to a file
  $ docker images | %[1]s -o /path/to/docker-images.txt REPOSITORY "IMAGE ID"

  # use default file redirection
  $ docker images | %[1]s REPOSITORY "IMAGE ID" > /path/to/docker-images.txt

  # csv files
  $ cat /path/to/file.csv | %[1]s -f ";"
  $ cat /path/to/file.csv | %[1]s -f ";" -n
  $ cat /path/to/file.csv | %[1]s -f ";" -n -nb
  $ cat /path/to/file.csv | %[1]s -f ";" -n -nb -nh
  $ cat /path/to/file.csv | %[1]s -f ";" -n -nb -nh <HEADER>

`

func getUsage() {
	binaryName := os.Args[0]
	versionInformation := Version

	if os.Getenv("PRINT_HELP_FOR_README") != "" {
		binaryName = "tablo"
		versionInformation = "X.X.X"
	}

	args := []any{
		binaryName,
		versionInformation,
		helpFieldDelimiterChar,
		string(defaultFieldDelimiter),
		helpLineDelimiterChar,
		helpNoSeparateRows,
		helpNoBorders,
		helpNoHeaders,
		helpFilterIndexes,
		helpOutput,
	}
	fmt.Fprintf(os.Stdout, usage, args...)

	if os.Getenv("PRINT_DEFAULTS") != "" {
		flag.PrintDefaults()
	}
}
