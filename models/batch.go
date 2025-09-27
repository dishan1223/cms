package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Batch struct {
    ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    BatchName     string             `bson:"batch_name" json:"batch_name"`
    Time          string             `bson:"time" json:"time"`
    Days          []string           `bson:"days" json:"days"`
    Class         string             `bson:"class" json:"class"`
    Subject       string             `bson:"subject" json:"subject"`
    TotalStudents int                `bson:"total_students" json:"total_students"`
    Payment_amount float64            `bson:"payment_amount" json:"payment_amount"`
}
