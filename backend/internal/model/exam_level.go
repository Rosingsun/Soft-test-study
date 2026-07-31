package model

type ExamLevel struct {
	ID        uint      `gorm:"primarykey"`
	Name      string    `gorm:"column:name;type:varchar(50);not null"`
	SortOrder int       `gorm:"column:sort_order;default:0"`
	Subjects  []Subject `gorm:"foreignKey:LevelID"`
}

func (ExamLevel) TableName() string {
	return "exam_levels"
}
