package seeders

type DatabaseSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *DatabaseSeeder) Signature() string {
	return "DatabaseSeeder"
}

// Run executes the seeder logic.
// 注意：现在所有 seeder 都在 database/kernel.go 中按顺序注册，
// DatabaseSeeder 作为主入口，可以在这里添加一些全局的初始化逻辑
func (s *DatabaseSeeder) Run() error {
	// 所有具体的 seeder 都在 kernel.go 中按顺序注册执行
	// 这里可以添加一些全局的初始化逻辑，如果需要的话
	return nil
}
