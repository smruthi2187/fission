package main

import (
	"fmt"
	"os"
	"github.com/fission/fission/crd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/client-go/rest"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

func runportForward(localPort string) error {

	//KUBECONFIG needs to be set to the correct path i.e ~/.kube/config
	config, PodClient, _, err := crd.GetKubernetesClient()
	if err != nil {
		msg := fmt.Sprint("%v\n", err)
		fatal(msg)
	}

	//get the podname for the controller
	PodList, err := PodClient.CoreV1().Pods("").List(meta_v1.ListOptions{LabelSelector:"svc=controller"})
	if err != nil {
		fatal("Error getting PodList with selector")
	}
	var podName string
	//there should only be one Pod in this list, the controller pod
	for _, item := range PodList.Items {

		podName = item.Name
		break
	}


	RESTClient, err := rest.RESTClientFor(config)

	if err != nil {
		msg := fmt.Sprintf("%v", err)
		fatal(msg)
	}

	StopChannel := make(chan struct{}, 1)
	ReadyChannel := make(chan struct{})

	//create request URL
	req := RESTClient.Post().Resource("pods").Namespace("").Name(podName).SubResource("portforward")

	url := req.URL()

	//create ports slice
	ports := []string{localPort, "8888"}

	//actually start the port-forwarding process here
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		fatal(msg)
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport:transport}, "POST", url)
	fw, err := portforward.New(dialer, ports , StopChannel, ReadyChannel, os.Stdout, os.Stderr)

	if err != nil {
		msg := fmt.Sprintf("%v", err)
		fatal(msg)
	}
	return fw.ForwardPorts()
}