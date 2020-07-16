// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/blues/note-go/note"
	"github.com/blues/note-go/notecard"
)

// Default number of days for old contacts
const contactDefaultPurgeDays = 14
const contactPathSep = "|"

// Get the user's contact
func contactGet() (contact note.MessageContact, err error) {
	req := notecard.Request{Req: notecard.ReqNoteGet}
	req.NotefileID = note.ContactStore
	req.NoteID = note.ContactOwnerNoteID
	rsp, err := card.TransactionRequest(req)
	if err != nil && (note.ErrorContains(err, note.ErrNotefileNoExist) || note.ErrorContains(err, note.ErrNoteNoExist)) {
		err = nil
	}
	if err != nil {
		return
	}
	note.BodyToObject(rsp.Body, &contact)
	return
}

// Add a device to your contact, by path
func contactAddDevice(path string) (err error) {

	// Make sure we HAVE a contact
	var contact note.MessageContact
	contact, err = contactSet("", "", 0)
	if err != nil {
		return
	}

	// Extract the components of the path
	var ma note.MessageAddress
	ma.DeviceUID, ma.ProductUID, ma.Hub, err = contactPathComponents(contact, path)
	if err != nil {
		return
	}

	// Update the contact if the device is already there
	found := false
	for i, c := range contact.Addresses {
		if c.DeviceUID == ma.DeviceUID {
			contact.Addresses[i].DeviceUID = ma.DeviceUID
			contact.Addresses[i].ProductUID = ma.ProductUID
			contact.Addresses[i].Hub = ma.Hub
			found = true
			break
		}
	}

	if !found {
		contact.Addresses = append(contact.Addresses, ma)
	}

	// Convert back to a map
	var body map[string]interface{}
	body, err = note.ObjectToBody(contact)
	if err != nil {
		return
	}

	// Save the contact
	req := notecard.Request{Req: notecard.ReqNoteUpdate}
	req.NotefileID = note.ContactStore
	req.NoteID = note.ContactOwnerNoteID
	req.Body = &body
	_, err = card.TransactionRequest(req)
	if err != nil {
		return
	}

	// Done
	return

}

// Remove a device from your contact, by device UID
func contactRemoveDevice(deviceUID string) (err error) {

	// Make sure we HAVE a contact
	var contact note.MessageContact
	contact, err = contactSet("", "", 0)
	if err != nil {
		return
	}

	// Update the contact if the device is already there
	found := false
	for i, c := range contact.Addresses {
		if c.DeviceUID == deviceUID {
			contact.Addresses = append(contact.Addresses[0:i], contact.Addresses[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("deviceUID %s not found in contact", deviceUID)
	}

	// Convert back to a map
	var body map[string]interface{}
	body, err = note.ObjectToBody(contact)
	if err != nil {
		return
	}

	// Save the contact
	req := notecard.Request{Req: notecard.ReqNoteUpdate}
	req.NotefileID = note.ContactStore
	req.NoteID = note.ContactOwnerNoteID
	req.Body = &body
	_, err = card.TransactionRequest(req)
	if err != nil {
		return
	}

	// Done
	return

}

// Set the user's name and mail in their contact
func contactSet(name string, email string, daysInactive int) (contact note.MessageContact, err error) {

	// Read the existing note for the user's name
	contact, err = contactGet()

	// Name
	if name == "-" {
		contact.Name = ""
	} else if name != "" && name != "?" {
		contact.Name = name
	}

	// Email
	if email == "-" {
		contact.Email = ""
	} else if email != "" && email != "?" {
		contact.Email = email
	}

	// Get the current device's info
	var rsp notecard.Request
	rsp, err = card.TransactionRequest(notecard.Request{Req: notecard.ReqHubGet})
	if err != nil {
		return
	}
	address := note.MessageAddress{}
	address.DeviceUID = rsp.DeviceUID
	address.DeviceSN = rsp.SN
	address.ProductUID = rsp.ProductUID
	address.Active = uint32(time.Now().Unix())

	// Clean up the host by extracting just the hostname assuming that the format
	// of the expanded host string is scheme:host[:port]|scheme:host[:port]|...
	address.Hub = rsp.Host
	components := strings.Split(rsp.Host, "|")
	if len(components) >= 1 && strings.Contains(components[0], ":") {
		components = strings.Split(components[0], ":")
		address.Hub = components[1]
	}

	// Iterate over the contact, looking for this device.  If found, update it, else add it.
	var newAddresses []note.MessageAddress
	newAddresses = append(newAddresses, address)
	expiration := uint32(time.Now().Unix()) - uint32(daysInactive*60*60*24)
	for _, addr := range contact.Addresses {
		if daysInactive > 0 && addr.Active < expiration {
			continue
		}
		if addr.DeviceUID == "" {
			continue
		}
		if addr.DeviceUID == address.DeviceUID {
			continue
		}
		newAddresses = append(newAddresses, addr)
	}
	contact.Addresses = newAddresses

	// Convert back to a map
	var body map[string]interface{}
	body, err = note.ObjectToBody(contact)
	if err != nil {
		return
	}

	// Save the contact
	req := notecard.Request{Req: notecard.ReqNoteUpdate}
	req.NotefileID = note.ContactStore
	req.NoteID = note.ContactOwnerNoteID
	req.Body = &body
	_, err = card.TransactionRequest(req)
	if err != nil {
		return
	}

	// Done
	return

}

// See if a contact is ours, even if the devices have changed, focusing on devices
func contactNewer(cbase note.MessageContact, cnew note.MessageContact) (same bool, changed bool, whatChanged string, best note.MessageContact) {

	// Begin to generate a new contact by starting with the base but by regenerating addresses
	best = cbase
	best.Addresses = []note.MessageAddress{}
	whatChanged = "no change"

	// If we're filling in the name or email, take it
	if best.Name == "" && cnew.Name != "" {
		changed = true
		whatChanged = "name changed"
		best.Name = cnew.Name
	}
	if best.Email == "" && cnew.Email != "" {
		changed = true
		whatChanged = "name changed"
		best.Email = cnew.Email
	}

	// Generate the best output contact
	for _, addrNew := range cnew.Addresses {
		foundNewDeviceInBase := false
		for _, addrBase := range cbase.Addresses {
			if addrBase.DeviceUID == addrNew.DeviceUID {
				same = true
				foundNewDeviceInBase = true
				if addrNew.Active > addrBase.Active {
					changed = true
					whatChanged = "more recent device address"
					best.Addresses = append(best.Addresses, addrNew)
				} else {
					best.Addresses = append(best.Addresses, addrBase)
				}
			}
		}
		if !foundNewDeviceInBase {
			changed = true
			whatChanged = "addition of a new device address"
			best.Addresses = append(best.Addresses, addrNew)
		}
	}

	// If not found, make the output contact the new contact
	if !same {
		whatChanged = "new"
		best = cnew
	}

	// Done
	return

}

// Get the simplest possible name for the contact
func contactName(ourContact note.MessageContact, c note.MessageContact) (name string) {
	if c.Name != "" && c.Email != "" {
		return c.Name + " <" + c.Email + ">"
	}
	if c.Name != "" {
		return c.Name
	}
	if c.Email != "" {
		return c.Email
	}
	if len(c.Addresses) < 1 || len(ourContact.Addresses) < 1 {
		return "(unknown)"
	}
	oca := ourContact.Addresses[0]
	ca := c.Addresses[0]
	if oca.Hub == ca.Hub && oca.ProductUID == ca.ProductUID {
		return ca.DeviceUID
	}
	if oca.Hub == ca.Hub {
		return ca.ProductUID + contactPathSep + ca.DeviceUID
	}
	return ca.Hub + contactPathSep + ca.ProductUID + contactPathSep + ca.DeviceUID

}

// Update the contact list from a list of contacts
func contactListUpdate(updates []note.MessageContact) (err error) {

	// Get all notes in the contact database
	var notes []note.Info
	notes, err = getAllNotes(note.ContactStore)
	if err != nil {
		return
	}

	// Convert them into an array of contacts
	var clist []note.MessageContact
	var nlist []string
	for _, n := range notes {
		if n.Body == nil {
			continue
		}
		var c note.MessageContact
		err = note.BodyToObject(n.Body, &c)
		if err != nil {
			continue
		}
		clist = append(clist, c)
		nlist = append(nlist, n.NoteID)
	}

	// Iterate over them, updating or adding as appropriate
	for _, cnew := range updates {

		// Update or add it as appropriate
		found := false
		for i, c := range clist {
			same, changed, whatChanged, best := contactNewer(c, cnew)
			if !same {
				continue
			}
			found = true
			if changed {
				var body map[string]interface{}
				body, err = note.ObjectToBody(best)
				if err != nil {
					return
				}
				req := notecard.Request{Req: notecard.ReqNoteUpdate}
				req.NotefileID = note.ContactStore
				req.NoteID = nlist[i]
				req.Body = &body
				_, err = card.TransactionRequest(req)
				if err != nil {
					return
				}
				fmt.Printf("%s contact updated (%s)\n", contactName(note.MessageContact{}, best), whatChanged)
			}

		}

		// If new contact not found in contact store, add it
		if !found {
			var body map[string]interface{}
			body, err = note.ObjectToBody(cnew)
			if err != nil {
				return
			}
			req := notecard.Request{Req: notecard.ReqNoteAdd}
			req.NotefileID = note.ContactStore
			req.Body = &body
			_, err = card.TransactionRequest(req)
			if err != nil {
				return
			}
			fmt.Printf("%s contact added\n", contactName(note.MessageContact{}, cnew))
		}

	}

	// Done
	return
}

// Show contacts
func contactShowOthers() (err error) {

	// Get all notes in the contact database
	var notes []note.Info
	notes, err = getAllNotes(note.ContactStore)
	if err != nil {
		return
	}

	// Show them
	for _, n := range notes {
		if n.NoteID != note.ContactOwnerNoteID {
			contactJSON, _ := note.JSONMarshalIndent(n, "", "  ")
			fmt.Printf("%s\n", contactJSON)
		}
	}

	// Done
	return
}

// Remove a specific contact
func contactRemove(who string) (err error) {

	// Get all notes in the contact database
	var notes []note.Info
	notes, err = getAllNotes(note.ContactStore)
	if err != nil {
		return
	}

	// Scan them looking for the specified contact
	noteID := ""
	for _, n := range notes {
		if n.NoteID == note.ContactOwnerNoteID {
			continue
		}
		var c note.MessageContact
		note.BodyToObject(n.Body, &c)
		if c.Email == who || c.Name == who {
			noteID = n.NoteID
		}
		for _, addr := range c.Addresses {
			if addr.DeviceUID == who || addr.DeviceSN == who {
				noteID = n.NoteID
				break
			}
		}
		if noteID != "" {
			break
		}
	}

	// If no note to delete, err
	if noteID == "" {
		err = fmt.Errorf("contact not found")
		return
	}

	// Delete the note
	req := notecard.Request{Req: notecard.ReqNoteDelete}
	req.NotefileID = note.ContactStore
	req.NoteID = noteID
	_, err = card.TransactionRequest(req)
	if err != nil {
		return
	}

	// Done
	return
}

// Get all notes from the specified notefile
func getAllNotes(notefileID string) (notes []note.Info, err error) {

	// Go into a loop requesting all notes, from the start of the file
	tracker := "all"
	for {

		// Get the next batch of notes
		req := notecard.Request{Req: notecard.ReqNoteChanges}
		req.NotefileID = notefileID
		req.Max = 25
		req.TrackerID = tracker
		req.Start = (len(notes) == 0)
		rsp := notecard.Request{}
		rsp, err = card.TransactionRequest(req)
		if err != nil {
			break
		}
		if rsp.Notes == nil || len(*rsp.Notes) == 0 {
			break
		}

		// Append to results
		for k, v := range *rsp.Notes {
			v.NoteID = k
			notes = append(notes, v)
		}

	}

	// Delete the tracker; this is extremely important else all deleted stubs will be retained indefinitely!
	req := notecard.Request{Req: notecard.ReqNoteChanges}
	req.NotefileID = notefileID
	req.TrackerID = tracker
	req.Stop = true
	card.TransactionRequest(req)

	// Done; return error.
	return

}

// Expand the components of a contact path
func contactPathComponents(ourContact note.MessageContact, who string) (DeviceUID string, ProductUID string, Hub string, err error) {
	components := strings.Split(who, contactPathSep)
	switch len(components) {
	case 1: // imei:1324234234234
		Hub = ourContact.Addresses[0].Hub
		ProductUID = ourContact.Addresses[0].ProductUID
		DeviceUID = components[0]
	case 2: // product:com.blues.test!imei:1324234234234
		Hub = ourContact.Addresses[0].Hub
		ProductUID = components[0]
		DeviceUID = components[1]
	case 3: // a.notefile.net!product:com.blues.test!imei:1324234234234
		Hub = components[0]
		ProductUID = components[1]
		DeviceUID = components[2]
	default:
		err = fmt.Errorf("%s is an invalid address format (notehubHOST!productUID!deviceUID)", who)
		return
	}

	return
}

// Expand a shortcut list of "to" addresses to a list of recipients, using the supplied contact to default
// the route for those contacts where a physical address expansion is required
func contactExpand(ourContact note.MessageContact, to string) (toContacts []note.MessageContact, err error) {

	// Exit if our contact is not usable for address expansion
	if len(ourContact.Addresses) == 0 {
		err = fmt.Errorf("unable to find the DeviceUID of our own notecard")
		return
	}

	// Next, read in the entire address book so we can do a lookup.  Note that we clear "active" because
	// we're going to overload it as a flag indicating that we encountered it already, so we don't send dups
	var notes []note.Info
	var contacts []note.MessageContact
	notes, err = getAllNotes(note.ContactStore)
	if err != nil {
		return
	}
	for _, n := range notes {
		var c note.MessageContact
		note.BodyToObject(n.Body, &c)
		contacts = append(contacts, c)
	}

	// First, expand the list into all possible recipients using all legal separators
	to = strings.Replace(to, ",", ";", -1)
	toList := strings.Split(to, ";")

	// Now, generate the output contacts based on that list
	for _, who := range toList {

		// If it's a note #, extract the contacts from that message
		if strings.HasPrefix(who, "#") {
			var mc []note.MessageContact
			mc, err = messageContacts(ourContact, who)
			if err != nil {
				return
			}
			for _, c := range mc {
				toContacts = append(toContacts, c)
			}
			continue
		}

		// Look at the entire contact list
		found := false
		for _, c := range contacts {
			if contactMatch(c, who) {
				toContacts = append(toContacts, c)
				found = true
				break
			}
		}

		// Continue if found
		if found {
			continue
		}

		// See if we can make it into a path
		if strings.Contains(who, ":") {

			// Break down the components
			var ma note.MessageAddress
			ma.DeviceUID, ma.ProductUID, ma.Hub, err = contactPathComponents(ourContact, who)
			if err != nil {
				return
			}

			var mc note.MessageContact
			mc.Addresses = append(mc.Addresses, ma)
			toContacts = append(toContacts, mc)
			continue

		}

		err = fmt.Errorf("%s is not found in your contact list", who)
		return

	}

	// Done
	return
}

// See if a contact matches a string pattern
func contactMatch(c note.MessageContact, who string) bool {

	// Check the basic fields
	if c.Name == who {
		return true
	}
	if c.Email == who {
		return true
	}

	// Check for device UID being specified
	for _, addr := range c.Addresses {
		if addr.DeviceUID == who {
			return true
		}
		if addr.DeviceSN == who {
			return true
		}
	}

	// Not found
	return false

}
