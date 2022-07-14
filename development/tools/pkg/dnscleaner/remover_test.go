package dnscleaner

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/pkg/dnscleaner/automock"
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

	testProject = "testProject"
	testZone    = "testRegion"
)

func TestNew(t *testing.T) {
	t.Run("Should return and delete DNS entry", func(t *testing.T) {
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)

		record := createDNSResourceRecordSet(shouldDeleteDNSName, shouldDeleteDNSIP, shouldDeleteDNSRecordType, shouldDeleteDNSTTL)

		ctx := context.Background()

		//Given
		mockDNSAPI.On("LookupDNSEntry", ctx, testProject, testZone, shouldDeleteDNSName, shouldDeleteDNSIP, shouldDeleteDNSRecordType, shouldDeleteDNSTTL).Return(record, nil)
		mockDNSAPI.On("RemoveDNSEntry", ctx, testProject, testZone, record).Return(nil)

		//When
		der := New(mockDNSAPI, 3, 2, true)
		err := der.Run(testProject, testZone, shouldDeleteDNSName, shouldDeleteDNSIP, shouldDeleteDNSRecordType, shouldDeleteDNSTTL)

		//Then
		mockDNSAPI.AssertNumberOfCalls(t, "LookupDNSEntry", 1)
		mockDNSAPI.AssertNumberOfCalls(t, "RemoveDNSEntry", 1)
		require.NoError(t, err)
	})

	t.Run("Should not delete DNS entry", func(t *testing.T) {
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)

		ctx := context.Background()

		//Given
		mockDNSAPI.On("LookupDNSEntry", ctx, testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL).Return(nil, errors.New("test error"))

		//When
		der := New(mockDNSAPI, 3, 2, true)
		err := der.Run(testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL)

		//Then
		mockDNSAPI.AssertNotCalled(t, "RemoveDNSEntry")
		mockDNSAPI.AssertNumberOfCalls(t, "LookupDNSEntry", 3)
		require.Error(t, err)
	})
}

func TestRetryableCalls(t *testing.T) {
	t.Run("Should retry 3 times and then throw error", func(t *testing.T) {
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)

		ctx := context.Background()

		//Given
		mockDNSAPI.On("LookupDNSEntry", ctx, testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL).Return(nil, errors.New("test error")).Times(3)

		//When
		der := New(mockDNSAPI, 3, 2, true)
		err := der.Run(testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL)

		//Then
		mockDNSAPI.AssertNotCalled(t, "RemoveDNSEntry")
		mockDNSAPI.AssertNumberOfCalls(t, "LookupDNSEntry", 3)
		require.Error(t, err)
	})
}

func TestRetryableCallsPartTwo(t *testing.T) {
	t.Run("Should retry 3 times and then succeed lookup and fail on remove", func(t *testing.T) {
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)

		record := createDNSResourceRecordSet(shouldDeleteDNSName, shouldDeleteDNSIP, shouldDeleteDNSRecordType, shouldDeleteDNSTTL)

		ctx := context.Background()

		//Given
		mockDNSAPI.On("LookupDNSEntry", ctx, testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL).Return(nil, errors.New("test error")).Times(2)
		mockDNSAPI.On("LookupDNSEntry", ctx, testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL).Return(record, nil)
		mockDNSAPI.On("RemoveDNSEntry", ctx, testProject, testZone, record).Return(errors.New("test error")).Times(3)

		//When
		der := New(mockDNSAPI, 3, 2, true)
		err := der.Run(testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL)

		//Then
		mockDNSAPI.AssertNumberOfCalls(t, "LookupDNSEntry", 3)
		mockDNSAPI.AssertNumberOfCalls(t, "RemoveDNSEntry", 3)
		require.Error(t, err)
	})
}

func TestRetryableCallsPartThree(t *testing.T) {
	t.Run("Should retry 3 times and then succeed lookup and retry 3 times for delete and succeed", func(t *testing.T) {
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)

		record := createDNSResourceRecordSet(shouldDeleteDNSName, shouldDeleteDNSIP, shouldDeleteDNSRecordType, shouldDeleteDNSTTL)

		ctx := context.Background()

		//Given
		mockDNSAPI.On("LookupDNSEntry", ctx, testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL).Return(nil, errors.New("test error")).Times(2)
		mockDNSAPI.On("LookupDNSEntry", ctx, testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL).Return(record, nil)
		mockDNSAPI.On("RemoveDNSEntry", ctx, testProject, testZone, record).Return(errors.New("test error")).Times(2)
		mockDNSAPI.On("RemoveDNSEntry", ctx, testProject, testZone, record).Return(nil)

		//When
		der := New(mockDNSAPI, 3, 1, true)
		err := der.Run(testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL)

		//Then
		mockDNSAPI.AssertNumberOfCalls(t, "RemoveDNSEntry", 3)
		mockDNSAPI.AssertNumberOfCalls(t, "LookupDNSEntry", 3)
		require.NoError(t, err)
	})
}

func TestDryRunBehaviour(t *testing.T) {
	t.Run("Should not delete on dryrun", func(t *testing.T) {
		mockDNSAPI := &automock.DNSAPI{}
		defer mockDNSAPI.AssertExpectations(t)

		record := createDNSResourceRecordSet(shouldDeleteDNSName, shouldDeleteDNSIP, shouldDeleteDNSRecordType, shouldDeleteDNSTTL)

		ctx := context.Background()

		//Given
		mockDNSAPI.On("LookupDNSEntry", ctx, testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL).Return(record, nil)

		//When
		der := New(mockDNSAPI, 3, 2, false)
		err := der.Run(testProject, testZone, shouldNotDeleteDNSName, shouldNotDeleteDNSIP, shouldNotDeleteDNSRecordType, shouldNotDeleteDNSTTL)

		//Then
		mockDNSAPI.AssertNumberOfCalls(t, "RemoveDNSEntry", 0)
		mockDNSAPI.AssertNumberOfCalls(t, "LookupDNSEntry", 1)
		require.NoError(t, err)
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
