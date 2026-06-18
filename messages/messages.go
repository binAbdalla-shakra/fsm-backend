package messages

// Standard success notification messages
const (
	SuccessTicketCreated     = "Ticket reported successfully and queued for automated dispatch."
	SuccessTicketDispatched  = "Technician matched and dispatched to your location."
	SuccessTicketStarted     = "Work has started on the ticket."
	SuccessTicketCompleted   = "Ticket completed successfully. OTP verified."
	SuccessTechStatusUpdated = "Technician status and online status updated successfully."
	SuccessCustomerCreated   = "Customer profile created successfully."
	SuccessTechCreated       = "Technician profile registered successfully."
	SuccessTicketCancelled   = "Ticket cancelled successfully."
)

// Standard error messages
const (
	ErrInvalidPayload      = "Invalid request format or payload validation failed."
	ErrTechnicianNotFound  = "Requested technician does not exist."
	ErrCustomerNotFound    = "Requested customer does not exist."
	ErrTicketNotFound      = "Requested ticket does not exist."
	ErrDispatchFailed      = "No available technicians match the required skill and location criteria."
	ErrInvalidOTP          = "Invalid verification OTP. Please try again."
	ErrUnauthorizedTransit = "Invalid transition state requested for the ticket."

	ErrCustAccountConflict       = "Customer account number already exists"
	ErrCustIDRequired            = "Customer ID is required"
	ErrTechSkillRequired         = "At least one skill must be assigned to the technician profile"
	ErrTechIDRequired            = "Technician ID is required"
	ErrTechStatusInvalid         = "Invalid status value provided"
	ErrUnauthorizedTech          = "Unauthorized technician action on this ticket"
	ErrImageCompression          = "Image compression failed"
	ErrTicketReportDispatchFail  = "Ticket reported, but automated dispatch failed to find a matching online technician."
	ErrTechIDRequiredQuery       = "technician_id is required"
	ErrOTPRequired               = "OTP code is required"
)

// Log and audit trail notes
const (
	LogSearchInitiated = "Initiated automated proximity search"
	LogDispatchFailed  = "Proximity dispatch failed: No matching online technicians found"
	LogTechAssigned    = "Automated spatial dispatch assigned technician"
	LogTechArrived     = "Technician arrived and marked ticket in-progress"
	LogTechCompleted   = "Technician completed ticket, OTP verified"
)
