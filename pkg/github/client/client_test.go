package client

import (
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSapToolsClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SapToolsClient Suite")
}

var _ = Describe("SapToolsClient", func() {
	var (
		client *Client
	)

	BeforeEach(func() {
		client = &Client{
			WrapperClientMu: sync.RWMutex{},
		}
	})

	It("should Lock the mutex", func() {

		client.MuLock()
		Expect(client.WrapperClientMu.TryLock()).To(BeFalse(), "Mutex should be locked")
		client.MuUnlock()
		Expect(client.WrapperClientMu.TryLock()).To(BeTrue(), "Mutex should be unlocked")
	})

	It("should Lock the Read mutex", func() {
		client.MuRLock()
		Expect(client.WrapperClientMu.TryLock()).To(BeFalse(), "Read mutex should be locked")
		client.MuRUnlock()
		Expect(client.WrapperClientMu.TryLock()).To(BeTrue(), "Read mutex should be unlocked")
	})
})
