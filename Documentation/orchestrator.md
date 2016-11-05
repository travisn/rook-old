# Storage Cluster Orchestration
Rook provides automated configuration and management of a ceph cluster. The deployment of the ceph services in a cluster, and automated maintenance of those services is what we define as cluster orchestration. Orchestration provides support for:
- Adding new nodes to the cluster results in a cluster with more capacity
- Removing nodes from the cluster results in less capacity, with ceph rebalancing and repairing the data
- Failover of internal services to healthy nodes in the cluster when nodes become unhealthy

The cluster is designed to be aware of both the current state of the services, as well as changes that need to be made. The changes could be due to infrastructure changes such as adding a new node or replacing a disk, or adding or updating the service. This document describes the orchestration of the services across the cluster infrastructure.

## Cluster Foundation
Etcd is the distributed KV store that provides the foundation for deploying services in the cluster.

## Leader
One of the nodes in the cluster is elected to be the leader of the orchestration. The leader is the central brain of deploying services in the cluster. 

When a leader is elected, the leader starts a deployment of the cluster services.

When the leader goes down, a new leader is elected. The new leader starts a new deployment of the cluster services to ensure services are current. 

## Service Managers
A service manager owns the deployment of a distributed service in the cluster. The manager is aware of the configuration that needs to run the distributed service. 

A service is expected to arrive at a desired state. Each time the service configuration is run, achieving the desired state is the goal of the service. Thus, on the first configuration the service will initialize all of its services, as defined by the desired state of the service. If no desired state changes before the next configuration, the service will not make any changes on the next configuration.

A service manager is only active on the machine that is elected to be the leader.

There are only two service managers. Their purpose is only to provide the basics that are needed to run a storage cluster. 
- etcd
- ceph

Other service managers could be implemented in the future if needed. However, it would be advantageous to use a higher-level orchestrator such as kubernetes to manage the orchestration of applications. The storage cluster cannot build on top of an orchestrator such as kubernetes because it must run in minimal environments where only rook is available.

### Interface
A service manager implements:
- `HandleRefresh(e *RefreshEvent)`
  - Implements the distributed service configuration
- `RefreshKeys() []*RefreshKey`
  - Gets the list of etcd keys that when changed should trigger an orchestration

## Service Agents
On each node where rook is running, rook watches for instructions from the service managers to deploy services on the local node. The communication between the managers and the agent is through watching an etcd key.

When a node is notified that configuration is needed, it is given the name of a service agent to call. The service agents are the unit for deploying a service on a single node of the cluster. 

The service agents are currently:
- *ceph mon*: generates the ceph config and starts a ceph mon
- *ceph osd*: generate the ceph config and starts a ceph osd for each available device on the node
- *etcd*: starts or stops embedded etcd instances to ensure availability of the etcd cluster

If configuration of a service agent requires settings to be determined by the central service manager, the settings are stored in etcd by the service manager, then read from etcd by the service agent.

### Interface
Each service agent implements: 
- `Initialize(context *Context) error`
  - Initialize the agent. Allows the agent to store desired state in etcd before orchestration starts.
- `ConfigureLocalService(context *Context) error`
  - Configure and start the service on this node
- `DestroyLocalService(context *Context) error`
  - Stop a service on the node that is no longer needed
- `Name() string`
  - Gets the name of the service agent. Used for logging.
