package jobsuite

import "testing"

type Suite interface {
	Run(t *testing.T)
}
