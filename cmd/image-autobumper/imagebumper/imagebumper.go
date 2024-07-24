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

// Constants to indicate parts of the image and tag regex matches.
const (
	tagVersionPart = 1 // Index of the main version part of the tag in regex match
	tagExtraPart   = 2 // Index of the extra part of the tag in regex match
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
		// Check if it's a commit hash
		if len(parts[0]) == 40 || !strings.HasPrefix(parts[0], "v") {
			return "", 0, strings.TrimPrefix(parts[0], "g")
		}
		// It's a tag
		return parts[0], 0, ""
	}

	// If there are two parts, we need to handle it based on the second part
	if len(parts) == 2 {
		// If the second part starts with 'g', it's a git commit hash
		if strings.HasPrefix(parts[1], "g") {
			return parts[0], 0, strings.TrimPrefix(parts[1], "g")
		}
		// Otherwise, it's a version tag
		return parts[0], 0, ""
	}

	// If there are three parts, it should be in the form v0.0.30-14-gdeadbeef
	if len(parts) == 3 {
		// Parse the middle part as an integer (number of commits)
		n, err := strconv.Atoi(parts[1])
		if err != nil {
			panic(err)
		}
		// The last part should start with 'g' and followed by a commit hash
		return parts[0], n, strings.TrimPrefix(parts[2], "g")
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

// updateAllTags updates all image tags in the given content based on the tagPicker function.
// If imageFilter is provided, only images matching the filter will be updated.
func updateAllTags(tagPicker func(string, string, string) (string, error), content []byte, imageFilter *regexp.Regexp) []byte {
	return imageRegexp.ReplaceAllFunc(content, func(image []byte) []byte {
		matches := imageRegexp.FindSubmatch(image)
		if len(matches) != 4 {
			return image // Should not happen, but for safety
		}

		imageHost := string(matches[1])
		imageName := string(matches[2])
		imageTag := string(matches[3])

		// Apply image filter if provided
		if imageFilter != nil && !imageFilter.Match(image) {
			return image
		}

		newTag, err := tagPicker(imageHost, imageName, imageTag)
		if err != nil {
			return image // If there is an error getting the new tag, return the original image
		}

		// Construct the updated image with the new tag
		return []byte(fmt.Sprintf("%s/%s:%s", imageHost, imageName, newTag))
	})
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
