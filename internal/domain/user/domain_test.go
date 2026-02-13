package user

import (
	"testing"

	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockUserDB := mock_outbound_port.NewMockUserDatabasePort(ctrl)

	domain := NewUserDomain(mockDB)

	t.Run("success", func(t *testing.T) {
		userID := uint(1)
		expectedUser := &model.User{ID: userID, Name: "Test", Email: "test@example.com"}

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().FindByID(userID).Return(expectedUser, nil)

		user, err := domain.GetProfile(userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})
}
