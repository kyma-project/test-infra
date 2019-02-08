package notifier

import (
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/prow-addons-ctrl-manager/pkg/controller/notifier/automock"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	prowjobsv1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const timeout = time.Second * 5

func TestReconcile(t *testing.T) {
	// given
	var (
		g = NewGomegaWithT(t)

		givenPJ          = fixFailureProwJob()
		expRequest       = reconcileReqForPJ(givenPJ)
		mockReporterName = "mock-reporter"
	)

	reporterMock := &automock.ReportClient{}
	defer reporterMock.AssertExpectations(t)

	reporterMock.On("GetName").Return(mockReporterName)

	reporterMock.On("ShouldReport", mock.AnythingOfType("*v1.ProwJob")).Return(true).Once()
	reporterMock.On("ShouldReport", mock.AnythingOfType("*v1.ProwJob")).Return(false).Once()

	reporterMock.On("Report", mock.AnythingOfType("*v1.ProwJob")).Return(nil).Once()

	// Setup the Manager and Controller. Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(HaveOccurred())
	c := mgr.GetClient()

	r := &ReconcileProwJob{
		k8sCli:   c,
		reporter: reporterMock,
		scheme:   mgr.GetScheme(),
		log:      log.Log.WithName("ctrl:notifier"),
	}

	recFn, requests := SetupTestReconcile(r)
	err = add(mgr, recFn)
	g.Expect(err).NotTo(HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// when
	err = c.Create(context.TODO(), givenPJ)

	// then
	g.Expect(err).NotTo(HaveOccurred())
	defer c.Delete(context.TODO(), givenPJ)

	// first call after ProwJob creation
	g.Eventually(requests, timeout).Should(Receive(Equal(expRequest)))
	// second call after ProwJob Status update
	g.Eventually(requests, timeout).Should(Receive(Equal(expRequest)))

	// Check if Status was updated
	updatePJ := prowjobsv1.ProwJob{}
	err = c.Get(context.TODO(), client.ObjectKey{Name: givenPJ.Name, Namespace: givenPJ.Namespace}, &updatePJ)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(updatePJ.Status.PrevReportStates).Should(HaveKeyWithValue(mockReporterName, givenPJ.Status.State))
}

func fixFailureProwJob() *prowjobsv1.ProwJob {
	return &prowjobsv1.ProwJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: prowjobsv1.ProwJobSpec{
			Job: "test-prow-job",
		},
		Status: prowjobsv1.ProwJobStatus{
			State: "failure",
		},
	}
}

func reconcileReqForPJ(job *prowjobsv1.ProwJob) reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{Name: job.Name, Namespace: job.Namespace}}
}
