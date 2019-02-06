package dnscollector

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/dnscollector/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	compute "google.golang.org/api/compute/v1"
	dns "google.golang.org/api/dns/v1"
)

const regexPattern = "(remoteenvs-)?gkeint-(pr|commit)-.*"
const sampleAddressName = "gkeint-pr-2271-0v5rfpux6w"

var (
	addressNameRegex         = regexp.MustCompile(regexPattern)
	filterFunc               = DefaultIPAddressRemovalPredicate([]*regexp.Regexp{addressNameRegex}, 1) //age is 1 hour
	timeNow                  = time.Now()
	timeNowFormatted         = timeNow.Format(time.RFC3339Nano)
	timeTwoHoursAgo          = timeNow.Add(time.Duration(-1) * time.Hour)
	timeTwoHoursAgoFormatted = timeTwoHoursAgo.Format(time.RFC3339Nano)
	nonEmptyUsers            = []string{"someUser"}
	emptyUsers               = []string{}
)

func TestDefaultIPAddressRemovalPredicate(t *testing.T) {
	//given
	var testCases = []struct {
		name                     string
		expectedFilterValue      bool
		addressName              string
		addressCreationTimestamp string
	}{
		{name: "Should find matching address",
			expectedFilterValue:      true,
			addressName:              sampleAddressName,
			addressCreationTimestamp: timeTwoHoursAgoFormatted},
		{name: "Should skip address with non matching name",
			expectedFilterValue:      false,
			addressName:              "otherName",
			addressCreationTimestamp: timeTwoHoursAgoFormatted},
		{name: "Should skip address recently created",
			expectedFilterValue:      false,
			addressName:              sampleAddressName,
			addressCreationTimestamp: timeNowFormatted},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//when
			address := createAddress(testCase.addressName, "192.168.0.1", testCase.addressCreationTimestamp)
			collected, err := filterFunc(address)

			//then
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedFilterValue, collected)
		})
	}

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

func TestGarbageDNSCollector(t *testing.T) {
	addressMatching1 := createAddress(sampleAddressName+"1", "192.168.0.1", timeTwoHoursAgoFormatted) //matches removal filter
	addressNotMatching2 := createAddress(sampleAddressName+"2", "192.168.0.2", timeNowFormatted)      //does not match removal filter
	addressMatching3 := createAddress(sampleAddressName+"3", "192.168.0.2", timeTwoHoursAgoFormatted) //matches removal filter

	t.Run("listRegionIPs() should find matching addresses in region", func(t *testing.T) {
		mockComputeAPI := &automock.ComputeAPI{}
		defer mockComputeAPI.AssertExpectations(t)

		testProject := "testProject"
		testRegion := "testRegion"
		mockComputeAPI.On("LookupIPAddresses", testProject, testRegion).Return([]*compute.Address{addressMatching1, addressNotMatching2, addressMatching3}, nil)

		gdc := New(mockComputeAPI, nil, filterFunc)

		res, allSucceeded, err := gdc.listRegionIPs(testProject, testRegion)
		require.NoError(t, err)
		assert.True(t, allSucceeded)
		assert.Len(t, res, 2)
		assert.Equal(t, testRegion, res[0].region)
		assert.Equal(t, addressMatching1, res[0].data)
		assert.Equal(t, testRegion, res[1].region)
		assert.Equal(t, addressMatching3, res[1].data)
	})
}

func createAddress(name, address, creationTimestamp string) *compute.Address {
	return &compute.Address{
		Name:              name,
		Address:           address,
		CreationTimestamp: creationTimestamp,
	}
}

func createDNSRecord(name, address string) *dns.ResourceRecordSet {
	return nil
}
