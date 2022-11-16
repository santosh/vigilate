package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tsawler/vigilate/internal/models"
)

const (
	HTTP           = 1
	HTTPS          = 2
	SSLCertificate = 3
)

type jsonResp struct {
	OK            bool      `json:"ok"`
	Message       string    `json:"message"`
	ServiceID     int       `json:"service_id"`
	HostServiceID int       `json:"host_service_id"`
	HostID        int       `json:"host_id"`
	OldStatus     string    `json:"old_status"`
	NewStatus     string    `json:"new_status"`
	LastCheck     time.Time `json:"last_check"`
}

// ScheduledCheck performs a scheduled check on a host service by id
func (repo *DBRepo) ScheduledCheck(hostServiceID int) {

}

func (repo *DBRepo) TestCheck(w http.ResponseWriter, r *http.Request) {
	hostServiceId, _ := strconv.Atoi(chi.URLParam(r, "id"))
	oldStatus := chi.URLParam(r, "oldStatus")
	okay := true

	// get host service
	hs, err := repo.DB.GetHostServiceById(hostServiceId)
	if err != nil {
		log.Println(err)
		okay = false
	}

	// get host?
	h, err := repo.DB.GetHostByID(hs.HostID)
	if err != nil {
		log.Println(err)
		okay = false
	}

	// test the service
	msg, newStatus := repo.testServiceForHost(h, hs)

	// update the host service in the database with status (if changed) and last check
	hs.Status = newStatus
	hs.LastCheck = time.Now()
	hs.UpdatedAt = time.Now()
	err = repo.DB.UpdateHostService(hs)
	if err != nil {
		log.Println(err)
		okay = false
	}

	// broadcast serivce status changed event

	// create json
	var resp jsonResp
	if okay {
		resp = jsonResp{
			OK:            okay,
			Message:       msg,
			ServiceID:     hs.ServiceID,
			HostServiceID: hs.ID,
			HostID:        hs.HostID,
			OldStatus:     oldStatus,
			NewStatus:     newStatus,
			LastCheck:     time.Now(),
		}
	} else {
		resp.OK = false
		resp.Message = "Something went wrong"
	}

	// send json to client
	out, _ := json.MarshalIndent(resp, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (repo *DBRepo) testServiceForHost(h models.Host, hs models.HostService) (string, string) {
	var msg, newStatus string

	switch hs.ServiceID {
	case HTTP:
		msg, newStatus = testHTTPForHost(h.URL)
	}

	return msg, newStatus
}

func testHTTPForHost(url string) (string, string) {
	url = strings.TrimSuffix(url, "/")
	url = strings.Replace(url, "https://", "http://", -1)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("%s - %s", url, "error connecting"), "problem"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("%s - %s", url, resp.Status), "problem"
	}

	return fmt.Sprintf("%s - %d", url, resp.StatusCode), "healthy"
}
