#ingo

ingo is a simple Go library helping you to persist flags in a ini-like config
file.

##Features and limitations

* Requires Go 1.5
* automatically creates config file, if it does not exist yet
* option value priorities (from high to low):
  0. flags given on the commandline
  0. flags read from the config file
  0. defaults given when flags are initialized
* write defaults to config file, if they are not set there already
* every flag in the config file has the flag usage prepended as a comment
* old flags, which are not used anymore are not removed
* flags must not contain the characters `:` and `=` as they are used as
  separator in the config file
* no sections, namespaces or FlagSets other than the default one

##Installation

```shell
go get -u github.com/schachmat/ingo
```

##Usage example

Just setup your flags with defaults like you are used to do. Instead of
`flag.Parse()` you have to call `ingo.Parse(CONFIG_FILE_PATH)`. Thats all.

```go
package main

import (
	"flag"
	"fmt"
	"github.com/schachmat/ingo"
)

func main() {
	num := flag.Int("num", 3, "`NUMBER` of times to\n    \tdo a barrel roll")
	location := flag.String("location", "space", "`WHERE` to do the barrel roll")
	ingo.Parse("lol")
	fmt.Println(*num, *location)
}
```

The `\n    \t` separator (one newline, four spaces, one tab) will ensure that
multi-line usage strings will be lay out correctly in the config file *and* in
the `-h` help message. The code will create the following config file `lol`:

```shell
# WHERE to do the barrel roll
location=space
# NUMBER of times to
# do a barrel roll
num=3
```

If you change num to 5 in the config file, it will be persistent on all future
runs:

```shell
# WHERE to do the barrel roll
location=space
# NUMBER of times to
# do a barrel roll
num=5
```

If you add a new flag `style` to your programm, it will be added to the config
file on the first run using the default value from the flag:

```shell
# WHERE to do the barrel roll
location=space
# NUMBER of times to
# do a barrel roll
num=5
# HOW to do the barrel roll
style=epic
```

If you remove the location flag from your programm, the config entry will be
rewritten to this:

```shell
# NUMBER of times to
# do a barrel roll
num=5
# HOW to do the barrel roll
style=epic


# The following flags are not used anymore!
location=space
```

##License - ISC

Copyright (c) 2016,  <teichm@in.tum.de>

Permission to use, copy, modify, and/or distribute this software for any purpose
with or without fee is hereby granted, provided that the above copyright notice
and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
THIS SOFTWARE.
