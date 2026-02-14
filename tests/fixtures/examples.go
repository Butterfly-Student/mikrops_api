package fixtures

// This file contains examples of how to use the test data factory in tests.
// It serves as documentation and can be referenced for best practices.

/*
Example 1: Using fixtures in unit tests with mocks

func TestUserService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	factory := NewTestDataFactory()
	mockDB := mock_outbound_port.NewMockUserDatabasePort(ctrl)

	// Use factory to create test data
	testUser := factory.User.ValidUser()

	mockDB.EXPECT().FindByID(testUser.ID).Return(&testUser, nil)

	service := NewUserService(mockDB)
	result, err := service.GetUser(testUser.ID)

	assert.NoError(t, err)
	assert.Equal(t, testUser, result)
}

// Example 2: Using fixtures in integration tests

func TestUserIntegration(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	factory := NewTestDataFactory()
	users := factory.User.MultipleUsers(5)

	// Insert test data
	for _, user := range users {
		result := pgContainer.DB.Create(&user)
		if result.Error != nil {
			t.Fatalf("Failed to insert user: %v", result.Error)
		}
	}

	// Test retrieval
	adapter := postgres.NewUserAdapter(pgContainer.DB)
	found, err := adapter.FindByEmail(users[0].Email)

	assert.NoError(t, err)
	assert.Equal(t, users[0].Email, found.Email)
}

// Example 3: Using factories with custom variants

func TestUserRoles(t *testing.T) {
	factory := NewTestDataFactory()

	// Create users with specific roles
	adminUser := factory.User.UserWithRole("admin")
	regularUser := factory.User.UserWithRole("user")

	assert.Equal(t, "admin", adminUser.Role)
	assert.Equal(t, "user", regularUser.Role)
}

// Example 4: Using multiple factories together

func TestPppoeWithRouter(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	factory := NewTestDataFactory()

	// Create router
	router := factory.Mikrotik.ValidMikrotikRouter()
	pgContainer.DB.Create(&router)

	// Create PPPoE data for this router
	secret := factory.Pppoe.ValidPppoeSecret()

	// Test integration between router and PPPoE
	assert.NotNil(t, router.ID)
	assert.Equal(t, "pppoe", secret.Service)
}

// Example 5: Creating bulk test data efficiently

func TestBulkUserCreation(t *testing.T) {
	factory := NewTestDataFactory()

	// Create multiple users efficiently
	users := factory.User.MultipleUsers(100)
	assert.Equal(t, 100, len(users))

	// Bulk insert (pseudo-code)
	// db.CreateInBatches(users, 10)

	for i, user := range users {
		assert.Equal(t, uint(i+1), user.ID)
	}
}

// Example 6: Testing with queues and stats

func TestQueueStats(t *testing.T) {
	factory := NewTestDataFactory()

	queue := factory.Queue.ValidPppoeQueue()
	stats := factory.Queue.ValidQueueStats()

	assert.NotEmpty(t, queue.Name)
	assert.NotEmpty(t, stats.Name)
	assert.Greater(t, stats.BytesIn, int64(0))
}

// Example 7: Testing PPPoE session lifecycle with WebSocket events

func TestPppoeSessionLifecycle(t *testing.T) {
	factory := NewTestDataFactory()

	username := "test-user"

	// Session up event
	sessionUpMsg := factory.Pppoe.WebSocketMessageSessionUp(username)
	assert.Equal(t, "session_up", sessionUpMsg.Event)

	// Session down event
	sessionDownMsg := factory.Pppoe.WebSocketMessageSessionDown(username)
	assert.Equal(t, "session_down", sessionDownMsg.Event)
}

// Example 8: Using factories with table-driven tests

func TestUserValidation(t *testing.T) {
	factory := NewTestDataFactory()

	tests := []struct {
		name    string
		user    *User
		wantErr bool
	}{
		{
			name:    "valid user",
			user:    &factory.User.ValidUser(),
			wantErr: false,
		},
		{
			name:    "admin user",
			user:    &factory.User.AdminUser(),
			wantErr: false,
		},
		{
			name:    "inactive user",
			user:    &factory.User.InactiveUser(),
			wantErr: true, // assuming inactive users don't validate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUser(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
*/
