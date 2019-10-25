package costmodel

import (
	costAnalyzerCloud "github.com/kubecost/cost-model/cloud"
	"k8s.io/klog"
)

// NetworkUsageVNetworkUsageDataector contains the network usage values for egress network traffic
type NetworkUsageData struct {
	ClusterID             string
	PodName               string
	Namespace             string
	NetworkZoneEgress     []*Vector
	NetworkRegionEgress   []*Vector
	NetworkInternetEgress []*Vector
}

// NetworkUsageVector contains a network usage vector for egress network traffic
type NetworkUsageVector struct {
	ClusterID string
	PodName   string
	Namespace string
	Values    []*Vector
}

// GetNetworkUsageData performs a join of the the results of zone, region, and internet usage queries to return a single
// map containing network costs for each namespace+pod
func GetNetworkUsageData(zr interface{}, rr interface{}, ir interface{}, defaultClusterID string) (map[string]*NetworkUsageData, error) {
	zoneNetworkMap, err := getNetworkUsage(zr, defaultClusterID)
	if err != nil {
		return nil, err
	}

	regionNetworkMap, err := getNetworkUsage(rr, defaultClusterID)
	if err != nil {
		return nil, err
	}

	internetNetworkMap, err := getNetworkUsage(ir, defaultClusterID)
	if err != nil {
		return nil, err
	}

	usageData := make(map[string]*NetworkUsageData)
	for k, v := range zoneNetworkMap {
		existing, ok := usageData[k]
		if !ok {
			usageData[k] = &NetworkUsageData{
				ClusterID:         v.ClusterID,
				PodName:           v.PodName,
				Namespace:         v.Namespace,
				NetworkZoneEgress: v.Values,
			}
			continue
		}

		existing.NetworkZoneEgress = v.Values
	}

	for k, v := range regionNetworkMap {
		existing, ok := usageData[k]
		if !ok {
			usageData[k] = &NetworkUsageData{
				ClusterID:           v.ClusterID,
				PodName:             v.PodName,
				Namespace:           v.Namespace,
				NetworkRegionEgress: v.Values,
			}
			continue
		}

		existing.NetworkRegionEgress = v.Values
	}

	for k, v := range internetNetworkMap {
		existing, ok := usageData[k]
		if !ok {
			usageData[k] = &NetworkUsageData{
				ClusterID:             v.ClusterID,
				PodName:               v.PodName,
				Namespace:             v.Namespace,
				NetworkInternetEgress: v.Values,
			}
			continue
		}

		existing.NetworkInternetEgress = v.Values
	}

	return usageData, nil
}

// GetNetworkCost computes the actual cost for NetworkUsageData based on data provided by the Provider.
func GetNetworkCost(usage *NetworkUsageData, cloud costAnalyzerCloud.Provider) ([]*Vector, error) {
	var results []*Vector

	pricing, err := cloud.NetworkPricing()
	if err != nil {
		return nil, err
	}
	zoneCost := pricing.ZoneNetworkEgressCost
	regionCost := pricing.RegionNetworkEgressCost
	internetCost := pricing.InternetNetworkEgressCost

	zlen := len(usage.NetworkZoneEgress)
	rlen := len(usage.NetworkRegionEgress)
	ilen := len(usage.NetworkInternetEgress)

	l := max(zlen, rlen, ilen)
	for i := 0; i < l; i++ {
		var cost float64 = 0
		var timestamp float64

		if i < zlen {
			cost += usage.NetworkZoneEgress[i].Value * zoneCost
			timestamp = usage.NetworkZoneEgress[i].Timestamp
		}

		if i < rlen {
			cost += usage.NetworkRegionEgress[i].Value * regionCost
			timestamp = usage.NetworkRegionEgress[i].Timestamp
		}

		if i < ilen {
			cost += usage.NetworkInternetEgress[i].Value * internetCost
			timestamp = usage.NetworkInternetEgress[i].Timestamp
		}

		results = append(results, &Vector{
			Value:     cost,
			Timestamp: timestamp,
		})
	}

	return results, nil
}

func getNetworkUsage(qr interface{}, defaultClusterID string) (map[string]*NetworkUsageVector, error) {
	ncdmap := make(map[string]*NetworkUsageVector)
	result, err := NewQueryResults(qr)
	if err != nil {
		return nil, err
	}

	for _, val := range result {
		podName, err := val.GetString("pod_name")
		if err != nil {
			return nil, err
		}

		namespace, err := val.GetString("namespace")
		if err != nil {
			return nil, err
		}

		clusterID, err := val.GetString("cluster_id")
		if clusterID == "" {
			klog.V(4).Info("Prometheus vector does not have cluster id")
			clusterID = defaultClusterID
		}

		key := namespace + "," + podName + "," + clusterID
		ncdmap[key] = &NetworkUsageVector{
			ClusterID: clusterID,
			Namespace: namespace,
			PodName:   podName,
			Values:    val.Values,
		}
	}
	return ncdmap, nil
}

func max(x int, rest ...int) int {
	curr := x
	for _, v := range rest {
		if v > curr {
			curr = v
		}
	}
	return curr
}
