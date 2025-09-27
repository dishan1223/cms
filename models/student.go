package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Student struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	PhoneNumber   string             `bson:"phone_number" json:"phone_number"`
	BatchTime     string             `bson:"batch_time" json:"batch_time"`
	Class         string             `bson:"class" json:"class"`
	Subject       string             `bson:"subject" json:"subject"`
	PaymentStatus bool               `bson:"payment_status" json:"payment_status"`
    PaymentAmount float64            `bson:"payment_amount" json:"payment_amount"`
    PaidMonths    []string           `bson:"paid_months" json:"paid_months"`
    DueMonths     []string           `bson:"due_months" json:"due_months"`
    // satureday , monday, wednesday - smw 
    // sunday , tuesday, thursday - stt
    StudyDays     string             `bson:"study_days" json:"study_days"`
    BatchID       string             `bson:"batch_id" json:"batch_id"`
}
