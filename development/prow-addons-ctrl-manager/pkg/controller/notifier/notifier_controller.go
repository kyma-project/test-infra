package notifier

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/test-infra/development/prow-addons-ctrl-manager/pkg/slack"
	apiSlack "github.com/nlopes/slack"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	prowjobsv1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Config struct {
	SlackToken         string
	SlackReportChannel string
	ActOnProwJobType   []prowjobsv1.ProwJobType
}

// Add creates a new ProwJob Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	var cfg Config
	if err := envconfig.InitWithPrefix(&cfg, "NOTIFIER"); err != nil {
		return errors.Wrap(err, "while initializing configuration for notifier controller")
	}

	slackCli := apiSlack.New(cfg.SlackToken)
	r := &ReconcileProwJob{
		k8sCli:   mgr.GetClient(),
		reporter: slack.NewReporter(slackCli, cfg.SlackReportChannel, cfg.ActOnProwJobType),
		scheme:   mgr.GetScheme(),
		log:      log.Log.WithName("ctrl:notifier"),
	}

	return add(mgr, r)
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("prowjob-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to ProwJob
	err = c.Watch(&source.Kind{Type: &prowjobsv1.ProwJob{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileProwJob{}

// ReconcileProwJob reconciles a ProwJob object
type ReconcileProwJob struct {
	k8sCli   client.Client
	scheme   *runtime.Scheme
	reporter *slack.Client
	log      logr.Logger
}

// Reconcile reads that state of the cluster for a ProwJob object and makes changes based on the state read
// and what is in the ProwJob.Spec
// +kubebuilder:rbac:groups=prowjobs.prow.k8s.io,resources=prowjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=prowjobs.prow.k8s.io,resources=prowjobs/status,verbs=get;update;patch
func (r *ReconcileProwJob) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	var (
		ctx         = context.TODO()
		pj          = &prowjobsv1.ProwJob{}
		infoLog     = r.log.WithValues("prowjob", request.NamespacedName).Info
		debugLog    = r.log.WithValues("prowjob", request.NamespacedName).V(1).Info
		actionTaken = false
	)

	//infoLog("Start reconcile")
	//defer func() { infoLog("End reconcile", "actionTaken", actionTaken) }()

	if err := r.k8sCli.Get(ctx, request.NamespacedName, pj); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	infoLog("Start reconcile")
	defer func() { infoLog("End reconcile", "actionTaken", actionTaken, "name", pj.Spec.Job) }()

	if !r.reporter.ShouldReport(pj) {
		debugLog("Decided that reporter should not act on ProwJob", "reporter", r.reporter.GetName())
		return reconcile.Result{}, nil
	}

	if r.alreadyReported(pj) {
		debugLog("Reporter already act on that ProwJob state", "reporter", r.reporter.GetName(), "state", pj.Status.State)
		return reconcile.Result{}, nil
	}

	debugLog("Report ProwJob", "reporter", r.reporter.GetName(), "state", pj.Status.State)
	if err := r.reporter.Report(pj); err != nil {
		return reconcile.Result{}, err
	}
	actionTaken = true

	debugLog("Report finished, now updating ProwJob to mark reporter as informed", "reporter", r.reporter.GetName())
	err := retry.RetryOnConflict(retry.DefaultBackoff, r.updateProwJobFn(ctx, request.NamespacedName))
	if err != nil {
		// may be conflict if max retries were hit
		return reconcile.Result{}, err
	}

	debugLog("ProwJob updated", "reporter", r.reporter.GetName())
	return reconcile.Result{}, nil
}
func (r *ReconcileProwJob) updateProwJobFn(ctx context.Context, name types.NamespacedName) func() error {
	return func() error {
		pj := &prowjobsv1.ProwJob{}
		if err := r.k8sCli.Get(ctx, name, pj); err != nil {
			return err
		}

		// update pj report status
		cpy := pj.DeepCopy()
		// we set omitempty on PrevReportStates, so here we need to init it if is nil
		if cpy.Status.PrevReportStates == nil {
			cpy.Status.PrevReportStates = map[string]prowjobsv1.ProwJobState{}
		}
		cpy.Status.PrevReportStates[r.reporter.GetName()] = cpy.Status.State

		return r.k8sCli.Update(ctx, cpy)
	}
}
func (r *ReconcileProwJob) alreadyReported(pj *prowjobsv1.ProwJob) bool {
	// we set omitempty on PrevReportStates, so here we need to init it if is nil
	if pj.Status.PrevReportStates == nil {
		pj.Status.PrevReportStates = map[string]prowjobsv1.ProwJobState{}
	}

	// already reported current state
	if pj.Status.PrevReportStates[r.reporter.GetName()] == pj.Status.State {
		return true
	}

	return false
}
