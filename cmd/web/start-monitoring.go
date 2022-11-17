package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

type job struct {
	HostServiceID int
}

// Run runs the scheduled job
func (j job) Run() {
	repo.ScheduledCheck(j.HostServiceID)
}

// startMonitoring starts the monitoring process
func startMonitoring() {
	if preferenceMap["monitoring_live"] == "1" {
		log.Println("************* starting monitoring **************")
		// trigger a message to broadcast to all clients that app is strating to monitor
		data := make(map[string]string)
		data["message"] = "Monitoring is started..."
		err := app.WsClient.Trigger("public-channel", "app-starting", data)
		if err != nil {
			log.Println(err)
		}

		// get all of the services that we want to monitor
		servicesToMonitor, err := repo.DB.GetServicesToMonitor()
		if err != nil {
			log.Println(err)
		}

		log.Println("len(servicesToMonitor):", len(servicesToMonitor))
		// range through the services
		for _, hs := range servicesToMonitor {
			log.Println("Service to monitor on", hs.HostName, "is", hs.Service.ServiceName)

			// get the schedule unit and number
			var sch string
			if hs.ScheduleUnit == "d" {
				sch = fmt.Sprintf("@every %d%s", hs.ScheduleNumber*24, "h")
			} else {
				sch = fmt.Sprintf("@every %d%s", hs.ScheduleNumber, hs.ScheduleUnit)
			}

			// create a job
			var j job
			j.HostServiceID = hs.ID
			scheduleID, err := app.Scheduler.AddJob(sch, j)
			if err != nil {
				log.Println(err)
			}

			// save the id of the job so we can start/stop it
			app.MonitorMap[hs.ID] = scheduleID

			// broadcast over websockets the fact that the service is scheduled
			payload := make(map[string]string)
			payload["message"] = "scheduling"
			payload["host_service_id"] = strconv.Itoa(hs.ID)
			yearOne := time.Date(0001, 11, 17, 20, 34, 58, 65138737, time.UTC)
			if app.Scheduler.Entry(app.MonitorMap[hs.ID]).Next.After(yearOne) {
				payload["next_run"] = app.Scheduler.Entry(app.MonitorMap[hs.ID]).Next.Format("2006-01-02 3:04:05 PM")
			} else {
				payload["next_run"] = "Pending..."
			}

			payload["host"] = hs.HostName
			payload["service"] = hs.Service.ServiceName
			if hs.LastCheck.After(yearOne) {
				payload["last_run"] = hs.LastCheck.Format("2006-01-02 3:04:05 PM")
			} else {
				payload["last_run"] = "Pending..."
			}
			payload["schedule"] = fmt.Sprintf("@every %d%s", hs.ScheduleNumber, hs.ScheduleUnit)

			err = app.WsClient.Trigger("public-channel", "next-run-event", payload)
			if err != nil {
				log.Println(err)
			}

			err = app.WsClient.Trigger("public-channel", "schedule-changed-event", payload)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
