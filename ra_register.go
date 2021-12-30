package FabricEmu

import (
	"log"

	"github.com/tg123/phabrik/transport"
)

func (h *ReplicaAgent) registerApplicationHost(c transport.Conn, headers transport.MessageHeaders, req StartRegisterApplicationHostRequest) {
	c.SendOneWay(&transport.Message{
		Headers: transport.MessageHeaders{
			RelatesTo: headers.Id,
		},
		Body: &StartRegisterApplicationHostReply{
			IpAddressOrFQDN:     "127.0.0.1",
			NodeId:              h.nodeid,
			NodeLeaseHandle:     1,
			DeploymentDirectory: h.deploymentDirectory,
		},
	})
}

func (h *ReplicaAgent) finishRegisterApplicationHost(c transport.Conn, headers transport.MessageHeaders, req FinishRegisterApplicationHostRequest) {
	c.SendOneWay(&transport.Message{
		Headers: transport.MessageHeaders{
			RelatesTo: headers.Id,
		},
		Body: &FinishRegisterApplicationHostReply{
			ErrorCode: transport.FabricErrorCodeSuccess,
		},
	})
}

func (h *ReplicaAgent) registerFabricRuntimeRequest(c transport.Conn, headers transport.MessageHeaders, req RegisterFabricRuntimeRequest) {
	c.SendOneWay(&transport.Message{
		Headers: transport.MessageHeaders{
			RelatesTo: headers.Id,
		},
		Body: &RegisterFabricRuntimeReply{
			ErrorCode: transport.FabricErrorCodeSuccess,
		},
	})
}

func (h *ReplicaAgent) registerServiceType(c transport.Conn, headers transport.MessageHeaders, req RegisterServiceTypeRequest) {
	c.SendOneWay(&transport.Message{
		Headers: transport.MessageHeaders{
			RelatesTo: headers.Id,
			Action:    "RegisterServiceTypeReply", // why only this msg has action
		},
		Body: &struct {
			ErrorCode transport.FabricErrorCode
		}{
			ErrorCode: transport.FabricErrorCodeSuccess,
		},
	})

	replica, err := h.newReplica(c, req)
	if err != nil {
		log.Printf("create new replica failed %v", err)
		return
	}

	if err := replica.open(); err != nil {
		log.Printf("replica open failed %v", err)
		return
	}

	h.onNewReplicaOpened(replica)
}
