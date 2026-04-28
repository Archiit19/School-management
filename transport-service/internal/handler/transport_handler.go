package handler

import (
	"errors"
	"net/http"

	"github.com/avaneeshravat/school-management/transport-service/internal/apierrors"
	"github.com/avaneeshravat/school-management/transport-service/internal/model"
	"github.com/avaneeshravat/school-management/transport-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func writeErr(c *gin.Context, err error) {
	var he *apierrors.HTTP
	if errors.As(err, &he) {
		c.JSON(he.Status, gin.H{"error": he.Error()})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

type TransportHandler struct {
	svc *service.TransportService
}

func NewTransportHandler(svc *service.TransportService) *TransportHandler {
	return &TransportHandler{svc: svc}
}

// Vehicle Handlers

func (h *TransportHandler) CreateVehicle(c *gin.Context) {
	var req model.CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	vehicle, err := h.svc.CreateVehicle(req, schoolID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, vehicle)
}

func (h *TransportHandler) GetVehicle(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vehicle id"})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	vehicle, err := h.svc.GetVehicle(id, schoolID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, vehicle)
}

func (h *TransportHandler) UpdateVehicle(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vehicle id"})
		return
	}

	var req model.UpdateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	vehicle, err := h.svc.UpdateVehicle(id, req, schoolID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, vehicle)
}

func (h *TransportHandler) DeleteVehicle(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vehicle id"})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	if err := h.svc.DeleteVehicle(id, schoolID); err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vehicle deleted"})
}

func (h *TransportHandler) GetVehicles(c *gin.Context) {
	var query model.VehicleQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	resp, err := h.svc.GetVehicles(schoolID, query)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Route Handlers

func (h *TransportHandler) CreateRoute(c *gin.Context) {
	var req model.CreateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	route, err := h.svc.CreateRoute(req, schoolID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, route)
}

func (h *TransportHandler) GetRoute(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid route id"})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	route, err := h.svc.GetRoute(id, schoolID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, route)
}

func (h *TransportHandler) UpdateRoute(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid route id"})
		return
	}

	var req model.UpdateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	route, err := h.svc.UpdateRoute(id, req, schoolID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, route)
}

func (h *TransportHandler) DeleteRoute(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid route id"})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	if err := h.svc.DeleteRoute(id, schoolID); err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "route deleted"})
}

func (h *TransportHandler) GetRoutes(c *gin.Context) {
	var query model.RouteQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	resp, err := h.svc.GetRoutes(schoolID, query)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Stop Handlers

func (h *TransportHandler) CreateStop(c *gin.Context) {
	var req model.CreateStopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	stop, err := h.svc.CreateStop(req, schoolID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, stop)
}

func (h *TransportHandler) GetStop(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stop id"})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	stop, err := h.svc.GetStop(id, schoolID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, stop)
}

func (h *TransportHandler) UpdateStop(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stop id"})
		return
	}

	var req model.UpdateStopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	stop, err := h.svc.UpdateStop(id, req, schoolID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, stop)
}

func (h *TransportHandler) DeleteStop(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stop id"})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	if err := h.svc.DeleteStop(id, schoolID); err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stop deleted"})
}

func (h *TransportHandler) GetStops(c *gin.Context) {
	var query model.StopQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	resp, err := h.svc.GetStops(schoolID, query)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// StudentTransport Handlers

func (h *TransportHandler) CreateStudentTransport(c *gin.Context) {
	var req model.CreateStudentTransportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	st, err := h.svc.CreateStudentTransport(req, schoolID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, st)
}

func (h *TransportHandler) GetStudentTransport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student transport id"})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	st, err := h.svc.GetStudentTransport(id, schoolID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, st)
}

func (h *TransportHandler) UpdateStudentTransport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student transport id"})
		return
	}

	var req model.UpdateStudentTransportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	st, err := h.svc.UpdateStudentTransport(id, req, schoolID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, st)
}

func (h *TransportHandler) DeleteStudentTransport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student transport id"})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	if err := h.svc.DeleteStudentTransport(id, schoolID); err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "student transport assignment deleted"})
}

func (h *TransportHandler) GetStudentTransports(c *gin.Context) {
	var query model.StudentTransportQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	resp, err := h.svc.GetStudentTransports(schoolID, query)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
