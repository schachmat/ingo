package ingo

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	obsoleteKeys = make(map[string]string)
)

func Parse(configPath string) {
	if flag.Parsed() {
		return
	}

	loadConfig(configPath)

	saveConfig(configPath)

	flag.Parse()
}

func loadConfig(configPath string) {
	fin, err := os.Open(configPath)
	if _, ok := err.(*os.PathError); ok {
		log.Printf("No config file found. Creating %s ...", configPath)
		return
	} else if err != nil {
		log.Fatalf("Unable to read config file %v: %v", configPath, err)
	}
	defer fin.Close()

	scanner := bufio.NewScanner(fin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}

		split := strings.SplitN(line, "=", 2)
		if len(split) != 2 {
			continue
		}
		key, val := split[0], split[1]

		if err := flag.Set(key, val); err != nil {
			obsoleteKeys[key] = val
			continue
		}
	}
}

func saveConfig(configPath string) {
	fout, err := os.Create(configPath)
	if err != nil {
		log.Fatalf("Unable to open file %v for writing: %v", configPath, err)
	}
	defer fout.Close()

	writer := bufio.NewWriter(fout)
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Fprintln(writer, "# %v", strings.Replace(f.Usage, "\n", "\n# ", -1))
		fmt.Fprintln(writer, "%v=%v", f.Name, f.Value.String())
	})
	fmt.Fprintln(writer, "\n\n# The following flags are not used anymore!")
	for key, val := range obsoleteKeys {
		fmt.Fprintln(writer, "%v=%v", key, val)
	}
	writer.Flush()
}
