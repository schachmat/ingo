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
}

func saveConfig(configPath string) {
	fout, err := os.Create(configPath)
	if err != nil {
		log.Fatalf("Unable to open file %v for writing: %v", configPath, err)
	}
	defer fout.Close()

	writer := bufio.NewWriter(fout)
	defer writer.Flush()
	flag.VisitAll(func(f *flag.Flag) {
		_, usage := flag.UnquoteUsage(f)
		fmt.Fprintln(writer, "#", strings.Replace(usage, "\n", "\n# ", -1))
		fmt.Fprintf(writer, "%v=%v\n", f.Name, f.Value.String())
	})
	if len(obsoleteKeys) == 0 {
		return
	}
	fmt.Fprintln(writer, "\n\n# The following flags are not used anymore!")
	for key, val := range obsoleteKeys {
		fmt.Fprintf(writer, "%v=%v\n", key, val)
	}
}
