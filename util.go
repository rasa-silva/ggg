package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v2"
)

func TitleFromBody(body string) string {
	title := strings.Split(body, "\n")[0]
	title = strings.TrimPrefix(title, "#") //strip out the markdown formating for the title
	title = strings.TrimSpace(title)
	return title
}

func OpenTempFileInEditor(contents string) (string, error) {
	f, err := os.CreateTemp("", "ggg_note-*.md")
	if err != nil {
		return "", err
	}
	defer f.Close()

	editorEnv := os.Getenv("EDITOR")
	if editorEnv == "" {
		return "", errors.New("EDITOR not found")
	}
	editorPath, err := exec.LookPath(editorEnv)
	if err != nil {
		return "", err
	}

	LogOnVerbose(fmt.Sprint("$EDITOR is", editorPath))

	if contents != "" {
		err := os.WriteFile(f.Name(), []byte(contents), os.ModeAppend)
		if err != nil {
			return "", fmt.Errorf("could not write initial file content: %w", err.Error())
		}
	}

	cmd := exec.Command(editorPath, f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return "", err
	}
	err = cmd.Wait()
	if err != nil {
		return "", err
	}

	LogOnVerbose("Successfully edited.")

	return f.Name(), nil
}

func LogOnVerbose(msg string) {
	if verbose {
		fmt.Printf("%v %v\n", verbosePrefix, msg)
	}
}

func ExitWithError(err error, errorMsg string) cli.ExitCoder {
	LogOnVerbose(err.Error())
	return cli.Exit(fmt.Sprintf("%v %v", errorPrefix, errorMsg), 1)
}

func ExitWithMessage(errorMsg string) cli.ExitCoder {
	return cli.Exit(fmt.Sprintf("%v %v", errorPrefix, errorMsg), 1)
}
