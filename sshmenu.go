package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v3"
)

// discardWriteCloser wraps io.Discard to satisfy io.WriteCloser
type discardWriteCloser struct{}

func (d discardWriteCloser) Write(p []byte) (int, error) {
	return io.Discard.Write(p)
}
func (d discardWriteCloser) Close() error { return nil }

// bellFilter filters out BEL ('\a') characters from the output stream.
type bellFilter struct {
	w io.Writer
}

// Close implements io.Closer for bellFilter, but does nothing.
func (b bellFilter) Close() error {
	return nil
}

func (b bellFilter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	// Remove all BEL characters
	filtered := make([]byte, 0, len(p))
	for _, c := range p {
		if c != '\a' {
			filtered = append(filtered, c)
		}
	}
	// Write the filtered bytes to the underlying writer.
	n, err := b.w.Write(filtered)
	// We return the length of the original slice (as Write contract), but if there's an error,
	// return that error. Many callers expect n == len(p); here we return the number of bytes
	// successfully "accepted" from the original slice; most callers ignore the exact n on success.
	if err != nil {
		return n, err
	}
	return len(p), nil
}

// Config structures

type Server struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Command     string `yaml:"command,omitempty"`
}

type Project struct {
	Name    string   `yaml:"name"`
	Servers []Server `yaml:"servers"`
}

type Config struct {
	GlobalCommand string    `yaml:"global_command"`
	Projects      []Project `yaml:"projects"`
}

func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg Config
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	// Clear the terminal screen at the start
	fmt.Print("\033[2J\033[H")

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error determining working directory:", err)
		os.Exit(1)
	}
	configPath := cwd + string(os.PathSeparator) + "sshmenu.yaml"
	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	// Create a bell-filtered writer that wraps the real stdout
	filteredStdout := bellFilter{w: os.Stdout}

	searchString := ""
	// Support search parameter from command line
	if len(os.Args) > 1 {
		searchString = strings.Join(os.Args[1:], " ")
	}
	for {
		// Reload config for inplace update
		cfg, err = loadConfig(configPath)
		if err != nil {
			fmt.Println("Error loading config:", err)
			os.Exit(1)
		}

		if searchString != "" {
			// Flat filtered list
			flatServers := []Server{}
			for _, p := range cfg.Projects {
				for _, s := range p.Servers {
					if strings.Contains(strings.ToLower(s.Name), strings.ToLower(searchString)) || strings.Contains(strings.ToLower(s.Description), strings.ToLower(searchString)) {
						flatServers = append(flatServers, s)
					}
				}
			}
			if len(flatServers) == 0 {
				fmt.Println("No servers found matching:", searchString)
				return
			}
			serverNames := []string{}
			for _, s := range flatServers {
				serverNames = append(serverNames, s.Name+" - "+s.Description)
			}
			serverNames = append(serverNames, "\u23FB Quit") // ⏻ Quit

			// Select server from flat list
			serverPrompt := promptui.Select{
				Label:    "Select Server",
				Items:    serverNames,
				HideHelp: true,
				Size:     50,
				Stdout:   filteredStdout,
			}
			sidx, sresult, err := serverPrompt.Run()
			fmt.Print("\r\033[K")
			if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
				fmt.Println("Exiting.")
				return
			}
			if err != nil {
				fmt.Println("Prompt failed:", err)
				return
			}
			if sresult == "\u23FB Quit" {
				fmt.Println("Exiting.")
				return
			}
			// Only proceed if a real server was selected
			if sidx < 0 || sidx >= len(flatServers) {
				return
			}
			server := flatServers[sidx]
			cmdStr := server.Command
			if cmdStr == "" {
				cmdStr = cfg.GlobalCommand
				cmdStr = replaceServer(cmdStr, server.Name)
			}
			fmt.Println("Running:", cmdStr)
			cmd := exec.Command("bash", "-c", cmdStr)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				fmt.Println("Command failed:", err)
			}
			return
		} else {
			// Two-step UI: project → server
			projectNames := []string{}
			for _, p := range cfg.Projects {
				projectNames = append(projectNames, p.Name)
			}
			projectNames = append(projectNames, "\u23FB Quit") // ⏻ Quit

			fmt.Println("Use ↑/↓ to navigate, Enter to select. Select '⏻ Quit' to exit.")
			projectPrompt := promptui.Select{
				Label:    "Select Project",
				Items:    projectNames,
				HideHelp: true,
				Size:     50,
				Stdout:   filteredStdout,
			}
			pidx, presult, err := projectPrompt.Run()
			fmt.Print("\r\033[K")
			if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
				fmt.Println("Exiting.")
				return
			}
			if err != nil {
				fmt.Println("Prompt failed:", err)
				return
			}
			if presult == "\u23FB Quit" {
				fmt.Println("Exiting.")
				return
			}
			// Only proceed if a real project was selected
			if pidx < 0 || pidx >= len(cfg.Projects) {
				return
			}
			project := cfg.Projects[pidx]
			serverNames := []string{}
			for _, s := range project.Servers {
				serverNames = append(serverNames, s.Name+" - "+s.Description)
			}
			serverNames = append(serverNames, "\u2B05 Back") // ⬅ Back
			for {
				serverPrompt := promptui.Select{
					Label:    "Select Server",
					Items:    serverNames,
					HideHelp: true,
					Size:     50,
					Stdout:   filteredStdout,
				}
				sidx, sresult, err := serverPrompt.Run()
				fmt.Print("\r\033[K")
				if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
					fmt.Println("Exiting.")
					return
				}
				if err != nil {
					fmt.Println("Prompt failed:", err)
					break
				}
				if sresult == "\u2B05 Back" {
					// Return to project selection
					goto ProjectSelect
				}
				// Only proceed if a real server was selected
				if sidx < 0 || sidx >= len(project.Servers) {
					continue
				}
				server := project.Servers[sidx]
				cmdStr := server.Command
				if cmdStr == "" {
					cmdStr = cfg.GlobalCommand
					cmdStr = replaceServer(cmdStr, server.Name)
				}
				fmt.Println("Running:", cmdStr)
				cmd := exec.Command("bash", "-c", cmdStr)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Stdin = os.Stdin
				if err := cmd.Run(); err != nil {
					fmt.Println("Command failed:", err)
				}
			}
			// After server selection, exit
			return
ProjectSelect:
			// Restart project selection loop
			continue
		}
	}
}

func replaceServer(template, server string) string {
	return stringReplace(template, "{server}", server)
}

func stringReplace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}
