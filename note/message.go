// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package note

// MessageAddress is the network routing information for a message
type MessageAddress struct {
	Hub        string `json:"hub,omitempty"`
	ProductUID string `json:"product,omitempty"`
	DeviceUID  string `json:"device,omitempty"`
	DeviceSN   string `json:"sn,omitempty"`
	Active     uint32 `json:"active,omitempty"`
}

// MessageContact is the entity sending a message, who may have multiple devices/addresses
type MessageContact struct {
	Name      string           `json:"name,omitempty"`
	Email     string           `json:"email,omitempty"`
	StoreTags []string         `json:"stags,omitempty"`
	Addresses []MessageAddress `json:"addresses,omitempty"`
}

// Message is the core message data structure.  Note that when stored in a map or a note,
// the UID is not present but rather is the map key or noteID.
type Message struct {
	UID         string                  `json:"id,omitempty"`
	Sent        uint32                  `json:"sent,omitempty"`
	Received    uint32                  `json:"received,omitempty"`
	From        MessageContact          `json:"from,omitempty"`
	To          []MessageContact        `json:"to,omitempty"`
	Tags        []string                `json:"tags,omitempty"`
	StoreTags   []string                `json:"stags,omitempty"`
	ContentType string                  `json:"type,omitempty"`
	Content     string                  `json:"content,omitempty"`
	Body        *map[string]interface{} `json:"body,omitempty"`
}

// MessageOutbox is the place from which messages are sent
const MessageOutbox = "messages.qo"

// MessageInbox is the place into which messages are received
const MessageInbox = "messages.qi"

// MessageStore is the place where the user retains messages
const MessageStore = "messages.db"

// ContactStore is the place where the user retains contact info
const ContactStore = "contacts.db"

// MessageContentASCII is just simple ASCII text
const MessageContentASCII = ""

// MessageTagImportant indicates that the sender feels that this is an important message
const MessageTagImportant = "important"

// MessageTagUrgent indicates that the sender feels that this is an urgent message
const MessageTagUrgent = "urgent"

// MessageSTagSent indicates that this was a sent message
const MessageSTagSent = "sent"

// MessageSTagReceived indicates that this was a received message
const MessageSTagReceived = "received"

// ContactOwnerNoteID indicates that this is my contact
const ContactOwnerNoteID = "owner"
