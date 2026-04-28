package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/avaneeshravat/school-management/transport-service/internal/apierrors"
	"github.com/avaneeshravat/school-management/transport-service/internal/model"
	"github.com/avaneeshravat/school-management/transport-service/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransportService struct {
	repo *repository.TransportRepository
}

func NewTransportService(repo *repository.TransportRepository) *TransportService {
	return &TransportService{repo: repo}
}

// Vehicle operations
func (s *TransportService) CreateVehicle(req model.CreateVehicleRequest, schoolID uuid.UUID) (*model.Vehicle, error) {
	if _, err := s.repo.GetVehicleByNumber(req.VehicleNumber, schoolID); err == nil {
		return nil, apierrors.Conflict("vehicle with this number already exists")
	}

	vehicleType := strings.ToLower(strings.TrimSpace(req.VehicleType))
	if !isValidVehicleType(vehicleType) {
		return nil, errors.New("invalid vehicle_type, allowed: bus, van, mini-bus")
	}

	vehicle := &model.Vehicle{
		SchoolID:       schoolID,
		VehicleNumber:  strings.ToUpper(strings.TrimSpace(req.VehicleNumber)),
		VehicleType:    vehicleType,
		Capacity:       req.Capacity,
		DriverName:     req.DriverName,
		DriverPhone:    req.DriverPhone,
		DriverLicense:  req.DriverLicense,
		ConductorName:  req.ConductorName,
		ConductorPhone: req.ConductorPhone,
		IsActive:       true,
	}

	if req.InsuranceExpiry != "" {
		t, err := time.Parse("2006-01-02", req.InsuranceExpiry)
		if err != nil {
			return nil, errors.New("invalid insurance_expiry format, use YYYY-MM-DD")
		}
		vehicle.InsuranceExpiry = &t
	}
	if req.FitnessExpiry != "" {
		t, err := time.Parse("2006-01-02", req.FitnessExpiry)
		if err != nil {
			return nil, errors.New("invalid fitness_expiry format, use YYYY-MM-DD")
		}
		vehicle.FitnessExpiry = &t
	}

	if err := s.repo.CreateVehicle(vehicle); err != nil {
		return nil, fmt.Errorf("failed to create vehicle: %w", err)
	}
	return vehicle, nil
}

func (s *TransportService) GetVehicle(id, schoolID uuid.UUID) (*model.Vehicle, error) {
	v, err := s.repo.GetVehicleByID(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierrors.NotFound("vehicle not found")
		}
		return nil, fmt.Errorf("failed to fetch vehicle: %w", err)
	}
	return v, nil
}

func (s *TransportService) UpdateVehicle(id uuid.UUID, req model.UpdateVehicleRequest, schoolID uuid.UUID) (*model.Vehicle, error) {
	v, err := s.repo.GetVehicleByID(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierrors.NotFound("vehicle not found")
		}
		return nil, fmt.Errorf("failed to fetch vehicle: %w", err)
	}

	if req.VehicleNumber != nil {
		v.VehicleNumber = strings.ToUpper(strings.TrimSpace(*req.VehicleNumber))
	}
	if req.VehicleType != nil {
		vt := strings.ToLower(strings.TrimSpace(*req.VehicleType))
		if !isValidVehicleType(vt) {
			return nil, errors.New("invalid vehicle_type, allowed: bus, van, mini-bus")
		}
		v.VehicleType = vt
	}
	if req.Capacity != nil {
		v.Capacity = *req.Capacity
	}
	if req.DriverName != nil {
		v.DriverName = *req.DriverName
	}
	if req.DriverPhone != nil {
		v.DriverPhone = *req.DriverPhone
	}
	if req.DriverLicense != nil {
		v.DriverLicense = *req.DriverLicense
	}
	if req.ConductorName != nil {
		v.ConductorName = *req.ConductorName
	}
	if req.ConductorPhone != nil {
		v.ConductorPhone = *req.ConductorPhone
	}
	if req.InsuranceExpiry != nil {
		t, err := time.Parse("2006-01-02", *req.InsuranceExpiry)
		if err != nil {
			return nil, errors.New("invalid insurance_expiry format, use YYYY-MM-DD")
		}
		v.InsuranceExpiry = &t
	}
	if req.FitnessExpiry != nil {
		t, err := time.Parse("2006-01-02", *req.FitnessExpiry)
		if err != nil {
			return nil, errors.New("invalid fitness_expiry format, use YYYY-MM-DD")
		}
		v.FitnessExpiry = &t
	}
	if req.IsActive != nil {
		v.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateVehicle(v); err != nil {
		return nil, fmt.Errorf("failed to update vehicle: %w", err)
	}
	return v, nil
}

func (s *TransportService) DeleteVehicle(id, schoolID uuid.UUID) error {
	if _, err := s.repo.GetVehicleByID(id, schoolID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apierrors.NotFound("vehicle not found")
		}
		return fmt.Errorf("failed to fetch vehicle: %w", err)
	}

	routeCount, err := s.repo.CountRoutesByVehicle(id, schoolID)
	if err != nil {
		return fmt.Errorf("failed to check vehicle dependencies: %w", err)
	}
	if routeCount > 0 {
		return apierrors.Conflict(fmt.Sprintf("cannot delete vehicle: %d route(s) are assigned to it", routeCount))
	}

	return s.repo.DeleteVehicle(id, schoolID)
}

func (s *TransportService) GetVehicles(schoolID uuid.UUID, query model.VehicleQuery) (*model.VehicleListResponse, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	vehicles, total, err := s.repo.GetVehicles(schoolID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vehicles: %w", err)
	}

	return &model.VehicleListResponse{
		Vehicles: vehicles,
		Total:    total,
		Page:     query.Page,
		Limit:    query.Limit,
	}, nil
}

// Route operations
func (s *TransportService) CreateRoute(req model.CreateRouteRequest, schoolID uuid.UUID) (*model.Route, error) {
	if _, err := s.repo.GetRouteByCode(req.RouteCode, schoolID); err == nil {
		return nil, apierrors.Conflict("route with this code already exists")
	}

	route := &model.Route{
		SchoolID:   schoolID,
		RouteName:  strings.TrimSpace(req.RouteName),
		RouteCode:  strings.ToUpper(strings.TrimSpace(req.RouteCode)),
		StartPoint: req.StartPoint,
		EndPoint:   req.EndPoint,
		Distance:   req.Distance,
		Duration:   req.Duration,
		MonthlyFee: req.MonthlyFee,
		IsActive:   true,
	}

	if req.VehicleID != "" {
		vid, err := uuid.Parse(req.VehicleID)
		if err != nil {
			return nil, errors.New("invalid vehicle_id")
		}
		if _, err := s.repo.GetVehicleByID(vid, schoolID); err != nil {
			return nil, errors.New("vehicle not found")
		}
		route.VehicleID = &vid
	}

	if err := s.repo.CreateRoute(route); err != nil {
		return nil, fmt.Errorf("failed to create route: %w", err)
	}
	return route, nil
}

func (s *TransportService) GetRoute(id, schoolID uuid.UUID) (*model.Route, error) {
	r, err := s.repo.GetRouteByID(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierrors.NotFound("route not found")
		}
		return nil, fmt.Errorf("failed to fetch route: %w", err)
	}
	return r, nil
}

func (s *TransportService) UpdateRoute(id uuid.UUID, req model.UpdateRouteRequest, schoolID uuid.UUID) (*model.Route, error) {
	r, err := s.repo.GetRouteByID(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierrors.NotFound("route not found")
		}
		return nil, fmt.Errorf("failed to fetch route: %w", err)
	}

	if req.RouteName != nil {
		r.RouteName = strings.TrimSpace(*req.RouteName)
	}
	if req.RouteCode != nil {
		r.RouteCode = strings.ToUpper(strings.TrimSpace(*req.RouteCode))
	}
	if req.VehicleID != nil {
		if *req.VehicleID == "" {
			r.VehicleID = nil
		} else {
			vid, err := uuid.Parse(*req.VehicleID)
			if err != nil {
				return nil, errors.New("invalid vehicle_id")
			}
			r.VehicleID = &vid
		}
	}
	if req.StartPoint != nil {
		r.StartPoint = *req.StartPoint
	}
	if req.EndPoint != nil {
		r.EndPoint = *req.EndPoint
	}
	if req.Distance != nil {
		r.Distance = *req.Distance
	}
	if req.Duration != nil {
		r.Duration = *req.Duration
	}
	if req.MonthlyFee != nil {
		r.MonthlyFee = *req.MonthlyFee
	}
	if req.IsActive != nil {
		r.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateRoute(r); err != nil {
		return nil, fmt.Errorf("failed to update route: %w", err)
	}
	return r, nil
}

func (s *TransportService) DeleteRoute(id, schoolID uuid.UUID) error {
	if _, err := s.repo.GetRouteByID(id, schoolID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apierrors.NotFound("route not found")
		}
		return fmt.Errorf("failed to fetch route: %w", err)
	}

	stopCount, err := s.repo.CountStopsByRoute(id, schoolID)
	if err != nil {
		return fmt.Errorf("failed to check route dependencies: %w", err)
	}
	if stopCount > 0 {
		return apierrors.Conflict(fmt.Sprintf("cannot delete route: %d stop(s) are associated with it", stopCount))
	}

	studentCount, err := s.repo.CountStudentsByRoute(id, schoolID)
	if err != nil {
		return fmt.Errorf("failed to check route dependencies: %w", err)
	}
	if studentCount > 0 {
		return apierrors.Conflict(fmt.Sprintf("cannot delete route: %d active student assignment(s) exist", studentCount))
	}

	return s.repo.DeleteRoute(id, schoolID)
}

func (s *TransportService) GetRoutes(schoolID uuid.UUID, query model.RouteQuery) (*model.RouteListResponse, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	routes, total, err := s.repo.GetRoutes(schoolID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch routes: %w", err)
	}

	return &model.RouteListResponse{
		Routes: routes,
		Total:  total,
		Page:   query.Page,
		Limit:  query.Limit,
	}, nil
}

// Stop operations
func (s *TransportService) CreateStop(req model.CreateStopRequest, schoolID uuid.UUID) (*model.Stop, error) {
	routeID, err := uuid.Parse(req.RouteID)
	if err != nil {
		return nil, errors.New("invalid route_id")
	}

	if _, err := s.repo.GetRouteByID(routeID, schoolID); err != nil {
		return nil, errors.New("route not found")
	}

	stop := &model.Stop{
		SchoolID:   schoolID,
		RouteID:    routeID,
		StopName:   strings.TrimSpace(req.StopName),
		StopOrder:  req.StopOrder,
		PickupTime: req.PickupTime,
		DropTime:   req.DropTime,
		Landmark:   req.Landmark,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
	}

	if err := s.repo.CreateStop(stop); err != nil {
		return nil, fmt.Errorf("failed to create stop: %w", err)
	}
	return stop, nil
}

func (s *TransportService) GetStop(id, schoolID uuid.UUID) (*model.Stop, error) {
	stop, err := s.repo.GetStopByID(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierrors.NotFound("stop not found")
		}
		return nil, fmt.Errorf("failed to fetch stop: %w", err)
	}
	return stop, nil
}

func (s *TransportService) UpdateStop(id uuid.UUID, req model.UpdateStopRequest, schoolID uuid.UUID) (*model.Stop, error) {
	stop, err := s.repo.GetStopByID(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierrors.NotFound("stop not found")
		}
		return nil, fmt.Errorf("failed to fetch stop: %w", err)
	}

	if req.StopName != nil {
		stop.StopName = strings.TrimSpace(*req.StopName)
	}
	if req.StopOrder != nil {
		stop.StopOrder = *req.StopOrder
	}
	if req.PickupTime != nil {
		stop.PickupTime = *req.PickupTime
	}
	if req.DropTime != nil {
		stop.DropTime = *req.DropTime
	}
	if req.Landmark != nil {
		stop.Landmark = *req.Landmark
	}
	if req.Latitude != nil {
		stop.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		stop.Longitude = req.Longitude
	}

	if err := s.repo.UpdateStop(stop); err != nil {
		return nil, fmt.Errorf("failed to update stop: %w", err)
	}
	return stop, nil
}

func (s *TransportService) DeleteStop(id, schoolID uuid.UUID) error {
	if _, err := s.repo.GetStopByID(id, schoolID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apierrors.NotFound("stop not found")
		}
		return fmt.Errorf("failed to fetch stop: %w", err)
	}

	studentCount, err := s.repo.CountStudentTransportsByStop(id, schoolID)
	if err != nil {
		return fmt.Errorf("failed to check stop dependencies: %w", err)
	}
	if studentCount > 0 {
		return apierrors.Conflict(fmt.Sprintf("cannot delete stop: %d student assignment(s) use this stop", studentCount))
	}

	return s.repo.DeleteStop(id, schoolID)
}

func (s *TransportService) GetStops(schoolID uuid.UUID, query model.StopQuery) (*model.StopListResponse, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 50
	}

	stops, total, err := s.repo.GetStops(schoolID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stops: %w", err)
	}

	return &model.StopListResponse{
		Stops: stops,
		Total: total,
		Page:  query.Page,
		Limit: query.Limit,
	}, nil
}

// StudentTransport operations
func (s *TransportService) CreateStudentTransport(req model.CreateStudentTransportRequest, schoolID uuid.UUID) (*model.StudentTransport, error) {
	studentID, err := uuid.Parse(req.StudentID)
	if err != nil {
		return nil, errors.New("invalid student_id")
	}
	routeID, err := uuid.Parse(req.RouteID)
	if err != nil {
		return nil, errors.New("invalid route_id")
	}
	stopID, err := uuid.Parse(req.StopID)
	if err != nil {
		return nil, errors.New("invalid stop_id")
	}

	if _, err := s.repo.GetRouteByID(routeID, schoolID); err != nil {
		return nil, errors.New("route not found")
	}
	if _, err := s.repo.GetStopByID(stopID, schoolID); err != nil {
		return nil, errors.New("stop not found")
	}

	transportType := strings.ToLower(strings.TrimSpace(req.TransportType))
	if !isValidTransportType(transportType) {
		return nil, errors.New("invalid transport_type, allowed: pickup, drop, both")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.New("invalid start_date format, use YYYY-MM-DD")
	}

	st := &model.StudentTransport{
		SchoolID:      schoolID,
		StudentID:     studentID,
		RouteID:       routeID,
		StopID:        stopID,
		TransportType: transportType,
		StartDate:     startDate,
		IsActive:      true,
	}

	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, errors.New("invalid end_date format, use YYYY-MM-DD")
		}
		st.EndDate = &endDate
	}

	if err := s.repo.CreateStudentTransport(st); err != nil {
		return nil, fmt.Errorf("failed to create student transport: %w", err)
	}
	return st, nil
}

func (s *TransportService) GetStudentTransport(id, schoolID uuid.UUID) (*model.StudentTransport, error) {
	st, err := s.repo.GetStudentTransportByID(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierrors.NotFound("student transport assignment not found")
		}
		return nil, fmt.Errorf("failed to fetch student transport: %w", err)
	}
	return st, nil
}

func (s *TransportService) UpdateStudentTransport(id uuid.UUID, req model.UpdateStudentTransportRequest, schoolID uuid.UUID) (*model.StudentTransport, error) {
	st, err := s.repo.GetStudentTransportByID(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierrors.NotFound("student transport assignment not found")
		}
		return nil, fmt.Errorf("failed to fetch student transport: %w", err)
	}

	if req.RouteID != nil {
		routeID, err := uuid.Parse(*req.RouteID)
		if err != nil {
			return nil, errors.New("invalid route_id")
		}
		st.RouteID = routeID
	}
	if req.StopID != nil {
		stopID, err := uuid.Parse(*req.StopID)
		if err != nil {
			return nil, errors.New("invalid stop_id")
		}
		st.StopID = stopID
	}
	if req.TransportType != nil {
		tt := strings.ToLower(strings.TrimSpace(*req.TransportType))
		if !isValidTransportType(tt) {
			return nil, errors.New("invalid transport_type, allowed: pickup, drop, both")
		}
		st.TransportType = tt
	}
	if req.StartDate != nil {
		startDate, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, errors.New("invalid start_date format, use YYYY-MM-DD")
		}
		st.StartDate = startDate
	}
	if req.EndDate != nil {
		if *req.EndDate == "" {
			st.EndDate = nil
		} else {
			endDate, err := time.Parse("2006-01-02", *req.EndDate)
			if err != nil {
				return nil, errors.New("invalid end_date format, use YYYY-MM-DD")
			}
			st.EndDate = &endDate
		}
	}
	if req.IsActive != nil {
		st.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateStudentTransport(st); err != nil {
		return nil, fmt.Errorf("failed to update student transport: %w", err)
	}
	return st, nil
}

func (s *TransportService) DeleteStudentTransport(id, schoolID uuid.UUID) error {
	if _, err := s.repo.GetStudentTransportByID(id, schoolID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apierrors.NotFound("student transport assignment not found")
		}
		return fmt.Errorf("failed to fetch student transport: %w", err)
	}
	return s.repo.DeleteStudentTransport(id, schoolID)
}

func (s *TransportService) GetStudentTransports(schoolID uuid.UUID, query model.StudentTransportQuery) (*model.StudentTransportListResponse, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	assignments, total, err := s.repo.GetStudentTransports(schoolID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch student transports: %w", err)
	}

	return &model.StudentTransportListResponse{
		Assignments: assignments,
		Total:       total,
		Page:        query.Page,
		Limit:       query.Limit,
	}, nil
}

// Helpers
func isValidVehicleType(t string) bool {
	switch t {
	case "bus", "van", "mini-bus":
		return true
	default:
		return false
	}
}

func isValidTransportType(t string) bool {
	switch t {
	case "pickup", "drop", "both":
		return true
	default:
		return false
	}
}
