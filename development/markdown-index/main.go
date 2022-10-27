package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var repositoryName = "test-infra"

func main() {
	f, err := os.Create("index.md")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	startPath, err := os.Getwd()
	filepath.Walk(startPath, func(path string, info os.FileInfo, e error) error {
		pathFromRepositoryRoot := strings.Split(path, repositoryName)[1]
		if strings.Contains(path, ".md") && !strings.Contains(path, ".github") &&
			!strings.Contains(path, ".githooks") && pathFromRepositoryRoot != "/CODE_OF_CONDUCT.md" &&
			pathFromRepositoryRoot != "/CONTRIBUTING.md" && pathFromRepositoryRoot != "/NOTICE.md" &&
			pathFromRepositoryRoot != "/README.md" && pathFromRepositoryRoot != "/index.md" {

			mdLine := getDescription(path) + "\n[" + pathFromRepositoryRoot + "](" + pathFromRepositoryRoot + ")\n\n"
			fmt.Println(mdLine)
			//write line to file
			_, err = f.WriteString(mdLine)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("ERROR:", err)
	}
}

func getDescription(path string) string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return "# " + strings.Split(path, repositoryName)[1]
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)

	fileScanner.Split(bufio.ScanLines)

	var description = ""
	for fileScanner.Scan() {
		if len(description) == 0 && strings.Contains(fileScanner.Text(), "#") {
			description = fileScanner.Text() + "\n"
		} else if len(description) > 0 && !strings.Contains(fileScanner.Text(), "#") && len(fileScanner.Text()) > 0 {
			description += fileScanner.Text() + "\n"
			break
		}
	}

	if len(description) > 0 {
		return description
	}
	return "# " + strings.Split(path, repositoryName)[1]
}
