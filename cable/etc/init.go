package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// Check if PROJECT parameter is provided
	project := os.Getenv("PROJECT")
	if project == "" {
		fmt.Print("Enter project name (e.g. my-org/my-project): ")
		var input string
		_, err := fmt.Scanln(&input)

		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			os.Exit(1)
		}

		if input == "" {
			fmt.Println("Error: Project name is required")
			os.Exit(1)
		}
		project = input
	}

	// Get binary name from project
	binName := filepath.Base(project)

	// Replace in all .go files
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			replaceInFile(path, "github.com/palkan/twilio-ai-cable", "github.com/"+project)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error scanning .go files: %v\n", err)
		os.Exit(1)
	}

	// Replace in go.mod
	replaceInFile("go.mod", "github.com/palkan/twilio-ai-cable", "github.com/"+project)

	// Replace in go.sum
	replaceInFile("go.sum", "github.com/palkan/twilio-ai-cable", "github.com/"+project)

	// Replace in Makefile
	replaceInFile("Makefile", "PROJECT=anycable/mycable", "PROJECT="+project)

	// Rename cmd directory
	err = os.Rename("cmd/mycable", "cmd/"+binName)
	if err != nil {
		fmt.Printf("Error renaming cmd directory: %v\n", err)
		os.Exit(1)
	}

	// Run go mod tidy
	if err := execCommand("go", "mod", "tidy"); err != nil {
		fmt.Printf("Error running go mod tidy: %v\n", err)
		os.Exit(1)
	}

	// Run make lint
	if err := execCommand("make", "lint"); err != nil {
		fmt.Printf("Error running make lint: %v\n", err)
		os.Exit(1)
	}

	// Run make test
	if err := execCommand("make", "test"); err != nil {
		fmt.Printf("Error running make test: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Project renamed successfully!")
}

func replaceInFile(filename, old, new string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		os.Exit(1)
	}

	newContent := strings.ReplaceAll(string(content), old, new)

	err = os.WriteFile(filename, []byte(newContent), 0600)
	if err != nil {
		fmt.Printf("Error writing file %s: %v\n", filename, err)
		os.Exit(1)
	}
}

func execCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
