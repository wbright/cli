package testhelpers

import "cf/api"

type FakePasswordRepo struct {
	Score string
	ScoredPassword string

	UpdateUnauthorized bool
	UpdateNewPassword string
	UpdateOldPassword string
}



func (repo *FakePasswordRepo) GetScore(password string) (score string, apiStatus api.ApiStatus){
	repo.ScoredPassword = password
	score = repo.Score
	return
}

func (repo *FakePasswordRepo) UpdatePassword(old string, new string) (apiStatus api.ApiStatus) {
	repo.UpdateOldPassword = old
	repo.UpdateNewPassword = new

	if repo.UpdateUnauthorized {
		apiStatus = api.NewApiStatus("Authorization Failed", "unauthorized", 401)
	}

	return
}

