package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sigs.k8s.io/yaml"

	"github.com/spf13/cobra"
)

const version = "v0.1.0"

var verbose bool

type SessionSpecification struct {
	Session string   `yaml:"session"`
	Windows []Window `yaml:"windows"`
}

type Window struct {
	Name   string `yaml:"name"`
	Layout string `yaml:"layout,omitempty"`
	Panes  []Pane `yaml:"panes"`
}

type Pane struct {
	Command string `yaml:"command"`
	Path    string `yaml:"path,omitempty"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:     "sesh [spec.yaml]",
		Short:   "sesh - CLI tool to declaratively launch tmux sessions from a YAML spec",
		Args:    cobra.ExactArgs(1),
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			run(args[0])
		},
	}

	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	rootCmd.SetVersionTemplate("sesh {{.Version}}\n")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(specPath string) {
	f, err := os.Open(specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open spec: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read spec: %v\n", err)
		os.Exit(1)
	}

	var spec SessionSpecification
	if err := yaml.Unmarshal(data, &spec); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse YAML: %v\n", err)
		os.Exit(1)
	}

	if err := tmux("new-session", "-d", "-s", spec.Session, "-n", spec.Windows[0].Name); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create session: %v\n", err)
		os.Exit(1)
	}

	for i, w := range spec.Windows {
		windowTarget := fmt.Sprintf("%s:%d", spec.Session, i)
		if i != 0 {
			args := []string{"new-window", "-t", spec.Session, "-n", w.Name}
			if err := tmux(args...); err != nil {
				fmt.Fprintf(os.Stderr, "failed to create window %s: %v\n", w.Name, err)
				os.Exit(1)
			}
		}

		for j, p := range w.Panes {
			paneTarget := fmt.Sprintf("%s.%d", windowTarget, j)
			if j != 0 {
				splitArgs := []string{"split-window", "-t", windowTarget, "-h"}
				if p.Path != "" {
					splitArgs = append(splitArgs, "-c", p.Path)
				}
				if err := tmux(splitArgs...); err != nil {
					fmt.Fprintf(os.Stderr, "failed to split pane in window %s: %v\n", w.Name, err)
					os.Exit(1)
				}
			}
			sendCmd := fmt.Sprintf("cd %s && %s", p.Path, p.Command)
			if p.Path == "" {
				sendCmd = p.Command
			}
			if err := tmux("send-keys", "-t", paneTarget, sendCmd, "C-m"); err != nil {
				fmt.Fprintf(os.Stderr, "failed to send command to pane in window %s: %v\n", w.Name, err)
				os.Exit(1)
			}
		}
		if w.Layout != "" {
			if err := tmux("select-layout", "-t", windowTarget, w.Layout); err != nil {
				fmt.Fprintf(os.Stderr, "failed to set layout for window %s: %v\n", w.Name, err)
				os.Exit(1)
			}
		}
	}
}

func tmux(args ...string) error {
	if verbose {
		fmt.Printf("tmux %v\n", args)
	}
	cmd := exec.Command("tmux", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
