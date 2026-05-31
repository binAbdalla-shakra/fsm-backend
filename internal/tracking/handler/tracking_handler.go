package handler

import (
	"context"
	"encoding/json"
	"fsm-backend/internal/tracking/dto"
	"fsm-backend/internal/tracking/service"
	"log"

	"github.com/gofiber/contrib/websocket"
)

type TrackingHandler struct {
	service service.TrackingService
}

func NewTrackingHandler(service service.TrackingService) *TrackingHandler {
	return &TrackingHandler{service: service}
}

func (h *TrackingHandler) TechTelemetryWS(c *websocket.Conn) {
	techID := c.Params("tech_id")
	ticketID := c.Query("ticket_id")

	if techID == "" {
		_ = c.WriteMessage(websocket.TextMessage, []byte("Error: Missing technician_id parameter"))
		c.Close()
		return
	}

	defer c.Close()

	for {
		messageType, messageBytes, err := c.ReadMessage()
		if err != nil {
			log.Printf("Technician %s disconnected telemetry session: %v", techID, err)
			break
		}

		if messageType == websocket.TextMessage {
			var ping dto.LocationPing
			if err := json.Unmarshal(messageBytes, &ping); err != nil {
				_ = c.WriteMessage(websocket.TextMessage, []byte("Error: Invalid ping schema format"))
				continue
			}

			ping.TechnicianID = techID

			err = h.service.ProcessTechPing(context.Background(), &ping, ticketID)
			if err != nil {
				log.Printf("Error processing coordinates ping for tech %s: %v", techID, err)
				_ = c.WriteMessage(websocket.TextMessage, []byte("Error: Telemetry processing failure"))
			}
		}
	}
}

func (h *TrackingHandler) CustomerTrackWS(c *websocket.Conn) {
	ticketID := c.Params("ticket_id")
	if ticketID == "" {
		_ = c.WriteMessage(websocket.TextMessage, []byte("Error: Missing ticket_id parameter"))
		c.Close()
		return
	}

	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pubsub := h.service.GetSubscription(ctx, ticketID)
	defer pubsub.Close()

	ch := pubsub.Channel()

	go func() {
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			err := c.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
			if err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
