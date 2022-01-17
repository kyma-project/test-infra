package app

import (
	"fmt"
	"os"
	"time"

	logf "github.com/sirupsen/logrus"
	slackGo "github.com/slack-go/slack"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"

	pkgConfig "github.com/kyma-project/test-infra/development/test-log-collector/pkg/config"
	"github.com/kyma-project/test-infra/development/test-log-collector/pkg/slack"

	pkgSlack "github.com/kyma-project/test-infra/development/test-log-collector/pkg/slack"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	restclient "k8s.io/client-go/rest"

	"github.com/kyma-project/test-infra/development/test-log-collector/pkg/resources/clustertestsuite"
	octopusTypes "github.com/kyma-project/test-infra/development/test-log-collector/pkg/resources/clustertestsuite/types"
)

type config struct {
	SlackToken     string
	ConfigLocation string
	ProwJobName    string `envconfig:"default=not-specified"` // for example post-main-kyma-gke-integration
}

func Mainerr() error {
	conf := &config{}
	if err := envconfig.InitWithPrefix(conf, "APP"); err != nil {
		return errors.Wrap(err, "while loading env config")
	}

	dispatchingConfig, err := pkgConfig.LoadDispatchingConfig(conf.ConfigLocation)
	if err != nil {
		return errors.Wrap(err, "while loading config for dispatching")
	}

	if err := dispatchingConfig.Validate(); err != nil {
		return errors.Wrap(err, "while validating dispatching configuration")
	}

	slackClient := slack.New(slackGo.New(conf.SlackToken))

	client := getRestConfigOrDie()

	clientset, err := kubernetes.NewForConfig(client)
	if err != nil {
		return errors.Wrap(err, "while creating clientset")
	}

	dynamicCli, err := dynamic.NewForConfig(client)
	if err != nil {
		return errors.Wrap(err, "while creating dynamicCli")
	}

	ctsCli := clustertestsuite.New(dynamicCli, 20*time.Second)

	ctsList, err := ctsCli.List()
	if err != nil {
		return errors.Wrapf(err, "while listing ClusterTestSuites")
	}

	cts, err := getClusterTestSuite(ctsList)
	if err != nil {
		return errors.Wrap(err, "while getting newest ClusterTestSuite")
	}

	logf.Infof("Newest ClusterTestSuite name: %s", cts.Name)

	logf.Info("Listing test pods")

	selector := labels.SelectorFromSet(map[string]string{
		octopusTypes.LabelKeyCreatedByOctopus: "true",
		octopusTypes.LabelKeySuiteName:        cts.Name,
	})

	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return errors.Wrapf(err, "while listing pods by %s selector", selector)
	}

	var messages []pkgSlack.Message

	completionTime := "Not yet completed"
	if cts.Status.CompletionTime != nil {
		completionTime = cts.Status.CompletionTime.String()
	}

	for _, pod := range pods.Items {
		testName, ok := pod.Labels[octopusTypes.LabelKeyTestDefName]
		if !ok {
			return fmt.Errorf("there's no `%s` label on a pod %s in namespace %s", octopusTypes.LabelKeyTestDefName, pod.Name, pod.Namespace)
		}

		testConfig, err := dispatchingConfig.GetConfigByNameWithFallback(testName)
		if err != nil {
			return errors.Wrapf(err, "while getting dispatching config for %s test suite", testName)
		}

		status, err := extractTestStatus(testName, cts)
		if err != nil {
			return errors.Wrapf(err, "while extracting test status from ClusterTestSuite for label %s", testName)
		}

		if status == octopusTypes.TestSucceeded && testConfig.OnlyReportFailure {
			logf.Infof("skipping report of %s test suite because it has status %s", testName, string(status))
			continue
		}

		container, err := getTestContainerName(pod)
		if err != nil {
			return errors.Wrapf(err, "while extracting test container name from pod %s in namespace %s", pod.Name, pod.Namespace)
		}
		logf.Info(fmt.Sprintf("Extracting logs from container %s from pod %s from namespace %s", container, pod.Name, pod.Namespace))
		req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
			Container: container,
		})

		msg := pkgSlack.Message{
			Data:        "",
			PodName:     pod.Name,
			ChannelName: testConfig.ChannelName,
			ChannelID:   testConfig.ChannelID,
			Attributes: pkgSlack.Attributes{
				Name:             testName,
				Status:           string(status),
				ClusterTestSuite: cts.Name,
				CompletionTime:   completionTime,
			},
		}

		data, err := req.DoRaw()
		if err != nil {
			// one of the possible things in test is like a failed init container, in which case we send error as string
			reqErr := errors.Wrapf(err, "while reading request from container %s in pod %s in namespace %s", container, pod.Name, pod.Namespace)
			logf.Warnf("couldn't read logs from pod %s, error: %s", pod.Name, reqErr)
			msg.Data = reqErr.Error()
		} else {
			msg.Data = string(data)
		}

		logf.Infof("Created msg for pod %s that will be sent to channel %s", pod.Name, msg.ChannelName)

		messages = append(messages, msg)
	}

	logf.Info("Uploading logs to appropriate slack thread")
	if err := slackClient.UploadLogFiles(messages, conf.ProwJobName, cts.Name, completionTime); err != nil {
		return errors.Wrap(err, "while uploading files to slack thread")
	}
	return nil
}

func getRestConfigOrDie() *restclient.Config {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		client, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(errors.Wrapf(err, "while creating restclient based on KUBECONFIG=%s", kubeconfig))
		}
		return client
	}

	client, err := restclient.InClusterConfig()
	if err != nil {
		panic(errors.Wrap(err, "while creating in cluster config"))
	}
	return client
}

func getTestContainerName(pod corev1.Pod) (string, error) {
	names := []string{}
	for _, cont := range pod.Spec.Containers {
		if cont.Name != "istio-proxy" {
			names = append(names, cont.Name)
		}
	}

	if len(names) != 1 {
		return "", fmt.Errorf("found more than 1 non-istio containers in pod %s in namespace %s", pod.Name, pod.Namespace)
	}

	return names[0], nil
}

func getClusterTestSuite(ctsList octopusTypes.ClusterTestSuiteList) (octopusTypes.ClusterTestSuite, error) {
	if len(ctsList.Items) == 0 {
		return octopusTypes.ClusterTestSuite{}, errors.New("there's no ClusterTestSuites")
	}

	// in our pipelines there's only 1 cts at a time
	return ctsList.Items[0], nil
}

func extractTestStatus(defName string, cts octopusTypes.ClusterTestSuite) (octopusTypes.TestStatus, error) {
	for _, result := range cts.Status.Results {
		if defName == result.Name {
			return result.Status, nil
		}
	}
	return "", fmt.Errorf("couldn't find %s test in %s ClusterTestSuite status", defName, cts.Name)
}
