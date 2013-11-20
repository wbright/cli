package api

import (
	"cf"
	"cf/net"
)

type FakeAppEventsRepo struct{
	AppGuid string
	Events []cf.Event
}


func (repo FakeAppEventsRepo)ListEvents(appGuid string) (events chan []cf.Event, statusChan chan net.ApiResponse) {
	repo.AppGuid = appGuid

	events = make(chan []cf.Event, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		for _, event := range repo.Events {
			events <- []cf.Event{event}
		}
		close(events)
		close(statusChan)
	}()

	return
}
