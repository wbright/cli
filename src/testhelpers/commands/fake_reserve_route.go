package commands

import (
	"cf"
)

type FakeReserveRoute struct {
	RouteToReserve cf.Route
	ReservedRoute cf.Route
}

func (cmd *FakeReserveRoute) ReserveRoute(route cf.Route) (reservedRoute cf.Route, err error) {
	cmd.RouteToReserve = route
	reservedRoute = cmd.ReservedRoute
	return
}
