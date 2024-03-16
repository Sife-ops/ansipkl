package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

type cfg struct {
	Options options
}

type options struct {
	Exclude     *[]string
	Executables *[]string
}

var flagCfg = flag.String("c", "./ansipkl.toml", "cfg path")
var flagVer = flag.Bool("v", false, "version")

func main() {
	if err := mainErr(); err != nil {
		log.Fatal(err)
	}
}

func mainErr() error {
	flag.Parse()

	if *flagVer {
		fmt.Print("0.0.8") // VERSION
		return nil
	}

	if _, err := os.Stat(*flagCfg); err != nil {
		return err
	}

	cfgBytes, err := os.ReadFile(*flagCfg)
	if err != nil {
		return err
	}

	var cfg cfg
	if err := toml.Unmarshal(cfgBytes, &cfg); err != nil {
		return err
	}

	err = filepath.Walk("./", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if cfg.Options.Exclude != nil {
			for _, p := range *cfg.Options.Exclude {
				match, err := regexp.MatchString(p, path)
				if err != nil {
					return err
				}
				if match {
					return nil
				}
			}
		}

		match, err := regexp.MatchString("^.*\\.pkl$", path)
		if err != nil {
			return err
		}
		if !match {
			return nil
		}

		cmd := exec.CommandContext(context.TODO(), "pkl", "eval", "-f", "yaml", path)
		cmd.Stderr = os.Stderr
		outBytes, err := cmd.Output()
		if err != nil {
			return err
		}

		var outputMap map[string]interface{}
		if err := yaml.Unmarshal(outBytes, &outputMap); err != nil {
			return err
		}

		playbookMap := map[string][]int{}
		for playbookName := range outputMap {
			scanner := bufio.NewScanner(bytes.NewReader(outBytes))
			i := 0
			var s *int
			var e *int
		Outer:
			for scanner.Scan() {
				if s != nil {
					for playbookName := range outputMap {
						if playbookName+":" == scanner.Text() {
							ii := i - 1
							e = &ii
							break Outer
						}
					}
				}
				if scanner.Text() == playbookName+":" {
					ii := i + 1
					s = &ii
				}
				i++
			}

			if e == nil {
				ii := i
				e = &ii
			}

			if s != nil && e != nil {
				playbookMap[playbookName] = []int{*s, *e}
			}
		}

		for playbookName, lines := range playbookMap {
			playbookPath := filepath.Dir(path) + "/" + playbookName + ".yml"
			file, err := os.Create(playbookPath)
			if err != nil {
				return err
			}
			defer file.Close()

			if cfg.Options.Executables != nil {
				for _, v := range *cfg.Options.Executables {
					match, err := regexp.MatchString(v, playbookName)
					if err != nil {
						return err
					}
					if !match {
						continue
					}
					os.Chmod(playbookPath, 0755)
					if _, err := file.Write([]byte("#!/usr/bin/env ansible-playbook\n")); err != nil {
						return err
					}
				}
			}

			if _, err := file.Write([]byte("---\n")); err != nil {
				return err
			}

			scanner := bufio.NewScanner(bytes.NewReader(outBytes))
			i := 0
			for scanner.Scan() {
				if i < lines[0] || i > lines[1] {
					i++
					continue
				}
				if _, err := file.Write(scanner.Bytes()); err != nil {
					return err
				}
				if _, err := file.Write([]byte("\n")); err != nil {
					return err
				}
				i++
			}
		}

		return nil
	})

	return err
}
