package dnscollector

import (
	"errors"
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
	testProject              = "testProject"
	testRegion1              = "testRegion1"
	testRegion2              = "testRegion2"
	addressMatching1         = createAddress(sampleAddressName+"1", "192.168.0.1", timeTwoHoursAgoFormatted) //matches removal filter
	addressNotMatching2      = createAddress(sampleAddressName+"2", "192.168.0.2", timeNowFormatted)         //does not match removal filter
	addressMatching3         = createAddress(sampleAddressName+"3", "192.168.0.3", timeTwoHoursAgoFormatted) //matches removal filter
	addressMatching4         = createAddress(sampleAddressName+"4", "192.168.0.4", timeTwoHoursAgoFormatted) //matches removal filter
	testRecord1              = createDNSRecord("", "192.168.0.1")
	testRecord2              = createDNSRecord("", "192.168.0.2")
	testRecord3              = createDNSRecord("", "192.168.0.3")
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

func TestListRegionIPs(t *testing.T) {

	t.Run("listRegionIPs() should find matching addresses in region", func(t *testing.T) {
		mockComputeAPI := &automock.ComputeAPI{}
		defer mockComputeAPI.AssertExpectations(t)

		mockComputeAPI.On("LookupIPAddresses", testProject, testRegion1).Return([]*compute.Address{addressMatching1, addressNotMatching2, addressMatching3}, nil)

		gdc := New(mockComputeAPI, nil, filterFunc)

		res, allSucceeded, err := gdc.listRegionIPs(testProject, testRegion1)
		require.NoError(t, err)
		assert.True(t, allSucceeded)
		assert.Len(t, res, 2)
		assert.Equal(t, testRegion1, res[0].region)
		assert.Equal(t, addressMatching1, res[0].data)
		assert.Equal(t, testRegion1, res[1].region)
		assert.Equal(t, addressMatching3, res[1].data)
	})

	t.Run("listRegionIPs() should return error correctly", func(t *testing.T) {
		mockComputeAPI := &automock.ComputeAPI{}
		defer mockComputeAPI.AssertExpectations(t)

		mockComputeAPI.On("LookupIPAddresses", testProject, testRegion1).Return(nil, errors.New("testError"))

		gdc := New(mockComputeAPI, nil, filterFunc)

		res, allSucceeded, err := gdc.listRegionIPs(testProject, testRegion1)
		assert.Nil(t, res)
		assert.False(t, allSucceeded)
		assert.Equal(t, "testError", err.Error())
	})

	t.Run("listRegionIPs() should find matching addresses in case predicate function fails", func(t *testing.T) {
		mockComputeAPI := &automock.ComputeAPI{}
		defer mockComputeAPI.AssertExpectations(t)

		addressWithInvalidTimestamp := createAddress(sampleAddressName, "192.168.0.4", timeTwoHoursAgoFormatted) //matches removal filter
		addressWithInvalidTimestamp.CreationTimestamp = "abc"

		mockComputeAPI.On("LookupIPAddresses", testProject, testRegion1).Return([]*compute.Address{addressWithInvalidTimestamp, addressMatching1, addressNotMatching2, addressMatching3}, nil)

		gdc := New(mockComputeAPI, nil, filterFunc)

		res, allSucceeded, err := gdc.listRegionIPs(testProject, testRegion1)
		require.NoError(t, err)
		assert.False(t, allSucceeded)
		assert.Len(t, res, 2)
		assert.Equal(t, testRegion1, res[0].region)
		assert.Equal(t, addressMatching1, res[0].data)
		assert.Equal(t, testRegion1, res[1].region)
		assert.Equal(t, addressMatching3, res[1].data)
	})
}

func TestListIPs(t *testing.T) {

	t.Run("listIPs() should find matching addresses in regions", func(t *testing.T) {
		mockComputeAPI := &automock.ComputeAPI{}
		defer mockComputeAPI.AssertExpectations(t)

		mockComputeAPI.On("LookupIPAddresses", testProject, testRegion1).Return([]*compute.Address{addressMatching1, addressNotMatching2}, nil)
		mockComputeAPI.On("LookupIPAddresses", testProject, testRegion2).Return([]*compute.Address{addressMatching3, addressMatching4}, nil)

		gdc := New(mockComputeAPI, nil, filterFunc)

		res, allSucceeded := gdc.listIPs(testProject, []string{testRegion1, testRegion2})
		assert.True(t, allSucceeded)
		assert.Len(t, res, 3)
		assert.Equal(t, testRegion1, res[0].region)
		assert.Equal(t, addressMatching1, res[0].data)
		assert.Equal(t, testRegion2, res[1].region)
		assert.Equal(t, addressMatching3, res[1].data)
		assert.Equal(t, testRegion2, res[2].region)
		assert.Equal(t, addressMatching4, res[2].data)
	})

	t.Run("listIPs() should correctly report partial failures in case some region lookups fail", func(t *testing.T) {
		mockComputeAPI := &automock.ComputeAPI{}
		defer mockComputeAPI.AssertExpectations(t)

		testRegionError := "testRegionError"

		mockComputeAPI.On("LookupIPAddresses", testProject, testRegion1).Return([]*compute.Address{addressMatching1, addressNotMatching2}, nil)
		mockComputeAPI.On("LookupIPAddresses", testProject, testRegionError).Return(nil, errors.New("test error"))
		mockComputeAPI.On("LookupIPAddresses", testProject, testRegion2).Return([]*compute.Address{addressMatching3, addressMatching4}, nil)

		gdc := New(mockComputeAPI, nil, filterFunc)

		res, allSucceeded := gdc.listIPs(testProject, []string{testRegion1, testRegionError, testRegion2})
		assert.False(t, allSucceeded)
		assert.Len(t, res, 3)
		assert.Equal(t, testRegion1, res[0].region)
		assert.Equal(t, addressMatching1, res[0].data)
		assert.Equal(t, testRegion2, res[1].region)
		assert.Equal(t, addressMatching3, res[1].data)
		assert.Equal(t, testRegion2, res[2].region)
		assert.Equal(t, addressMatching4, res[2].data)
	})

	t.Run("listIPs() should correctly report partial failures in case predicate function fails", func(t *testing.T) {
		mockComputeAPI := &automock.ComputeAPI{}
		defer mockComputeAPI.AssertExpectations(t)

		addressWithInvalidTimestamp := createAddress(sampleAddressName, "192.168.0.4", timeTwoHoursAgoFormatted) //matches removal filter
		addressWithInvalidTimestamp.CreationTimestamp = "abc"

		mockComputeAPI.On("LookupIPAddresses", testProject, testRegion1).Return([]*compute.Address{addressMatching1, addressWithInvalidTimestamp}, nil)
		mockComputeAPI.On("LookupIPAddresses", testProject, testRegion2).Return([]*compute.Address{addressNotMatching2, addressMatching3, addressMatching4}, nil)

		gdc := New(mockComputeAPI, nil, filterFunc)

		res, allSucceeded := gdc.listIPs(testProject, []string{testRegion1, testRegion2})
		assert.False(t, allSucceeded)
		assert.Len(t, res, 3)
		assert.Equal(t, testRegion1, res[0].region)
		assert.Equal(t, addressMatching1, res[0].data)
		assert.Equal(t, testRegion2, res[1].region)
		assert.Equal(t, addressMatching3, res[1].data)
		assert.Equal(t, testRegion2, res[2].region)
		assert.Equal(t, addressMatching4, res[2].data)
	})
}

func TestMatchDNSRecord(t *testing.T) {
	t.Run("matchDNSRecord() should return true for matching record", func(t *testing.T) {
		//given
		record := testRecord1
		//when
		res := matchDNSRecord(record)
		//then
		assert.True(t, res)
	})

	t.Run("matchDNSRecord() should return false for record with invalid type", func(t *testing.T) {
		//given
		record := createDNSRecord("", "192.168.0.1")
		record.Type = "TXT"
		//when
		res := matchDNSRecord(record)
		//then
		assert.False(t, res)
	})

	t.Run("matchDNSRecord() should return false for record with multiple IP addresses", func(t *testing.T) {
		//given
		record := dns.ResourceRecordSet{
			Type:    "A",
			Rrdatas: []string{"192.168.0.1", "192.168.0.2"},
		}
		//when
		res := matchDNSRecord(&record)
		//then
		assert.False(t, res)
	})
}

func TestListDNSRecords(t *testing.T) {
	testManagedZone := "testManagedZone"

	t.Run("listDNSRecords() should return error correctly", func(t *testing.T) {
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)

		mockDNSAPI.On("LookupDNSRecords", testProject, testManagedZone).Return(nil, errors.New("testError"))

		gdc := New(nil, mockDNSAPI, nil)

		res, err := gdc.listDNSRecords(testProject, testManagedZone)
		assert.Nil(t, res)
		assert.Equal(t, "testError", err.Error())
	})

	t.Run("listDNSRecords() should return only matching DNS records", func(t *testing.T) {
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)

		recordMatching1 := testRecord1
		recordNotMatching2 := createDNSRecord("", "192.168.0.2")
		recordNotMatching2.Type = "TXT"
		recordMatching3 := testRecord3

		mockDNSAPI.On("LookupDNSRecords", testProject, testManagedZone).Return([]*dns.ResourceRecordSet{recordMatching1, recordNotMatching2, recordMatching3}, nil)

		gdc := New(nil, mockDNSAPI, nil)

		res, err := gdc.listDNSRecords(testProject, testManagedZone)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res, 2)
		assert.Equal(t, recordMatching1, res[0])
		assert.Equal(t, recordMatching3, res[1])
	})
}

func TestFindAssociatedRecords(t *testing.T) {
	t.Run("findAssociatedRecords() should find a record that matches an IP Address", func(t *testing.T) {

		dnsRecords := []*dns.ResourceRecordSet{testRecord1, testRecord2, testRecord3}
		res := findAssociatedRecords("192.168.0.2", dnsRecords)
		assert.Len(t, res, 1)
		assert.Equal(t, testRecord2, res[0])
	})

	t.Run("findAssociatedRecords() should work correctly when no record matches an IP address", func(t *testing.T) {

		dnsRecords := []*dns.ResourceRecordSet{testRecord1, testRecord2, testRecord3}
		res := findAssociatedRecords("192.168.0.4", dnsRecords)
		assert.Len(t, res, 0)
	})
}

func TestRun(t *testing.T) {

	testManagedZone := "testManagedZone"
	testRegions := []string{testRegion1}

	t.Run("Run() should return early when no IP Addresses are found", func(t *testing.T) {
		mockComputeAPI := &automock.ComputeAPI{}
		defer mockComputeAPI.AssertExpectations(t)
		mockComputeAPI.On("LookupIPAddresses", testProject, testRegions[0]).Return([]*compute.Address{}, nil)

		gdc := New(mockComputeAPI, nil, nil)

		allSucceeded, err := gdc.Run(testProject, testManagedZone, testRegions, true)

		assert.True(t, allSucceeded)
		assert.Nil(t, err)
	})

	t.Run("Run() should return error when DNS records can't be listed", func(t *testing.T) {
		mockComputeAPI := &automock.ComputeAPI{}
		defer mockComputeAPI.AssertExpectations(t)
		mockComputeAPI.On("LookupIPAddresses", testProject, testRegions[0]).Return([]*compute.Address{addressMatching1}, nil)
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)
		mockDNSAPI.On("LookupDNSRecords", testProject, testManagedZone).Return(nil, errors.New("testError"))

		gdc := New(mockComputeAPI, mockDNSAPI, filterFunc)
		allSucceeded, err := gdc.Run(testProject, testManagedZone, testRegions, true)

		assert.False(t, allSucceeded)
		assert.Equal(t, "testError", err.Error())
	})

	t.Run("Run() should delete associated DNS record before IP Address", func(t *testing.T) {

		mockComputeAPI := &automock.ComputeAPI{}
		defer mockComputeAPI.AssertExpectations(t)
		mockComputeAPI.On("LookupIPAddresses", testProject, testRegions[0]).Return([]*compute.Address{addressMatching1, addressMatching3}, nil)
		mockComputeAPI.On("DeleteIPAddress", testProject, testRegions[0], addressMatching1.Name).Return(nil)
		mockComputeAPI.On("DeleteIPAddress", testProject, testRegions[0], addressMatching3.Name).Return(nil)

		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)
		mockDNSAPI.On("LookupDNSRecords", testProject, testManagedZone).Return([]*dns.ResourceRecordSet{testRecord1, testRecord2, testRecord3}, nil)
		mockDNSAPI.On("DeleteDNSRecord", testProject, testManagedZone, testRecord1).Return(nil)
		mockDNSAPI.On("DeleteDNSRecord", testProject, testManagedZone, testRecord3).Return(nil)

		gdc := New(mockComputeAPI, mockDNSAPI, filterFunc)
		allSucceeded, err := gdc.Run(testProject, testManagedZone, testRegions, true)

		assert.True(t, allSucceeded)
		assert.Nil(t, err)
	})

	t.Run("Run() should continue if a call to delete IP Address fails", func(t *testing.T) {

		mockComputeAPI := &automock.ComputeAPI{}
		defer mockComputeAPI.AssertExpectations(t)
		mockComputeAPI.On("LookupIPAddresses", testProject, testRegions[0]).Return([]*compute.Address{addressMatching1, addressMatching3}, nil)
		mockComputeAPI.On("DeleteIPAddress", testProject, testRegions[0], addressMatching1.Name).Return(errors.New("testError"))
		mockComputeAPI.On("DeleteIPAddress", testProject, testRegions[0], addressMatching3.Name).Return(nil)

		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)
		mockDNSAPI.On("LookupDNSRecords", testProject, testManagedZone).Return([]*dns.ResourceRecordSet{testRecord1, testRecord2, testRecord3}, nil)
		mockDNSAPI.On("DeleteDNSRecord", testProject, testManagedZone, testRecord1).Return(nil)
		mockDNSAPI.On("DeleteDNSRecord", testProject, testManagedZone, testRecord3).Return(nil)

		gdc := New(mockComputeAPI, mockDNSAPI, filterFunc)
		allSucceeded, err := gdc.Run(testProject, testManagedZone, testRegions, true)

		assert.False(t, allSucceeded)
		assert.Nil(t, err)
	})

	t.Run("Run() should continue if a call to delete DNS Record fails", func(t *testing.T) {

		mockComputeAPI := &automock.ComputeAPI{}
		defer mockComputeAPI.AssertExpectations(t)
		mockComputeAPI.On("LookupIPAddresses", testProject, testRegions[0]).Return([]*compute.Address{addressMatching1, addressMatching3}, nil)
		//mockComputeAPI.On("DeleteIPAddress", testProject, testRegions[0], addressMatching1.Name).Return(nil) <- this call is skipped!
		mockComputeAPI.On("DeleteIPAddress", testProject, testRegions[0], addressMatching3.Name).Return(nil)

		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)
		mockDNSAPI.On("LookupDNSRecords", testProject, testManagedZone).Return([]*dns.ResourceRecordSet{testRecord1, testRecord2, testRecord3}, nil)
		mockDNSAPI.On("DeleteDNSRecord", testProject, testManagedZone, testRecord1).Return(errors.New("testError"))
		mockDNSAPI.On("DeleteDNSRecord", testProject, testManagedZone, testRecord3).Return(nil)

		gdc := New(mockComputeAPI, mockDNSAPI, filterFunc)
		allSucceeded, err := gdc.Run(testProject, testManagedZone, testRegions, true)

		assert.False(t, allSucceeded)
		assert.Nil(t, err)
	})

	t.Run("Run() should perform dry run correctly", func(t *testing.T) {

		mockComputeAPI := &automock.ComputeAPI{}
		defer mockComputeAPI.AssertExpectations(t)
		mockComputeAPI.On("LookupIPAddresses", testProject, testRegions[0]).Return([]*compute.Address{addressMatching1, addressMatching3}, nil)

		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)
		mockDNSAPI.On("LookupDNSRecords", testProject, testManagedZone).Return([]*dns.ResourceRecordSet{testRecord1, testRecord2, testRecord3}, nil)

		gdc := New(mockComputeAPI, mockDNSAPI, filterFunc)
		allSucceeded, err := gdc.Run(testProject, testManagedZone, testRegions, false)

		assert.True(t, allSucceeded)
		assert.Nil(t, err)
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
	return &dns.ResourceRecordSet{
		Name:    name,
		Type:    "A",
		Rrdatas: []string{address},
	}
}
