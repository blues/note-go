// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package note

// The full URN of a resource for permissioning purposes is:
// app:xxx-xxxx-xxxx-xxxx:dev:xxxxxxxxxxx:file:xxxx

// ACResourceApp is the app (project) resource, which is the appUID that always begins with this string
const ACResourceApp = "app:"

// ACResourceApps is the resource for all apps
const ACResourceApps = "app:*"

// ACResourceDevice is the device resource, which is the deviceUID that always begins with this string
const ACResourceDevice = "dev:"

// ACResourceDevices is the resource for all devices
const ACResourceDevices = "dev:*"

// ACResourceNotefile is the notefile resource and its note-level actions,
// which is the notefileID prefixed with this string
const ACResourceNotefile = "file:"

// ACResourceNotefiles is the resource for all notefiles and all meta-notefile-level actions
const ACResourceNotefiles = "file:*"

// ACResourceAccount is an account resource, which is the accountUID that always begins with this string
const ACResourceAccount = "account:"

// ACResourceAccounts is the resource for all accounts and all meta-account-level actions
const ACResourceAccounts = "account:*"

// ACResourceRoute is an route resource, which is the routeUID that always begins with this string
const ACResourceRoute = "route:"

// ACResourceRoutes is the resource for all routes and all meta-route-level actions
const ACResourceRoutes = "route:*"

// ACResourceNotecardFirmwares is the resource for all notecard firmware
const ACResourceNotecardFirmwares = "notecard:*"

// ACResourceUserFirmwares is the resource for all user firmware
const ACResourceUserFirmwares = "firmware:*"

// ACResourceSep is the separator for building compound resource names
const ACResourceSep = ":"

// Entire vocabulary of allowed actions on resources

// ACActionRead (golint)
const ACActionRead = "read"

// ACActionUpdate (golint)
const ACActionUpdate = "update"

// ACActionCreate (golint)
const ACActionCreate = "create"

// ACActionDelete (golint)
const ACActionDelete = "delete"

// ACActionMonitor (golint)
const ACActionMonitor = "monitor"

// Ways of combining actions into one

// ACActionAnd ensures that all of these actions are allowed
const ACActionAnd = "&"

// ACActionOr ensures that any of these actions are allowed
const ACActionOr = "|"

// The entire palette of valid actions, as a comma-separated list

// ACValidActionsApp are actions allowed on apps
const ACValidActionsApp = "app:create,app:read,app:update,app:delete,app:monitor"

// ACValidActionsDev are actions allowed on devices
const ACValidActionsDev = "dev:read,dev:update,dev:delete,dev:monitor"

// ACValidActionsFile are actions allowed on notefiles
const ACValidActionsFile = "file:create,file:read,file:update,file:delete"

// ACValidActionsAccount are actions allowed on accounts
const ACValidActionsAccount = "account:create,account:read,account:update,account:delete"

// ACValidActionsRoute are actions allowed on routes
const ACValidActionsRoute = "route:create,route:read,route:update,route:delete"

// ACValidActionsNotecard are actions allowed on notecard firmware
const ACValidActionsNotecard = "notecard:create,notecard:read,notecard:update,notecard:delete"

// ACValidActionsFirmware are actions allowed on user firmware
const ACValidActionsFirmware = "firmware:create,firmware:read,firmware:update,firmware:delete"
