package handlers

import (
	"context"
	"testing"

	"github.com/NdoleStudio/httpsms/pkg/requests"
	"github.com/NdoleStudio/httpsms/pkg/responses"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
)

func createTestPhone(t *testing.T) *responses.PhoneResponse {
	t.Helper()

	fake := faker.New()
	response := new(responses.PhoneResponse)
	err := testClient().
		Put().
		Path("/v1/phones").
		BodyJSON(requests.PhoneUpsert{
			PhoneNumber: "+1213" + fake.RandomStringWithLength(7),
			SIM:         "SIM1",
		}).
		ToJSON(response).
		Fetch(context.Background())

	assert.Nil(t, err)
	return response
}

func deleteTestPhone(id string) {
	_ = testClient().
		Delete().
		Path("/v1/phones/" + id).
		Fetch(context.Background())
}

func TestPhoneHandler_updateOperators(t *testing.T) {
	// Arrange
	phone := createTestPhone(t)
	defer deleteTestPhone(phone.Data.ID.String())

	// Act
	response := new(responses.PhoneResponse)
	err := testClient().
		Put().
		Path("/v1/phones/" + phone.Data.ID.String() + "/operators").
		BodyJSON(requests.PhoneOperatorsUpdate{
			SupportedOperators: []string{"Orange", " mtn "},
		}).
		ToJSON(response).
		Fetch(context.Background())

	// Assert
	assert.Nil(t, err)
	assert.ElementsMatch(t, []string{"orange", "mtn"}, []string(response.Data.SupportedOperators))
	assert.True(t, response.Data.SupportsOperator("orange"))
	assert.True(t, response.Data.SupportsOperator("MTN"))
	assert.False(t, response.Data.SupportsOperator("moov"))
}

func TestPhoneHandler_updateOperators_unknownOperatorRejected(t *testing.T) {
	// Arrange
	phone := createTestPhone(t)
	defer deleteTestPhone(phone.Data.ID.String())

	// Act
	err := testClient().
		Put().
		Path("/v1/phones/" + phone.Data.ID.String() + "/operators").
		BodyJSON(requests.PhoneOperatorsUpdate{
			SupportedOperators: []string{"wanadoo"},
		}).
		Fetch(context.Background())

	// Assert
	assert.NotNil(t, err)
}

func TestPhoneHandler_updateOperators_unknownPhoneNotFound(t *testing.T) {
	// Act
	err := testClient().
		Put().
		Path("/v1/phones/00000000-0000-0000-0000-000000000000/operators").
		BodyJSON(requests.PhoneOperatorsUpdate{
			SupportedOperators: []string{"orange"},
		}).
		Fetch(context.Background())

	// Assert
	assert.NotNil(t, err)
}

func TestPhoneHandler_updateOperators_emptyListMeansGeneric(t *testing.T) {
	// Arrange
	phone := createTestPhone(t)
	defer deleteTestPhone(phone.Data.ID.String())

	// Act - a phone that never had its operators configured supports any operator
	// (SupportsOperator's zero-value/empty-slice behavior).
	assert.True(t, phone.Data.SupportsOperator("orange"))
	assert.True(t, phone.Data.SupportsOperator("mtn"))
	assert.True(t, phone.Data.SupportsOperator("moov"))
}
