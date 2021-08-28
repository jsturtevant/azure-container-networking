// Package npm Copyright 2018 Microsoft. All rights reserved.
// MIT License
package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Azure/azure-container-networking/aitelemetry"

	"github.com/Azure/azure-container-networking/npm/ipsm"
	"github.com/Azure/azure-container-networking/npm/metrics"
	"github.com/Azure/azure-container-networking/npm/util"
	"github.com/Azure/azure-container-networking/telemetry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	networkinginformers "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	utilexec "k8s.io/utils/exec"
)

var aiMetadata string

const (
	heartbeatIntervalInMinutes = 30
	// TODO: consider increasing thread number later when logics are correct
	// threadness = 1
)

// Cache to store namespace struct in nameSpaceController.go.
// Since this cache is shared between podController and NameSpaceController,
// it has mutex for avoiding racing condition between them.
type npmNamespaceCache struct {
	sync.Mutex
	nsMap map[string]*Namespace // Key is ns-<nsname>
}

func (n *npmNamespaceCache) MarshalJSON() ([]byte, error) {
	n.Lock()
	defer n.Unlock()
	return json.Marshal(n.nsMap)
}

// NetworkPolicyManager contains informers for pod, namespace and networkpolicy.
type NetworkPolicyManager struct {
	clientset       kubernetes.Interface
	informerFactory informers.SharedInformerFactory
	podInformer     coreinformers.PodInformer
	podController   *podController

	nsInformer          coreinformers.NamespaceInformer
	nameSpaceController *nameSpaceController

	npInformer       networkinginformers.NetworkPolicyInformer
	netPolController *networkPolicyController

	// ipsMgr are shared in all controllers. Thus, only one ipsMgr is created for simple management
	// and uses lock to avoid unintentional race condictions in IpsetManager.
	ipsMgr            *ipsm.IpsetManager
	npmNamespaceCache *npmNamespaceCache
	// Azure-specific variables
	clusterState     telemetry.ClusterState
	k8sServerVersion *version.Info
	NodeName         string
	version          string
	TelemetryEnabled bool
}

// NewNetworkPolicyManager creates a NetworkPolicyManager
func NewNetworkPolicyManager(clientset kubernetes.Interface, informerFactory informers.SharedInformerFactory,
	exec utilexec.Interface, npmVersion string, k8sServerVersion *version.Info) *NetworkPolicyManager {
	klog.Infof("API server version: %+v ai meta data %+v", k8sServerVersion, aiMetadata)

	npMgr := &NetworkPolicyManager{
		clientset:         clientset,
		informerFactory:   informerFactory,
		podInformer:       informerFactory.Core().V1().Pods(),
		nsInformer:        informerFactory.Core().V1().Namespaces(),
		npInformer:        informerFactory.Networking().V1().NetworkPolicies(),
		ipsMgr:            ipsm.NewIpsetManager(exec),
		npmNamespaceCache: &npmNamespaceCache{nsMap: make(map[string]*Namespace)},
		clusterState: telemetry.ClusterState{
			PodCount:      0,
			NsCount:       0,
			NwPolicyCount: 0,
		},
		k8sServerVersion: k8sServerVersion,
		NodeName:         os.Getenv("HOSTNAME"),
		version:          npmVersion,
		TelemetryEnabled: true,
	}

	// create pod controller
	npMgr.podController = NewPodController(npMgr.podInformer, clientset, npMgr.ipsMgr, npMgr.npmNamespaceCache)
	// create NameSpace controller
	npMgr.nameSpaceController = NewNameSpaceController(npMgr.nsInformer, clientset, npMgr.ipsMgr, npMgr.npmNamespaceCache)
	// create network policy controller
	npMgr.netPolController = NewNetworkPolicyController(npMgr.npInformer, clientset, npMgr.ipsMgr)

	return npMgr
}

func (npMgr *NetworkPolicyManager) MarshalJSON() ([]byte, error) {
	m := map[string]json.RawMessage{}
	npmNamespaceCacheRaw, err := json.Marshal(npMgr.npmNamespaceCache)
	if err != nil {
		return nil, err
	}
	m["NsMap"] = npmNamespaceCacheRaw

	podControllerRaw, err := json.Marshal(npMgr.podController)
	if err != nil {
		return nil, err
	}
	m["PodMap"] = podControllerRaw

	listMapRaw, err := npMgr.ipsMgr.MarshalListMapJSON()
	if err != nil {
		return nil, err
	}
	m["ListMap"] = listMapRaw

	setMapRaw, err := npMgr.ipsMgr.MarshalSetMapJSON()
	if err != nil {
		return nil, err
	}
	m["SetMap"] = setMapRaw

	nodeNameRaw, err := json.Marshal(npMgr.NodeName)
	if err != nil {
		return nil, err
	}
	m["Nodename"] = nodeNameRaw
	return json.Marshal(m)
}

// GetClusterState returns current cluster state.
func (npMgr *NetworkPolicyManager) GetClusterState() telemetry.ClusterState {
	pods, err := npMgr.clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Info("Error: Failed to list pods in GetClusterState")
	}

	namespaces, err := npMgr.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Info("Error: Failed to list namespaces in GetClusterState")
	}

	networkpolicies, err := npMgr.clientset.NetworkingV1().NetworkPolicies("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Info("Error: Failed to list networkpolicies in GetClusterState")
	}

	npMgr.clusterState.PodCount = len(pods.Items)
	npMgr.clusterState.NsCount = len(namespaces.Items)
	npMgr.clusterState.NwPolicyCount = len(networkpolicies.Items)

	return npMgr.clusterState
}

// GetAppVersion returns network policy manager app version
func (npMgr *NetworkPolicyManager) GetAppVersion() string {
	return npMgr.version
}

// GetAIMetadata returns ai metadata number
func GetAIMetadata() string {
	return aiMetadata
}

// SendClusterMetrics :- send NPM cluster metrics using AppInsights
// TODO(jungukcho): need to move codes into metrics packages
func (npMgr *NetworkPolicyManager) SendClusterMetrics() {
	var (
		heartbeat        = time.NewTicker(time.Minute * heartbeatIntervalInMinutes).C
		customDimensions = map[string]string{
			"ClusterID": util.GetClusterID(npMgr.NodeName),
			"APIServer": npMgr.k8sServerVersion.String(),
		}
		podCount = aitelemetry.Metric{
			Name:             "PodCount",
			CustomDimensions: customDimensions,
		}
		nsCount = aitelemetry.Metric{
			Name:             "NsCount",
			CustomDimensions: customDimensions,
		}
		nwPolicyCount = aitelemetry.Metric{
			Name:             "NwPolicyCount",
			CustomDimensions: customDimensions,
		}
	)

	for {
		<-heartbeat

		// Reducing one to remove all-namespaces ns obj
		lenOfNsMap := len(npMgr.npmNamespaceCache.nsMap)
		nsCount.Value = float64(lenOfNsMap - 1)

		lenOfRawNpMap := npMgr.netPolController.lengthOfRawNpMap()
		nwPolicyCount.Value += float64(lenOfRawNpMap)

		lenOfPodMap := npMgr.podController.lengthOfPodMap()
		podCount.Value += float64(lenOfPodMap)

		metrics.SendMetric(podCount)
		metrics.SendMetric(nsCount)
		metrics.SendMetric(nwPolicyCount)
	}
}

// Start starts shared informers and waits for the shared informer cache to sync.
func (npMgr *NetworkPolicyManager) Start(stopCh <-chan struct{}) error {
	// Do initialization of data plane before starting syncup of each controller to avoid heavy call to api-server
	if err := npMgr.netPolController.resetDataPlane(); err != nil {
		return fmt.Errorf("Failed to initialized data plane")
	}

	// Starts all informers manufactured by npMgr's informerFactory.
	npMgr.informerFactory.Start(stopCh)

	// Wait for the initial sync of local cache.
	if !cache.WaitForCacheSync(stopCh, npMgr.podInformer.Informer().HasSynced) {
		return fmt.Errorf("Pod informer failed to sync")
	}

	if !cache.WaitForCacheSync(stopCh, npMgr.nsInformer.Informer().HasSynced) {
		return fmt.Errorf("Namespace informer failed to sync")
	}

	if !cache.WaitForCacheSync(stopCh, npMgr.npInformer.Informer().HasSynced) {
		return fmt.Errorf("Network policy informer failed to sync")
	}

	// start controllers after synced
	go npMgr.podController.Run(stopCh)
	go npMgr.nameSpaceController.Run(stopCh)
	go npMgr.netPolController.Run(stopCh)
	go npMgr.netPolController.runPeriodicTasks(stopCh)
	return nil
}
