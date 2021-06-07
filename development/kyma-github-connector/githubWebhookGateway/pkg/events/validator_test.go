package events

import (
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
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
		event := cloudevents.NewEvent()
		event.SetType("")
		event.SetSource("github-connector-app")
		event.SetData(cloudevents.ApplicationJSON, sampleData)

		//when

		err := v.Validate(event)

		//then
		assert.Error(t, err)
		assert.Equal(t, "cloudevent type should not be empty", err.Error())

	})

	t.Run("Should return an error on empty data", func(t *testing.T) {
		//given
		v := validator{}
		event := cloudevents.NewEvent()
		event.SetType("sampleEventType")
		event.SetSource("github-connector-app")
		event.SetData(cloudevents.ApplicationJSON, []byte(``))

		//when
		err := v.Validate(event)

		//then
		assert.Error(t, err)
		assert.Equal(t, "cloudevent data should not be empty", err.Error())

	})

	t.Run("Should return an error on empty SourceID", func(t *testing.T) {
		//given
		v := validator{}
		event := cloudevents.NewEvent()
		event.SetType("sampleEventType")
		event.SetSource("")
		event.SetData(cloudevents.ApplicationJSON, sampleData)

		//when
		err := v.Validate(event)

		//then
		assert.Error(t, err)
		assert.Equal(t, "cloudevent source should not be empty", err.Error())

	})
}
