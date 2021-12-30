package FabricEmu

import (
	"log"
	"time"

	"github.com/tg123/phabrik/common"
	"github.com/tg123/phabrik/serialization"
	"github.com/tg123/phabrik/transport"
)

func (h *ReplicaAgent) handler(c transport.Conn, bam *transport.ByteArrayMessage) {
	switch bam.Headers.Actor {
	case transport.MessageActorTypeHosting:
		h.replyLeaseDuration(c, bam.Headers)
		return

	case transport.MessageActorTypeApplicationHostManager:
		switch bam.Headers.Action {
		case "StartRegisterApplicationHostRequest":
			var b StartRegisterApplicationHostRequest
			if err := serialization.Unmarshal(bam.Body, &b); err != nil {
				log.Printf("unmarshal StartRegisterApplicationHostRequest err %v", err)
				return
			}

			h.registerApplicationHost(c, bam.Headers, b)
			return

		case "FinishRegisterApplicationHostRequest":
			var b FinishRegisterApplicationHostRequest
			if err := serialization.Unmarshal(bam.Body, &b); err != nil {
				log.Printf("unmarshal FinishRegisterApplicationHostRequest err %v", err)
				return
			}

			h.finishRegisterApplicationHost(c, bam.Headers, b)
			return
		}

	case transport.MessageActorTypeFabricRuntimeManager:
		switch bam.Headers.Action {
		case "RegisterFabricRuntimeRequest":
			var b RegisterFabricRuntimeRequest
			if err := serialization.Unmarshal(bam.Body, &b); err != nil {
				log.Printf("unmarshal RegisterFabricRuntimeRequest err %v", err)
				return
			}

			h.registerFabricRuntimeRequest(c, bam.Headers, b)
			return

		case "RegisterServiceTypeRequest":
			var b RegisterServiceTypeRequest
			if err := serialization.Unmarshal(bam.Body, &b); err != nil {
				log.Printf("unmarshal RegisterServiceTypeRequest err %v", err)
				return
			}

			h.registerServiceType(c, bam.Headers, b)
			return
		}

	case transport.MessageActorTypeRA:
		switch bam.Headers.Action {
		case "UpdateConfigurationReply":
			var b ProxyReplyMessageBody
			if err := serialization.Unmarshal(bam.Body, &b); err != nil {
				log.Printf("unmarshal ProxyReplyMessageBody err %v %v", err, bam.Body)
				return
			}

			return
		case "ReplicaOpenReply":
			var b ProxyReplyMessageBody // the body name is really confusing
			if err := serialization.Unmarshal(bam.Body, &b); err != nil {
				log.Printf("unmarshal ProxyReplyMessageBody err %v %v", err, bam.Body)
				return
			}

			replica, ok := h.replicas.Load(b.FuDesc.FailoverUnitId.Guid)
			if ok {
				replica.(*Replica).replicaOpen <- &b
			}

			return
		case "ProxyReplicaEndpointUpdated":
			// TODO support this msg
			return
		case "RAReportFault":
			var b ReportFaultMessageBody
			if err := serialization.Unmarshal(bam.Body, &b); err != nil {
				log.Printf("unmarshal ReportFaultMessageBody err %v %v", err, bam.Body)
				return
			}

			log.Println("ra report fault", b)
			return
		}
	}

	log.Printf("unknown msg, %v", bam)
}

func (h *ReplicaAgent) replyLeaseDuration(c transport.Conn, headers transport.MessageHeaders) {
	c.SendOneWay(&transport.Message{
		Headers: transport.MessageHeaders{
			RelatesTo: headers.Id,
		},
		Body: &LeaseReply{
			TTL: common.TimeSpanFromDuration(1 * time.Hour),
		},
	})
}
