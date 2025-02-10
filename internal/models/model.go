package models

type PolishWord struct {
	ID   uint   `gorm:"primaryKey"`
	Text string `gorm:"unique;not null"`
}

type EnglishWord struct {
	ID   uint   `gorm:"primaryKey"`
	Text string `gorm:"unique;not null"`
}

type Translation struct {
	ID            uint        `gorm:"primaryKey"`
	PolishWordID  uint        `gorm:"not null;index"`
	EnglishWordID uint        `gorm:"not null;index"`
	PolishWord    PolishWord  `gorm:"foreignKey:PolishWordID;constraint:OnDelete:CASCADE"`
	EnglishWord   EnglishWord `gorm:"foreignKey:EnglishWordID;constraint:OnDelete:CASCADE"`
	Examples      []Example   `gorm:"foreignKey:TranslationID"`
}

type Example struct {
	ID            uint        `gorm:"primaryKey"`
	TranslationID uint        `gorm:"not null;index"`
	Text          string      `gorm:"not null"`
	Translation   Translation `gorm:"foreignKey:TranslationID;constraint:OnDelete:CASCADE"`
}
