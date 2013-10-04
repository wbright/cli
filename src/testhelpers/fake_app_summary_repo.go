package testhelpers

import (
	"cf"
	"cf/api"
)

type FakeAppSummaryRepo struct{
	GetSummaryApp cf.Application
	GetSummarySummary cf.AppSummary
}


func (repo *FakeAppSummaryRepo)GetSummary(app cf.Application) (summary cf.AppSummary, apiStatus api.ApiStatus) {
	repo.GetSummaryApp= app
	summary = repo.GetSummarySummary

	return
}
