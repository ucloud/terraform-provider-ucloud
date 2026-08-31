package uads

const (
	statusPending     = "pending"
	statusInitialized = "initialized"
)

const (
	uadsAllowedDomainStatusAdding   = 1
	uadsAllowedDomainStatusSuccess  = 2
	uadsAllowedDomainStatusDeleting = 3
	uadsAllowedDomainStatusFailure  = 4
	uadsAllowedDomainStatusDeleted  = 5
)

var uadsAllowedDomainStatusCvt = newIntConverter(map[int]string{
	uadsAllowedDomainStatusAdding:   "Adding",
	uadsAllowedDomainStatusSuccess:  "Success",
	uadsAllowedDomainStatusDeleting: "Deleting",
	uadsAllowedDomainStatusFailure:  "Failure",
	uadsAllowedDomainStatusDeleted:  "Deleted",
})

const (
	uadsBGPServiceIPStatusSuccess      = "Success"
	uadsBGPServiceIPStatusPending      = "Pending"
	uadsBGPServiceFwdRuleStatusSuccess = "Success"
	uadsBGPServiceFwdRuleStatusPending = "Pending"
)
