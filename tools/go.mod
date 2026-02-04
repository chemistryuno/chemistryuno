module tools

go 1.24.0

replace chemistryuno => ../backend

require (
	chemistryuno v0.0.0
	gorm.io/driver/sqlite v1.5.5
	gorm.io/gorm v1.25.7
)
