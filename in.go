package ingo

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"
	"unicode/utf8"
)

var (
	obsoleteKeys = make(map[string]string)
)

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
		usr, err := user.Current()
		if err != nil {
			return fmt.Errorf("%v\nYou can set the environment variable %s to point to your config file as a workaround", err, envname)
		}
		cPath = path.Join(usr.HomeDir, "."+strings.ToLower(appName)+"rc")
	}

	if err := loadConfig(appName, cPath); err != nil {
		return err
	}
	if err := saveConfig(appName, cPath); err != nil {
		return err
	}
	flag.Parse()
	return nil
}

func loadConfig(appName, configPath string) error {
	fin, err := os.Open(configPath)
	if _, ok := err.(*os.PathError); ok {
		fmt.Fprintf(os.Stderr, "No config file found for %s. Creating %s ...\n", appName, configPath)
		return nil
	} else if err != nil {
		return fmt.Errorf("Unable to read %s config file %v: %v", appName, configPath, err)
	}
	defer fin.Close()

	scanner := bufio.NewScanner(fin)
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
			obsoleteKeys[key] = val
			continue
		}
	}
	return nil
}

func saveConfig(appName, configPath string) error {
	fout, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("Unable to open %s config file %v for writing: %v", appName, configPath, err)
	}
	defer fout.Close()

	writer := bufio.NewWriter(fout)
	defer writer.Flush()

	// header
	fmt.Fprintf(writer, "# %s configuration\n# \n", appName)
	fmt.Fprintln(writer, "# This config has https://github.com/schachmat/ingo syntax.")
	fmt.Fprintln(writer, "# Empty lines or lines starting with # will be ignored.")
	fmt.Fprintln(writer, "# All other lines must look like `KEY=VALUE` (without the quotes).")
	fmt.Fprintln(writer, "# The VALUE must not be enclosed in quotes!")

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
			fmt.Fprintf(writer, "\n# %s (default %v)\n", strings.Replace(usage, "\n    \t", "\n# ", -1), f.DefValue)
			fmt.Fprintf(writer, "%v=%v\n", f.Name, f.Value.String())
		}
	})

	// if we have obsolete keys left from the old config, preserve them in an
	// additional section at the end of the file
	if len(obsoleteKeys) == 0 {
		return nil
	}
	fmt.Fprintln(os.Stderr, "!!!!!!!!!!")
	fmt.Fprintln(os.Stderr, "! WARNING: The application was probably updated,")
	fmt.Fprintln(os.Stderr, "! Check and update", configPath, " as necessary and")
	fmt.Fprintln(os.Stderr, "! remove the last \"deprecated\" paragraph to disable this message!")
	fmt.Fprintln(os.Stderr, "!!!!!!!!!!")
	fmt.Fprintln(writer, "\n\n# The following options are probably deprecated and not used currently!")
	for key, val := range obsoleteKeys {
		fmt.Fprintf(writer, "%v=%v\n", key, val)
	}
	return nil
}
