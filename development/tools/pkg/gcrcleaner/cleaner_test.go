package gcrcleaner

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/gcrcleaner/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gcrgoogle "github.com/google/go-containerregistry/pkg/v1/google"
)

var (
	gcrNameIgnoreRegexPattern = "test/filtered/repo"
	regexRepo                 = regexp.MustCompile(gcrNameIgnoreRegexPattern)
	emptyRegex                = regexp.MustCompile("")
	repoFilter                = NewRepoFilter(regexRepo)
	emptyFilter               = NewRepoFilter(emptyRegex)
	imageFilter               = NewImageFilter(1) // age is 1 hour
	timeNow                   = time.Now()
	timeTwoHoursAgo           = timeNow.Add(time.Duration(-1) * time.Hour)
)

func TestNewRepoFilter(t *testing.T) {
	var testCases = []struct {
		name                string
		expectedFilterValue bool
		repo                string
	}{
		{
			name:                "should filter matching repo",
			expectedFilterValue: false,
			repo:                "test/filtered/repos",
		},
		{
			name:                "should skip repo without matching name",
			expectedFilterValue: true,
			repo:                "test/unfiltered/repo",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//when
			collected := repoFilter(testCase.repo)

			//then
			assert.Equal(t, testCase.expectedFilterValue, collected)
		})

		// test that empty filter will take all images into consideration
		t.Run("Should delete all images", func(t *testing.T) {
			//when
			collected := emptyFilter(testCase.repo)

			//then
			assert.Equal(t, true, collected)
		})
	}
}

func TestNewImageFilter(t *testing.T) {
	var testCases = []struct {
		name                string
		expectedFilterValue bool
		created             time.Time
	}{
		{
			name:                "should filter older image",
			expectedFilterValue: true,
			created:             timeTwoHoursAgo,
		},
		{
			name:                "should skip recently created image",
			expectedFilterValue: false,
			created:             timeNow,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//when
			manifest := createImageManifest(testCase.created, make([]string, 0))
			collected := imageFilter(&manifest)

			//then
			assert.Equal(t, testCase.expectedFilterValue, collected)
		})
	}
}

func TestImageRemoval(t *testing.T) {

	imageCorrect := createImageManifest(timeTwoHoursAgo, make([]string, 0))

	t.Run("ListSubrepositories() should find addresses to remove", func(t *testing.T) {
		mockRepoAPI := &automock.RepoAPI{}
		defer mockRepoAPI.AssertExpectations(t)

		mockImageAPI := &automock.ImageAPI{}
		defer mockImageAPI.AssertExpectations(t)

		registry := "eu.gcr.io"
		repo := "test"
		repository := registry + "/" + repo
		mockRepoAPI.On("ListSubrepositories", "eu.gcr.io/test").Return([]string{"test/repo", "test/another"}, nil)

		repos, err := mockRepoAPI.ListSubrepositories(repository)
		require.NoError(t, err)

		assert.Len(t, repos, 2)
		assert.Equal(t, "test/repo", repos[0])
		assert.Equal(t, "test/another", repos[1])
	})

	t.Run("Run(makeChanges=true) should continue process if a call fails", func(t *testing.T) {
		mockRepoAPI := &automock.RepoAPI{}
		defer mockRepoAPI.AssertExpectations(t)

		mockImageAPI := &automock.ImageAPI{}
		defer mockImageAPI.AssertExpectations(t)

		registry := "eu.gcr.io"
		repo := "test"
		repository := registry + "/" + repo
		mockRepoAPI.On("ListSubrepositories", "eu.gcr.io/test").Return([]string{"test/repo", "test/another"}, nil)

		mockImageAPI.On("ListImages", "eu.gcr.io", "test/repo").Return(map[string]gcrgoogle.ManifestInfo{"sha256:abcd": imageCorrect}, nil)
		mockImageAPI.On("ListImages", "eu.gcr.io", "test/another").Return(map[string]gcrgoogle.ManifestInfo{"sha256:efgh": imageCorrect}, nil)

		mockImageAPI.On("DeleteImage", registry, "test/repo", "sha256:abcd", imageCorrect).Return(errors.New("test error")) // Called first, returns error
		mockImageAPI.On("DeleteImage", registry, "test/another", "sha256:efgh", imageCorrect).Return(nil)                   // Called although the previous call failed

		gcrc := New(mockRepoAPI, mockImageAPI, repoFilter, imageFilter)

		allSucceeded, err := gcrc.Run(repository, true)
		require.NoError(t, err)
		assert.False(t, allSucceeded)
	})

	t.Run("Run(makeChanges=false) should not remove anything (dry run)", func(t *testing.T) {
		mockRepoAPI := &automock.RepoAPI{}
		defer mockRepoAPI.AssertExpectations(t)

		mockImageAPI := &automock.ImageAPI{}
		defer mockImageAPI.AssertExpectations(t)

		repository := "eu.gcr.io/test"
		mockRepoAPI.On("ListSubrepositories", "eu.gcr.io/test").Return([]string{"test/repo", "test/another"}, nil)

		mockImageAPI.On("ListImages", "eu.gcr.io", "test/repo").Return(map[string]gcrgoogle.ManifestInfo{"sha256:abcd": imageCorrect}, nil)
		mockImageAPI.On("ListImages", "eu.gcr.io", "test/another").Return(map[string]gcrgoogle.ManifestInfo{"sha256:efgh": imageCorrect}, nil)

		gcrc := New(mockRepoAPI, mockImageAPI, repoFilter, imageFilter)

		allSucceeded, err := gcrc.Run(repository, false)
		require.NoError(t, err)
		assert.True(t, allSucceeded)
	})
}

func createImageManifest(created time.Time, tags []string) gcrgoogle.ManifestInfo {
	return gcrgoogle.ManifestInfo{
		Created: created,
		Tags:    tags,
	}
}
