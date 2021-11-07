package mutations

import (
	"context"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ToVersion: Transform an emote into a versioned emote
func (em *EmoteMutation) ToVersion(
	ctx context.Context,
	inst mongo.Instance,
	parentEmote *structures.Emote,
	opt CreateVersionOptions,
) (*EmoteMutation, error) {
	if em.EmoteBuilder == nil || em.EmoteBuilder.Emote == nil {
		return nil, structures.ErrIncompleteMutation
	}
	targetEmote := em.EmoteBuilder.Emote
	actor := em.Actor

	// Check for the permission to edit emotes
	if actor != nil && !actor.HasPermission(structures.RolePermissionEditEmote) {
		return nil, structures.ErrInsufficientPrivilege
	} // end general permission check

	checkEmoteRights := func(emote *structures.Emote) (bool, error) {
		// Check the user's rights to update the emote
		if emote.OwnerID.IsZero() && !actor.HasPermission(structures.RolePermissionEditAnyEmote) {
			// No permission if the emote has no owner and the user lacks privilege
			return false, structures.ErrInsufficientPrivilege
		}
		if emote.OwnerID != actor.ID {
			ok := false
			// Fetch emote owner?
			if emote.Owner == nil {
				if err := inst.Collection(mongo.CollectionNameUsers).FindOne(ctx, bson.M{
					"_id": emote.OwnerID,
				}).Decode(emote.Owner); err != nil {
					if err != mongo.ErrNoDocuments {
						return false, err
					}
					// Emote's owner couldn't be found, it can't be edited unless the user is privileged
					if actor.HasPermission(structures.RolePermissionEditAnyEmote) {
						ok = true // ok if the user can edit any emote
					}
				} else {
					// Check for permission as an editor
					for _, ed := range emote.Owner.Editors {
						if ed.ID == actor.ID && ed.HasPermission(structures.UserEditorPermissionManageOwnedEmotes) {
							ok = true // ok if the actor is an editor with the permission to manage owned emotes
						}
					}

					ok = true // ok if no error
				}
			}
			if !ok { // error if not ok
				return false, structures.ErrInsufficientPrivilege
			}
		} // end check for right to edit the emote
		return true, nil
	}
	// Check rights for the target emote
	if ok, err := checkEmoteRights(targetEmote); !ok {
		return nil, err
	}
	// Check the rights for the parent emote
	if ok, err := checkEmoteRights(parentEmote); !ok {
		return nil, err
	}

	// Update the emote
	em.EmoteBuilder.SetParentID(parentEmote.ID)
	em.EmoteBuilder.SetVersioningData(structures.EmoteVersioning{
		Tag:       opt.Tag,
		Diverged:  opt.Diverges,
		Timestamp: time.Now(),
	})

	// Write the changes to DB
	if err := inst.Collection(mongo.CollectionNameEmotes).FindOneAndUpdate(
		ctx,
		bson.M{"_id": targetEmote.ID},
		em.EmoteBuilder.Update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(targetEmote); err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, structures.ErrInternalError
	}

	// If version is not diverged, the current version should change
	if !em.EmoteBuilder.Emote.Versioning.Diverged {
		if _, err := em.SetCurrentVersion(ctx, inst, parentEmote); err != nil {
			return nil, err
		}
		em.EmoteBuilder.Emote.ParentID = nil
		em.EmoteBuilder.Emote.Versioning = nil
	}

	return em, nil
}

// SetCurrentVersion: Change the current version of an emote.
// This will change the "parent" emote, meaning the previous emote and all its versions will belong to this one
func (em *EmoteMutation) SetCurrentVersion(ctx context.Context, inst mongo.Instance, emote *structures.Emote) (*EmoteMutation, error) {
	if em.EmoteBuilder == nil || em.EmoteBuilder.Emote == nil {
		return nil, structures.ErrIncompleteMutation
	}

	// Update parent of versioned emotes
	if _, err := inst.Collection(mongo.CollectionNameEmotes).UpdateMany(ctx, bson.M{
		"_id": bson.M{
			"$not": bson.M{"$eq": emote.ID},
		},
		"parent_id": em.EmoteBuilder.Emote.ID,
		"status":    structures.EmoteStatusLive,
	}, bson.M{
		"parent_id": emote.ID,
	}); err != nil {
		return nil, err
	}

	// Remove versioning data for the emote that became current version
	if _, err := inst.Collection(mongo.CollectionNameEmotes).UpdateOne(ctx, bson.M{
		"_id": emote.ID,
	}, bson.M{
		"$unset": bson.M{
			"parent_id": 1,
			"version":   1,
		},
	}); err != nil {
		return nil, err
	}

	return em, nil
}

type CreateVersionOptions struct {
	Tag      string // The version tag
	Diverges bool   // whether or not the emote diverges from the original and should not be treated as an update
}
