package postgres

type repository struct {
	//db *gorm.DB
	db *string // để đỡ để có mock data test mấy cái res
}

func NewMovementRepository(db *string) *repository {
	return &repository{
		db: db,
	}
}
