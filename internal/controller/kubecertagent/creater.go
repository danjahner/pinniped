// Copyright 2020 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kubecertagent

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/clock"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	pinnipedclientset "go.pinniped.dev/generated/1.19/client/clientset/versioned"
	"go.pinniped.dev/internal/constable"
	pinnipedcontroller "go.pinniped.dev/internal/controller"
	"go.pinniped.dev/internal/controllerlib"
)

type createrController struct {
	agentPodConfig                       *AgentPodConfig
	credentialIssuerConfigLocationConfig *CredentialIssuerConfigLocationConfig
	credentialIssuerConfigLabels         map[string]string
	clock                                clock.Clock
	k8sClient                            kubernetes.Interface
	pinnipedAPIClient                    pinnipedclientset.Interface
	kubeSystemPodInformer                corev1informers.PodInformer
	agentPodInformer                     corev1informers.PodInformer
}

// NewCreaterController returns a controller that creates new kube-cert-agent pods for every known
// kube-controller-manager pod.
//
// It also is tasked with updating the CredentialIssuerConfig, located via the provided
// credentialIssuerConfigLocationConfig, with any errors that it encounters.
func NewCreaterController(
	agentPodConfig *AgentPodConfig,
	credentialIssuerConfigLocationConfig *CredentialIssuerConfigLocationConfig,
	credentialIssuerConfigLabels map[string]string,
	clock clock.Clock,
	k8sClient kubernetes.Interface,
	pinnipedAPIClient pinnipedclientset.Interface,
	kubeSystemPodInformer corev1informers.PodInformer,
	agentPodInformer corev1informers.PodInformer,
	withInformer pinnipedcontroller.WithInformerOptionFunc,
	withInitialEvent pinnipedcontroller.WithInitialEventOptionFunc,
) controllerlib.Controller {
	return controllerlib.New(
		controllerlib.Config{
			//nolint: misspell
			Name: "kube-cert-agent-creater-controller",
			Syncer: &createrController{
				agentPodConfig:                       agentPodConfig,
				credentialIssuerConfigLocationConfig: credentialIssuerConfigLocationConfig,
				credentialIssuerConfigLabels:         credentialIssuerConfigLabels,
				clock:                                clock,
				k8sClient:                            k8sClient,
				pinnipedAPIClient:                    pinnipedAPIClient,
				kubeSystemPodInformer:                kubeSystemPodInformer,
				agentPodInformer:                     agentPodInformer,
			},
		},
		withInformer(
			kubeSystemPodInformer,
			pinnipedcontroller.SimpleFilter(isControllerManagerPod),
			controllerlib.InformerOption{},
		),
		withInformer(
			agentPodInformer,
			pinnipedcontroller.SimpleFilter(isAgentPod),
			controllerlib.InformerOption{},
		),
		// Be sure to run once even to make sure the CIC is updated if there are no controller manager
		// pods. We should be able to pass an empty key since we don't use the key in the sync (we sync
		// the world).
		withInitialEvent(controllerlib.Key{}),
	)
}

// Sync implements controllerlib.Syncer.
func (c *createrController) Sync(ctx controllerlib.Context) error {
	controllerManagerSelector, err := labels.Parse("component=kube-controller-manager")
	if err != nil {
		return fmt.Errorf("cannot create controller manager selector: %w", err)
	}

	controllerManagerPods, err := c.kubeSystemPodInformer.Lister().List(controllerManagerSelector)
	if err != nil {
		return fmt.Errorf("informer cannot list controller manager pods: %w", err)
	}

	if len(controllerManagerPods) == 0 {
		// If there are no controller manager pods, we alert the user that we can't find the keypair via
		// the CredentialIssuerConfig.
		return createOrUpdateCredentialIssuerConfig(
			ctx.Context,
			*c.credentialIssuerConfigLocationConfig,
			c.credentialIssuerConfigLabels,
			c.clock,
			c.pinnipedAPIClient,
			constable.Error("did not find kube-controller-manager pod(s)"),
		)
	}

	for _, controllerManagerPod := range controllerManagerPods {
		agentPod, err := findAgentPodForSpecificControllerManagerPod(
			controllerManagerPod,
			c.kubeSystemPodInformer,
			c.agentPodInformer,
			c.agentPodConfig.Labels(),
		)
		if err != nil {
			return err
		}
		if agentPod == nil {
			agentPod = newAgentPod(controllerManagerPod, c.agentPodConfig.PodTemplate())

			klog.InfoS(
				"creating agent pod",
				"pod",
				klog.KObj(agentPod),
				"controller",
				klog.KObj(controllerManagerPod),
			)
			_, err := c.k8sClient.CoreV1().
				Pods(c.agentPodConfig.Namespace).
				Create(ctx.Context, agentPod, metav1.CreateOptions{})
			if err != nil {
				err = fmt.Errorf("cannot create agent pod: %w", err)
				strategyResultUpdateErr := createOrUpdateCredentialIssuerConfig(
					ctx.Context,
					*c.credentialIssuerConfigLocationConfig,
					c.credentialIssuerConfigLabels,
					c.clock,
					c.pinnipedAPIClient,
					err,
				)
				if strategyResultUpdateErr != nil {
					// If the CIC update fails, then we probably want to try again. This controller will get
					// called again because of the pod create failure, so just try the CIC update again then.
					klog.ErrorS(strategyResultUpdateErr, "could not create or update CredentialIssuerConfig")
				}

				return err
			}
		}

		// The deleter controller handles the case where the expected fields do not match in the agent pod.
	}

	return nil
}

func findAgentPodForSpecificControllerManagerPod(
	controllerManagerPod *corev1.Pod,
	kubeSystemPodInformer corev1informers.PodInformer,
	agentPodInformer corev1informers.PodInformer,
	agentLabels map[string]string,
) (*corev1.Pod, error) {
	agentSelector := labels.SelectorFromSet(agentLabels)
	agentPods, err := agentPodInformer.
		Lister().
		List(agentSelector)
	if err != nil {
		return nil, fmt.Errorf("informer cannot list agent pods: %w", err)
	}

	for _, maybeAgentPod := range agentPods {
		maybeControllerManagerPod, err := findControllerManagerPodForSpecificAgentPod(
			maybeAgentPod,
			kubeSystemPodInformer,
		)
		if err != nil {
			return nil, err
		}
		if maybeControllerManagerPod != nil &&
			maybeControllerManagerPod.UID == controllerManagerPod.UID {
			return maybeAgentPod, nil
		}
	}

	return nil, nil
}
