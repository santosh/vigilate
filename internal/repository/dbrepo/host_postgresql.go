package dbrepo

import (
	"context"
	"log"
	"time"

	"github.com/tsawler/vigilate/internal/models"
)

// InsertHost inserts a host into hosts table
func (m *postgresDBRepo) InsertHost(h models.Host) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO hosts (host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning id`

	var newID int

	err := m.DB.QueryRowContext(ctx, query,
		h.HostName,
		h.CanonicalName,
		h.URL,
		h.IP,
		h.IPV6,
		h.Location,
		h.OS,
		h.Active,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		log.Println(err)
		return newID, err
	}

	// add HostServices and set to inactive
	stmt := `insert into host_services (host_id, service_id, active, schedule_number, schedule_unit, status, created_at, updated_at) values ($1, 1, 0, 3, 'm', 'pending', $2, $3)`

	_, err = m.DB.ExecContext(ctx, stmt, newID, time.Now(), time.Now())
	if err != nil {
		return newID, err
	}
	return newID, nil
}

// GetHostByID returns a host of given HostID
func (m *postgresDBRepo) GetHostByID(id int) (models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at from hosts where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var h models.Host

	err := row.Scan(
		&h.ID,
		&h.HostName,
		&h.CanonicalName,
		&h.URL,
		&h.IP,
		&h.IPV6,
		&h.Location,
		&h.OS,
		&h.Active,
		&h.CreatedAt,
		&h.UpdatedAt,
	)

	if err != nil {
		log.Println(err)
		return h, err
	}

	// get all services for host
	query = `select
		hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit, hs.last_check, hs.status, hs.created_at, hs.updated_at,
		s.id, s.service_name, s.active, s.icon, s.created_at, s.updated_at
	from
		host_services hs
		left join services s on (s.id = hs.service_id)
	where
		host_id = $1`

	rows, err := m.DB.QueryContext(ctx, query, h.ID)
	if err != nil {
		log.Println(err)
		return h, err
	}
	defer rows.Close()

	var hostServices []models.HostService
	for rows.Next() {
		var hs models.HostService
		err := rows.Scan(
			&hs.ID,
			&hs.HostID,
			&hs.ServiceID,
			&hs.Active,
			&hs.ScheduleNumber,
			&hs.ScheduleUnit,
			&hs.LastCheck,
			&hs.Status,
			&hs.CreatedAt,
			&hs.UpdatedAt,
			&hs.Service.ID,
			&hs.Service.ServiceName,
			&hs.Service.Active,
			&hs.Service.Icon,
			&hs.Service.CreatedAt,
			&hs.Service.UpdatedAt,
		)
		if err != nil {
			log.Println(err)
			return h, err
		}

		hostServices = append(hostServices, hs)
	}

	h.HostServices = hostServices

	return h, nil
}

// UpdateHost takes in a models.Host and update relevant record on db
func (m *postgresDBRepo) UpdateHost(h models.Host) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update hosts set host_name = $1, canonical_name = $2, url = $3, ip = $4, ipv6 = $5, location = $6, os = $7, active = $8, updated_at = $9 where id = $10`

	_, err := m.DB.ExecContext(ctx, stmt,
		h.HostName,
		h.CanonicalName,
		h.URL,
		h.IP,
		h.IPV6,
		h.Location,
		h.OS,
		h.Active,
		time.Now(),
		h.ID,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (m *postgresDBRepo) GetAllServiceStatusCounts() (int, int, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `select
		(select count(id) from host_services where active = 1 and status = 'pending') as pending,
		(select count(id) from host_services where active = 1 and status = 'healthy') as healthy,
		(select count(id) from host_services where active = 1 and status = 'warning') as warning,
		(select count(id) from host_services where active = 1 and status = 'problem') as problem`

	var pending, healthy, warning, problem int

	row := m.DB.QueryRowContext(ctx, query)
	err := row.Scan(
		&pending,
		&healthy,
		&warning,
		&problem,
	)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return pending, healthy, warning, problem, nil
}

// AllHosts returns a slice of hosts
func (m *postgresDBRepo) AllHosts() ([]models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at from hosts order by host_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []models.Host
	for rows.Next() {
		var h models.Host
		err = rows.Scan(
			&h.ID,
			&h.HostName,
			&h.CanonicalName,
			&h.URL,
			&h.IP,
			&h.IPV6,
			&h.Location,
			&h.OS,
			&h.Active,
			&h.CreatedAt,
			&h.UpdatedAt,
		)

		if err != nil {
			log.Println(err)
			return nil, err
		}

		// get all services for host
		serviceQuery := `select
					hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit, hs.last_check, hs.status, hs.created_at, hs.updated_at,
					s.id, s.service_name, s.active, s.icon, s.created_at, s.updated_at
				from
					host_services hs
					left join services s on (s.id = hs.service_id)
				where
					host_id = $1`

		serviceRows, err := m.DB.QueryContext(ctx, serviceQuery, h.ID)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		var hostServices []models.HostService
		for serviceRows.Next() {
			var hs models.HostService
			serviceRows.Scan(
				&hs.ID,
				&hs.HostID,
				&hs.ServiceID,
				&hs.Active,
				&hs.ScheduleNumber,
				&hs.ScheduleUnit,
				&hs.LastCheck,
				&hs.Status,
				&hs.CreatedAt,
				&hs.UpdatedAt,
				&hs.Service.ID,
				&hs.Service.ServiceName,
				&hs.Service.Active,
				&hs.Service.Icon,
				&hs.Service.CreatedAt,
				&hs.Service.UpdatedAt,
			)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			hostServices = append(hostServices, hs)
			serviceRows.Close()
		}
		h.HostServices = hostServices
		hosts = append(hosts, h)
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return hosts, nil
}

// UpdateHostServiceStatus updates the active status of a host service
func (m *postgresDBRepo) UpdateHostServiceStatus(hostID, serviceID, active int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update host_services set active = $1 where host_id = $2 and service_id = $3`

	_, err := m.DB.ExecContext(ctx, stmt, active, hostID, serviceID)
	if err != nil {
		return err
	}

	return nil
}
