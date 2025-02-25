// pkg/dal/models.go
type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex"`
	Password string
}

type Product struct {
	gorm.Model
	Name  string
	Price float64
	Stock int
}