package vmscollector

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/vmscollector/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	compute "google.golang.org/api/compute/v1"
)

const sampleInstanceNameRegexp = "^kyma-integration-test-.*"
const sampleJobLabelRegexp = "^kyma-integration$"

const sampleInstanceName = "kyma-integration-test-abc"
const sampleJobLabel = "kyma-integration"
const sampleStatus = "RUNNING"

var (
	instanceNameRegexp       = regexp.MustCompile(sampleInstanceNameRegexp)
	jobLabelRegexp           = regexp.MustCompile(sampleJobLabelRegexp)
	filterFunc               = DefaultInstanceRemovalPredicate(instanceNameRegexp, jobLabelRegexp, 1) //age is 1 hour
	timeNow                  = time.Now()
	timeNowFormatted         = timeNow.Format(time.RFC3339Nano)
	timeTwoHoursAgo          = timeNow.Add(time.Duration(-1) * time.Hour)
	timeTwoHoursAgoFormatted = timeTwoHoursAgo.Format(time.RFC3339Nano)
)

func TestNewInstanceRemovalPredicate(t *testing.T) {

	//given
	var testCases = []struct {
		name                string
		expectedFilterValue bool
		instanceName        string
		instanceCreateTime  string
		instanceJobLabel    string
		instanceStatus      string
	}{
		{name: "Should select matching instance",
			expectedFilterValue: true,
			instanceName:        sampleInstanceName,
			instanceCreateTime:  timeTwoHoursAgoFormatted,
			instanceJobLabel:    sampleJobLabel,
			instanceStatus:      sampleStatus},
		{name: "Should skip instance with non matching name",
			expectedFilterValue: false,
			instanceName:        "otherName",
			instanceCreateTime:  timeTwoHoursAgoFormatted,
			instanceJobLabel:    sampleJobLabel,
			instanceStatus:      sampleStatus},
		{name: "Should skip instance recently created",
			expectedFilterValue: false,
			instanceName:        sampleInstanceName,
			instanceCreateTime:  timeNowFormatted,
			instanceJobLabel:    sampleJobLabel,
			instanceStatus:      sampleStatus},
		{name: "Should skip instance with invalid label",
			expectedFilterValue: false,
			instanceName:        sampleInstanceName,
			instanceCreateTime:  timeTwoHoursAgoFormatted,
			instanceJobLabel:    "otherLabel",
			instanceStatus:      sampleStatus},
		{name: "Should skip instance in STOPPING status",
			expectedFilterValue: false,
			instanceName:        sampleInstanceName,
			instanceCreateTime:  timeTwoHoursAgoFormatted,
			instanceJobLabel:    sampleJobLabel,
			instanceStatus:      "STOPPING"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//when
			instance := createInstance(testCase.instanceName, testCase.instanceCreateTime, testCase.instanceJobLabel, testCase.instanceStatus)
			collected, err := filterFunc(instance)

			//then
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedFilterValue, collected)
		})
	}

	t.Run("Should return error on invalid timestamp", func(t *testing.T) {

		//given
		badTime := "@@@"
		instance := createInstance(sampleInstanceName, badTime, sampleJobLabel, sampleStatus)

		_, err := filterFunc(instance)
		assert.Contains(t, err.Error(), fmt.Sprintf("parsing time \"%s\" as", badTime))
	})
}

func TestInstancesGarbageCollector(t *testing.T) {

	instanceMatching1 := createInstance(sampleInstanceName+"1", timeTwoHoursAgoFormatted, sampleJobLabel, sampleStatus)      //matches removal filter
	instanceNonMatchingName := createInstance("otherName"+"2", timeTwoHoursAgoFormatted, sampleJobLabel, sampleStatus)       //non matching name
	instanceNonMatchingLabel := createInstance(sampleInstanceName+"3", timeTwoHoursAgoFormatted, "otherLabel", sampleStatus) //non matching label
	instanceCreatedTooRecently := createInstance(sampleInstanceName+"4", timeNowFormatted, sampleJobLabel, sampleStatus)     //not old enough
	instanceMatching2 := createInstance(sampleInstanceName+"5", timeTwoHoursAgoFormatted, sampleJobLabel, sampleStatus)      //matches removal filter

	t.Run("list() should find two instances out of five", func(t *testing.T) {

		mockInstancesAPI := &automock.InstancesAPI{}
		defer mockInstancesAPI.AssertExpectations(t)

		testProject := "testProject"

		//Given
		mockInstancesAPI.On("ListInstances", testProject).Return([]*compute.Instance{instanceMatching1, instanceNonMatchingName, instanceNonMatchingLabel, instanceCreatedTooRecently, instanceMatching2}, nil)

		//When
		gdc := NewInstancesGarbageCollector(mockInstancesAPI, filterFunc)
		res, err := gdc.list(testProject)

		//Then
		require.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, instanceMatching1, res[0])
		assert.Equal(t, instanceMatching2, res[1])
	})

	t.Run("Run() should fail if list() fails", func(t *testing.T) {

		mockInstancesAPI := &automock.InstancesAPI{}
		defer mockInstancesAPI.AssertExpectations(t)

		testError := errors.New("testError")
		testProject := "testProject"
		mockInstancesAPI.On("ListInstances", testProject).Return(nil, testError)

		gdc := NewInstancesGarbageCollector(mockInstancesAPI, filterFunc)

		_, err := gdc.Run(testProject, true)
		require.Error(t, err)
		assert.Equal(t, testError, err)
	})

	t.Run("Run(makeChanges=true) should remove matching instances", func(t *testing.T) {

		mockInstancesAPI := &automock.InstancesAPI{}
		defer mockInstancesAPI.AssertExpectations(t)

		testProject := "testProject"
		mockInstancesAPI.On("ListInstances", testProject).Return([]*compute.Instance{instanceMatching1, instanceNonMatchingName, instanceNonMatchingLabel, instanceCreatedTooRecently, instanceMatching2}, nil)

		mockInstancesAPI.On("RemoveInstance", testProject, instanceMatching1.Name+"-zone", instanceMatching1.Name).Return(nil)
		mockInstancesAPI.On("RemoveInstance", testProject, instanceMatching2.Name+"-zone", instanceMatching2.Name).Return(nil)

		gdc := NewInstancesGarbageCollector(mockInstancesAPI, filterFunc)

		allSucceeded, err := gdc.Run(testProject, true)
		require.NoError(t, err)
		assert.True(t, allSucceeded)
	})

	t.Run("Run(makeChanges=true) should continue processing if a previous call failed", func(t *testing.T) {

		mockInstancesAPI := &automock.InstancesAPI{}
		defer mockInstancesAPI.AssertExpectations(t)

		testProject := "testProject"

		//Given
		mockInstancesAPI.On("ListInstances", testProject).Return([]*compute.Instance{instanceMatching1, instanceNonMatchingName, instanceNonMatchingLabel, instanceCreatedTooRecently, instanceMatching2}, nil)

		mockInstancesAPI.On("RemoveInstance", testProject, instanceMatching1.Name+"-zone", instanceMatching1.Name).Return(errors.New("testError"))
		mockInstancesAPI.On("RemoveInstance", testProject, instanceMatching2.Name+"-zone", instanceMatching2.Name).Return(nil)
		//When
		gdc := NewInstancesGarbageCollector(mockInstancesAPI, filterFunc)
		allSucceeded, err := gdc.Run(testProject, true)

		//Then
		mockInstancesAPI.AssertCalled(t, "RemoveInstance", mock.Anything, mock.Anything, mock.Anything)
		require.NoError(t, err)
		assert.False(t, allSucceeded)
	})

	t.Run("Run(makeChanges=false) should not invoke RemoveInstance() (dry run)", func(t *testing.T) {

		mockInstancesAPI := &automock.InstancesAPI{}
		defer mockInstancesAPI.AssertExpectations(t)

		testProject := "testProject"

		//Given
		mockInstancesAPI.On("ListInstances", testProject).Return([]*compute.Instance{instanceMatching1, instanceNonMatchingName, instanceNonMatchingLabel, instanceCreatedTooRecently, instanceMatching2}, nil)

		//When
		gdc := NewInstancesGarbageCollector(mockInstancesAPI, filterFunc)
		allSucceeded, err := gdc.Run(testProject, false)

		//Then
		mockInstancesAPI.AssertCalled(t, "ListInstances", testProject)
		mockInstancesAPI.AssertNotCalled(t, "RemoveInstance", mock.Anything, mock.Anything, mock.Anything)
		require.NoError(t, err)
		assert.True(t, allSucceeded)
	})
}

func createInstance(name, creationTimestamp, jobLabel, status string) *compute.Instance {
	return &compute.Instance{
		Name:              name,
		Zone:              fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/test/zones/%s-zone", name), //This is what GCloud API returns in "Zone" attribute on List() call
		CreationTimestamp: creationTimestamp,
		Labels:            map[string]string{jobLabelName: jobLabel},
		Status:            status,
	}
}
