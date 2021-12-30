package FabricEmu

import (
	"github.com/tg123/phabrik/common"
	"github.com/tg123/phabrik/federation"
	"github.com/tg123/phabrik/serialization"
	"github.com/tg123/phabrik/transport"
)

type ApplicationHostType int64

const (
	ApplicationHostTypeNonActivated ApplicationHostType = iota
	ApplicationHostTypeActivatedSingleCodePackage
	ApplicationHostTypeActivatedMultiCodePackage
	ApplicationHostTypeActivatedInProcess
)

type StartRegisterApplicationHostRequest struct {
	Id                         string
	Type                       ApplicationHostType
	ProcessId                  uint32
	Timeout                    common.TimeSpan
	IsContainerHost            bool
	IsCodePackageActivatorHost bool
}

type SharedLogSettings struct {
	PathWrapper          []uint16
	DiskId               serialization.GUID
	LogContainerId       serialization.GUID
	LogSize              int64
	MaximumNumberStreams uint32
	MaximumRecordSize    uint32
	Flags                uint32
}

type StartRegisterApplicationHostReply struct {
	NodeId                          string
	NodeHostProcessId               uint32
	NodeLeaseHandle                 uint64
	NodeLeaseDuration               common.TimeSpan
	ErrorCode                       transport.FabricErrorCode
	NodeInstanceId                  uint64
	NodeName                        string
	NodeType                        string
	ClientConnectionAddress         string
	DeploymentDirectory             string
	IpAddressOrFQDN                 string
	NodeWorkFolder                  string
	InitialLeaseTTL                 common.TimeSpan
	LogicalApplicationDirectories   map[string]string
	LogicalNodeDirectories          map[string]string
	ApplicationSharedLogSettings    *SharedLogSettings
	SystemServicesSharedLogSettings *SharedLogSettings
}

type FinishRegisterApplicationHostRequest struct {
	Id string
}

type FinishRegisterApplicationHostReply struct {
	ErrorCode transport.FabricErrorCode
}

type LeaseReply struct {
	TTL common.TimeSpan
}

type RegisterFabricRuntimeRequest struct {
	RuntimeContext FabricRuntimeContext
}

type FabricRuntimeContext struct {
	RuntimeId   string
	HostContext ApplicationHostContext
	CodeContext CodePackageContext
}

type ApplicationHostContext struct {
	HostId                     string
	HostType                   ApplicationHostType
	ProcessId                  uint32
	IsContainerHost            bool
	IsCodePackageActivatorHost bool
}

type CodePackageContext struct {
	CodePackageInstanceId         CodePackageInstanceIdentifier
	CodePackageInstanceSeqNum     int64
	ServicePackageInstanceSeqNum  int64
	ServicePackageVersionInstance ServicePackageVersionInstance
	ApplicationName               string
}

type CodePackageInstanceIdentifier struct {
	ServicePackageInstanceId ServicePackageInstanceIdentifier
	CodePackageName          string
}

type ServicePackageInstanceIdentifier struct {
	ServicePackageId           ServicePackageIdentifier
	ActivationContext          ServicePackageActivationContext
	ServicePackageActivationId string
}

type ServicePackageIdentifier struct {
	ApplicationIdentifier ApplicationIdentifier
	ServicePackageName    string
}

type ApplicationIdentifier struct {
	ApplicationTypeName string
	ApplicationNumber   uint32
}

type ServicePackageActivationContext struct {
	ActivationGuid   serialization.GUID
	FailoverUnitGuid serialization.GUID
	ReplicaId        int64
	ActivationMode   int64 // TODO ServicePackageActivationMode::Enum
}

type ServicePackageVersionInstance struct {
	Version    ServicePackageVersion
	InstanceId uint64
}

type ServicePackageVersion struct {
	AppVersion ApplicationVersion
	Value      RolloutVersion
}

type ApplicationVersion struct {
	Value RolloutVersion
}

type RolloutVersion struct {
	Major uint32
	Minor uint32
}

type RegisterFabricRuntimeReply struct {
	ErrorCode transport.FabricErrorCode
}

type RegisterServiceTypeRequest struct {
	RuntimeContext        FabricRuntimeContext
	ServiceTypeInstanceId ServiceTypeInstanceIdentifier
}

type ServiceTypeInstanceIdentifier struct {
	PackageInstanceIdentifier ServicePackageInstanceIdentifier
	ServiceTypeName           string
}

// ra

type ProxyRequestMessageBody struct {
	RuntimeId      string
	FuDesc         FailoverUnitDescription
	LocalReplica   ReplicaDescription
	RemoteReplicas []ReplicaDescription
	Service        ServiceDescription
	Flags          int64 // ProxyMessageFlags::Enum
	ActivationId   string
}

type FailoverUnitDescription struct {
	FailoverUnitId             FailoverUnitId
	ConsistencyUnitDescription ConsistencyUnitDescription
	CcEpoch                    Epoch
	PcEpoch                    Epoch
	TargetReplicaSetSize       int32
	MinReplicaSetSize          int32
	AuxiliaryReplicaSetSize    int32
}

type FailoverUnitId struct {
	Guid serialization.GUID
}

type ConsistencyUnitDescription struct {
	ConsistencyUnitId ConsistencyUnitId
	LowKeyInclusive   int64
	HighKeyInclusive  int64
	PartitionName     string
	PartitionKind     int64 // FABRICSERVICEPARTITIONKIND
}

type ConsistencyUnitId struct {
	Guid serialization.GUID
}

type Epoch struct {
	ConfigurationVersion int64
	DataLossVersion      int64
}

type ReplicaRole int64

const (
	ReplicaRoleUnknown ReplicaRole = iota
	ReplicaRoleNone
	ReplicaRoleIdle
	ReplicaRoleSecondary
	ReplicaRolePrimary
	ReplicaRoleIdleAuxiliary
	ReplicaRoleAuxiliary
	ReplicaRolePrimaryAuxiliary
)

type ReplicaStates int64

const (
	ReplicaStatesStandBy ReplicaStates = iota
	ReplicaStatesInBuild
	ReplicaStatesReady
	ReplicaStatesDropped
)

type ReplicaDescription struct {
	FederationNodeInstance        federation.NodeInstance
	ReplicaId                     int64
	InstanceId                    int64
	CurrentConfigurationRole      ReplicaRole
	PreviousConfigurationRole     ReplicaRole
	IsUp                          bool
	LastAcknowledgedLSN           int64
	FirstAcknowledgedLSN          int64
	State                         ReplicaStates
	PackageVersionInstance        ServicePackageVersionInstance
	ServiceLocation               string
	ReplicationEndpoint           string
	HasRanToCompletion            bool
	ReadinessProbeServiceLocation string
}

type ServiceDescription struct {
	Name                         string
	Instance                     uint64
	Type                         ServiceTypeIdentifier
	ApplicationName              string
	PackageVersionInstance       ServicePackageVersionInstance
	PartitionCount               int32
	TargetReplicaSetSize         int32
	MinReplicaSetSize            int32
	IsStateful                   bool
	HasPersistedState            bool
	InitializationData           []byte
	IsServiceGroup               bool
	ServiceCorrelations          []ServiceCorrelationDescription
	PlacementConstraints         string
	ScaleoutCount                int32
	Metrics                      []ServiceLoadMetricDescription
	DefaultMoveCost              uint32
	ReplicaRestartWaitDuration   common.TimeSpan
	QuorumLossWaitDuration       common.TimeSpan
	UpdateVersion                uint64
	PlacementPolicies            []ServicePlacementPolicyDescription
	StandByReplicaKeepDuration   common.TimeSpan
	ServicePackageActivationMode int64 // ServiceModel::ServicePackageActivationMode::Enum
	ServiceDnsName               string
	ScalingPolicies              []ServiceScalingPolicyDescription
	ServicePlacementTimeLimit    common.TimeSpan
	MinInstanceCount             int32
	MinInstancePercentage        int32
	InstanceCloseDelayDuration   common.TimeSpan
	InstanceRestartWaitDuration  common.TimeSpan
	DropSourceReplicaOnMove      bool
	InstanceLifecycleDescription InstanceLifecycleDescription
	ReplicaLifecycleDescription  ReplicaLifecycleDescription
	ServiceTags                  ServiceTagsCollection
	AuxiliaryReplicaCount        int32
}

type ServiceTagsCollection struct {
	TagsRequiredToPlaceCollection []string
	TagsRequiredToRunCollection   []string
}

type ServiceTypeIdentifier struct {
	PackageIdentifier ServicePackageIdentifier
	ServiceTypeName   string
}

type ServiceCorrelationDescription struct {
	ServiceName string
	Scheme      int64 // FABRICSERVICECORRELATIONSCHEME
}

type ServiceLoadMetricDescription struct {
	Name                 string
	Weight               int64 // FABRICSERVICELOADMETRICWEIGHT
	PrimaryDefaultLoad   uint32
	SecondaryDefaultLoad uint32
	AuxiliaryDefaultLoad uint32
}

type ServicePlacementPolicyDescription struct {
	DomainName string
	Type       int64 // FABRICPLACEMENTPOLICYTYPE
}

type ServiceScalingPolicyDescription struct {
	Mechanism *ScalingMechanism
	Trigger   *ScalingTrigger
}

type ScalingMechanism struct {
	Kind int64 // ScalingMechanismKind::Enum
}

type ScalingTrigger struct {
	Kind int64 // ScalingTriggerKind::Enum
}

type InstanceLifecycleDescription struct {
	RestoreReplicaLocationAfterUpgrade *bool
}

type ReplicaLifecycleDescription struct {
	IsSingletonReplicaMoveAllowedDuringUpgrade *bool
	RestoreReplicaLocationAfterUpgrade         *bool
}

type ProxyReplyMessageBody struct {
	FuDesc         FailoverUnitDescription
	LocalReplica   ReplicaDescription
	RemoteReplicas []ReplicaDescription
	ProxyErrorCode ProxyErrorCode
	Flags          int64 // ProxyMessageFlags::Enum
}

type ProxyErrorCode struct {
	ErrorCode transport.FabricErrorCode
	Message   string
	Api       ApiNameDescription
}

type ApiNameDescription struct {
	ApiName       int64
	InterfaceName int64
	Metadata      string
}

type FaultType int64

const (
	FaultTypeInvalid FaultType = iota
	FaultTypeTransient
	FaultTypePermanent
)

type ReportFaultMessageBody struct {
	FuDesc              FailoverUnitDescription
	ReplicaDesc         ReplicaDescription
	FaultType           FaultType
	ActivityDescription ActivityDescription
}

type ActivityDescription struct {
	ActivityId   transport.ActivityId
	ActivityType int64 // TODO enum
}
