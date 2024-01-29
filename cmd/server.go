package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	types "github.com/no8geo/notify/cmd/types"

	router "github.com/no8geo/notify/internal/router"
	"github.com/no8geo/notify/pkg/k8s"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"github.com/olahol/melody"
)

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.PersistentFlags().StringP("port", "p", "8081", "port for notify service")
}

func server(port string) {
	r := gin.Default()
	m := melody.New()

	clientset, err := k8s.Client()
	if err != nil {
		log.Printf("Error creating Kubernetes client: %v", err)
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
			l, ok := pod.GetLabels()["atop.io/managed-by"]
			if ok && l == "core" {
				log.Printf("Add pod %s status is %s \n", pod.Name, pod.Status.Phase)
				c := types.Cast{
					Name:              pod.Name,
					Namespace:         pod.Namespace,
					CreationTimestamp: pod.ObjectMeta.GetCreationTimestamp().Time,
					Status:            pod.Status,
				}
				podsJson, err := json.Marshal(c)
				if err != nil {
					log.Printf("Error json marshal: %v", err)
				}
				m.BroadcastFilter(podsJson, func(s *melody.Session) bool {
					return s.Request.RequestURI == "/v1/ws/watch"
				})
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
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
					log.Printf("Error json marshal: %v", err)
				}
				m.BroadcastFilter(resp, func(s *melody.Session) bool {
					return s.Request.RequestURI == "/v1/ws/watch"
				})
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
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
					log.Printf("Error json marshal: %v", err)
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
			labelSelectorMap := map[string]string{
				"atop.io/managed-by": "core",
			}
			labelSelector := labels.SelectorFromSet(labels.Set(labelSelectorMap))
			pods, err := podInformer.Lister().List(labelSelector)
			if err != nil {
				log.Printf("Error get pods list: %v", err)
			}

			var detail []*types.Cast
			for _, v := range pods {
				c := types.Cast{
					Name:              v.Name,
					Namespace:         v.Namespace,
					CreationTimestamp: v.ObjectMeta.GetCreationTimestamp().Time,
					Status:            v.Status,
				}
				detail = append(detail, &c)
			}
			metrics := types.Metrics{
				Total:  len(pods),
				Detail: detail,
			}
			metricsJson, err := json.Marshal(metrics)
			if err != nil {
				log.Printf("Error json marshal: %v", err)
			}
			m.BroadcastFilter(metricsJson, func(s *melody.Session) bool {
				return s.Request.RequestURI == "/v1/ws/monitor"
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
			log.Printf("Error get pods list: %v", err)
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
					log.Printf("Error json marshal: %v", err)
				}
				m.BroadcastFilter(resp, func(s *melody.Session) bool {
					return s.Request.RequestURI == "/v1/ws/pull"
				})
			}
		}

	})

	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("execute notify service failed, %s", err.Error())
		os.Exit(1)
	}
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run notify as a service",
	Long:  `Run notify as a service`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		server(port)
	},
}
