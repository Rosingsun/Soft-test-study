package model

type SubSubject struct {
	ID        uint      `gorm:"primarykey"`
	SubjectID uint      `gorm:"column:subject_id;not null;index"`
	Name      string    `gorm:"column:name;type:varchar(200);not null"`
	SortOrder int       `gorm:"column:sort_order;default:0"`
	Subject   Subject   `gorm:"foreignKey:SubjectID"`
	Chapters  []Chapter `gorm:"foreignKey:SubSubjectID"`
}

func (SubSubject) TableName() string {
	return "sub_subjects"
}
