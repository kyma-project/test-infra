package clusterscollector

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/clusterscollector/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	container "google.golang.org/api/container/v1"
)

const sampleClusterNameRegexp = "^gkeint[-](pr|commit)[-].*"
const sampleJobLabelRegexp = "^kyma-gke-integration"

const sampleClusterName = "gkeint-pr-2991-abc"
const sampleJobLabel = "kyma-gke-integration"
const sampleStatus = "RUNNING"

var (
	clusterNameRegexp        = regexp.MustCompile(sampleClusterNameRegexp)
	jobLabelRegexp           = regexp.MustCompile(sampleJobLabelRegexp)
	filterFunc               = DefaultClusterRemovalPredicate(clusterNameRegexp, jobLabelRegexp, 1) //age is 1 hour
	timeNow                  = time.Now()
	timeNowFormatted         = timeNow.Format(time.RFC3339Nano)
	timeTwoHoursAgo          = timeNow.Add(time.Duration(-1) * time.Hour)
	timeTwoHoursAgoFormatted = timeTwoHoursAgo.Format(time.RFC3339Nano)
	nonEmptyUsers            = []string{"someUser"}
	emptyUsers               = []string{}
)

func TestNewClusterRemovalPredicate(t *testing.T) {

	//given
	var testCases = []struct {
		name                string
		expectedFilterValue bool
		clusterName         string
		clusterCreateTime   string
		clusterJobLabel     string
		clusterStatus       string
	}{
		{name: "Should filter matching cluster",
			expectedFilterValue: true,
			clusterName:         sampleClusterName,
			clusterCreateTime:   timeTwoHoursAgoFormatted,
			clusterJobLabel:     sampleJobLabel,
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster with non matching name",
			expectedFilterValue: false,
			clusterName:         "otherName",
			clusterCreateTime:   timeTwoHoursAgoFormatted,
			clusterJobLabel:     sampleJobLabel,
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster recently created",
			expectedFilterValue: false,
			clusterName:         sampleClusterName,
			clusterCreateTime:   timeNowFormatted,
			clusterJobLabel:     sampleJobLabel,
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster with invalid label",
			expectedFilterValue: false,
			clusterName:         sampleClusterName,
			clusterCreateTime:   timeTwoHoursAgoFormatted,
			clusterJobLabel:     "otherLabel",
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster in STOPPING status",
			expectedFilterValue: false,
			clusterName:         sampleClusterName,
			clusterCreateTime:   timeTwoHoursAgoFormatted,
			clusterJobLabel:     sampleJobLabel,
			clusterStatus:       "STOPPING"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//when
			cluster := createCluster(testCase.clusterName, testCase.clusterCreateTime, testCase.clusterJobLabel, testCase.clusterStatus)
			collected, err := filterFunc(cluster)

			//then
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedFilterValue, collected)
		})
	}

	t.Run("Should return error on invalid timestamp", func(t *testing.T) {

		//given
		badTime := "@@@"
		cluster := &container.Cluster{
			CreateTime: badTime,
		}

		_, err := filterFunc(cluster)
		assert.Contains(t, err.Error(), fmt.Sprintf("parsing time \"%s\" as", badTime))
	})
}

func TestClustersGarbageCollector(t *testing.T) {

	clusterMatching1 := createCluster(sampleClusterName+"1", timeTwoHoursAgoFormatted, sampleJobLabel, sampleStatus)      //matches removal filter
	clusterNonMatchingName := createCluster("otherName"+"2", timeTwoHoursAgoFormatted, sampleJobLabel, sampleStatus)      //non matching name
	clusterNonMatchingLabel := createCluster(sampleClusterName+"3", timeTwoHoursAgoFormatted, "otherLabel", sampleStatus) //non matching label
	clusterCreatedTooRecently := createCluster(sampleClusterName+"4", timeNowFormatted, sampleJobLabel, sampleStatus)     //not old enough
	clusterMatching2 := createCluster(sampleClusterName+"5", timeTwoHoursAgoFormatted, sampleJobLabel, sampleStatus)      //matches removal filter

	t.Run("list() should filter two clusters to remove out of five", func(t *testing.T) {

		mockClusterAPI := &automock.ClusterAPI{}
		defer mockClusterAPI.AssertExpectations(t)

		testProject := "testProject"

		//Given
		mockClusterAPI.On("ListClusters", testProject).Return([]*container.Cluster{clusterMatching1, clusterNonMatchingName, clusterNonMatchingLabel, clusterCreatedTooRecently, clusterMatching2}, nil)

		//When
		gdc := NewClustersGarbageCollector(mockClusterAPI, filterFunc)
		res, err := gdc.list(testProject)

		//Then
		require.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, clusterMatching1, res[0])
		assert.Equal(t, clusterMatching2, res[1])
	})

	t.Run("Run(makeChanges=true) should remove matching disks", func(t *testing.T) {

		mockClusterAPI := &automock.ClusterAPI{}
		defer mockClusterAPI.AssertExpectations(t)

		testProject := "testProject"
		mockClusterAPI.On("ListClusters", testProject).Return([]*container.Cluster{clusterMatching1, clusterNonMatchingName, clusterNonMatchingLabel, clusterCreatedTooRecently, clusterMatching2}, nil)

		mockClusterAPI.On("RemoveCluster", testProject, clusterMatching1.Zone, clusterMatching1.Name).Return(nil)
		mockClusterAPI.On("RemoveCluster", testProject, clusterMatching2.Zone, clusterMatching2.Name).Return(nil)

		gdc := NewClustersGarbageCollector(mockClusterAPI, filterFunc)

		allSucceeded, err := gdc.Run(testProject, true)
		require.NoError(t, err)
		assert.True(t, allSucceeded)
	})

	t.Run("Run() should fail if list() fails", func(t *testing.T) {

		mockClusterAPI := &automock.ClusterAPI{}
		defer mockClusterAPI.AssertExpectations(t)

		testError := errors.New("testError")
		testProject := "testProject"
		mockClusterAPI.On("ListClusters", testProject).Return(nil, testError)

		gdc := NewClustersGarbageCollector(mockClusterAPI, filterFunc)

		_, err := gdc.Run(testProject, true)
		require.Error(t, err)
		assert.Equal(t, testError, err)
	})

	t.Run("Run(makeChanges=true) should continue process if a previous call failed", func(t *testing.T) {

		mockClusterAPI := &automock.ClusterAPI{}
		defer mockClusterAPI.AssertExpectations(t)

		testProject := "testProject"

		//Given
		mockClusterAPI.On("ListClusters", testProject).Return([]*container.Cluster{clusterMatching1, clusterNonMatchingName, clusterNonMatchingLabel, clusterCreatedTooRecently, clusterMatching2}, nil)

		mockClusterAPI.On("RemoveCluster", testProject, clusterMatching1.Zone, clusterMatching1.Name).Return(errors.New("testError"))
		mockClusterAPI.On("RemoveCluster", testProject, clusterMatching2.Zone, clusterMatching2.Name).Return(nil)
		//When
		gdc := NewClustersGarbageCollector(mockClusterAPI, filterFunc)
		allSucceeded, err := gdc.Run(testProject, true)

		//Then
		mockClusterAPI.AssertCalled(t, "RemoveCluster", mock.Anything, mock.Anything, mock.Anything)
		require.NoError(t, err)
		assert.False(t, allSucceeded)
	})

	t.Run("Run(makeChanges=false) should not invoke RemoveCluster() (dry run)", func(t *testing.T) {

		mockClusterAPI := &automock.ClusterAPI{}
		defer mockClusterAPI.AssertExpectations(t)

		testProject := "testProject"

		//Given
		mockClusterAPI.On("ListClusters", testProject).Return([]*container.Cluster{clusterMatching1, clusterNonMatchingName, clusterNonMatchingLabel, clusterCreatedTooRecently, clusterMatching2}, nil)

		//When
		gdc := NewClustersGarbageCollector(mockClusterAPI, filterFunc)
		allSucceeded, err := gdc.Run(testProject, false)

		//Then
		mockClusterAPI.AssertCalled(t, "ListClusters", testProject)
		mockClusterAPI.AssertNotCalled(t, "RemoveCluster", mock.Anything, mock.Anything, mock.Anything)
		require.NoError(t, err)
		assert.True(t, allSucceeded)
	})
}

func createCluster(name, createTime, jobLabel, status string) *container.Cluster {
	return &container.Cluster{
		Name:           name,
		Zone:           name + "-zone",
		CreateTime:     createTime,
		ResourceLabels: map[string]string{"job": jobLabel},
		Status:         status,
	}
}
