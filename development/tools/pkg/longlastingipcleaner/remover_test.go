package ipcleaner

import (
	"errors"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/pkg/longlastingipcleaner/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	shouldDeleteIPByName    = "this-ip-is-a-delete-candidate"
	shouldNotDeleteIPByName = "this-ip-stays"
	testProject             = "testProject"
	testRegion              = "testRegion"
)

func TestIPRemover(t *testing.T) {
	t.Run("Should delete IP", func(t *testing.T) {
		mockAddressAPI := &automock.ComputeAPI{}
		defer mockAddressAPI.AssertExpectations(t)

		//Given
		mockAddressAPI.On("RemoveIP", testProject, testRegion, shouldDeleteIPByName).Return(false, nil)

		//When
		ipr := NewIPRemover(mockAddressAPI)
		success, err := ipr.Run(testProject, testRegion, shouldDeleteIPByName, 5, 20, true)

		//Then
		mockAddressAPI.AssertCalled(t, "RemoveIP", testProject, testRegion, shouldDeleteIPByName)
		mockAddressAPI.AssertNumberOfCalls(t, "RemoveIP", 1)
		require.NoError(t, err)
		assert.Equal(t, success, true)
	})

	t.Run("Should not delete IP", func(t *testing.T) {
		mockAddressAPI := &automock.ComputeAPI{}
		defer mockAddressAPI.AssertExpectations(t)

		//Given
		mockAddressAPI.On("RemoveIP", testProject, testRegion, shouldNotDeleteIPByName).Return(false, errors.New("testError"))

		//When
		ipr := NewIPRemover(mockAddressAPI)
		success, err := ipr.Run(testProject, testRegion, shouldNotDeleteIPByName, 5, 20, true)

		//Then
		mockAddressAPI.AssertNumberOfCalls(t, "RemoveIP", 1)
		require.Error(t, err)
		assert.Equal(t, success, false)
	})

	t.Run("Should retry 3 times and then throw error", func(t *testing.T) {
		mockAddressAPI := &automock.ComputeAPI{}
		defer mockAddressAPI.AssertExpectations(t)

		//Given
		mockAddressAPI.On("RemoveIP", testProject, testRegion, shouldNotDeleteIPByName).Return(true, errors.New("testError")).Times(3)

		//When
		ipr := NewIPRemover(mockAddressAPI)
		success, err := ipr.Run(testProject, testRegion, shouldNotDeleteIPByName, 3, 2, true)

		//Then
		mockAddressAPI.AssertNumberOfCalls(t, "RemoveIP", 3)
		require.Error(t, err)
		assert.Equal(t, success, false)
	})

	t.Run("Should retry 3 times and then succeed", func(t *testing.T) {
		mockAddressAPI := &automock.ComputeAPI{}
		defer mockAddressAPI.AssertExpectations(t)

		//Given
		mockAddressAPI.On("RemoveIP", testProject, testRegion, shouldDeleteIPByName).Return(true, errors.New("testError")).Twice()
		mockAddressAPI.On("RemoveIP", testProject, testRegion, shouldDeleteIPByName).Return(false, nil)

		//When
		ipr := NewIPRemover(mockAddressAPI)
		success, err := ipr.Run(testProject, testRegion, shouldDeleteIPByName, 3, 2, true)

		//Then
		mockAddressAPI.AssertNumberOfCalls(t, "RemoveIP", 3)
		require.NoError(t, err)
		assert.Equal(t, success, true)
	})

	t.Run("Should not delete on dryrun", func(t *testing.T) {
		mockAddressAPI := &automock.ComputeAPI{}
		defer mockAddressAPI.AssertExpectations(t)

		//When
		ipr := NewIPRemover(mockAddressAPI)
		success, err := ipr.Run(testProject, testRegion, shouldDeleteIPByName, 3, 2, false)

		//Then
		mockAddressAPI.AssertNumberOfCalls(t, "RemoveIP", 0)
		require.NoError(t, err)
		assert.Equal(t, success, true)
	})
}
