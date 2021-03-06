package apperrors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {

	t.Run("should create error with proper code", func(t *testing.T) {
		assert.Equal(t, InternalError, Internal("error").Code())
		assert.Equal(t, NotFoundError, NotFound("error").Code())
		assert.Equal(t, AlreadyExistsError, AlreadyExists("error").Code())
		assert.Equal(t, WrongInputError, WrongInput("error").Code())
		assert.Equal(t, UpstreamServerCallFailedError, UpstreamServerCallFailed("error").Code())
		assert.Equal(t, AuthenticationFailedError, AuthenticationFailed("error").Code())
	})

	t.Run("should create error with simple message", func(t *testing.T) {
		assert.Equal(t, "error", Internal("error").Error())
		assert.Equal(t, "error", NotFound("error").Error())
		assert.Equal(t, "error", AlreadyExists("error").Error())
		assert.Equal(t, "error", WrongInput("error").Error())
		assert.Equal(t, "error", UpstreamServerCallFailed("error").Error())
		assert.Equal(t, "error", AuthenticationFailed("error").Error())
	})

	t.Run("should create error with formatted message", func(t *testing.T) {
		assert.Equal(t, "code: 1, error: bug", Internal("code: %d, error: %s", 1, "bug").Error())
		assert.Equal(t, "code: 1, error: bug", NotFound("code: %d, error: %s", 1, "bug").Error())
		assert.Equal(t, "code: 1, error: bug", AlreadyExists("code: %d, error: %s", 1, "bug").Error())
		assert.Equal(t, "code: 1, error: bug", WrongInput("code: %d, error: %s", 1, "bug").Error())
		assert.Equal(t, "code: 1, error: bug", UpstreamServerCallFailed("code: %d, error: %s", 1, "bug").Error())
		assert.Equal(t, "code: 1, error: bug", AuthenticationFailed("code: %d, error: %s", 1, "bug").Error())
	})

	t.Run("should append apperrors without changing error code", func(t *testing.T) {
		//given
		createdInternalErr := Internal("Some Internal apperror, %s", "Some pkg err")
		createdNotFoundErr := NotFound("Some NotFound apperror, %s", "Some pkg err")
		createdAlreadyExistsErr := AlreadyExists("Some AlreadyExists apperror, %s", "Some pkg err")
		createdWrongInputErr := WrongInput("Some WrongInput apperror, %s", "Some pkg err")
		createdUpstreamServerCallFailedErr := UpstreamServerCallFailed("Some UpstreamServerCallFailed apperror, %s", "Some pkg err")
		createdAuthenticationFailedErr := AuthenticationFailed("Some AuthenticationFailed apperror, %s", "Some pkg err")

		//when
		appendedInternalErr := createdInternalErr.Append("Some additional message")
		appendedNotFoundErr := createdNotFoundErr.Append("Some additional message")
		appendedAlreadyExistsErr := createdAlreadyExistsErr.Append("Some additional message")
		appendedWrongInputErr := createdWrongInputErr.Append("Some additional message")
		appendedUpstreamServerCallFailedErr := createdUpstreamServerCallFailedErr.Append("Some additional message")
		appendedAuthenticationFailedErr := createdAuthenticationFailedErr.Append("Some additional message")

		//then
		assert.Equal(t, InternalError, appendedInternalErr.Code())
		assert.Equal(t, NotFoundError, appendedNotFoundErr.Code())
		assert.Equal(t, AlreadyExistsError, appendedAlreadyExistsErr.Code())
		assert.Equal(t, WrongInputError, appendedWrongInputErr.Code())
		assert.Equal(t, UpstreamServerCallFailedError, appendedUpstreamServerCallFailedErr.Code())
		assert.Equal(t, AuthenticationFailedError, appendedAuthenticationFailedErr.Code())
	})

	t.Run("should append apperrors and chain messages correctly", func(t *testing.T) {
		//given
		createdInternalErr := Internal("Some Internal apperror, %s", "Some pkg err")
		createdNotFoundErr := NotFound("Some NotFound apperror, %s", "Some pkg err")
		createdAlreadyExistsErr := AlreadyExists("Some AlreadyExists apperror, %s", "Some pkg err")
		createdWrongInputErr := WrongInput("Some WrongInput apperror, %s", "Some pkg err")
		createdUpstreamServerCallFailedErr := UpstreamServerCallFailed("Some UpstreamServerCallFailed apperror, %s", "Some pkg err")
		createdAuthenticationFailedErr := AuthenticationFailed("Some AuthenticationFailed apperror, %s", "Some pkg err")

		//when
		appendedInternalErr := createdInternalErr.Append("Some additional message: %s", "error")
		appendedNotFoundErr := createdNotFoundErr.Append("Some additional message: %s", "error")
		appendedAlreadyExistsErr := createdAlreadyExistsErr.Append("Some additional message: %s", "error")
		appendedWrongInputErr := createdWrongInputErr.Append("Some additional message: %s", "error")
		appendedUpstreamServerCallFailedErr := createdUpstreamServerCallFailedErr.Append("Some additional message: %s", "error")
		appendedAuthenticationFailedErr := createdAuthenticationFailedErr.Append("Some additional message: %s", "error")

		//then
		assert.Equal(t, "Some additional message: error, Some Internal apperror, Some pkg err", appendedInternalErr.Error())
		assert.Equal(t, "Some additional message: error, Some NotFound apperror, Some pkg err", appendedNotFoundErr.Error())
		assert.Equal(t, "Some additional message: error, Some AlreadyExists apperror, Some pkg err", appendedAlreadyExistsErr.Error())
		assert.Equal(t, "Some additional message: error, Some WrongInput apperror, Some pkg err", appendedWrongInputErr.Error())
		assert.Equal(t, "Some additional message: error, Some UpstreamServerCallFailed apperror, Some pkg err", appendedUpstreamServerCallFailedErr.Error())
		assert.Equal(t, "Some additional message: error, Some AuthenticationFailed apperror, Some pkg err", appendedAuthenticationFailedErr.Error())
	})
}
