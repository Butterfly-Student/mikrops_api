Baik! Saya akan menjelaskan perbedaan jenis-jenis testing yang ada di project ini:                                                                                                                                                                                                                           
  📚 Tipe-Tipe Testing di Project Ini                                                                                                                                                                                                                                                                       
  
  1️ Unit Tests (internal/domain/*/domain_test.go)                                                                                                                                                                                                                                                        %                                                                                                                                                                                                                                                                                                                Lokasi: internal/domain/customer/domain_test.go, internal/domain/payment/domain_test.go, dll.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         
  Tujuan: Test business logic secara terisolasi

  Karakteristik:
  - ✅ Menggunakan mocks untuk semua dependencies
  - ✅ Tidak ada database real
  - ✅ Tidak ada network calls
  - ✅ Sangat cepat (milliseconds)
  - ✅ Test satu function/method secara isolated

  Contoh:
  func TestCreateCustomer(t *testing.T) {
      // Mock database
      mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
      mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)

      // Setup mock expectations
      mockDB.EXPECT().Customer().Return(mockCustomerDB)
      mockCustomerDB.EXPECT().Create(gomock.Any()).Return(nil)

      // Test domain logic
      domain := NewCustomerDomain(mockDB, mockMikrotikPort)
      result, err := domain.CreateCustomer(ctx, input)

      assert.NoError(t, err)
  }

  Apa yang di-test:
  - ✅ Business logic correctness
  - ✅ Error handling
  - ✅ Edge cases
  - ✅ Validation rules

  ---
  2️⃣HTTP Handler Tests (internal/adapter/inbound/gin/*_test.go)

  Lokasi: internal/adapter/inbound/gin/auth_test.go, internal/adapter/inbound/gin/user_test.go, dll.

  Tujuan: Test HTTP layer - request/response handling

  Karakteristik:
  - ✅ Menggunakan mocks untuk domain layer
  - ✅ Menggunakan httptest package untuk fake HTTP requests
  - ✅ Test routing, middleware, request parsing
  - ✅ Tidak ada database real
  - ✅ Fast (milliseconds to seconds)

  Contoh:
  func TestAuthAdapter(t *testing.T) {
      // Setup mock domain
      mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
      dom := domain.NewDomain(mockDatabasePort, ...)
      adapter := gin_inbound_adapter.NewAdapter(dom)

      // Setup fake HTTP router
      router := gin.New()
      router.POST("/auth/login", adapter.Auth().Login)

      // Create fake HTTP request
      reqBody := model.LoginRequest{Email: "test@example.com", Password: "password"}
      bodyBytes, _ := json.Marshal(reqBody)
      req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(bodyBytes))
      w := httptest.NewRecorder()

      // Execute request
      router.ServeHTTP(w, req)

      // Assert HTTP response
      assert.Equal(t, http.StatusOK, w.Code)
  }

  Apa yang di-test:
  - ✅ HTTP request parsing (JSON, query params, headers)
  - ✅ HTTP response format (status codes, JSON structure)
  - ✅ Routing correctness
  - ✅ Middleware behavior (auth, logging, etc.)
  - ✅ Input validation at HTTP layer
  - ✅ Error response formatting

  ---
  3️⃣Integration Tests (tests/integration/*_integration_test.go)

  Lokasi: tests/integration/auth_integration_test.go, tests/integration/pppoe_integration_test.go, dll.

  Tujuan: Test end-to-end flow dengan real dependencies

  Karakteristik:
  - ✅ Menggunakan real database (via testcontainers)
  - ✅ Menggunakan real adapters (postgres, redis, etc.)
  - ✅ Test full request flow: HTTP → Domain → Database → Response
  - ✅ Slower (seconds to minutes)
  - ✅ Requires Docker for testcontainers
  - ✅ Tagged with //go:build integration

  Contoh:
  //go:build integration
  // +build integration

  func TestAuthIntegration(t *testing.T) {
      // Setup REAL PostgreSQL container
      pgContainer, err := helpers.SetupPostgresContainer(ctx)
      defer pgContainer.Terminate(ctx)

      // Run REAL migrations
      pgContainer.DB.AutoMigrate(&model.User{})

      // Use REAL adapters
      adapter := postgres_outbound_adapter.NewUserAdapter(pgContainer.DB)
      dbAdapter := postgres_outbound_adapter.NewAdapter(pgContainer.DB)
      dom := domain.NewDomain(dbAdapter, nil, nil, nil, nil, enforcer)

      // Test real flow
      user := &model.User{Email: email, Password: hashedPassword}
      err := adapter.Create(user)  // Real database insert!

      // Verify in real database
      foundUser, err := adapter.FindByEmail(email)
      assert.NoError(t, err)
  }

  Apa yang di-test:
  - ✅ Full end-to-end workflow
  - ✅ Database interactions (real SQL queries)
  - ✅ Transaction handling
  - ✅ Data persistence
  - ✅ Integration between layers
  - ✅ Real error scenarios (DB constraints, etc.)

  ---
  📊 Perbandingan Detail

  ┌─────────────────┬───────────────────┬────────────────────┬──────────────────────────┐
  │      Aspek      │    Unit Tests     │ HTTP Handler Tests │    Integration Tests     │
  ├─────────────────┼───────────────────┼────────────────────┼──────────────────────────┤
  │ Speed           │ ⚡ Very Fast (ms) │ ⚡ Fast (ms-s)     │ 🐌 Slow (s-min)          │
  ├─────────────────┼───────────────────┼────────────────────┼──────────────────────────┤
  │ Dependencies    │ 🔸 All mocked     │ 🔸 Domain mocked   │ ✅ All real              │
  ├─────────────────┼───────────────────┼────────────────────┼──────────────────────────┤
  │ Database        │ ❌ Mock           │ ❌ Mock            │ ✅ Real (testcontainers) │
  ├─────────────────┼───────────────────┼────────────────────┼──────────────────────────┤
  │ Network         │ ❌ No             │ ❌ Fake HTTP       │ ✅ Real HTTP             │
  ├─────────────────┼───────────────────┼────────────────────┼──────────────────────────┤
  │ Isolation       │ ✅ Complete       │ ⚠️ Partial         │ ❌ Full stack            │
  ├─────────────────┼───────────────────┼────────────────────┼──────────────────────────┤
  │ Docker Required │ ❌ No             │ ❌ No              │ ✅ Yes                   │
  ├─────────────────┼───────────────────┼────────────────────┼──────────────────────────┤
  │ CI/CD Friendly  │ ✅ Yes            │ ✅ Yes             │ ⚠️ Slower                │
  ├─────────────────┼───────────────────┼────────────────────┼──────────────────────────┤
  │ Debugging       │ ✅ Easy           │ ✅ Medium          │ ⚠️ Complex               │
  ├─────────────────┼───────────────────┼────────────────────┼──────────────────────────┤
  │ Test Scope      │ Single function   │ HTTP layer         │ Entire flow              │
  └─────────────────┴───────────────────┴────────────────────┴──────────────────────────┘

  ---
  🎯 Kapan Menggunakan Yang Mana?

  Gunakan Unit Tests ketika:

  - ✅ Test business logic / domain rules
  - ✅ Test validation
  - ✅ Test error handling
  - ✅ Test edge cases
  - ✅ Development awal (TDD)

  Gunakan HTTP Handler Tests ketika:

  - ✅ Test API endpoints
  - ✅ Test request/response format
  - ✅ Test middleware behavior
  - ✅ Test routing logic
  - ✅ Test HTTP status codes

  Gunakan Integration Tests ketika:

  - ✅ Test database transactions
  - ✅ Test full user workflows
  - ✅ Test data persistence
  - ✅ Verify real SQL queries
  - ✅ Before production deployment

  ---
  🏗️ Testing Pyramid di Project Ini

          /\
         /  \  Integration Tests (Few)
        /    \  - Full E2E flows
       /      \ - Real dependencies
      /--------\
     /          \ HTTP Handler Tests (Medium)
    /            \ - API endpoints
   /              \ - Request/response
  /----------------\
  |                | Unit Tests (Many)
  |   Domain Logic | - Business rules
  |   Validation   | - Fast & isolated
  |________________|

  Best Practice:
  - 70% Unit Tests (fast, isolated)
  - 20% HTTP Handler Tests (API layer)
  - 10% Integration Tests (E2E verification)

  ---
  🚀 Cara Menjalankan

  # Unit tests (domain logic)
  go test ./internal/domain/... -v

  # HTTP Handler tests (API layer)
  go test ./internal/adapter/inbound/gin/... -v

  # Integration tests (E2E with real DB)
  go test -tags=integration ./tests/integration/... -v

  # All tests
  go test ./... -v

  # Skip integration tests
  go test ./... -short -v

  ---
  💡 Kesimpulan

  HTTP Handler Tests (inbound/gin):
  - Test API layer (HTTP request/response)
  - Mock domain layer
  - Fast & isolated
  - Focus on HTTP behavior

  Integration Tests (tests/integration):
  - Test full stack (HTTP → Domain → Database)
  - Real dependencies (PostgreSQL via Docker)
  - Slow but comprehensive
  - Focus on E2E workflows
