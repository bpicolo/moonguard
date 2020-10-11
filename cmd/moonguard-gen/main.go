package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

type langConfig struct {
	execFlag string
}

var supportedLangs = map[string]langConfig{
	"go": langConfig{
		execFlag: "--go_out=plugins=grpc:",
	},
}

func buildGrpcCommand(sources []string, langs []string, outDir string) (*exec.Cmd, error) {
	grpcExec, err := exec.LookPath("protoc")
	if err != nil {
		return nil, fmt.Errorf("unable to find command `protoc` in path")
	}
	args := []string{grpcExec}

	for _, lang := range langs {
		cfg := supportedLangs[lang]
		flag := cfg.execFlag + path.Join(outDir, "/go")
		args = append(args, flag)
	}

	args = append(args, sources...)

	cmd := &exec.Cmd{
		Path:   grpcExec,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Args:   args,
	}

	return cmd, nil
}

func findInputSources(pattern string) ([]string, error) {
	if pattern == "" {
		return nil, fmt.Errorf("first argument must be your protobuf source path")
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("unable to find input sources: %s", err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("found no protobuf sources matching `%s`", matches)
	}

	return matches, nil
}

func validateLangs(langs []string) error {
	for _, lang := range langs {
		if _, ok = supportedLangs[lang]; ok != true {
			return fmt.Errorf("`%s` is not yet supported by the moonguard client generator", lang)
		}
	}
	return nil
}

func main() {
	app := &cli.App{
		Name:  "moonguard-gen",
		Usage: "Generate gRPC client libraries",
		Action: func(c *cli.Context) error {
			inputSourcesStr := c.Args().Get(0)
			sources, err := findInputSources(inputSourcesStr)
			if err != nil {
				return err
			}

			langs := c.StringSlice("languages")
			err = validateLangs(langs)
			if err != nil {
				return err
			}

			cmd, err := buildGrpcCommand(sources, langs, c.String("out"))
			if err != nil {
				return err
			}

			return cmd.Run()
		},

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "out",
				Usage:       "Output directory for generated clients",
				DefaultText: "./moonguard-clients",
			},
			&cli.StringSliceFlag{
				Name:     "languages",
				Aliases:  []string{"l"},
				Usage:    "Build gRPC clients for this set of languages",
				Required: true,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}