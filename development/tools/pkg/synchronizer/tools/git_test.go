package tools

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_generateHistoryFromCommandResult(t *testing.T) {
	// Given
	commandResult := "\"bc120200aa9d0d42b1dca21d4f13fa9853529ec0-1549638124\"\n\"4965746e63fef409dd04ba39d2528958dd4dfea3-1548768400\"\n\"0f7a9d1a0110213344f19792d043dea826e1f0ea-1544519466\"\n"
	wrongCommandResult := "wrong:result"

	// When
	result, err := generateHistoryFromCommandResult(commandResult)

	// Then
	assert.Equal(t, map[int64]string{
		1549638124: "bc120200aa9d0d42b1dca21d4f13fa9853529ec0",
		1548768400: "4965746e63fef409dd04ba39d2528958dd4dfea3",
		1544519466: "0f7a9d1a0110213344f19792d043dea826e1f0ea",
	}, result)
	assert.Empty(t, err)

	// When
	resultWrong, errWrong := generateHistoryFromCommandResult(wrongCommandResult)

	// Then
	assert.NotEmpty(t, errWrong)
	assert.Empty(t, resultWrong)
}

func Test_limitGitHistoryTimePeriod(t *testing.T) {
	// Given
	history := map[int64]string{
		subtractDaysFromToday(1): "a1a1a1a1",
		subtractDaysFromToday(3): "b2b2b2b2",
		subtractDaysFromToday(2): "c3c3c3c3",
		subtractDaysFromToday(7): "d4d4d4d4",
		subtractDaysFromToday(6): "e5e5e5e5",
		subtractDaysFromToday(4): "f6f6f6f6",
		subtractDaysFromToday(0): "g7g7g7g7",
	}

	// When
	result := limitGitHistoryTimePeriod(history, 4)

	// Then
	assert.Len(t, result, 5)
	values := []string{}
	for _, val := range result {
		values = append(values, val)
	}
	assert.Contains(t, values, "a1a1a1a1")
	assert.Contains(t, values, "b2b2b2b2")
	assert.Contains(t, values, "c3c3c3c3")
	assert.Contains(t, values, "f6f6f6f6")
	assert.Contains(t, values, "g7g7g7g7")
}

func subtractDaysFromToday(subDays int) int64 {
	if subDays > 0 {
		return time.Now().AddDate(0, 0, -(subDays)).Unix()
	}

	return time.Now().Unix()
}
