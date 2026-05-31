package constants

const (
	TicketStatusReported        = "REPORTED"
	TicketStatusAutoDispatching = "AUTO_DISPATCHING"
	TicketStatusDispatched      = "DISPATCHED"
	TicketStatusInProgress      = "IN_PROGRESS"
	TicketStatusCompleted       = "COMPLETED"
)

const (
	TechStatusOnline  = "ONLINE"
	TechStatusOffline = "OFFLINE"
)

const (
	SkillHVAC       = "hvac"
	SkillPlumbing   = "plumbing"
	SkillElectrical = "electrical"
	SkillAppliance  = "appliance"
	SkillLocksmith  = "locksmith"
)

const (
	RedisGeoKeyPrefix   = "fsm:tech:locations"
	RedisStatusPrefix   = "fsm:tech:status:"
	RedisChannelPrefix  = "fsm:ticket:stream:"
	RedisTrackingBuffer = "fsm:tracking:buffer"
)
