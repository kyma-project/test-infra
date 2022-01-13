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

const sampleInstanceNameExcludeRegexp = "^nightly-.*"
const sampleJobLabelExcludeRegexp = "^kyma-nightly$"

const sampleInstanceName = "nightly-test-abc"
const sampleJobLabel = "kyma-nightly"

const otherName = "otherName"
const otherLabel = "otherlabel"
const sampleStatus = "RUNNING"

var (
	instanceNameExcludeRegexp = regexp.MustCompile(sampleInstanceNameExcludeRegexp)
	jobLabelExcludeRegexp     = regexp.MustCompile(sampleJobLabelExcludeRegexp)
	filterFunc                = DefaultInstanceRemovalPredicate(instanceNameExcludeRegexp, jobLabelExcludeRegexp, 1) //age is 1 hour
	timeNow                   = time.Now()
	timeNowFormatted          = timeNow.Format(time.RFC3339Nano)
	timeTwoHoursAgo           = timeNow.Add(time.Duration(-1) * time.Hour)
	timeTwoHoursAgoFormatted  = timeTwoHoursAgo.Format(time.RFC3339Nano)
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
		{name: "Should skip matching instance",
			expectedFilterValue: false,
			instanceName:        sampleInstanceName,
			instanceCreateTime:  timeTwoHoursAgoFormatted,
			instanceJobLabel:    sampleJobLabel,
			instanceStatus:      sampleStatus},
		{name: "Should delete instance with non-matching name and non-matching label",
			expectedFilterValue: true,
			instanceName:        otherName,
			instanceCreateTime:  timeTwoHoursAgoFormatted,
			instanceJobLabel:    otherLabel,
			instanceStatus:      sampleStatus},
		{name: "Should skip instance recently created",
			expectedFilterValue: false,
			instanceName:        otherName,
			instanceCreateTime:  timeNowFormatted,
			instanceJobLabel:    otherLabel,
			instanceStatus:      sampleStatus},
		{name: "Should skip instance with matching name and non-matching label",
			expectedFilterValue: false,
			instanceName:        sampleInstanceName,
			instanceCreateTime:  timeTwoHoursAgoFormatted,
			instanceJobLabel:    otherLabel,
			instanceStatus:      sampleStatus},
		{name: "Should skip instance with non-matching name and matching label",
			expectedFilterValue: false,
			instanceName:        otherName,
			instanceCreateTime:  timeTwoHoursAgoFormatted,
			instanceJobLabel:    sampleJobLabel,
			instanceStatus:      sampleStatus},
		{name: "Should skip instance in STOPPING status",
			expectedFilterValue: false,
			instanceName:        otherName,
			instanceCreateTime:  timeTwoHoursAgoFormatted,
			instanceJobLabel:    otherLabel,
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

	instanceExcluded := createInstance(sampleInstanceName+"1", timeTwoHoursAgoFormatted, sampleJobLabel, sampleStatus) //excluded name and label
	instanceExcludedName := createInstance(sampleInstanceName+"2", timeTwoHoursAgoFormatted, otherLabel, sampleStatus) //excluded name
	instanceExcludedLabel := createInstance(otherName+"1", timeTwoHoursAgoFormatted, sampleJobLabel, sampleStatus)     //excluded label
	instanceMatching1 := createInstance(otherName+"2", timeTwoHoursAgoFormatted, otherLabel, sampleStatus)             //matches removal filter
	instanceCreatedTooRecently := createInstance(otherName+"3", timeNowFormatted, otherLabel, sampleStatus)            //not old enough
	instanceMatching2 := createInstance(otherName+"4", timeTwoHoursAgoFormatted, otherLabel, sampleStatus)             //matches removal filter

	t.Run("list() should find two instances out of six", func(t *testing.T) {

		mockInstancesAPI := &automock.InstancesAPI{}
		defer mockInstancesAPI.AssertExpectations(t)

		testProject := "testProject"

		//Given
		mockInstancesAPI.On("ListInstances", testProject).Return([]*compute.Instance{instanceExcluded, instanceExcludedName, instanceExcludedLabel, instanceMatching1, instanceCreatedTooRecently, instanceMatching2}, nil)

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
		mockInstancesAPI.On("ListInstances", testProject).Return([]*compute.Instance{instanceExcluded, instanceExcludedName, instanceExcludedLabel, instanceMatching1, instanceCreatedTooRecently, instanceMatching2}, nil)

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
		mockInstancesAPI.On("ListInstances", testProject).Return([]*compute.Instance{instanceExcluded, instanceExcludedName, instanceExcludedLabel, instanceMatching1, instanceCreatedTooRecently, instanceMatching2}, nil)

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
		mockInstancesAPI.On("ListInstances", testProject).Return([]*compute.Instance{instanceExcluded, instanceExcludedName, instanceExcludedLabel, instanceMatching1, instanceCreatedTooRecently, instanceMatching2}, nil)

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
