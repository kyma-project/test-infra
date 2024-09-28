package tags

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestOption_PRNumber_success(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := struct {
		pr       string
		expected string
	}{
		pr:       "123",
		expected: "123",
	}
	tag := Tagger{}
	f := PRNumber(tc.pr)
	err := f(&tag)

	g.Expect(err).To(BeNil())
	g.Expect(tag.PRNumber).To(Equal(tc.expected))
}

func TestOption_PRNumber_return_error_when_empty_pr(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := struct {
		pr string
	}{
		pr: "",
	}
	tag := Tagger{}
	f := PRNumber(tc.pr)
	err := f(&tag)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("pr number cannot be empty"))
}

func TestOption_CommitSHA_success(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := struct {
		sha      string
		expected string
	}{
		sha:      "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		expected: "da39a3ee5e6b4b0d3255bfef95601890afd80709",
	}
	tag := Tagger{
		Time:      time.Now(),
		CommitSHA: "1edd8d99e07c726c2226713312ae9551162b825b",
	}
	f := CommitSHA(tc.sha)
	f(&tag)

	g.Expect(tag.CommitSHA).To(Equal(tc.expected))
}

func TestOption_CommitSHA_return_error_when_empty_sha(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := struct {
		sha string
	}{
		sha: "",
	}
	tag := Tagger{}
	f := CommitSHA(tc.sha)
	err := f(&tag)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("sha cannot be empty"))
}

func TestOption_DateFormat_success(t *testing.T) {
	g := NewGomegaWithT(t)

	now := time.Now()
	tc := struct {
		dateFormat   string
		expectedDate string
	}{
		dateFormat:   "2006-01-02",
		expectedDate: now.Format("2006-01-02"),
	}
	tag := Tagger{
		Time: now,
	}
	f := DateFormat(tc.dateFormat)
	f(&tag)

	g.Expect(tag.Date).To(Equal(tc.expectedDate))
}

func TestOption_DateFormat_return_error_when_empty_date_format(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := struct {
		dateFormat string
	}{
		dateFormat: "",
	}
	tag := Tagger{
		Time: time.Now(),
	}
	f := DateFormat(tc.dateFormat)
	err := f(&tag)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("date format cannot be empty"))
}
