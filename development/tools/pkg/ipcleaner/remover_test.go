package ipcleaner

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/ipcleaner/automock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	compute "google.golang.org/api/compute/v1"
)

const sampleAddressName = "gkekcpint-commit-abc-zyx"

var (
	ipNameIgnoreRegexPattern = "^otherName"
	ipNameIgnoreRegex        = regexp.MustCompile(ipNameIgnoreRegexPattern)
	filterFunc               = NewIPFilter(ipNameIgnoreRegex, 1) //age is 1 hour
	timeNow                  = time.Now()
	timeNowFormatted         = timeNow.Format(time.RFC3339Nano)
	timeTwoHoursAgo          = timeNow.Add(time.Duration(-1) * time.Hour)
	timeTwoHoursAgoFormatted = timeTwoHoursAgo.Format(time.RFC3339Nano)
	nonEmptyUsers            = []string{"someUser"}
	emptyUsers               = []string{}
)

func TestNewAddressFilter(t *testing.T) {
	//given
	var testCases = []struct {
		name                     string
		expectedFilterValue      bool
		addresskName             string
		addressCreationTimestamp string
		addressUsers             []string
	}{
		{name: "Should filter matching address",
			expectedFilterValue:      true,
			addresskName:             sampleAddressName,
			addressCreationTimestamp: timeTwoHoursAgoFormatted,
			addressUsers:             emptyUsers},
		{name: "Should skip address with matching name",
			expectedFilterValue:      false,
			addresskName:             "otherName",
			addressCreationTimestamp: timeTwoHoursAgoFormatted,
			addressUsers:             emptyUsers},
		{name: "Should skip address recently created",
			expectedFilterValue:      false,
			addresskName:             sampleAddressName,
			addressCreationTimestamp: timeNowFormatted,
			addressUsers:             emptyUsers},
		{name: "Should skip address with users",
			expectedFilterValue:      false,
			addresskName:             sampleAddressName,
			addressCreationTimestamp: timeTwoHoursAgoFormatted,
			addressUsers:             nonEmptyUsers},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//when
			address := createAddress(testCase.addresskName, testCase.addressCreationTimestamp, testCase.addressUsers)
			collected, err := filterFunc(address)

			//then
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedFilterValue, collected)
		})

		t.Run("Should return error on invalid timestamp", func(t *testing.T) {

			//given
			badTimestamp := "@@@"
			address := compute.Address{
				CreationTimestamp: badTimestamp,
			}

			_, err := filterFunc(&address)
			assert.Contains(t, err.Error(), fmt.Sprintf("parsing time \"%s\" as", badTimestamp))
		})
	}
}

func TestGarbageAddressCollector(t *testing.T) {
	addressMatching1 := createAddress(sampleAddressName+"1", timeTwoHoursAgoFormatted, []string{})   //matches removal filter
	addressNonMatchingName := createAddress("otherName"+"2", timeTwoHoursAgoFormatted, []string{})   //non matching name
	addressHasUsers := createAddress(sampleAddressName+"3", timeTwoHoursAgoFormatted, nonEmptyUsers) //still has users
	addressCreatedTooRecently := createAddress(sampleAddressName+"4", timeNowFormatted, []string{})  //not old enough
	addressMatching2 := createAddress(sampleAddressName+"5", timeTwoHoursAgoFormatted, []string{})   //matches removal filter

	t.Run("list() should find addresses to remove", func(t *testing.T) {
		mockRegionAPI := &automock.RegionAPI{}
		defer mockRegionAPI.AssertExpectations(t)

		mockAddressAPI := &automock.AddressAPI{}
		defer mockAddressAPI.AssertExpectations(t)

		testProject := "testProject"
		mockRegionAPI.On("ListRegions", testProject).Return([]string{"a", "b", "c"}, nil)
		mockAddressAPI.On("ListAddresses", testProject, "a").Return([]*compute.Address{addressMatching1}, nil)
		mockAddressAPI.On("ListAddresses", testProject, "b").Return([]*compute.Address{addressNonMatchingName, addressHasUsers, addressCreatedTooRecently}, nil)
		mockAddressAPI.On("ListAddresses", testProject, "c").Return([]*compute.Address{addressMatching2}, nil)

		gac := New(mockAddressAPI, mockRegionAPI, filterFunc)

		res, err := gac.list(testProject)
		require.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, "a", res[0].region)
		assert.Equal(t, addressMatching1, res[0].address)
		assert.Equal(t, "c", res[1].region)
		assert.Equal(t, addressMatching2, res[1].address)
	})

	t.Run("list() should collect addresses even when some regions fail", func(t *testing.T) {

		mockRegionAPI := &automock.RegionAPI{}
		defer mockRegionAPI.AssertExpectations(t)

		mockAddressAPI := &automock.AddressAPI{}
		defer mockAddressAPI.AssertExpectations(t)

		testProject := "testProject"
		mockRegionAPI.On("ListRegions", testProject).Return([]string{"a", "b", "c", "d"}, nil)
		mockAddressAPI.On("ListAddresses", testProject, "a").Return(nil, errors.New("No such region"))
		mockAddressAPI.On("ListAddresses", testProject, "b").Return([]*compute.Address{addressMatching2, addressCreatedTooRecently, addressHasUsers}, nil)
		mockAddressAPI.On("ListAddresses", testProject, "c").Return(nil, errors.New("No such region"))
		mockAddressAPI.On("ListAddresses", testProject, "d").Return([]*compute.Address{addressNonMatchingName, addressMatching1}, nil)

		gdc := New(mockAddressAPI, mockRegionAPI, filterFunc)

		res, err := gdc.list(testProject)
		require.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, "b", res[0].region)
		assert.Equal(t, addressMatching2, res[0].address)
		assert.Equal(t, "d", res[1].region)
		assert.Equal(t, addressMatching1, res[1].address)
	})

	t.Run("Run(makeChanges=true) should remove matching addresses", func(t *testing.T) {

		mockRegionAPI := &automock.RegionAPI{}
		defer mockRegionAPI.AssertExpectations(t)

		mockAddressAPI := &automock.AddressAPI{}
		defer mockAddressAPI.AssertExpectations(t)

		testProject := "testProject"
		mockRegionAPI.On("ListRegions", testProject).Return([]string{"a", "b"}, nil)
		mockAddressAPI.On("ListAddresses", testProject, "a").Return([]*compute.Address{addressMatching1, addressNonMatchingName, addressHasUsers}, nil)
		mockAddressAPI.On("ListAddresses", testProject, "b").Return([]*compute.Address{addressCreatedTooRecently, addressMatching2}, nil)

		mockAddressAPI.On("RemoveIP", testProject, "a", addressMatching1.Name).Return(nil)
		mockAddressAPI.On("RemoveIP", testProject, "b", addressMatching2.Name).Return(nil)

		gdc := New(mockAddressAPI, mockRegionAPI, filterFunc)

		allSucceeded, err := gdc.Run(testProject, true)
		require.NoError(t, err)
		assert.True(t, allSucceeded)
	})

	t.Run("Run(makeChanges=true) should continue process if a call fails", func(t *testing.T) {

		mockRegionAPI := &automock.RegionAPI{}
		defer mockRegionAPI.AssertExpectations(t)

		mockAddressAPI := &automock.AddressAPI{}
		defer mockAddressAPI.AssertExpectations(t)

		testProject := "testProject"
		mockRegionAPI.On("ListRegions", testProject).Return([]string{"a", "b"}, nil)
		mockAddressAPI.On("ListAddresses", testProject, "a").Return([]*compute.Address{addressMatching1, addressNonMatchingName, addressHasUsers}, nil)
		mockAddressAPI.On("ListAddresses", testProject, "b").Return([]*compute.Address{addressCreatedTooRecently, addressMatching2}, nil)

		mockAddressAPI.On("RemoveIP", testProject, "a", addressMatching1.Name).Return(errors.New("test error")) //Called first, returns error
		mockAddressAPI.On("RemoveIP", testProject, "b", addressMatching2.Name).Return(nil)                      //Called although the previous call failed

		gdc := New(mockAddressAPI, mockRegionAPI, filterFunc)

		allSucceeded, err := gdc.Run(testProject, true)
		require.NoError(t, err)
		assert.False(t, allSucceeded)
	})

	t.Run("Run(makeChanges=false) should not remove anything (dry run)", func(t *testing.T) {

		mockRegionAPI := &automock.RegionAPI{}
		defer mockRegionAPI.AssertExpectations(t)

		mockAddressAPI := &automock.AddressAPI{}
		defer mockAddressAPI.AssertExpectations(t)

		testProject := "testProject"
		mockRegionAPI.On("ListRegions", testProject).Return([]string{"a", "b"}, nil)
		mockAddressAPI.On("ListAddresses", testProject, "a").Return([]*compute.Address{addressMatching1, addressNonMatchingName, addressHasUsers}, nil)
		mockAddressAPI.On("ListAddresses", testProject, "b").Return([]*compute.Address{addressCreatedTooRecently, addressMatching2}, nil)

		gdc := New(mockAddressAPI, mockRegionAPI, filterFunc)

		allSucceeded, err := gdc.Run(testProject, false)
		require.NoError(t, err)
		assert.True(t, allSucceeded)
	})

}

func createAddress(name string, creationTimestamp string, users []string) *compute.Address {
	return &compute.Address{
		Name:              name,
		CreationTimestamp: creationTimestamp,
		Users:             users,
	}
}
