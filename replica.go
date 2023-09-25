package FabricEmu

import (
	"fmt"
	"math"
	"sync"

	"github.com/tg123/phabrik/serialization"
	"github.com/tg123/phabrik/transport"
)

type Replica struct {
	// States ReplicaStates // TODO expose
	lock sync.Mutex

	parent          *ReplicaAgent
	registeringConn transport.Conn
	rapClient       *transport.Client

	runtimeId           string
	applicationTypeName string
	applicationNumber   uint32
	servicePackageName  string
	serviceTypeName     string

	serviceName          string
	incarnationId        serialization.GUID
	replicaId            int64
	partitionId          serialization.GUID
	configurationVersion int
	replicaOpen          chan *ProxyReplyMessageBody
	stateful             bool
}

func (h *ReplicaAgent) newReplica(c transport.Conn, req RegisterServiceTypeRequest) (*Replica, error) {

	failoverId, err := serialization.NewGuidV4()
	if err != nil {
		return nil, err
	}

	r := &Replica{
		parent:          h,
		registeringConn: c,

		runtimeId:           req.RuntimeContext.RuntimeId,
		applicationTypeName: req.RuntimeContext.CodeContext.CodePackageInstanceId.ServicePackageInstanceId.ServicePackageId.ApplicationIdentifier.ApplicationTypeName,
		applicationNumber:   req.RuntimeContext.CodeContext.CodePackageInstanceId.ServicePackageInstanceId.ServicePackageId.ApplicationIdentifier.ApplicationNumber,
		servicePackageName:  req.RuntimeContext.CodeContext.CodePackageInstanceId.ServicePackageInstanceId.ServicePackageId.ServicePackageName,
		serviceTypeName:     req.ServiceTypeInstanceId.ServiceTypeName,

		serviceName:          fmt.Sprintf("fabric:/app/%v", req.ServiceTypeInstanceId.ServiceTypeName),
		partitionId:          failoverId,
		replicaOpen:          make(chan *ProxyReplyMessageBody),
		configurationVersion: 1,
		stateful:             h.stateful,
	}

	h.replicas.Store(failoverId, r)

	return r, nil
}

func (r *Replica) ID() serialization.GUID {
	return r.partitionId
}

func (r *Replica) open() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	err := r.registeringConn.SendOneWay(&transport.Message{
		Headers: transport.MessageHeaders{
			Actor:  transport.MessageActorTypeRA,
			Action: "ReplicaOpen",
		},
		Body: &ProxyRequestMessageBody{
			RuntimeId: r.runtimeId,
			LocalReplica: ReplicaDescription{
				IsUp:                     true,
				CurrentConfigurationRole: ReplicaRoleIdle,
				State:                    ReplicaStatesReady,
				LastAcknowledgedLSN:      1,
			},
			FuDesc: FailoverUnitDescription{
				FailoverUnitId: FailoverUnitId{
					Guid: r.partitionId,
				},
				ConsistencyUnitDescription: ConsistencyUnitDescription{
					// PartitionKind: 1,
					// Singleton = 1,
					// Int64Range = 2,
					// Named = 3
					PartitionKind: 2, //
					// LowKeyInclusive:  -9223372036854775808, // min int64
					LowKeyInclusive: math.MinInt64,
					// HighKeyInclusive: 9223372036854775807,
					HighKeyInclusive: math.MaxInt64,
				},
				CcEpoch: Epoch{
					ConfigurationVersion: int64(r.configurationVersion),
					DataLossVersion:      0,
				},
				PcEpoch: Epoch{
					ConfigurationVersion: 0,
					DataLossVersion:      0,
				},
			},
			Service: ServiceDescription{
				Name: r.serviceName,
				IsStateful: r.stateful,
				Type: ServiceTypeIdentifier{
					PackageIdentifier: ServicePackageIdentifier{
						ApplicationIdentifier: ApplicationIdentifier{
							ApplicationTypeName: r.applicationTypeName,
							ApplicationNumber:   r.applicationNumber,
						},
						ServicePackageName: r.servicePackageName,
					},
					ServiceTypeName: r.serviceTypeName,
				},
			},
		},
	})

	if err != nil {
		return err
	}

	reply := <-r.replicaOpen

	if reply.ProxyErrorCode.ErrorCode != 0 {
		return fmt.Errorf(reply.ProxyErrorCode.Message)
	}

	// ss := strings.Split(reply.LocalReplica.ReplicationEndpoint, "/")
	// log.Printf("reply %v", reply)
	// endpoint := ss[0]
	// log.Printf("replica open endpoint %v", endpoint)

	// // this is an unsecure client
	// rapClient, err := transport.DialTCP(endpoint, transport.ClientConfig{
	// 	// MessageCallback: func(c transport.Conn, bam *transport.ByteArrayMessage) {
	// 	// },
	// })
	// if err != nil {
	// 	return err
	// }

	// r.rapClient = rapClient

	// if len(ss) < 2 {
	// 	return fmt.Errorf("bad replica endpoint string %v", reply.LocalReplica.ReplicationEndpoint)
	// }

	// ss = strings.Split(ss[1], ";")
	// if len(ss) < 2 {
	// 	return fmt.Errorf("bad replica ids in endpoint string %v", reply.LocalReplica.ReplicationEndpoint)
	// }

	// incarnationId, err := serialization.GUIDFromString(ss[1])
	// if err != nil {
	// 	return err
	// }

	// r.incarnationId = incarnationId

	// if err := r.startCopy(); err != nil {
	// 	return err
	// }

	// if err := r.copy(); err != nil {
	// 	return err
	// }

	return nil
}

// func (r *Replica) updateConfiguration() error {
func (r *Replica) ChangeRole(role ReplicaRole) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.configurationVersion++

	err := r.registeringConn.SendOneWay(&transport.Message{
		Headers: transport.MessageHeaders{
			Actor:  transport.MessageActorTypeRA,
			Action: "UpdateConfiguration",
		},
		Body: &ProxyRequestMessageBody{
			RuntimeId: r.runtimeId,
			LocalReplica: ReplicaDescription{
				IsUp:                     true,
				CurrentConfigurationRole: role,
				// CurrentConfigurationRole: ReplicaRoleIdle,
				State: ReplicaStatesReady,
			},
			FuDesc: FailoverUnitDescription{
				FailoverUnitId: FailoverUnitId{
					Guid: r.partitionId,
				},
				ConsistencyUnitDescription: ConsistencyUnitDescription{
					PartitionKind: 1,
				},
				CcEpoch: Epoch{
					ConfigurationVersion: int64(r.configurationVersion),
					DataLossVersion:      0,
				},
				PcEpoch: Epoch{
					ConfigurationVersion: 1,
					DataLossVersion:      0,
				},
			},
			Service: ServiceDescription{
				Name:       r.serviceName,
				IsStateful: true,
				Type: ServiceTypeIdentifier{
					PackageIdentifier: ServicePackageIdentifier{
						ApplicationIdentifier: ApplicationIdentifier{
							ApplicationTypeName: r.applicationTypeName,
							ApplicationNumber:   r.applicationNumber,
						},
						ServicePackageName: r.servicePackageName,
					},
					ServiceTypeName: r.serviceTypeName,
				},
			},
			Flags: 2,
		},
	})

	if err != nil {
		return err
	}

	return nil
}

// TODO very basic replica status
func (r *Replica) startCopy() error {
	msgstartcopy := &transport.Message{
		Headers: transport.MessageHeaders{
			// Actor:  transport.MessageActorTypeRA,
			Action: "StartCopy",
		},
	}

	{
		rh := &struct {
			PartitionId   serialization.GUID
			ReplicaId     int64
			IncarnationId serialization.GUID
		}{}
		rh.PartitionId = r.partitionId
		rh.ReplicaId = r.replicaId
		rh.IncarnationId = r.incarnationId

		msgstartcopy.Headers.SetCustomHeader(transport.MessageHeaderIdTypeReplicationActor, rh)
	}

	{
		b := &struct {
			Epoch                          bytearray
			ReplicaId                      int64
			ReplicationStartSequenceNumber int64
		}{
			ReplicationStartSequenceNumber: 1,
		}

		b.Epoch.data = &struct {
			DataLossNumber      int64
			ConfigurationNumber int64
		}{
			DataLossNumber:      0,
			ConfigurationNumber: 1,
		}
		msgstartcopy.Body = b
	}

	{
		rh := &struct {
			Address       string
			PartitionId   serialization.GUID
			ReplicaId     int64
			IncarnationId serialization.GUID
		}{}
		rh.PartitionId = r.partitionId
		rh.ReplicaId = r.replicaId
		rh.IncarnationId = r.incarnationId
		rh.Address = "dummyaddr" // TODO placeholder
		msgstartcopy.Headers.SetCustomHeader(transport.MessageHeaderIdTypeREFrom, rh)
	}

	return r.rapClient.SendOneWay(msgstartcopy)
}

func (r *Replica) copy() error {

	msgcopy := &transport.Message{
		Headers: transport.MessageHeaders{
			// Actor:  transport.MessageActorTypeRA,
			Action: "CopyOperation",
		},
	}

	{
		rh := &struct {
			PartitionId   serialization.GUID
			ReplicaId     int64
			IncarnationId serialization.GUID
		}{}
		rh.PartitionId = r.partitionId
		rh.ReplicaId = r.replicaId
		rh.IncarnationId = r.incarnationId
		msgcopy.Headers.SetCustomHeader(transport.MessageHeaderIdTypeReplicationActor, rh)
	}

	{
		rh := &struct {
			ReplicaId         int64
			PrimaryEpoch      bytearray
			OperationMetadata bytearray
			SegmentSizes      []uint32
			IsLast            bool
		}{}

		rh.PrimaryEpoch.data = &struct {
			DataLossNumber      int64
			ConfigurationNumber int64
		}{
			DataLossNumber:      0,
			ConfigurationNumber: 1,
		}

		rh.OperationMetadata.data = &struct {
			Type           int64
			SequenceNumber int64
			AtomicGroupId  int64
		}{
			SequenceNumber: 1,
		}
		rh.IsLast = true

		msgcopy.Headers.SetCustomHeader(transport.MessageHeaderIdTypeCopyOperation, rh)
	}

	{
		rh := &struct {
			Address       string
			PartitionId   serialization.GUID
			ReplicaId     int64
			IncarnationId serialization.GUID
		}{}
		rh.PartitionId = r.partitionId
		rh.ReplicaId = r.replicaId
		rh.IncarnationId = r.incarnationId
		rh.Address = "dummyaddr"
		msgcopy.Headers.SetCustomHeader(transport.MessageHeaderIdTypeREFrom, rh)
	}

	return r.rapClient.SendOneWay(msgcopy)
}

// TODO
// func (r *Replica) Close() error {
// 	return nil
// }
