package repository

import (
	"time"

	"github.com/avaneeshravat/school-management/transport-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransportRepository struct {
	db *gorm.DB
}

func NewTransportRepository(db *gorm.DB) *TransportRepository {
	return &TransportRepository{db: db}
}

// Vehicle CRUD
func (r *TransportRepository) CreateVehicle(v *model.Vehicle) error {
	return r.db.Create(v).Error
}

func (r *TransportRepository) GetVehicleByID(id, schoolID uuid.UUID) (*model.Vehicle, error) {
	var v model.Vehicle
	err := r.db.Where("id = ? AND school_id = ?", id, schoolID).First(&v).Error
	return &v, err
}

func (r *TransportRepository) GetVehicleByNumber(number string, schoolID uuid.UUID) (*model.Vehicle, error) {
	var v model.Vehicle
	err := r.db.Where("vehicle_number = ? AND school_id = ?", number, schoolID).First(&v).Error
	return &v, err
}

func (r *TransportRepository) UpdateVehicle(v *model.Vehicle) error {
	return r.db.Save(v).Error
}

func (r *TransportRepository) DeleteVehicle(id, schoolID uuid.UUID) error {
	return r.db.Where("id = ? AND school_id = ?", id, schoolID).Delete(&model.Vehicle{}).Error
}

func (r *TransportRepository) GetVehicles(schoolID uuid.UUID, query model.VehicleQuery) ([]model.Vehicle, int64, error) {
	var vehicles []model.Vehicle
	var total int64

	q := r.db.Model(&model.Vehicle{}).Where("school_id = ?", schoolID)

	if query.VehicleType != "" {
		q = q.Where("vehicle_type = ?", query.VehicleType)
	}
	if query.IsActive == "true" {
		q = q.Where("is_active = ?", true)
	} else if query.IsActive == "false" {
		q = q.Where("is_active = ?", false)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.Limit
	err := q.Order("vehicle_number asc").Offset(offset).Limit(query.Limit).Find(&vehicles).Error
	return vehicles, total, err
}

// Route CRUD
func (r *TransportRepository) CreateRoute(route *model.Route) error {
	return r.db.Create(route).Error
}

func (r *TransportRepository) GetRouteByID(id, schoolID uuid.UUID) (*model.Route, error) {
	var route model.Route
	err := r.db.Where("id = ? AND school_id = ?", id, schoolID).First(&route).Error
	return &route, err
}

func (r *TransportRepository) GetRouteByCode(code string, schoolID uuid.UUID) (*model.Route, error) {
	var route model.Route
	err := r.db.Where("route_code = ? AND school_id = ?", code, schoolID).First(&route).Error
	return &route, err
}

func (r *TransportRepository) UpdateRoute(route *model.Route) error {
	return r.db.Save(route).Error
}

func (r *TransportRepository) DeleteRoute(id, schoolID uuid.UUID) error {
	return r.db.Where("id = ? AND school_id = ?", id, schoolID).Delete(&model.Route{}).Error
}

func (r *TransportRepository) GetRoutes(schoolID uuid.UUID, query model.RouteQuery) ([]model.Route, int64, error) {
	var routes []model.Route
	var total int64

	q := r.db.Model(&model.Route{}).Where("school_id = ?", schoolID)

	if query.IsActive == "true" {
		q = q.Where("is_active = ?", true)
	} else if query.IsActive == "false" {
		q = q.Where("is_active = ?", false)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.Limit
	err := q.Order("route_name asc").Offset(offset).Limit(query.Limit).Find(&routes).Error
	return routes, total, err
}

// Stop CRUD
func (r *TransportRepository) CreateStop(stop *model.Stop) error {
	return r.db.Create(stop).Error
}

func (r *TransportRepository) GetStopByID(id, schoolID uuid.UUID) (*model.Stop, error) {
	var stop model.Stop
	err := r.db.Where("id = ? AND school_id = ?", id, schoolID).First(&stop).Error
	return &stop, err
}

func (r *TransportRepository) UpdateStop(stop *model.Stop) error {
	return r.db.Save(stop).Error
}

func (r *TransportRepository) DeleteStop(id, schoolID uuid.UUID) error {
	return r.db.Where("id = ? AND school_id = ?", id, schoolID).Delete(&model.Stop{}).Error
}

func (r *TransportRepository) GetStops(schoolID uuid.UUID, query model.StopQuery) ([]model.Stop, int64, error) {
	var stops []model.Stop
	var total int64

	q := r.db.Model(&model.Stop{}).Where("school_id = ?", schoolID)

	if query.RouteID != "" {
		q = q.Where("route_id = ?", query.RouteID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.Limit
	err := q.Order("route_id asc, stop_order asc").Offset(offset).Limit(query.Limit).Find(&stops).Error
	return stops, total, err
}

func (r *TransportRepository) GetStopsByRoute(routeID, schoolID uuid.UUID) ([]model.Stop, error) {
	var stops []model.Stop
	err := r.db.Where("route_id = ? AND school_id = ?", routeID, schoolID).
		Order("stop_order asc").Find(&stops).Error
	return stops, err
}

// StudentTransport CRUD
func (r *TransportRepository) CreateStudentTransport(st *model.StudentTransport) error {
	return r.db.Create(st).Error
}

func (r *TransportRepository) GetStudentTransportByID(id, schoolID uuid.UUID) (*model.StudentTransport, error) {
	var st model.StudentTransport
	err := r.db.Where("id = ? AND school_id = ?", id, schoolID).First(&st).Error
	return &st, err
}

func (r *TransportRepository) GetActiveStudentTransport(studentID, schoolID uuid.UUID) (*model.StudentTransport, error) {
	var st model.StudentTransport
	today := time.Now().Format("2006-01-02")
	err := r.db.Where(
		"student_id = ? AND school_id = ? AND is_active = ? AND start_date <= ? AND (end_date IS NULL OR end_date >= ?)",
		studentID, schoolID, true, today, today,
	).First(&st).Error
	return &st, err
}

func (r *TransportRepository) UpdateStudentTransport(st *model.StudentTransport) error {
	return r.db.Save(st).Error
}

func (r *TransportRepository) DeleteStudentTransport(id, schoolID uuid.UUID) error {
	return r.db.Where("id = ? AND school_id = ?", id, schoolID).Delete(&model.StudentTransport{}).Error
}

func (r *TransportRepository) GetStudentTransports(schoolID uuid.UUID, query model.StudentTransportQuery) ([]model.StudentTransport, int64, error) {
	var assignments []model.StudentTransport
	var total int64

	q := r.db.Model(&model.StudentTransport{}).Where("school_id = ?", schoolID)

	if query.StudentID != "" {
		q = q.Where("student_id = ?", query.StudentID)
	}
	if query.RouteID != "" {
		q = q.Where("route_id = ?", query.RouteID)
	}
	if query.StopID != "" {
		q = q.Where("stop_id = ?", query.StopID)
	}
	if query.IsActive == "true" {
		q = q.Where("is_active = ?", true)
	} else if query.IsActive == "false" {
		q = q.Where("is_active = ?", false)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.Limit
	err := q.Order("created_at desc").Offset(offset).Limit(query.Limit).Find(&assignments).Error
	return assignments, total, err
}

// Stats
func (r *TransportRepository) CountStudentsByRoute(routeID, schoolID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.StudentTransport{}).
		Where("route_id = ? AND school_id = ? AND is_active = ?", routeID, schoolID, true).
		Count(&count).Error
	return count, err
}

func (r *TransportRepository) CountStopsByRoute(routeID, schoolID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.Stop{}).
		Where("route_id = ? AND school_id = ?", routeID, schoolID).
		Count(&count).Error
	return count, err
}

func (r *TransportRepository) CountRoutesByVehicle(vehicleID, schoolID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.Route{}).
		Where("vehicle_id = ? AND school_id = ?", vehicleID, schoolID).
		Count(&count).Error
	return count, err
}

func (r *TransportRepository) CountStudentTransportsByStop(stopID, schoolID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.StudentTransport{}).
		Where("stop_id = ? AND school_id = ?", stopID, schoolID).
		Count(&count).Error
	return count, err
}
