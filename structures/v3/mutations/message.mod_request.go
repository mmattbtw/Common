package mutations

import (
	"context"
	"time"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (m *Mutate) SendModRequestMessage(ctx context.Context, mb *structures.MessageBuilder[structures.MessageDataModRequest]) error {
	if mb == nil {
		return errors.ErrInternalIncompleteMutation()
	} else if mb.IsTainted() {
		return errors.ErrMutateTaintedObject()
	}

	// Get the message
	req := mb.Message.Data

	// Verify that the target item exists
	var target interface{}
	filter := bson.M{"_id": req.TargetID}
	switch req.TargetKind {
	case structures.ObjectKindEmote:
		filter = bson.M{"versions.id": req.TargetID}
	}
	coll := mongo.CollectionName(req.TargetKind.CollectionName())
	if err := m.mongo.Collection(coll).FindOne(ctx, filter).Decode(&target); err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.ErrInvalidRequest().SetDetail("Target item doesn't exist")
		}
		return errors.ErrInternalServerError().SetDetail(err.Error())
	}

	// Create the message
	result, err := m.mongo.Collection(mongo.CollectionNameMessages).InsertOne(ctx, mb.Message)
	if err != nil {
		return err
	}
	msgID := result.InsertedID.(primitive.ObjectID)
	mb.Message.ID = msgID

	// Create a read state
	_, err = m.mongo.Collection(mongo.CollectionNameMessagesRead).InsertOne(ctx, &structures.MessageRead{
		MessageID: msgID,
		Kind:      structures.MessageKindModRequest,
		Timestamp: time.Now(),
	})

	mb.MarkAsTainted()
	return err
}
