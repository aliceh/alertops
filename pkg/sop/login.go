package sop

import (
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RosaCluster struct {
	client    client.Client
	hive      client.Client
	hiveAdmin client.Client

	cluster   *cmv1.Cluster
	clusterId string
}
