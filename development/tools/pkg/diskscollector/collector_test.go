package diskscollector

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/diskscollector/automock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	compute "google.golang.org/api/compute/v1"
)

const regexPattern = "^gke-gkeint.*[-]pvc[-]"
const sampleDiskName = "gke-gkeint-abc-pvc-xyz"

var (
	diskNameRegex            = regexp.MustCompile(regexPattern)
	filterFunc               = NewDiskFilter(diskNameRegex, 1) //age is 1 hour
	timeNow                  = time.Now()
	timeNowFormatted         = timeNow.Format(time.RFC3339Nano)
	timeTwoHoursAgo          = timeNow.Add(time.Duration(-1) * time.Hour)
	timeTwoHoursAgoFormatted = timeTwoHoursAgo.Format(time.RFC3339Nano)
	nonEmptyUsers            = []string{"someUser"}
	emptyUsers               = []string{}
)

func TestNewDiskFilter(t *testing.T) {
	//given
	var testCases = []struct {
		name                  string
		expectedFilterValue   bool
		diskName              string
		diskCreationTimestamp string
		diskUsers             []string
	}{
		{name: "Should filter matching disk",
			expectedFilterValue:   true,
			diskName:              sampleDiskName,
			diskCreationTimestamp: timeTwoHoursAgoFormatted,
			diskUsers:             emptyUsers},
		{name: "Should skip disk with non matching name",
			expectedFilterValue:   false,
			diskName:              "otherName",
			diskCreationTimestamp: timeTwoHoursAgoFormatted,
			diskUsers:             emptyUsers},
		{name: "Should skip disk recently created",
			expectedFilterValue:   false,
			diskName:              sampleDiskName,
			diskCreationTimestamp: timeNowFormatted,
			diskUsers:             emptyUsers},
		{name: "Should skip disk with users",
			expectedFilterValue:   false,
			diskName:              sampleDiskName,
			diskCreationTimestamp: timeTwoHoursAgoFormatted,
			diskUsers:             nonEmptyUsers},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//when
			disk := createDisk(testCase.diskName, testCase.diskCreationTimestamp, testCase.diskUsers)
			collected, err := filterFunc(disk)

			//then
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedFilterValue, collected)
		})
	}

	t.Run("Should return error on invalid timestamp", func(t *testing.T) {

		//given
		badTimestamp := "@@@"
		disk := compute.Disk{
			CreationTimestamp: badTimestamp,
		}

		_, err := filterFunc(&disk)
		assert.Contains(t, err.Error(), fmt.Sprintf("parsing time \"%s\" as", badTimestamp))
	})
}

func TestGarbageDiskCollector(t *testing.T) {

	diskMatching1 := createDisk(sampleDiskName+"1", timeTwoHoursAgoFormatted, []string{})    //matches removal filter
	diskNonMatchingName := createDisk("otherName"+"2", timeTwoHoursAgoFormatted, []string{}) //non matching name
	diskHasUsers := createDisk(sampleDiskName+"3", timeTwoHoursAgoFormatted, nonEmptyUsers)  //still has users
	diskCreatedTooRecently := createDisk(sampleDiskName+"4", timeNowFormatted, []string{})   //not old enough
	diskMatching2 := createDisk(sampleDiskName+"5", timeTwoHoursAgoFormatted, []string{})    //matches removal filter

	t.Run("list() should find disks to remove", func(t *testing.T) {
		mockZoneAPI := &automock.ZoneAPI{}
		defer mockZoneAPI.AssertExpectations(t)

		mockDiskAPI := &automock.DiskAPI{}
		defer mockDiskAPI.AssertExpectations(t)

		testProject := "testProject"
		mockZoneAPI.On("ListZones", testProject).Return([]string{"a", "b", "c"}, nil)
		mockDiskAPI.On("ListDisks", testProject, "a").Return([]*compute.Disk{diskMatching1}, nil)
		mockDiskAPI.On("ListDisks", testProject, "b").Return([]*compute.Disk{diskNonMatchingName, diskHasUsers, diskCreatedTooRecently}, nil)
		mockDiskAPI.On("ListDisks", testProject, "c").Return([]*compute.Disk{diskMatching2}, nil)

		gdc := NewDisksGarbageCollector(mockZoneAPI, mockDiskAPI, filterFunc)

		res, err := gdc.list(testProject)
		require.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, "a", res[0].zone)
		assert.Equal(t, diskMatching1, res[0].disk)
		assert.Equal(t, "c", res[1].zone)
		assert.Equal(t, diskMatching2, res[1].disk)
	})

	t.Run("list() should collect disks even when some zones fail", func(t *testing.T) {

		mockZoneAPI := &automock.ZoneAPI{}
		defer mockZoneAPI.AssertExpectations(t)

		mockDiskAPI := &automock.DiskAPI{}
		defer mockDiskAPI.AssertExpectations(t)

		testProject := "testProject"
		mockZoneAPI.On("ListZones", testProject).Return([]string{"a", "b", "c", "d"}, nil)
		mockDiskAPI.On("ListDisks", testProject, "a").Return(nil, errors.New("No such zone"))
		mockDiskAPI.On("ListDisks", testProject, "b").Return([]*compute.Disk{diskMatching2, diskCreatedTooRecently, diskHasUsers}, nil)
		mockDiskAPI.On("ListDisks", testProject, "c").Return(nil, errors.New("No such zone"))
		mockDiskAPI.On("ListDisks", testProject, "d").Return([]*compute.Disk{diskNonMatchingName, diskMatching1}, nil)

		gdc := NewDisksGarbageCollector(mockZoneAPI, mockDiskAPI, filterFunc)

		res, err := gdc.list(testProject)
		require.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, "b", res[0].zone)
		assert.Equal(t, diskMatching2, res[0].disk)
		assert.Equal(t, "d", res[1].zone)
		assert.Equal(t, diskMatching1, res[1].disk)
	})

	t.Run("Run(makeChanges=true) should remove matching disks", func(t *testing.T) {

		mockZoneAPI := &automock.ZoneAPI{}
		defer mockZoneAPI.AssertExpectations(t)

		mockDiskAPI := &automock.DiskAPI{}
		defer mockDiskAPI.AssertExpectations(t)

		testProject := "testProject"
		mockZoneAPI.On("ListZones", testProject).Return([]string{"a", "b"}, nil)
		mockDiskAPI.On("ListDisks", testProject, "a").Return([]*compute.Disk{diskMatching1, diskNonMatchingName, diskHasUsers}, nil)
		mockDiskAPI.On("ListDisks", testProject, "b").Return([]*compute.Disk{diskCreatedTooRecently, diskMatching2}, nil)

		mockDiskAPI.On("RemoveDisk", testProject, "a", diskMatching1.Name).Return(nil)
		mockDiskAPI.On("RemoveDisk", testProject, "b", diskMatching2.Name).Return(nil)

		gdc := NewDisksGarbageCollector(mockZoneAPI, mockDiskAPI, filterFunc)

		err := gdc.Run(testProject, true)
		require.NoError(t, err)
	})

	t.Run("Run(makeChanges=true) should continue process if a call fails", func(t *testing.T) {

		mockZoneAPI := &automock.ZoneAPI{}
		defer mockZoneAPI.AssertExpectations(t)

		mockDiskAPI := &automock.DiskAPI{}
		defer mockDiskAPI.AssertExpectations(t)

		testProject := "testProject"
		mockZoneAPI.On("ListZones", testProject).Return([]string{"a", "b"}, nil)
		mockDiskAPI.On("ListDisks", testProject, "a").Return([]*compute.Disk{diskMatching1, diskNonMatchingName, diskHasUsers}, nil)
		mockDiskAPI.On("ListDisks", testProject, "b").Return([]*compute.Disk{diskCreatedTooRecently, diskMatching2}, nil)

		mockDiskAPI.On("RemoveDisk", testProject, "a", diskMatching1.Name).Return(errors.New("test error")) //Called first, returns error
		mockDiskAPI.On("RemoveDisk", testProject, "b", diskMatching2.Name).Return(nil)                      //Called although the previous call failed

		gdc := NewDisksGarbageCollector(mockZoneAPI, mockDiskAPI, filterFunc)

		err := gdc.Run(testProject, true)
		require.NoError(t, err)
	})

	t.Run("Run(makeChanges=false) should not remove anything (dry run)", func(t *testing.T) {

		mockZoneAPI := &automock.ZoneAPI{}
		defer mockZoneAPI.AssertExpectations(t)

		mockDiskAPI := &automock.DiskAPI{}
		defer mockDiskAPI.AssertExpectations(t)

		testProject := "testProject"
		mockZoneAPI.On("ListZones", testProject).Return([]string{"a", "b"}, nil)
		mockDiskAPI.On("ListDisks", testProject, "a").Return([]*compute.Disk{diskMatching1, diskNonMatchingName, diskHasUsers}, nil)
		mockDiskAPI.On("ListDisks", testProject, "b").Return([]*compute.Disk{diskCreatedTooRecently, diskMatching2}, nil)

		gdc := NewDisksGarbageCollector(mockZoneAPI, mockDiskAPI, filterFunc)

		err := gdc.Run(testProject, false)
		require.NoError(t, err)
	})
}

func createDisk(name, creationTimestamp string, users []string) *compute.Disk {
	return &compute.Disk{
		Name:              name,
		CreationTimestamp: creationTimestamp,
		Users:             users,
	}
}
