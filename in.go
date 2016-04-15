// Package ingo provides a drop-in replacement for flag.Parse() with flags
// persisted in a user editable configuration file.
package ingo

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"strings"
	"unicode/utf8"
)

const updateWarning = `!!!!!!!!!!
! WARNING: %s was probably updated,
! Check and update %s as necessary
! and remove the last "deprecated" paragraph to disable this message!
!!!!!!!!!!
`
const configHeader = `# %s configuration
# 
# This config has https://github.com/schachmat/ingo syntax.
# Empty lines or lines starting with # will be ignored.
# All other lines must look like "KEY=VALUE" (without the quotes).
# The VALUE must not be enclosed in quotes as well!
`

var openOrCreate = os.OpenFile

// fallbackHome can be used to determine home if user.Current() fails.
func fallbackHome() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	// For Windows.
	return os.Getenv("UserProfile")
}

// Parse should be called by the user instead of `flag.Parse()` after all flags
// have been added. It will read the following sources with the given priority:
//   1. flags given on the command line
//   2. values read from the config file
//   3. default values from the  flags
// Then it will update the config file only using sources 2 and 3, so given flag
// values will not be persisted in the config file. The default value from the
// flag will only be used if the flag could not be found in the config file
// already. Values from the config file, which don't have a corresponding flag
// anymore will be appended in a special section at the end of the new version
// of the config file so their values won't get lost and a warning message will
// be printed to stderr.
//
// The location of the config file depends on the appName argument. An appName
// of `Ingo` would resolve to the config file path `$HOME/.ingorc`. This default
// location can also be overwritten temporarily by setting an environment
// variable like `INGORC` to point to the config file path.
func Parse(appName string) error {
	if flag.Parsed() {
		return fmt.Errorf("flags have been parsed already")
	}

	envname := strings.ToUpper(appName) + "RC"
	cPath := os.Getenv(envname)
	if cPath == "" {
		var homeDir string
		usr, err := user.Current()
		if err == nil {
			homeDir = usr.HomeDir
		} else {
			if homeDir = fallbackHome(); homeDir == "" {
				return fmt.Errorf("%v\nYou can set the environment variable %s to point to your config file as a workaround", err, envname)
			}
		}
		cPath = path.Join(homeDir, "."+strings.ToLower(appName)+"rc")
	}

	cf, err := openOrCreate(cPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("unable to open %s config file %v for reading and writing: %v", appName, cPath, err)
	}
	defer cf.Close()

	// read config to buffer and parse
	oldConf := new(bytes.Buffer)
	obsoleteKeys := parseConfig(io.TeeReader(cf, oldConf))
	if len(obsoleteKeys) > 0 {
		fmt.Fprintf(os.Stderr, updateWarning, appName, cPath)
	}

	// write updated config to another buffer
	newConf := new(bytes.Buffer)
	fmt.Fprintf(newConf, configHeader, appName)
	saveConfig(newConf, obsoleteKeys)

	// only write the file if it changed
	if !bytes.Equal(oldConf.Bytes(), newConf.Bytes()) {
		if ofs, err := cf.Seek(0, 0); err != nil || ofs != 0 {
			return fmt.Errorf("failed to seek to beginning of %s: %v", cPath, err)
		} else if err = cf.Truncate(0); err != nil {
			return fmt.Errorf("failed to truncate %s: %v", cPath, err)
		} else if _, err = newConf.WriteTo(cf); err != nil {
			return fmt.Errorf("failed to write %s: %v", cPath, err)
		}
	}

	flag.Parse()
	return nil
}

func parseConfig(r io.Reader) map[string]string {
	obsKeys := make(map[string]string)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}

		// find first assignment symbol and parse key, val
		i := strings.IndexAny(line, "=:")
		if i == -1 {
			continue
		}
		key, val := strings.TrimSpace(line[:i]), strings.TrimSpace(line[i+1:])

		if err := flag.Set(key, val); err != nil {
			obsKeys[key] = val
			continue
		}
	}
	return obsKeys
}

func saveConfig(w io.Writer, obsKeys map[string]string) {
	// find flags pointing to the same variable. We will only write the longest
	// named flag to the config file, the shorthand version is ignored.
	deduped := make(map[flag.Value]flag.Flag)
	flag.VisitAll(func(f *flag.Flag) {
		if cur, ok := deduped[f.Value]; !ok || utf8.RuneCountInString(f.Name) > utf8.RuneCountInString(cur.Name) {
			deduped[f.Value] = *f
		}
	})
	flag.VisitAll(func(f *flag.Flag) {
		if cur, ok := deduped[f.Value]; ok && cur.Name == f.Name {
			_, usage := flag.UnquoteUsage(f)
			usage = strings.Replace(usage, "\n    \t", "\n# ", -1)
			fmt.Fprintf(w, "\n# %s (default %v)\n", usage, f.DefValue)
			fmt.Fprintf(w, "%s=%v\n", f.Name, f.Value.String())
		}
	})

	// if we have obsolete keys left from the old config, preserve them in an
	// additional section at the end of the file
	if obsKeys != nil && len(obsKeys) > 0 {
		fmt.Fprintln(w, "\n\n# The following options are probably deprecated and not used currently!")
		for key, val := range obsKeys {
			fmt.Fprintf(w, "%v=%v\n", key, val)
		}
	}
}
