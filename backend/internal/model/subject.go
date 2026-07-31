package model

type Subject struct {
	ID          uint         `gorm:"primarykey"`
	LevelID     uint         `gorm:"column:level_id;not null;index"`
	Name        string       `gorm:"column:name;type:varchar(100);not null"`
	ShortName   string       `gorm:"column:short_name;type:varchar(20)"`
	Description string       `gorm:"column:description;type:text"`
	Icon        string       `gorm:"column:icon;type:varchar(255)"`
	SortOrder   int          `gorm:"column:sort_order;default:0"`
	Status      int          `gorm:"column:status;default:1"`
	Level       ExamLevel    `gorm:"foreignKey:LevelID"`
	SubSubjects []SubSubject `gorm:"foreignKey:SubjectID"`
}

func (Subject) TableName() string {
	return "subjects"
}
