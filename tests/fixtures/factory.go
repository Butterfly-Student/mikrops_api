package fixtures

// TestDataFactory provides access to all test data factories
type TestDataFactory struct {
	Client   *ClientTestData
	User     *UserTestData
	Mikrotik *MikrotikTestData
	Pppoe    *PppoeTestData
	Queue    *QueueTestData
}

// NewTestDataFactory creates a new instance of TestDataFactory with all sub-factories initialized
func NewTestDataFactory() *TestDataFactory {
	return &TestDataFactory{
		Client:   NewClientTestData(),
		User:     NewUserTestData(),
		Mikrotik: NewMikrotikTestData(),
		Pppoe:    NewPppoeTestData(),
		Queue:    NewQueueTestData(),
	}
}
