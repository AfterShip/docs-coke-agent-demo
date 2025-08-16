package types

import (
	"time"

	"eino/pkg/langfuse/api/resources/utils/pagination/types"
)

// Organization represents an organization in Langfuse
type Organization struct {
	// Unique identifier for the organization
	ID string `json:"id"`

	// Name of the organization
	Name string `json:"name"`

	// Display name for the organization
	DisplayName *string `json:"displayName,omitempty"`

	// Description of the organization
	Description *string `json:"description,omitempty"`

	// Organization settings
	Settings *OrganizationSettings `json:"settings,omitempty"`

	// Plan information
	Plan *OrganizationPlan `json:"plan,omitempty"`

	// Whether the organization is active
	IsActive bool `json:"isActive"`

	// Metadata associated with the organization
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Timestamp when the organization was created
	CreatedAt time.Time `json:"createdAt"`

	// Timestamp when the organization was last updated
	UpdatedAt time.Time `json:"updatedAt"`

	// User who created the organization
	CreatedBy *string `json:"createdBy,omitempty"`

	// Organization statistics
	Stats *OrganizationStats `json:"stats,omitempty"`
}

// OrganizationSettings represents settings for an organization
type OrganizationSettings struct {
	// Default data retention period in days
	DefaultDataRetentionDays *int `json:"defaultDataRetentionDays,omitempty"`

	// Whether to allow project creation
	AllowProjectCreation bool `json:"allowProjectCreation"`

	// Maximum number of projects allowed
	MaxProjects *int `json:"maxProjects,omitempty"`

	// Whether to enforce SSO for all members
	RequireSSO bool `json:"requireSSO"`

	// Allowed domains for auto-joining
	AllowedDomains []string `json:"allowedDomains,omitempty"`

	// Billing settings
	BillingSettings *BillingSettings `json:"billingSettings,omitempty"`

	// Security settings
	SecuritySettings *SecuritySettings `json:"securitySettings,omitempty"`

	// Feature flags for the organization
	FeatureFlags map[string]bool `json:"featureFlags,omitempty"`

	// Custom configuration
	CustomConfig map[string]interface{} `json:"customConfig,omitempty"`
}

// BillingSettings represents billing settings for an organization
type BillingSettings struct {
	// Billing email
	BillingEmail string `json:"billingEmail"`

	// Payment method information
	PaymentMethod *PaymentMethod `json:"paymentMethod,omitempty"`

	// Billing address
	BillingAddress *Address `json:"billingAddress,omitempty"`

	// Tax information
	TaxInfo *TaxInfo `json:"taxInfo,omitempty"`

	// Whether billing is enabled
	BillingEnabled bool `json:"billingEnabled"`

	// Invoice settings
	InvoiceSettings *InvoiceSettings `json:"invoiceSettings,omitempty"`
}

// PaymentMethod represents payment method information
type PaymentMethod struct {
	Type        string `json:"type"` // "card", "bank_transfer", "invoice"
	LastFour    string `json:"lastFour"`
	Brand       string `json:"brand"`
	ExpiryMonth int    `json:"expiryMonth"`
	ExpiryYear  int    `json:"expiryYear"`
	IsDefault   bool   `json:"isDefault"`
}

// Address represents a billing address
type Address struct {
	Line1      string  `json:"line1"`
	Line2      *string `json:"line2,omitempty"`
	City       string  `json:"city"`
	State      *string `json:"state,omitempty"`
	PostalCode string  `json:"postalCode"`
	Country    string  `json:"country"`
}

// TaxInfo represents tax information
type TaxInfo struct {
	TaxID     *string `json:"taxId,omitempty"`
	TaxType   *string `json:"taxType,omitempty"` // "vat", "gst", "sales_tax"
	TaxExempt bool    `json:"taxExempt"`
	TaxRegion *string `json:"taxRegion,omitempty"`
}

// InvoiceSettings represents invoice settings
type InvoiceSettings struct {
	AutoSend            bool     `json:"autoSend"`
	PaymentTerms        int      `json:"paymentTerms"` // Days
	AdditionalEmails    []string `json:"additionalEmails,omitempty"`
	InvoicePrefix       *string  `json:"invoicePrefix,omitempty"`
	PurchaseOrderNumber *string  `json:"purchaseOrderNumber,omitempty"`
}

// SecuritySettings represents security settings
type SecuritySettings struct {
	// Whether to require 2FA for all members
	Require2FA bool `json:"require2fa"`

	// Session timeout in minutes
	SessionTimeout *int `json:"sessionTimeout,omitempty"`

	// IP whitelist
	IPWhitelist []string `json:"ipWhitelist,omitempty"`

	// Whether to enable audit logging
	AuditLogging bool `json:"auditLogging"`

	// Data encryption settings
	EncryptionSettings *EncryptionSettings `json:"encryptionSettings,omitempty"`
}

// EncryptionSettings represents encryption settings
type EncryptionSettings struct {
	AtRestEncryption    bool `json:"atRestEncryption"`
	InTransitEncryption bool `json:"inTransitEncryption"`
	KeyRotationEnabled  bool `json:"keyRotationEnabled"`
	CustomerManagedKeys bool `json:"customerManagedKeys"`
}

// OrganizationPlan represents plan information
type OrganizationPlan struct {
	// Plan name
	Name string `json:"name"`

	// Plan type
	Type PlanType `json:"type"`

	// Billing interval
	BillingInterval BillingInterval `json:"billingInterval"`

	// Plan limits
	Limits *PlanLimits `json:"limits,omitempty"`

	// Current usage
	Usage *PlanUsage `json:"usage,omitempty"`

	// Plan cost
	Cost *PlanCost `json:"cost,omitempty"`

	// Plan start date
	StartDate time.Time `json:"startDate"`

	// Plan end date (for temporary plans)
	EndDate *time.Time `json:"endDate,omitempty"`

	// Whether the plan auto-renews
	AutoRenew bool `json:"autoRenew"`

	// Trial information
	Trial *TrialInfo `json:"trial,omitempty"`
}

// PlanType represents plan types
type PlanType string

const (
	PlanTypeFree       PlanType = "free"
	PlanTypeStarter    PlanType = "starter"
	PlanTypePro        PlanType = "pro"
	PlanTypeTeam       PlanType = "team"
	PlanTypeEnterprise PlanType = "enterprise"
)

// BillingInterval represents billing intervals
type BillingInterval string

const (
	BillingIntervalMonthly BillingInterval = "monthly"
	BillingIntervalYearly  BillingInterval = "yearly"
)

// PlanLimits represents limits for a plan
type PlanLimits struct {
	MaxProjects        *int    `json:"maxProjects,omitempty"`
	MaxTracesPerMonth  *int    `json:"maxTracesPerMonth,omitempty"`
	MaxUsers           *int    `json:"maxUsers,omitempty"`
	MaxDatasetItems    *int    `json:"maxDatasetItems,omitempty"`
	MaxPrompts         *int    `json:"maxPrompts,omitempty"`
	DataRetentionDays  *int    `json:"dataRetentionDays,omitempty"`
	APIRateLimit       *int    `json:"apiRateLimit,omitempty"` // requests per minute
	StorageLimit       *int64  `json:"storageLimit,omitempty"` // bytes
	SupportLevel       *string `json:"supportLevel,omitempty"`
	CustomIntegrations bool    `json:"customIntegrations"`
	AdvancedAnalytics  bool    `json:"advancedAnalytics"`
	SSO                bool    `json:"sso"`
	AuditLogs          bool    `json:"auditLogs"`
}

// PlanUsage represents current usage against plan limits
type PlanUsage struct {
	Projects             int   `json:"projects"`
	TracesThisMonth      int   `json:"tracesThisMonth"`
	Users                int   `json:"users"`
	DatasetItems         int   `json:"datasetItems"`
	Prompts              int   `json:"prompts"`
	StorageUsed          int64 `json:"storageUsed"` // bytes
	APIRequestsThisMonth int   `json:"apiRequestsThisMonth"`
}

// PlanCost represents plan cost information
type PlanCost struct {
	BaseCost      float64 `json:"baseCost"`
	UsageCost     float64 `json:"usageCost"`
	TotalCost     float64 `json:"totalCost"`
	Currency      string  `json:"currency"`
	BillingPeriod string  `json:"billingPeriod"`
}

// TrialInfo represents trial information
type TrialInfo struct {
	IsTrialActive  bool       `json:"isTrialActive"`
	TrialStartDate *time.Time `json:"trialStartDate,omitempty"`
	TrialEndDate   *time.Time `json:"trialEndDate,omitempty"`
	TrialDaysLeft  *int       `json:"trialDaysLeft,omitempty"`
}

// OrganizationStats represents statistics for an organization
type OrganizationStats struct {
	TotalProjects   int        `json:"totalProjects"`
	ActiveProjects  int        `json:"activeProjects"`
	TotalUsers      int        `json:"totalUsers"`
	ActiveUsers     int        `json:"activeUsers"`
	TotalTraces     int        `json:"totalTraces"`
	TracesThisMonth int        `json:"tracesThisMonth"`
	TotalCost       *float64   `json:"totalCost,omitempty"`
	CostThisMonth   *float64   `json:"costThisMonth,omitempty"`
	Currency        *string    `json:"currency,omitempty"`
	LastActivity    *time.Time `json:"lastActivity,omitempty"`
	StorageUsed     *int64     `json:"storageUsed,omitempty"`
	MonthlyGrowth   *float64   `json:"monthlyGrowth,omitempty"` // percentage
}

// OrganizationMember represents a member of an organization
type OrganizationMember struct {
	ID           string                   `json:"id"`
	UserID       string                   `json:"userId"`
	Email        string                   `json:"email"`
	Name         *string                  `json:"name,omitempty"`
	Role         OrganizationRole         `json:"role"`
	Status       OrganizationMemberStatus `json:"status"`
	JoinedAt     time.Time                `json:"joinedAt"`
	LastActive   *time.Time               `json:"lastActive,omitempty"`
	Permissions  []Permission             `json:"permissions,omitempty"`
	ProjectRoles []ProjectRole            `json:"projectRoles,omitempty"`
}

// OrganizationRole represents roles within an organization
type OrganizationRole string

const (
	OrganizationRoleOwner  OrganizationRole = "owner"
	OrganizationRoleAdmin  OrganizationRole = "admin"
	OrganizationRoleMember OrganizationRole = "member"
	OrganizationRoleViewer OrganizationRole = "viewer"
)

// OrganizationMemberStatus represents member status
type OrganizationMemberStatus string

const (
	OrganizationMemberStatusActive   OrganizationMemberStatus = "active"
	OrganizationMemberStatusInvited  OrganizationMemberStatus = "invited"
	OrganizationMemberStatusInactive OrganizationMemberStatus = "inactive"
)

// Permission represents a specific permission
type Permission struct {
	Resource string   `json:"resource"` // "projects", "datasets", "prompts", etc.
	Actions  []string `json:"actions"`  // "read", "write", "delete", "admin"
}

// ProjectRole represents a role within a specific project
type ProjectRole struct {
	ProjectID string `json:"projectId"`
	Role      string `json:"role"`
}

// GetOrganizationsRequest represents a request to list organizations
type GetOrganizationsRequest struct {
	Page          *int       `json:"page,omitempty"`
	Limit         *int       `json:"limit,omitempty"`
	Name          *string    `json:"name,omitempty"`
	IsActive      *bool      `json:"isActive,omitempty"`
	PlanType      *PlanType  `json:"planType,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
	IncludeStats  *bool      `json:"includeStats,omitempty"`
}

// GetOrganizationsResponse represents the response from listing organizations
type GetOrganizationsResponse struct {
	Data []Organization     `json:"data"`
	Meta types.MetaResponse `json:"meta"`
}

// CreateOrganizationRequest represents a request to create an organization
type CreateOrganizationRequest struct {
	Name        string                 `json:"name"`
	DisplayName *string                `json:"displayName,omitempty"`
	Description *string                `json:"description,omitempty"`
	Settings    *OrganizationSettings  `json:"settings,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CreateOrganizationResponse represents the response from creating an organization
type CreateOrganizationResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	DisplayName *string                `json:"displayName,omitempty"`
	Description *string                `json:"description,omitempty"`
	Settings    *OrganizationSettings  `json:"settings,omitempty"`
	Plan        *OrganizationPlan      `json:"plan,omitempty"`
	IsActive    bool                   `json:"isActive"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	CreatedBy   *string                `json:"createdBy,omitempty"`
}

// UpdateOrganizationRequest represents a request to update an organization
type UpdateOrganizationRequest struct {
	Name        *string                `json:"name,omitempty"`
	DisplayName *string                `json:"displayName,omitempty"`
	Description *string                `json:"description,omitempty"`
	IsActive    *bool                  `json:"isActive,omitempty"`
	Settings    *OrganizationSettings  `json:"settings,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// InviteMemberRequest represents a request to invite a member
type InviteMemberRequest struct {
	Email       string           `json:"email"`
	Role        OrganizationRole `json:"role"`
	Permissions []Permission     `json:"permissions,omitempty"`
	Message     *string          `json:"message,omitempty"`
}

// InviteMemberResponse represents the response from inviting a member
type InviteMemberResponse struct {
	ID          string                   `json:"id"`
	Email       string                   `json:"email"`
	Role        OrganizationRole         `json:"role"`
	Status      OrganizationMemberStatus `json:"status"`
	InvitedAt   time.Time                `json:"invitedAt"`
	InviteToken string                   `json:"inviteToken"`
	ExpiresAt   time.Time                `json:"expiresAt"`
}

// UpdateMemberRequest represents a request to update a member
type UpdateMemberRequest struct {
	Role        *OrganizationRole `json:"role,omitempty"`
	Permissions []Permission      `json:"permissions,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Validate validates the GetOrganizationsRequest
func (req *GetOrganizationsRequest) Validate() error {
	if req.Limit != nil && (*req.Limit < 1 || *req.Limit > 1000) {
		return &ValidationError{Field: "limit", Message: "limit must be between 1 and 1000"}
	}

	if req.Page != nil && *req.Page < 1 {
		return &ValidationError{Field: "page", Message: "page must be greater than 0"}
	}

	if req.FromTimestamp != nil && req.ToTimestamp != nil && req.FromTimestamp.After(*req.ToTimestamp) {
		return &ValidationError{Field: "timestamps", Message: "fromTimestamp cannot be after toTimestamp"}
	}

	if req.PlanType != nil {
		validPlanTypes := map[PlanType]bool{
			PlanTypeFree:       true,
			PlanTypeStarter:    true,
			PlanTypePro:        true,
			PlanTypeTeam:       true,
			PlanTypeEnterprise: true,
		}
		if !validPlanTypes[*req.PlanType] {
			return &ValidationError{Field: "planType", Message: "invalid plan type"}
		}
	}

	return nil
}

// Validate validates the CreateOrganizationRequest
func (req *CreateOrganizationRequest) Validate() error {
	if req.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}

	if len(req.Name) > 255 {
		return &ValidationError{Field: "name", Message: "name must be 255 characters or less"}
	}

	if req.DisplayName != nil && len(*req.DisplayName) > 255 {
		return &ValidationError{Field: "displayName", Message: "displayName must be 255 characters or less"}
	}

	if req.Description != nil && len(*req.Description) > 2000 {
		return &ValidationError{Field: "description", Message: "description must be 2000 characters or less"}
	}

	if req.Settings != nil {
		if err := req.Settings.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the InviteMemberRequest
func (req *InviteMemberRequest) Validate() error {
	if req.Email == "" {
		return &ValidationError{Field: "email", Message: "email is required"}
	}

	// Basic email validation
	if len(req.Email) > 255 {
		return &ValidationError{Field: "email", Message: "email must be 255 characters or less"}
	}

	validRoles := map[OrganizationRole]bool{
		OrganizationRoleOwner:  true,
		OrganizationRoleAdmin:  true,
		OrganizationRoleMember: true,
		OrganizationRoleViewer: true,
	}
	if !validRoles[req.Role] {
		return &ValidationError{Field: "role", Message: "invalid role"}
	}

	return nil
}

// Validate validates the OrganizationSettings
func (settings *OrganizationSettings) Validate() error {
	if settings.DefaultDataRetentionDays != nil && *settings.DefaultDataRetentionDays < 1 {
		return &ValidationError{Field: "settings.defaultDataRetentionDays", Message: "defaultDataRetentionDays must be greater than 0"}
	}

	if settings.MaxProjects != nil && *settings.MaxProjects < 1 {
		return &ValidationError{Field: "settings.maxProjects", Message: "maxProjects must be greater than 0"}
	}

	if settings.BillingSettings != nil {
		if err := settings.BillingSettings.Validate(); err != nil {
			return err
		}
	}

	if settings.SecuritySettings != nil {
		if err := settings.SecuritySettings.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the BillingSettings
func (billing *BillingSettings) Validate() error {
	if billing.BillingEmail == "" {
		return &ValidationError{Field: "billingSettings.billingEmail", Message: "billingEmail is required"}
	}

	if len(billing.BillingEmail) > 255 {
		return &ValidationError{Field: "billingSettings.billingEmail", Message: "billingEmail must be 255 characters or less"}
	}

	if billing.BillingAddress != nil {
		if err := billing.BillingAddress.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the Address
func (addr *Address) Validate() error {
	if addr.Line1 == "" {
		return &ValidationError{Field: "address.line1", Message: "line1 is required"}
	}

	if addr.City == "" {
		return &ValidationError{Field: "address.city", Message: "city is required"}
	}

	if addr.Country == "" {
		return &ValidationError{Field: "address.country", Message: "country is required"}
	}

	if addr.PostalCode == "" {
		return &ValidationError{Field: "address.postalCode", Message: "postalCode is required"}
	}

	return nil
}

// Validate validates the SecuritySettings
func (security *SecuritySettings) Validate() error {
	if security.SessionTimeout != nil && *security.SessionTimeout < 5 {
		return &ValidationError{Field: "securitySettings.sessionTimeout", Message: "sessionTimeout must be at least 5 minutes"}
	}

	return nil
}

// Active returns true if the organization is active
func (o *Organization) Active() bool {
	return o.IsActive
}

// HasFeature returns true if the organization has a specific feature enabled
func (o *Organization) HasFeature(feature string) bool {
	if o.Settings == nil || o.Settings.FeatureFlags == nil {
		return false
	}

	enabled, exists := o.Settings.FeatureFlags[feature]
	return exists && enabled
}

// IsOnTrial returns true if the organization is on a trial
func (o *Organization) IsOnTrial() bool {
	return o.Plan != nil && o.Plan.Trial != nil && o.Plan.Trial.IsTrialActive
}

// GetTrialDaysLeft returns the number of trial days left
func (o *Organization) GetTrialDaysLeft() int {
	if !o.IsOnTrial() || o.Plan.Trial.TrialDaysLeft == nil {
		return 0
	}
	return *o.Plan.Trial.TrialDaysLeft
}

// HasReachedLimit checks if the organization has reached a specific limit
func (o *Organization) HasReachedLimit(limitType string) bool {
	if o.Plan == nil || o.Plan.Limits == nil || o.Plan.Usage == nil {
		return false
	}

	limits := o.Plan.Limits
	usage := o.Plan.Usage

	switch limitType {
	case "projects":
		return limits.MaxProjects != nil && usage.Projects >= *limits.MaxProjects
	case "users":
		return limits.MaxUsers != nil && usage.Users >= *limits.MaxUsers
	case "traces":
		return limits.MaxTracesPerMonth != nil && usage.TracesThisMonth >= *limits.MaxTracesPerMonth
	case "storage":
		return limits.StorageLimit != nil && usage.StorageUsed >= *limits.StorageLimit
	default:
		return false
	}
}

// NewCreateOrganizationRequest creates a new organization creation request
func NewCreateOrganizationRequest(name string) *CreateOrganizationRequest {
	return &CreateOrganizationRequest{
		Name: name,
	}
}

// NewInviteMemberRequest creates a new member invitation request
func NewInviteMemberRequest(email string, role OrganizationRole) *InviteMemberRequest {
	return &InviteMemberRequest{
		Email: email,
		Role:  role,
	}
}
