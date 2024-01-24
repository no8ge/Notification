package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	types "github.com/no8geo/notify/pkg"

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
			var c types.Cast
			pod := obj.(*corev1.Pod)
			l, ok := pod.GetLabels()["atop.io/managed-by"]
			if ok && l == "core" {
				log.Printf("Add pod %s status is %s \n", pod.Name, pod.Status.Phase)
				c.Name = pod.Name
				c.Namespace = pod.Namespace
				c.Status = pod.Status
				c.CreationTimestamp = pod.ObjectMeta.GetCreationTimestamp().Time
				podsJson, err := json.Marshal(c)
				if err != nil {
					panic(err)
				}
				m.BroadcastFilter(podsJson, func(s *melody.Session) bool {
					return s.Request.RequestURI == "/v1/ws/watch"
				})
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			// var cs types.Castchange
			oldPod := old.(*corev1.Pod)
			newPod := new.(*corev1.Pod)
			l, ok := newPod.GetLabels()["atop.io/managed-by"]
			if newPod.Status.Phase != oldPod.Status.Phase && ok && l == "core" {
				log.Printf("Pod %s status updated %s -> %s \n",
					newPod.Name, oldPod.Status.Phase, newPod.Status.Phase)
				cs := &types.Castchange{
					Name:              newPod.Name,
					Namespace:         newPod.Namespace,
					CreationTimestamp: newPod.ObjectMeta.GetCreationTimestamp().Time,
					Changeset: types.Changeset{
						Before: oldPod.Status,
						After:  newPod.Status,
					},
				}
				resp, err := json.Marshal(cs)
				if err != nil {
					panic(err)
				}
				m.BroadcastFilter(resp, func(s *melody.Session) bool {
					return s.Request.RequestURI == "/v1/ws/watch"
				})
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			if err != nil {
				panic(err)
			}
			l, ok := pod.GetLabels()["atop.io/managed-by"]
			if ok && l == "core" {
				log.Printf("Pod %s delete \n", pod.Name)
				resp := map[string]interface{}{
					"name":              pod.Name,
					"namespace":         pod.Namespace,
					"creationTimestamp": pod.GetCreationTimestamp().Time,
					"deletionTimestamp": pod.GetCreationTimestamp().Time,
					"status":            "Terminating",
				}
				respJson, err := json.Marshal(resp)
				if err != nil {
					panic(err)
				}
				m.BroadcastFilter(respJson, func(s *melody.Session) bool {
					return s.Request.RequestURI == "/v1/ws/watch"
				})
			}
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
		var req types.Msg
		json.Unmarshal(msg, &req)

		labelSelectorMap := map[string]string{
			"atop.io/managed-by": "core",
		}
		labelSelector := labels.SelectorFromSet(labels.Set(labelSelectorMap))

		pods, err := podInformer.Lister().List(labelSelector)
		if err != nil {
			panic(err.Error())
		}

		for _, v := range pods {
			if v.Name == req.Name && v.Namespace == req.Namespace {
				c := types.Cast{
					Name:              v.Name,
					Namespace:         v.Namespace,
					CreationTimestamp: v.ObjectMeta.GetCreationTimestamp().Time,
					Status:            v.Status,
				}
				resp, err := json.Marshal(c)
				if err != nil {
					panic(err)
				}
				m.BroadcastFilter(resp, func(s *melody.Session) bool {
					return s.Request.RequestURI == "/v1/ws/pull"
				})
			}
		}

	})
	r.Run(":8081")

}
