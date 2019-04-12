package clusterscollector

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/clusterscollector/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	container "google.golang.org/api/container/v1"
)

const sampleClusterNameRegexp = "^gkeint[-](pr|commit)[-].*"

const sampleClusterName = "gkeint-pr-2991-abc"
const volatileLabel = "true"
const sampleStatus = "RUNNING"
const ttlLabel = "1" // max ttl of a cluster in hours

var (
	clusterNameRegexp       = regexp.MustCompile(sampleClusterNameRegexp)
	filterFunc              = DefaultClusterRemovalPredicate(clusterNameRegexp, 1) //age is 1 hour
	labelFilterFunc         = TimeBasedClusterRemovalPredicate()
	timeNow                 = time.Now()
	timeNowFormatted        = timeNow.Format(time.RFC3339Nano)
	timeNowUnix             = strconv.FormatInt(timeNow.Unix(), 10)
	timeOneHourAgo          = timeNow.Add(time.Duration(-1) * time.Hour)
	timeOneHourAgoFormatted = timeOneHourAgo.Format(time.RFC3339Nano)
	timeOneHourAgoUnix      = strconv.FormatInt(timeOneHourAgo.Unix(), 10)
	timeWellOver            = timeNow.Add(time.Duration(-24) * time.Hour)
	timeWellOverFormatted   = timeWellOver.Format(time.RFC3339Nano)
	timeWellOverUnix        = strconv.FormatInt(timeWellOver.Unix(), 10)
)

func TestDefaultClusterRemovalPredicate(t *testing.T) {

	//given
	var testCases = []struct {
		name                string
		expectedFilterValue bool
		clusterName         string
		clusterCreateTime   string
		volatileLabelValue  string
		clusterStatus       string
	}{
		{name: "Should filter matching cluster",
			expectedFilterValue: true,
			clusterName:         sampleClusterName,
			clusterCreateTime:   timeOneHourAgoFormatted,
			volatileLabelValue:  volatileLabel,
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster with non matching name",
			expectedFilterValue: false,
			clusterName:         "otherName",
			clusterCreateTime:   timeOneHourAgoFormatted,
			volatileLabelValue:  volatileLabel,
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster recently created",
			expectedFilterValue: false,
			clusterName:         sampleClusterName,
			clusterCreateTime:   timeNowFormatted,
			volatileLabelValue:  volatileLabel,
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster with invalid label",
			expectedFilterValue: false,
			clusterName:         sampleClusterName,
			clusterCreateTime:   timeOneHourAgoFormatted,
			volatileLabelValue:  "no",
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster in STOPPING status",
			expectedFilterValue: false,
			clusterName:         sampleClusterName,
			clusterCreateTime:   timeOneHourAgoFormatted,
			volatileLabelValue:  volatileLabel,
			clusterStatus:       "STOPPING"},
		{name: "Should skip cluster in with name kyma-prow",
			expectedFilterValue: false,
			clusterName:         "kyma-prow",
			clusterCreateTime:   timeWellOverFormatted,
			volatileLabelValue:  volatileLabel,
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster in with name workload-kyma-prow",
			expectedFilterValue: false,
			clusterName:         "workload-kyma-prow",
			clusterCreateTime:   timeWellOverFormatted,
			volatileLabelValue:  volatileLabel,
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster in with name nightly",
			expectedFilterValue: false,
			clusterName:         "nightly",
			clusterCreateTime:   timeWellOverFormatted,
			volatileLabelValue:  volatileLabel,
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster in with name weekly",
			expectedFilterValue: false,
			clusterName:         "weekly",
			clusterCreateTime:   timeWellOverFormatted,
			volatileLabelValue:  volatileLabel,
			clusterStatus:       sampleStatus},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//when
			cluster := createCluster(testCase.clusterName, testCase.clusterCreateTime, testCase.volatileLabelValue, "", "", testCase.clusterStatus)
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

func TestTimeBasedClusterRemovalPredicate(t *testing.T) {
	//given
	var testCases = []struct {
		name                string
		expectedFilterValue bool
		clusterName         string
		volatileLabelValue  string
		createdAtLabelValue string
		ttlLabelValue       string
		clusterStatus       string
	}{
		{name: "Should filter matching cluster",
			expectedFilterValue: true,
			volatileLabelValue:  volatileLabel,
			clusterName:         sampleClusterName,
			createdAtLabelValue: timeOneHourAgoUnix,
			ttlLabelValue:       ttlLabel,
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster recently created",
			expectedFilterValue: false,
			volatileLabelValue:  volatileLabel,
			clusterName:         sampleClusterName,
			createdAtLabelValue: timeNowUnix,
			ttlLabelValue:       ttlLabel,
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster with invalid label 1/?",
			expectedFilterValue: false,
			volatileLabelValue:  volatileLabel,
			clusterName:         sampleClusterName,
			createdAtLabelValue: timeNowUnix,
			ttlLabelValue:       "",
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster with invalid label 2/?",
			expectedFilterValue: false,
			volatileLabelValue:  volatileLabel,
			clusterName:         sampleClusterName,
			createdAtLabelValue: "",
			ttlLabelValue:       ttlLabel,
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster with invalid label 3/?",
			expectedFilterValue: false,
			volatileLabelValue:  "no",
			clusterName:         sampleClusterName,
			createdAtLabelValue: timeNowUnix,
			ttlLabelValue:       ttlLabel,
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster in STOPPING status",
			expectedFilterValue: false,
			volatileLabelValue:  volatileLabel,
			clusterName:         sampleClusterName,
			createdAtLabelValue: timeOneHourAgoUnix,
			ttlLabelValue:       ttlLabel,
			clusterStatus:       "STOPPING"},
		{name: "Should skip cluster in with name kyma-prow",
			expectedFilterValue: false,
			clusterName:         "kyma-prow",
			createdAtLabelValue: timeWellOverUnix,
			volatileLabelValue:  "no",
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster in with name workload-kyma-prow",
			expectedFilterValue: false,
			clusterName:         "workload-kyma-prow",
			createdAtLabelValue: timeWellOverUnix,
			volatileLabelValue:  "no",
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster in with name nightly",
			expectedFilterValue: false,
			clusterName:         "nightly",
			createdAtLabelValue: timeWellOverUnix,
			volatileLabelValue:  volatileLabel,
			clusterStatus:       sampleStatus},
		{name: "Should skip cluster in with name weekly",
			expectedFilterValue: false,
			clusterName:         "weekly",
			createdAtLabelValue: timeWellOverUnix,
			volatileLabelValue:  volatileLabel,
			clusterStatus:       sampleStatus},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//when
			cluster := createCluster(testCase.clusterName, "", testCase.volatileLabelValue, testCase.createdAtLabelValue, testCase.ttlLabelValue, testCase.clusterStatus)
			collected, err := labelFilterFunc(cluster)

			//then
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedFilterValue, collected)
		})
	}
	var errorMessageTestCases = []struct {
		name                string
		expectedFilterValue bool
		expectedErrorValue  string
		createdAtLabelValue string
		ttlLabelValue       string
	}{
		{name: "Should return error on invalid created-at label value",
			expectedFilterValue: false,
			expectedErrorValue:  "invalid timestamp value",
			createdAtLabelValue: "@@@",
			ttlLabelValue:       "1"},
		{name: "Should return error on invalid ttl label value",
			expectedFilterValue: false,
			expectedErrorValue:  "invalid ttl value",
			createdAtLabelValue: "1",
			ttlLabelValue:       "@@@"},
	}
	for _, errorTestCase := range errorMessageTestCases {
		t.Run(errorTestCase.name, func(t *testing.T) {
			//given
			cluster := &container.Cluster{
				ResourceLabels: map[string]string{createdAtLabelName: errorTestCase.createdAtLabelValue, ttlLabelName: errorTestCase.ttlLabelValue},
			}

			//test
			_, err := labelFilterFunc(cluster)
			assert.Contains(t, err.Error(), errorTestCase.expectedErrorValue)
		})
	}

}

func TestClustersGarbageCollector(t *testing.T) {

	clusterMatching1 := createCluster(sampleClusterName+"1", timeOneHourAgoFormatted, volatileLabel, timeOneHourAgoUnix, ttlLabel, sampleStatus)       //matches removal filter
	clusterNonMatchingName := createCluster("otherName"+"2", timeOneHourAgoFormatted, volatileLabel, timeOneHourAgoUnix, "", sampleStatus)             //non matching name
	clusterNonMatchingLabel := createCluster(sampleClusterName+"3", timeOneHourAgoFormatted, "otherLabel", timeOneHourAgoUnix, ttlLabel, sampleStatus) //non matching label
	clusterCreatedTooRecently := createCluster(sampleClusterName+"4", timeNowFormatted, volatileLabel, timeNowUnix, ttlLabel, sampleStatus)            //not old enough
	clusterMatching2 := createCluster(sampleClusterName+"5", timeOneHourAgoFormatted, volatileLabel, timeOneHourAgoUnix, ttlLabel, sampleStatus)       //matches removal filter

	t.Run("list() should select two clusters out of five - DefaultClusterRemovalPredicate", func(t *testing.T) {

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

	t.Run("list() should select two clusters out of five - TimeBasedClusterRemovalPredicate", func(t *testing.T) {

		mockClusterAPI := &automock.ClusterAPI{}
		defer mockClusterAPI.AssertExpectations(t)

		testProject := "testProject"

		//Given
		mockClusterAPI.On("ListClusters", testProject).Return([]*container.Cluster{clusterMatching1, clusterNonMatchingName, clusterNonMatchingLabel, clusterCreatedTooRecently, clusterMatching2}, nil)

		//When
		gdc := NewClustersGarbageCollector(mockClusterAPI, labelFilterFunc)
		res, err := gdc.list(testProject)

		//Then
		require.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, clusterMatching1, res[0])
		assert.Equal(t, clusterMatching2, res[1])
	})

	t.Run("Run() should fail if list() fails - DefaultClusterRemovalPredicate", func(t *testing.T) {

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

	t.Run("Run() should fail if list() fails - TimeBasedClusterRemovalPredicate", func(t *testing.T) {

		mockClusterAPI := &automock.ClusterAPI{}
		defer mockClusterAPI.AssertExpectations(t)

		testError := errors.New("testError")
		testProject := "testProject"
		mockClusterAPI.On("ListClusters", testProject).Return(nil, testError)

		gdc := NewClustersGarbageCollector(mockClusterAPI, labelFilterFunc)

		_, err := gdc.Run(testProject, true)
		require.Error(t, err)
		assert.Equal(t, testError, err)
	})

	t.Run("Run(makeChanges=true) should remove matching clusters - DefaultClusterRemovalPredicate", func(t *testing.T) {

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

	t.Run("Run(makeChanges=true) should remove matching clusters - TimeBasedClusterRemovalPredicate", func(t *testing.T) {

		mockClusterAPI := &automock.ClusterAPI{}
		defer mockClusterAPI.AssertExpectations(t)

		testProject := "testProject"
		mockClusterAPI.On("ListClusters", testProject).Return([]*container.Cluster{clusterMatching1, clusterNonMatchingName, clusterNonMatchingLabel, clusterCreatedTooRecently, clusterMatching2}, nil)

		mockClusterAPI.On("RemoveCluster", testProject, clusterMatching1.Zone, clusterMatching1.Name).Return(nil)
		mockClusterAPI.On("RemoveCluster", testProject, clusterMatching2.Zone, clusterMatching2.Name).Return(nil)

		gdc := NewClustersGarbageCollector(mockClusterAPI, labelFilterFunc)

		allSucceeded, err := gdc.Run(testProject, true)
		require.NoError(t, err)
		assert.True(t, allSucceeded)
	})

	t.Run("Run(makeChanges=true) should continue process if a previous call failed - DefaultClusterRemovalPredicate", func(t *testing.T) {

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

	t.Run("Run(makeChanges=true) should continue process if a previous call failed - TimeBasedClusterRemovalPredicate", func(t *testing.T) {

		mockClusterAPI := &automock.ClusterAPI{}
		defer mockClusterAPI.AssertExpectations(t)

		testProject := "testProject"

		//Given
		mockClusterAPI.On("ListClusters", testProject).Return([]*container.Cluster{clusterMatching1, clusterNonMatchingName, clusterNonMatchingLabel, clusterCreatedTooRecently, clusterMatching2}, nil)

		mockClusterAPI.On("RemoveCluster", testProject, clusterMatching1.Zone, clusterMatching1.Name).Return(errors.New("testError"))
		mockClusterAPI.On("RemoveCluster", testProject, clusterMatching2.Zone, clusterMatching2.Name).Return(nil)
		//When
		gdc := NewClustersGarbageCollector(mockClusterAPI, labelFilterFunc)
		allSucceeded, err := gdc.Run(testProject, true)

		//Then
		mockClusterAPI.AssertCalled(t, "RemoveCluster", mock.Anything, mock.Anything, mock.Anything)
		require.NoError(t, err)
		assert.False(t, allSucceeded)
	})

	t.Run("Run(makeChanges=false) should not invoke RemoveCluster() (dry run) - DefaultClusterRemovalPredicate", func(t *testing.T) {

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

	t.Run("Run(makeChanges=false) should not invoke RemoveCluster() (dry run) - TimeBasedClusterRemovalPredicate", func(t *testing.T) {

		mockClusterAPI := &automock.ClusterAPI{}
		defer mockClusterAPI.AssertExpectations(t)

		testProject := "testProject"

		//Given
		mockClusterAPI.On("ListClusters", testProject).Return([]*container.Cluster{clusterMatching1, clusterNonMatchingName, clusterNonMatchingLabel, clusterCreatedTooRecently, clusterMatching2}, nil)

		//When
		gdc := NewClustersGarbageCollector(mockClusterAPI, labelFilterFunc)
		allSucceeded, err := gdc.Run(testProject, false)

		//Then
		mockClusterAPI.AssertCalled(t, "ListClusters", testProject)
		mockClusterAPI.AssertNotCalled(t, "RemoveCluster", mock.Anything, mock.Anything, mock.Anything)
		require.NoError(t, err)
		assert.True(t, allSucceeded)
	})
}

func createCluster(name, createTime, volatileValue, createdAtValue, ttlValue, status string) *container.Cluster {
	return &container.Cluster{
		Name:           name,
		Zone:           name + "-zone",
		CreateTime:     createTime,
		ResourceLabels: map[string]string{volatileLabelName: volatileValue, createdAtLabelName: createdAtValue, ttlLabelName: ttlValue},
		Status:         status,
	}
}
