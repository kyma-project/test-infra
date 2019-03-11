package networkscollector

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/networkscollector/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	compute "google.golang.org/api/compute/v1"
)

const sampleNetworkNameRegexp = "^net-gkeint[-](pr|commit)[-].*"

const sampleNetworkName = "net-gkeint-pr-2991-abc"

var (
	networkNameRegexp        = regexp.MustCompile(sampleNetworkNameRegexp)
	filterFunc               = DefaultNetworkRemovalPredicate(networkNameRegexp, 1) //age is 1 hour
	timeNow                  = time.Now()
	timeNowFormatted         = timeNow.Format(time.RFC3339Nano)
	timeTwoHoursAgo          = timeNow.Add(time.Duration(-1) * time.Hour)
	timeTwoHoursAgoFormatted = timeTwoHoursAgo.Format(time.RFC3339Nano)
)

func TestDefaultNetworkRemovalPredicate(t *testing.T) {
	// given
	var testCases = []struct {
		name                     string
		expectedFilterValue      bool
		networkName              string
		networkCreationTimestamp string
	}{
		{name: "Should filter matching network",
			expectedFilterValue:      true,
			networkName:              sampleNetworkName,
			networkCreationTimestamp: timeTwoHoursAgoFormatted},
		{name: "Should skip network with non matching name",
			expectedFilterValue:      false,
			networkName:              "otherName",
			networkCreationTimestamp: timeTwoHoursAgoFormatted},
		{name: "Should skip network recently created",
			expectedFilterValue:      false,
			networkName:              sampleNetworkName,
			networkCreationTimestamp: timeNowFormatted},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//when
			network := createNetwork(testCase.networkName, testCase.networkCreationTimestamp)
			collected, err := filterFunc(network)

			//then
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedFilterValue, collected)
		})
	}
}

func TestNetworksGarbageCollector(t *testing.T) {

	networkMatching1 := createNetwork(sampleNetworkName+"1", timeTwoHoursAgoFormatted)  //matches removal filter
	networkNonMatchingName := createNetwork("otherName"+"2", timeTwoHoursAgoFormatted)  //non matching name
	networkCreatedTooRecently := createNetwork(sampleNetworkName+"3", timeNowFormatted) //not old enough
	networkMatching2 := createNetwork(sampleNetworkName+"4", timeTwoHoursAgoFormatted)  //matches removal filter

	t.Run("list() should select two networks out of four", func(t *testing.T) {

		mockNetworkAPI := &automock.NetworkAPI{}
		defer mockNetworkAPI.AssertExpectations(t)

		testProject := "testProject"

		//Given
		mockNetworkAPI.On("ListNetworks", testProject).Return([]*compute.Network{networkMatching1, networkNonMatchingName, networkCreatedTooRecently, networkMatching2}, nil)

		//When
		gdc := NewNetworksGarbageCollector(mockNetworkAPI, filterFunc)
		res, err := gdc.list(testProject)

		//Then
		require.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, networkMatching1, res[0])
		assert.Equal(t, networkMatching2, res[1])
	})

	t.Run("Run() should fail if list() fails", func(t *testing.T) {

		mockNetworkAPI := &automock.NetworkAPI{}
		defer mockNetworkAPI.AssertExpectations(t)

		testError := errors.New("testError")
		testProject := "testProject"
		mockNetworkAPI.On("ListNetworks", testProject).Return(nil, testError)

		gdc := NewNetworksGarbageCollector(mockNetworkAPI, filterFunc)

		_, err := gdc.Run(testProject, true)
		require.Error(t, err)
		assert.Equal(t, testError, err)
	})

	t.Run("Run(makeChanges=true) should remove matching networks", func(t *testing.T) {

		mockNetworkAPI := &automock.NetworkAPI{}
		defer mockNetworkAPI.AssertExpectations(t)

		testProject := "testProject"
		mockNetworkAPI.On("ListNetworks", testProject).Return([]*compute.Network{networkMatching1, networkNonMatchingName, networkCreatedTooRecently, networkMatching2}, nil)

		mockNetworkAPI.On("RemoveNetwork", testProject, networkMatching1.Name).Return(nil)
		mockNetworkAPI.On("RemoveNetwork", testProject, networkMatching2.Name).Return(nil)

		gdc := NewNetworksGarbageCollector(mockNetworkAPI, filterFunc)

		allSucceeded, err := gdc.Run(testProject, true)
		require.NoError(t, err)
		assert.True(t, allSucceeded)
	})

	t.Run("Run(makeChanges=true) should continue process if a previous call failed", func(t *testing.T) {

		mockNetworkAPI := &automock.NetworkAPI{}
		defer mockNetworkAPI.AssertExpectations(t)

		testProject := "testProject"

		//Given
		mockNetworkAPI.On("ListNetworks", testProject).Return([]*compute.Network{networkMatching1, networkNonMatchingName, networkCreatedTooRecently, networkMatching2}, nil)

		mockNetworkAPI.On("RemoveNetwork", testProject, networkMatching1.Name).Return(errors.New("testError"))
		mockNetworkAPI.On("RemoveNetwork", testProject, networkMatching2.Name).Return(nil)
		//When
		gdc := NewNetworksGarbageCollector(mockNetworkAPI, filterFunc)
		allSucceeded, err := gdc.Run(testProject, true)

		//Then
		mockNetworkAPI.AssertCalled(t, "RemoveNetwork", mock.Anything, mock.Anything, mock.Anything)
		require.NoError(t, err)
		assert.False(t, allSucceeded)
	})

	t.Run("Run(makeChanges=false) should not invoke RemoveNetwork() (dry run)", func(t *testing.T) {

		mockNetworkAPI := &automock.NetworkAPI{}
		defer mockNetworkAPI.AssertExpectations(t)

		testProject := "testProject"

		//Given
		mockNetworkAPI.On("ListNetworks", testProject).Return([]*compute.Network{networkMatching1, networkNonMatchingName, networkCreatedTooRecently, networkMatching2}, nil)

		//When
		gdc := NewNetworksGarbageCollector(mockNetworkAPI, filterFunc)
		allSucceeded, err := gdc.Run(testProject, false)

		//Then
		mockNetworkAPI.AssertCalled(t, "ListNetworks", testProject)
		mockNetworkAPI.AssertNotCalled(t, "RemoveNetwork", mock.Anything, mock.Anything, mock.Anything)
		require.NoError(t, err)
		assert.True(t, allSucceeded)
	})
}

func createNetwork(name, creationTimestamp string) *compute.Network {
	return &compute.Network{
		Name:              name,
		CreationTimestamp: creationTimestamp,
	}
}
