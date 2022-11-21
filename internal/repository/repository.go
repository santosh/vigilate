package repository

import "github.com/tsawler/vigilate/internal/models"

// DatabaseRepo is the database repository
type DatabaseRepo interface {
	// preferences
	AllPreferences() ([]models.Preference, error)
	SetSystemPref(name, value string) error
	UpdateSystemPref(name, value string) error
	InsertOrUpdateSitePreferences(pm map[string]string) error

	// users and authentication
	GetUserById(id int) (models.User, error)
	InsertUser(u models.User) (int, error)
	UpdateUser(u models.User) error
	DeleteUser(id int) error
	UpdatePassword(id int, newPassword string) error
	Authenticate(email, testPassword string) (int, string, error)
	AllUsers() ([]*models.User, error)
	InsertRememberMeToken(id int, token string) error
	DeleteToken(token string) error
	CheckForToken(id int, token string) bool

	// hosts
	InsertHost(h models.Host) (int, error)
	GetHostByID(id int) (models.Host, error)
	UpdateHost(h models.Host) error
	AllHosts() ([]models.Host, error)
	GetAllServiceStatusCounts() (int, int, int, int, error)

	// services
	GetHostServiceById(id int) (models.HostService, error)
	GetServicesToMonitor() ([]models.HostService, error)

	// host service
	GetServicesByStatus(status string) ([]models.HostService, error)
	UpdateHostService(hs models.HostService) error
	UpdateHostServiceStatus(hostID, serviceID, active int) error
	GetHostServiceByHostIdServiceID(hostID, serviceID int) (models.HostService, error)

	// events
	GetAllEvents() ([]models.Event, error)
}
