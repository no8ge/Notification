package tests

import (
	"context"
	"fmt"
	"testing"

	k8s "github.com/no8geo/notify/pkg/k8s"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClient(*testing.T) {

	clientset, err := k8s.Client()
	if err != nil {
		panic(err.Error())
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

}
