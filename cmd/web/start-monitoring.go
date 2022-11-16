package main

import "log"

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
			// create a job
			// save the id of the job so we can start/stop it
			// broadcast over websockets the fact that the service is scheduled
		}
	}
}
