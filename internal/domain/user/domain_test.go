package user

import (
	"testing"

	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockUserDB := mock_outbound_port.NewMockUserDatabasePort(ctrl)

	domain := NewUserDomain(mockDB)

	t.Run("success", func(t *testing.T) {
		userID := "550e8400-e29b-41d4-a716-446655440001"
		expectedUser := &model.AdminUser{
			ID:       uuid.MustParse(userID),
			FullName: "Test",
			Email:    "test@example.com",
		}

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().FindByID(userID).Return(expectedUser, nil)

		user, err := domain.GetProfile(userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})
}

func TestUpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockUserDB := mock_outbound_port.NewMockUserDatabasePort(ctrl)

	domain := NewUserDomain(mockDB)

	t.Run("success", func(t *testing.T) {
		userID := "550e8400-e29b-41d4-a716-446655440001"
		existingUser := &model.AdminUser{
			ID:       uuid.MustParse(userID),
			FullName: "Old Name",
			Email:    "old@example.com",
		}

		req := model.AdminUserInput{
			FullName: "New Name",
			Email:    "new@example.com",
		}

		mockDB.EXPECT().User().Return(mockUserDB).AnyTimes()
		mockUserDB.EXPECT().FindByID(userID).Return(existingUser, nil)
		mockUserDB.EXPECT().FindByEmail(req.Email).Return(nil, assert.AnError)
		mockUserDB.EXPECT().Update(gomock.Any()).Return(nil)

		err := domain.UpdateProfile(userID, req)
		assert.NoError(t, err)
	})

	t.Run("email already taken", func(t *testing.T) {
		userID := "550e8400-e29b-41d4-a716-446655440001"
		otherUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
		existingUser := &model.AdminUser{
			ID:       uuid.MustParse(userID),
			FullName: "Test User",
			Email:    "test@example.com",
		}

		req := model.AdminUserInput{
			FullName: "Test User",
			Email:    "new@example.com",
		}

		existingOtherUser := &model.AdminUser{
			ID:    otherUserID,
			Email: "new@example.com",
		}

		mockDB.EXPECT().User().Return(mockUserDB).AnyTimes()
		mockUserDB.EXPECT().FindByID(userID).Return(existingUser, nil)
		mockUserDB.EXPECT().FindByEmail(req.Email).Return(existingOtherUser, nil)

		err := domain.UpdateProfile(userID, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email already taken")
	})
}
