package main

import (
	"context"
	"fmt"
	"io/fs"
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
	Exclude *[]string
}

// var flagCfgPath = flag.String()
func main() {
	if err := mainErr(); err != nil {
		fmt.Print(err)
	}
}

func mainErr() error {
	// flag.Parse()

	if _, err := os.Stat("./ansipkl.toml"); err != nil {
		return err
	}

	cfgBytes, err := os.ReadFile("./ansipkl.toml")
	if err != nil {
		return err
	}

	var cfg cfg
	toml.Unmarshal(cfgBytes, &cfg)

	filepath.Walk("./", func(path string, info fs.FileInfo, err error) error {
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
		out, err := cmd.Output()
		if err != nil {
			return err
		}

		var m map[interface{}]interface{}
		if err := yaml.Unmarshal(out, &m); err != nil {
			return err
		}

		for k, v := range m {
			b, err := yaml.Marshal(&v)
			if err != nil {
				return err
			}

			asdf := filepath.Dir(path) + "/" + k.(string) + ".yml"
			file, err := os.Create(asdf)
			if err != nil {
				return err
			}
			defer file.Close()

            // todo condition
			{
				os.Chmod(asdf, 0755)
				if _, err := file.Write([]byte("#!/usr/bin/env ansible-playbook\n")); err != nil {
					return err
				}
			}

			if _, err := file.Write([]byte("---\n")); err != nil {
				return err
			}
			if _, err := file.Write(b); err != nil {
				return err
			}
		}

		return nil
	})

	return nil
}
