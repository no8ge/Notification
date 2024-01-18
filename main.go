package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/no8geo/notify/pkg/k8s"
	router "github.com/no8geo/notify/router"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"github.com/olahol/melody"
)

func main() {

	r := gin.Default()
	m := melody.New()

	clientset, err := k8s.Client()
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	router.V1(r, m)

	stopper := make(chan struct{})
	defer close(stopper)

	factory := informers.NewSharedInformerFactory(clientset, 0)
	podInformer := factory.Core().V1().Pods()
	informer := podInformer.Informer()
	defer runtime.HandleCrash()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			podsJson, err := json.Marshal(pod)
			if err != nil {
				panic(err)
			}
			m.BroadcastFilter(podsJson, func(s *melody.Session) bool {
				return s.Request.RequestURI == "/v1/ws/watch"
			})
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			oldPod := old.(*corev1.Pod)
			newPod := new.(*corev1.Pod)
			if newPod.Status.Phase != oldPod.Status.Phase {
				log.Printf("Pod %s status updated %s -> %s \n",
					newPod.Name, oldPod.Status.Phase, newPod.Status.Phase)

				diff := map[string]interface{}{
					"name": newPod.Name,
					"diff": fmt.Sprintf("Pod %s status updated %s to %s",
						newPod.Name, oldPod.Status.Phase, newPod.Status.Phase),
				}
				diffJson, err := json.Marshal(diff)
				if err != nil {
					panic(err)
				}
				m.BroadcastFilter(diffJson, func(s *melody.Session) bool {
					return s.Request.RequestURI == "/v1/ws/watch"
				})
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			if err != nil {
				panic(err)
			}
			resp := map[string]interface{}{
				"name":   pod.Name,
				"status": "deleted",
			}
			respJson, err := json.Marshal(resp)
			if err != nil {
				panic(err)
			}
			m.BroadcastFilter(respJson, func(s *melody.Session) bool {
				return s.Request.RequestURI == "/v1/ws/watch"
			})
		},
	})

	go informer.Run(stopper)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	go func(t *time.Ticker) {
		for {
			<-t.C
			pods, err := podInformer.Lister().List(labels.Everything())
			if err != nil {
				panic(err.Error())
			}
			podsJson, err := json.Marshal(pods)
			if err != nil {
				panic(err)
			}
			m.BroadcastFilter(podsJson, func(s *melody.Session) bool {
				return s.Request.RequestURI == "/v1/ws/push"
			})
		}
	}(ticker)

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		pods, err := podInformer.Lister().List(labels.Everything())
		if err != nil {
			panic(err.Error())
		}
		podsJson, err := json.Marshal(pods)
		if err != nil {
			panic(err)
		}
		m.BroadcastFilter(podsJson, func(s *melody.Session) bool {
			return s.Request.RequestURI == "/v1/ws/pull"
		})

	})
	r.Run(":8081")

}
