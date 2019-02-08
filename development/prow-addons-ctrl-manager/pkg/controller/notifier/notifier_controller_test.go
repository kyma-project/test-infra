package notifier

import (
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
	"time"

	prowjobsv1 "k8s.io/test-infra/prow/apis/prowjobs/v1"

	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

const timeout = time.Second * 5

func TestReconcile(t *testing.T) {
	// given
	g := NewGomegaWithT(t)
	expRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}}
	instance := &prowjobsv1.ProwJob{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"}}

	// Setup the Manager and Controller. Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(HaveOccurred())
	c = mgr.GetClient()

	r := &ReconcileProwJob{
		k8sCli:   c,
		reporter: nil,
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
	err = c.Create(context.TODO(), instance)

	// then
	g.Expect(err).NotTo(HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	g.Eventually(requests, timeout).Should(Receive(Equal(expRequest)))

}
