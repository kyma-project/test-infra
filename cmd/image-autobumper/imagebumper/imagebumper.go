/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package imagebumper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// imageRegexp matches image names with the following structure:
	// - The registry part can be:
	//     - gcr.io or docker.pkg.dev (optionally preceded by a subdomain)
	//     - e.g., `gcr.io`, `us.gcr.io`, `docker.pkg.dev`, `eu.docker.pkg.dev`
	// - The repository part:
	//     - Begins with a lowercase letter, followed by 5-29 lowercase letters, digits, or hyphens
	//     - e.g., `project-name`, `my-repo`
	// - The image name:
	//     - Begins with an alphanumeric character and can contain alphanumeric characters, underscores, dots, hyphens, and slashes
	//     - e.g., `my_image`, `some.image/name`
	// - The tag:
	//     - Contains alphanumeric characters, dots, underscores, or hyphens
	//     - e.g., `v1.0.0`, `latest`, `1.2.3-beta`
	imageRegexp = regexp.MustCompile(`\b((?:[a-z0-9]+\.)?gcr\.io|(?:[a-z0-9-]+)?docker\.pkg\.dev)/([a-z][a-z0-9-]{5,29}/[a-zA-Z0-9][a-zA-Z0-9_./-]+):([a-zA-Z0-9_.-]+)\b`)

	// tagRegexp matches version tags with the following structure:
	// - Version tags starting with an optional 'v' followed by an 8-digit date (YYYYMMDD)
	//     - Optionally followed by a dash, 'v', a version number, dash or dot-separated numeric parts, '-g', and a 6-10 character alphanumeric hash
	//     - e.g., `v20220714`, `20220714-v1.2.3-gabcdef1234`
	// - Alternatively, matches the string "latest"
	// - Optionally followed by a dash and any additional characters
	//     - e.g., `latest`, `v20220714-extra`
	tagRegexp = regexp.MustCompile(`(v?\d{8}-(?:v\d(?:[.-]\d+)*-g)?[0-9a-f]{6,10}|latest)(-.+)?`)
)

const (
	imageHostPartStart  = 0
	imageHostPartEnd    = 1
	imageImagePartStart = 2
	imageImagePartEnd   = 3
	imageTagPartStart   = 4
	imageTagPartEnd     = 5
)

type Client struct {
	// Keys are <imageHost>/<imageName>:<currentTag>. Values are corresponding tags.
	tagCache   map[string]string
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	// Shallow copy to adjust Timeout
	httpClientCopy := *httpClient
	httpClientCopy.Timeout = 1 * time.Minute

	return &Client{
		tagCache:   map[string]string{},
		httpClient: &httpClientCopy,
	}
}

type manifest map[string]struct {
	TimeCreatedMs string   `json:"timeCreatedMs"`
	Tags          []string `json:"tag"`
}

// commit | tag-n-gcommit
var commitRegexp = regexp.MustCompile(`^g?([\da-f]+)|(.+?)??(?:-(\d+)-g([\da-f]+))?$`)

// DeconstructCommit separates a git describe commit into its parts.
//
// Examples:
//
//	v0.0.30-14-gdeadbeef => (v0.0.30 14 deadbeef)
//	v0.0.30 => (v0.0.30 0 "")
//	deadbeef => ("", 0, deadbeef)
//
// See man git describe.
func DeconstructCommit(commit string) (string, int, string) {
	// Split the commit string by '-'
	parts := strings.Split(commit, "-")

	// If there's only one part, it can be a version tag or a commit hash
	if len(parts) == 1 {
		return "", 0, parts[0]
	}

	// If there are two parts, we need to handle it based on the second part
	if len(parts) == 2 {
		// If the second part starts with 'g', it's a git commit hash
		if strings.HasPrefix(parts[1], "g") {
			return parts[0], 0, parts[1][1:]
		}
		// Otherwise, it's a version tag
		return parts[0], 0, parts[1]
	}

	// If there are three parts, it should be in the form v0.0.30-14-gdeadbeef
	if len(parts) == 3 {
		// Parse the middle part as an integer (number of commits)
		n, err := strconv.Atoi(parts[1])
		if err != nil {
			panic(err) // Handle error appropriately in real code
		}
		// The last part should start with 'g' and followed by a commit hash
		if strings.HasPrefix(parts[2], "g") {
			return parts[0], n, parts[2][1:]
		}
	}

	// Fallback for unexpected formats
	return "", 0, ""
}

// DeconstructTag separates the tag into its vDATE-COMMIT-VARIANT components
//
// COMMIT may be in the form vTAG-NEW-gCOMMIT, use PureCommit to further process
// this down to COMMIT.
func DeconstructTag(tag string) (date, commit, variant string) {
	currentTagParts := tagRegexp.FindStringSubmatch(tag)
	if currentTagParts == nil {
		return "", "", ""
	}
	parts := strings.Split(currentTagParts[tagVersionPart], "-")
	return parts[0][1:], parts[len(parts)-1], currentTagParts[tagExtraPart]
}

// Constructs the URI for fetching the manifest
func constructManifestURI(imageHost, imageName string) string {
	return fmt.Sprintf("https://%s/v2/%s/tags/list", imageHost, imageName)
}

// Fetches the manifest for a given image from a registry
func (cli *Client) getManifest(imageHost, imageName string) (manifest, error) {
	uri := constructManifestURI(imageHost, imageName)
	resp, err := cli.httpClient.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("couldn't fetch tag list: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Manifest manifest `json:"manifest"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("couldn't parse tag information from registry: %w", err)
	}

	return result.Manifest, nil
}

// FindLatestTag returns the latest valid tag for the given image.
func (cli *Client) FindLatestTag(imageHost, imageName, currentTag string) (string, error) {
	k := imageHost + "/" + imageName + ":" + currentTag
	if result, ok := cli.tagCache[k]; ok {
		return result, nil
	}

	currentTagParts := tagRegexp.FindStringSubmatch(currentTag)
	if currentTagParts == nil {
		return "", fmt.Errorf("couldn't figure out the current tag in %q", currentTag)
	}
	if currentTagParts[tagVersionPart] == "latest" {
		return currentTag, nil
	}

	imageList, err := cli.getManifest(imageHost, imageName)
	if err != nil {
		return "", err
	}

	latestTag, err := pickBestTag(currentTagParts, imageList)
	if err != nil {
		return "", err
	}

	cli.tagCache[k] = latestTag

	return latestTag, nil
}

func (cli *Client) TagExists(imageHost, imageName, currentTag string) (bool, error) {
	imageList, err := cli.getManifest(imageHost, imageName)
	if err != nil {
		return false, err
	}

	for _, v := range imageList {
		for _, tag := range v.Tags {
			if tag == currentTag {
				return true, nil
			}
		}
	}

	return false, nil
}

// pickBestTag finds the most recently created image tag that matches the suffix of the current tag.
// If a tag called "latest" with the appropriate suffix is found, it is assumed to be the latest,
// regardless of its creation time.
func pickBestTag(currentTagParts []string, manifest manifest) (string, error) {
	var latestTime int64
	var latestTag string

	for _, entry := range manifest {
		var bestTag string
		var isLatest bool

		for _, tag := range entry.Tags {
			parts := tagRegexp.FindStringSubmatch(tag)
			if parts == nil {
				continue
			}
			if parts[tagExtraPart] != currentTagParts[tagExtraPart] {
				continue
			}
			if parts[tagVersionPart] == "latest" {
				isLatest = true
				continue
			}
			if bestTag == "" || len(tag) < len(bestTag) {
				bestTag = tag
			}
		}

		if bestTag == "" {
			continue
		}

		timeCreated, err := strconv.ParseInt(entry.TimeCreatedMs, 10, 64)
		if err != nil {
			return "", fmt.Errorf("couldn't parse timestamp %q: %w", entry.TimeCreatedMs, err)
		}

		if isLatest || timeCreated > latestTime {
			latestTime = timeCreated
			latestTag = bestTag
			if isLatest {
				break
			}
		}
	}

	if latestTag == "" {
		return "", fmt.Errorf("failed to find a suitable tag")
	}

	return latestTag, nil
}

// AddToCache keeps track of changed tags
func (cli *Client) AddToCache(image, newTag string) {
	cli.tagCache[image] = newTag
}

// updateAllTags updates all image tags in the content using the provided tagPicker function.
// It filters images based on the provided imageFilter regex, if not nil.
func updateAllTags(tagPicker func(host, image, tag string) (string, error), content []byte, imageFilter *regexp.Regexp) []byte {
	// Find all submatch indices for images in the content
	indexes := imageRegexp.FindAllSubmatchIndex(content, -1)
	// If no images are found, return the original content
	if indexes == nil {
		return content
	}

	// Initialize new content with the same capacity as the original
	newContent := make([]byte, 0, len(content))
	lastIndex := 0

	for _, m := range indexes {
		// Append content up to the current image tag part
		newContent = append(newContent, content[lastIndex:m[imageTagPartStart*2]]...)

		// Extract host, image, and tag parts from the content
		host := string(content[m[imageHostPartStart*2]:m[imageHostPartEnd*2+1]])
		image := string(content[m[imageImagePartStart*2]:m[imageImagePartEnd*2+1]])
		tag := string(content[m[imageTagPartStart*2]:m[imageTagPartEnd*2+1]])
		lastIndex = m[1]

		// If tag is empty or does not match the filter, append the original tag part and continue
		if tag == "" || (imageFilter != nil && !imageFilter.MatchString(host+"/"+image+":"+tag)) {
			newContent = append(newContent, content[m[imageTagPartStart*2]:m[1]]...)
			continue
		}

		// Get the latest tag using the tagPicker function
		latest, err := tagPicker(host, image, tag)
		if err != nil {
			log.Printf("Failed to update %s/%s:%s: %v.\n", host, image, tag, err)
			newContent = append(newContent, content[m[imageTagPartStart*2]:m[1]]...)
			continue
		}

		// Append the latest tag to the new content
		newContent = append(newContent, []byte(latest)...)
	}

	// Append the remaining content after the last match
	newContent = append(newContent, content[lastIndex:]...)

	return newContent
}

// UpdateFile updates a file in place.
func (cli *Client) UpdateFile(tagPicker func(imageHost, imageName, currentTag string) (string, error),
	path string, imageFilter *regexp.Regexp) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	newContent := updateAllTags(tagPicker, content, imageFilter)

	if err := os.WriteFile(path, newContent, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

// GetReplacements returns the tag replacements that have been made.
func (cli *Client) GetReplacements() map[string]string {
	return cli.tagCache
}
