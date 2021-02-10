package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var sampleData = []byte(`
{
	"action": "testing",
	"comment": {
		"author": "Rick Sanchez"
	}
}
`)

func TestValidatePayload(t *testing.T) {

	t.Run("Should return an error on empty eventType", func(t *testing.T) {
		//given
		v := validator{}
		payload := EventRequestPayload{
			"",
			"v1",
			"sampleEventID",
			"time",
			"github-connector-app",
			sampleData,
		}

		//when

		err := v.Validate(payload)

		//then
		assert.Error(t, err)
		assert.Equal(t, "eventType should not be empty", err.Error())

	})

	t.Run("Should return an error on empty eventTypeVersion", func(t *testing.T) {
		//given
		v := validator{}
		payload := EventRequestPayload{
			"sampleEventType",
			"",
			"sampleEventID",
			"time",
			"github-connector-app",
			sampleData,
		}

		//when
		err := v.Validate(payload)

		//then
		assert.Error(t, err)
		assert.Equal(t, "eventTypeVersion should not be empty", err.Error())

	})

	t.Run("Should return an error on empty data", func(t *testing.T) {
		//given
		v := validator{}
		payload := EventRequestPayload{
			"sampleEventType",
			"v1",
			"sampleEventID",
			"time",
			"github-connector-app",
			[]byte(``),
		}

		//when
		err := v.Validate(payload)

		//then
		assert.Error(t, err)
		assert.Equal(t, "data should not be empty", err.Error())

	})

	t.Run("Should return an error on empty SourceID", func(t *testing.T) {
		//given
		v := validator{}
		payload := EventRequestPayload{
			"sampleEventType",
			"v1",
			"sampleEventID",
			"time",
			"",
			sampleData,
		}

		//when
		err := v.Validate(payload)

		//then
		assert.Error(t, err)
		assert.Equal(t, "sourceID should not be empty", err.Error())

	})
}
