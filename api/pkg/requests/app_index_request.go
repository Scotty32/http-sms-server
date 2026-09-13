package requests

import "github.com/NdoleStudio/httpsms/pkg/repositories"

// AppIndex is the payload for listing Apps
type AppIndex struct {
	request
	Limit string `query:"limit"`
	Skip  string `query:"skip"`
	Query string `query:"query"`
}

// Sanitize sets defaults to AppIndex
func (input *AppIndex) Sanitize() AppIndex {
	if input.Limit == "" {
		input.Limit = "20"
	}
	if input.Skip == "" {
		input.Skip = "0"
	}
	return *input
}

// ToIndexParams converts AppIndex to repositories.IndexParams
func (input *AppIndex) ToIndexParams() repositories.IndexParams {
	return repositories.IndexParams{
		Limit: input.getInt(input.Limit),
		Skip:  input.getInt(input.Skip),
		Query: input.Query,
	}
}
