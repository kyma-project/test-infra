package dnscleaner

import (
	"errors"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/pkg/longlastingdnscleaner/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dns "google.golang.org/api/dns/v1"
)

var (
	// dns entry : 1
	shouldDeleteDNSName       = "this-dns-entry-is-a-delete-candidate"
	shouldDeleteDNSIP         = "10.1.1.1"
	shouldDeleteDNSRecordType = "A"
	shouldDeleteDNSTTL        = int64(300)

	// dns entry : 2
	shouldNotDeleteDNSName       = "this-dns-entry-stays"
	shouldNotDeleteDNSIP         = "10.1.1.2"
	shouldNotDeleteDNSRecordType = "A"
	shouldNotDeleteDNSTTL        = int64(300)

	// dns entry : 3
	shouldNotDeleteDNSNameTwo       = "this-dns-entry-stays-as-well"
	shouldNotDeleteDNSIPTwo         = "10.1.1.3"
	shouldNotDeleteDNSRecordTypeTwo = "B"
	shouldNotDeleteDNSTTLTwo        = int64(200)

	testProject = "testProject"
	testZone    = "testRegion"
)

func TestNewDNSEntryRemover(t *testing.T) {
	t.Run("Should return and delete DNS entry", func(t *testing.T) {
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)

		record := createDNSResourceRecordSet(shouldDeleteDNSName, shouldDeleteDNSIP, shouldDeleteDNSRecordType, shouldDeleteDNSTTL)

		//Given
		mockDNSAPI.On("LookupDNSEntry", testProject, testZone, shouldDeleteDNSName, shouldDeleteDNSIP, shouldDeleteDNSRecordType, shouldDeleteDNSTTL).Return(record, true, nil)
		mockDNSAPI.On("RemoveDNSEntry", testProject, testZone, record).Return(false, nil)

		//When
		der := NewDNSEntryRemover(mockDNSAPI)
		success, err := der.Run(testProject, testZone, shouldDeleteDNSName, shouldDeleteDNSIP, shouldDeleteDNSRecordType, shouldDeleteDNSTTL, 3, 2, true)

		//Then
		mockDNSAPI.AssertNumberOfCalls(t, "LookupDNSEntry", 1)
		mockDNSAPI.AssertNumberOfCalls(t, "RemoveDNSEntry", 1)
		require.NoError(t, err)
		assert.Equal(t, success, true)
	})

	t.Run("Should not delete DNS entry", func(t *testing.T) {
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)

		//Given
		mockDNSAPI.On("LookupDNSEntry", testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL).Return(nil, false, errors.New("test error"))

		//When
		der := NewDNSEntryRemover(mockDNSAPI)
		success, err := der.Run(testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL, 3, 2, true)

		//Then
		mockDNSAPI.AssertNotCalled(t, "RemoveDNSEntry")
		mockDNSAPI.AssertNumberOfCalls(t, "LookupDNSEntry", 1)
		require.Error(t, err)
		assert.Equal(t, success, false)
	})
}

func TestRetryableCalls(t *testing.T) {
	t.Run("Should retry 3 times and then throw error", func(t *testing.T) {
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)

		//Given
		mockDNSAPI.On("LookupDNSEntry", testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL).Return(nil, true, errors.New("test error")).Times(3)

		//When
		der := NewDNSEntryRemover(mockDNSAPI)
		success, err := der.Run(testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL, 3, 2, true)

		//Then
		mockDNSAPI.AssertNotCalled(t, "RemoveDNSEntry")
		mockDNSAPI.AssertNumberOfCalls(t, "LookupDNSEntry", 3)
		require.Error(t, err)
		assert.Equal(t, success, false)
	})
}

func TestRetryableCallsPartTwo(t *testing.T) {
	t.Run("Should retry 3 times and then succeed lookup and fail on remove", func(t *testing.T) {
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)

		record := createDNSResourceRecordSet(shouldDeleteDNSName, shouldDeleteDNSIP, shouldDeleteDNSRecordType, shouldDeleteDNSTTL)

		//Given
		mockDNSAPI.On("LookupDNSEntry", testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL).Return(nil, true, errors.New("test error")).Times(2)
		mockDNSAPI.On("LookupDNSEntry", testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL).Return(record, true, nil)
		mockDNSAPI.On("RemoveDNSEntry", testProject, testZone, record).Return(true, errors.New("test error")).Times(3)

		//When
		der := NewDNSEntryRemover(mockDNSAPI)
		success, err := der.Run(testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL, 3, 1, true)

		//Then
		mockDNSAPI.AssertNumberOfCalls(t, "LookupDNSEntry", 3)
		mockDNSAPI.AssertNumberOfCalls(t, "RemoveDNSEntry", 3)
		require.Error(t, err)
		assert.Equal(t, success, false)
	})
}

func TestRetryableCallsPartThree(t *testing.T) {
	t.Run("Should retry 3 times and then succeed lookup and retry 3 times for delete and succeed", func(t *testing.T) {
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)

		record := createDNSResourceRecordSet(shouldDeleteDNSName, shouldDeleteDNSIP, shouldDeleteDNSRecordType, shouldDeleteDNSTTL)

		//Given
		mockDNSAPI.On("LookupDNSEntry", testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL).Return(nil, true, errors.New("test error")).Times(2)
		mockDNSAPI.On("LookupDNSEntry", testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL).Return(record, true, nil)
		mockDNSAPI.On("RemoveDNSEntry", testProject, testZone, record).Return(true, errors.New("test error")).Times(2)
		mockDNSAPI.On("RemoveDNSEntry", testProject, testZone, record).Return(true, nil)

		//When
		der := NewDNSEntryRemover(mockDNSAPI)
		success, err := der.Run(testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL, 3, 1, true)

		//Then
		mockDNSAPI.AssertNumberOfCalls(t, "RemoveDNSEntry", 3)
		mockDNSAPI.AssertNumberOfCalls(t, "LookupDNSEntry", 3)
		require.NoError(t, err)
		assert.Equal(t, success, true)
	})
}
func TestDryRunBehaviour(t *testing.T) {
	t.Run("Should not delete on dryrun", func(t *testing.T) {
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)

		record := createDNSResourceRecordSet(shouldDeleteDNSName, shouldDeleteDNSIP, shouldDeleteDNSRecordType, shouldDeleteDNSTTL)

		//Given
		mockDNSAPI.On("LookupDNSEntry", testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL).Return(record, true, nil)

		//When
		der := NewDNSEntryRemover(mockDNSAPI)
		success, err := der.Run(testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL, 3, 2, false)

		//Then
		mockDNSAPI.AssertNumberOfCalls(t, "RemoveDNSEntry", 0)
		mockDNSAPI.AssertNumberOfCalls(t, "LookupDNSEntry", 1)
		require.NoError(t, err)
		assert.Equal(t, success, true)
	})
}

func createDNSResourceRecordSet(name, address, rtype string, ttl int64) *dns.ResourceRecordSet {
	return &dns.ResourceRecordSet{
		Name:    name,
		Rrdatas: []string{address},
		Ttl:     ttl,
		Type:    rtype,
	}
}
