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
	fmt.Println(startPath)
	filepath.Walk(startPath, func(path string, info os.FileInfo, e error) error {
		pathFromRepositoryRoot := strings.Split(path, repositoryName)[1]
		if filterByFileExtension(path) && filterByFolderName(path) && filterByFileName(pathFromRepositoryRoot) {
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

func filterByFileExtension(path string) bool {
	return strings.Contains(path, ".md")
}

func filterByFolderName(path string) bool {
	return !strings.Contains(path, ".github") && !strings.Contains(path, ".githooks")
}

func filterByFileName(path string) bool {
	return path != "/CODE_OF_CONDUCT.md" && path != "/CONTRIBUTING.md" && path != "/NOTICE.md" && path != "/README.md" && path != "/index.md"
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
