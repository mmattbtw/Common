package events

import (
	"encoding/json"

	"github.com/SevenTV/Common/structures/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventType string

const (
	// System

	EventTypeAnySystem          EventType = "system.*"
	EventTypeSystemAnnouncement EventType = "system.announcement"

	// Emote

	EventTypeAnyEmote    EventType = "emote.*"
	EventTypeCreateEmote EventType = "emote.create"
	EventTypeUpdateEmote EventType = "emote.update"
	EventTypeDeleteEmote EventType = "emote.delete"

	// Emote Set

	EventTypeAnyEmoteSet    EventType = "emote_set.*"
	EventTypeCreateEmoteSet EventType = "emote_set.create"
	EventTypeUpdateEmoteSet EventType = "emote_set.update"
	EventTypeDeleteEmoteSet EventType = "emote_set.delete"

	// User

	EventTypeAnyUser              EventType = "user.*"
	EventTypeCreateUser           EventType = "user.create"
	EventTypeUpdateUser           EventType = "user.update"
	EventTypeDeleteUser           EventType = "user.delete"
	EventTypeAddUserConnection    EventType = "user.add_connection"
	EventTypeUpdateUserConnection EventType = "user.update_connection"
	EventTypeDeleteUserConnection EventType = "user.delete_connection"
)

type EmptyObject = struct{}

type ChangeMap struct {
	// The object's ID
	ID primitive.ObjectID `json:"id"`
	// The type of the object
	Kind structures.ObjectKind `json:"kind"`
	// A list of added fields
	Added []ChangeField `json:"added,omitempty"`
	// A list of updated fields
	Updated []ChangeField `json:"updated,omitempty"`
	// A list of removed fields
	Removed []ChangeField `json:"removed,omitempty"`
	// A full object. Only available during a "create" event
	Object json.RawMessage `json:"object,omitempty"`
}

type ChangeField struct {
	Key      string `json:"key"`
	OldValue any    `json:"old_value"`
	NewValue any    `json:"new_value"`
}