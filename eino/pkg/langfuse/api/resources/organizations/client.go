package organizations

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/go-resty/resty/v2"
	"eino/pkg/langfuse/api/resources/organizations/types"
	commonErrors "eino/pkg/langfuse/api/resources/commons/errors"
)

const (
	organizationsBasePath       = "/api/public/organizations"
	organizationByIDPath        = "/api/public/organizations/%s"
	organizationMembersPath     = "/api/public/organizations/%s/members"
	organizationMemberByIDPath  = "/api/public/organizations/%s/members/%s"
	organizationMemberInvitePath = "/api/public/organizations/%s/members/invite"
)

// Client handles organization-related API operations
type Client struct {
	client *resty.Client
}

// NewClient creates a new organizations client
func NewClient(client *resty.Client) *Client {
	return &Client{
		client: client,
	}
}

// List retrieves a list of organizations
func (c *Client) List(ctx context.Context, req *types.GetOrganizationsRequest) (*types.GetOrganizationsResponse, error) {
	if req == nil {
		req = &types.GetOrganizationsRequest{}
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	// Build query parameters
	queryParams := make(map[string]string)
	
	if req.Page != nil {
		queryParams["page"] = strconv.Itoa(*req.Page)
	}
	
	if req.Limit != nil {
		queryParams["limit"] = strconv.Itoa(*req.Limit)
	}
	
	if req.Name != nil {
		queryParams["name"] = *req.Name
	}
	
	if req.IsActive != nil {
		queryParams["isActive"] = strconv.FormatBool(*req.IsActive)
	}
	
	if req.PlanType != nil {
		queryParams["planType"] = string(*req.PlanType)
	}
	
	if req.FromTimestamp != nil {
		queryParams["fromTimestamp"] = req.FromTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.ToTimestamp != nil {
		queryParams["toTimestamp"] = req.ToTimestamp.Format("2006-01-02T15:04:05.000Z")
	}
	
	if req.IncludeStats != nil {
		queryParams["includeStats"] = strconv.FormatBool(*req.IncludeStats)
	}
	
	response := &types.GetOrganizationsResponse{}
	
	request := c.client.R().
		SetContext(ctx).
		SetResult(response)
	
	// Add query parameters
	for key, value := range queryParams {
		request.SetQueryParam(key, value)
	}
	
	_, err := request.Get(organizationsBasePath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	
	return response, nil
}

// Get retrieves a specific organization by ID
func (c *Client) Get(ctx context.Context, organizationID string) (*types.Organization, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID cannot be empty")
	}
	
	response := &types.Organization{}
	
	path := fmt.Sprintf(organizationByIDPath, url.PathEscape(organizationID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetResult(response).
		Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get organization %s: %w", organizationID, err)
	}
	
	return response, nil
}

// Create creates a new organization
func (c *Client) Create(ctx context.Context, req *types.CreateOrganizationRequest) (*types.CreateOrganizationResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	response := &types.CreateOrganizationResponse{}
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(organizationsBasePath)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}
	
	return response, nil
}

// Update updates an existing organization
func (c *Client) Update(ctx context.Context, organizationID string, req *types.UpdateOrganizationRequest) (*types.Organization, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID cannot be empty")
	}
	
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}
	
	response := &types.Organization{}
	
	path := fmt.Sprintf(organizationByIDPath, url.PathEscape(organizationID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Patch(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update organization %s: %w", organizationID, err)
	}
	
	return response, nil
}

// Delete deletes an organization by ID
func (c *Client) Delete(ctx context.Context, organizationID string) error {
	if organizationID == "" {
		return fmt.Errorf("organization ID cannot be empty")
	}
	
	path := fmt.Sprintf(organizationByIDPath, url.PathEscape(organizationID))
	
	_, err := c.client.R().
		SetContext(ctx).
		Delete(path)
	
	if err != nil {
		return fmt.Errorf("failed to delete organization %s: %w", organizationID, err)
	}
	
	return nil
}

// ListMembers retrieves members of an organization
func (c *Client) ListMembers(ctx context.Context, organizationID string) ([]types.OrganizationMember, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID cannot be empty")
	}
	
	var response []types.OrganizationMember
	
	path := fmt.Sprintf(organizationMembersPath, url.PathEscape(organizationID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetResult(&response).
		Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to list members for organization %s: %w", organizationID, err)
	}
	
	return response, nil
}

// GetMember retrieves a specific member
func (c *Client) GetMember(ctx context.Context, organizationID, memberID string) (*types.OrganizationMember, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID cannot be empty")
	}
	
	if memberID == "" {
		return nil, fmt.Errorf("member ID cannot be empty")
	}
	
	response := &types.OrganizationMember{}
	
	path := fmt.Sprintf(organizationMemberByIDPath, url.PathEscape(organizationID), url.PathEscape(memberID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetResult(response).
		Get(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get member %s for organization %s: %w", memberID, organizationID, err)
	}
	
	return response, nil
}

// InviteMember invites a new member to the organization
func (c *Client) InviteMember(ctx context.Context, organizationID string, req *types.InviteMemberRequest) (*types.InviteMemberResponse, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID cannot be empty")
	}
	
	if req == nil {
		return nil, fmt.Errorf("invite request cannot be nil")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}
	
	response := &types.InviteMemberResponse{}
	
	path := fmt.Sprintf(organizationMemberInvitePath, url.PathEscape(organizationID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Post(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to invite member to organization %s: %w", organizationID, err)
	}
	
	return response, nil
}

// UpdateMember updates an existing member
func (c *Client) UpdateMember(ctx context.Context, organizationID, memberID string, req *types.UpdateMemberRequest) (*types.OrganizationMember, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID cannot be empty")
	}
	
	if memberID == "" {
		return nil, fmt.Errorf("member ID cannot be empty")
	}
	
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}
	
	response := &types.OrganizationMember{}
	
	path := fmt.Sprintf(organizationMemberByIDPath, url.PathEscape(organizationID), url.PathEscape(memberID))
	
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(response).
		Patch(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update member %s in organization %s: %w", memberID, organizationID, err)
	}
	
	return response, nil
}

// RemoveMember removes a member from the organization
func (c *Client) RemoveMember(ctx context.Context, organizationID, memberID string) error {
	if organizationID == "" {
		return fmt.Errorf("organization ID cannot be empty")
	}
	
	if memberID == "" {
		return fmt.Errorf("member ID cannot be empty")
	}
	
	path := fmt.Sprintf(organizationMemberByIDPath, url.PathEscape(organizationID), url.PathEscape(memberID))
	
	_, err := c.client.R().
		SetContext(ctx).
		Delete(path)
	
	if err != nil {
		return fmt.Errorf("failed to remove member %s from organization %s: %w", memberID, organizationID, err)
	}
	
	return nil
}

// Exists checks if an organization exists
func (c *Client) Exists(ctx context.Context, organizationID string) (bool, error) {
	if organizationID == "" {
		return false, fmt.Errorf("organization ID cannot be empty")
	}
	
	_, err := c.Get(ctx, organizationID)
	if err != nil {
		// Check if it's a "not found" error
		if _, ok := err.(*commonErrors.NotFoundError); ok {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}

// ListActive lists all active organizations
func (c *Client) ListActive(ctx context.Context) (*types.GetOrganizationsResponse, error) {
	req := &types.GetOrganizationsRequest{
		IsActive: func() *bool { b := true; return &b }(),
	}
	
	return c.List(ctx, req)
}

// ListByPlan lists organizations by plan type
func (c *Client) ListByPlan(ctx context.Context, planType types.PlanType) (*types.GetOrganizationsResponse, error) {
	req := &types.GetOrganizationsRequest{
		PlanType: &planType,
	}
	
	return c.List(ctx, req)
}

// FindByName finds organizations by name (partial match)
func (c *Client) FindByName(ctx context.Context, name string) (*types.GetOrganizationsResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	
	req := &types.GetOrganizationsRequest{
		Name: &name,
	}
	
	return c.List(ctx, req)
}

// CreateSimple creates a simple organization with just a name
func (c *Client) CreateSimple(ctx context.Context, name string) (*types.CreateOrganizationResponse, error) {
	req := types.NewCreateOrganizationRequest(name)
	return c.Create(ctx, req)
}

// ActivateOrganization activates an organization
func (c *Client) ActivateOrganization(ctx context.Context, organizationID string) (*types.Organization, error) {
	req := &types.UpdateOrganizationRequest{
		IsActive: func() *bool { b := true; return &b }(),
	}
	
	return c.Update(ctx, organizationID, req)
}

// DeactivateOrganization deactivates an organization
func (c *Client) DeactivateOrganization(ctx context.Context, organizationID string) (*types.Organization, error) {
	req := &types.UpdateOrganizationRequest{
		IsActive: func() *bool { b := false; return &b }(),
	}
	
	return c.Update(ctx, organizationID, req)
}

// UpdateSettings updates organization settings
func (c *Client) UpdateSettings(ctx context.Context, organizationID string, settings *types.OrganizationSettings) (*types.Organization, error) {
	req := &types.UpdateOrganizationRequest{
		Settings: settings,
	}
	
	return c.Update(ctx, organizationID, req)
}

// InviteAdmin invites a new admin to the organization
func (c *Client) InviteAdmin(ctx context.Context, organizationID, email string) (*types.InviteMemberResponse, error) {
	req := types.NewInviteMemberRequest(email, types.OrganizationRoleAdmin)
	return c.InviteMember(ctx, organizationID, req)
}

// InviteViewer invites a new viewer to the organization
func (c *Client) InviteViewer(ctx context.Context, organizationID, email string) (*types.InviteMemberResponse, error) {
	req := types.NewInviteMemberRequest(email, types.OrganizationRoleViewer)
	return c.InviteMember(ctx, organizationID, req)
}

// PromoteMember promotes a member to a higher role
func (c *Client) PromoteMember(ctx context.Context, organizationID, memberID string, newRole types.OrganizationRole) (*types.OrganizationMember, error) {
	req := &types.UpdateMemberRequest{
		Role: &newRole,
	}
	
	return c.UpdateMember(ctx, organizationID, memberID, req)
}

// GetActiveMembers retrieves only active members
func (c *Client) GetActiveMembers(ctx context.Context, organizationID string) ([]types.OrganizationMember, error) {
	members, err := c.ListMembers(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	
	var activeMembers []types.OrganizationMember
	for _, member := range members {
		if member.Status == types.OrganizationMemberStatusActive {
			activeMembers = append(activeMembers, member)
		}
	}
	
	return activeMembers, nil
}

// GetMembersByRole retrieves members with a specific role
func (c *Client) GetMembersByRole(ctx context.Context, organizationID string, role types.OrganizationRole) ([]types.OrganizationMember, error) {
	members, err := c.ListMembers(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	
	var roleMembers []types.OrganizationMember
	for _, member := range members {
		if member.Role == role {
			roleMembers = append(roleMembers, member)
		}
	}
	
	return roleMembers, nil
}

// GetOwners retrieves all organization owners
func (c *Client) GetOwners(ctx context.Context, organizationID string) ([]types.OrganizationMember, error) {
	return c.GetMembersByRole(ctx, organizationID, types.OrganizationRoleOwner)
}

// GetAdmins retrieves all organization admins
func (c *Client) GetAdmins(ctx context.Context, organizationID string) ([]types.OrganizationMember, error) {
	return c.GetMembersByRole(ctx, organizationID, types.OrganizationRoleAdmin)
}

// GetOrganizationStats gets statistics for an organization (with stats included)
func (c *Client) GetOrganizationStats(ctx context.Context, organizationID string) (*types.Organization, error) {
	// For a real implementation, this might be a separate endpoint
	// For now, we'll use the regular Get method
	return c.Get(ctx, organizationID)
}

// ListWithStats lists organizations including statistics
func (c *Client) ListWithStats(ctx context.Context) (*types.GetOrganizationsResponse, error) {
	req := &types.GetOrganizationsRequest{
		IncludeStats: func() *bool { b := true; return &b }(),
	}
	
	return c.List(ctx, req)
}

// EnableFeature enables a feature for an organization
func (c *Client) EnableFeature(ctx context.Context, organizationID, feature string) (*types.Organization, error) {
	// Get current organization to preserve existing settings
	org, err := c.Get(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current organization: %w", err)
	}
	
	settings := org.Settings
	if settings == nil {
		settings = &types.OrganizationSettings{}
	}
	
	if settings.FeatureFlags == nil {
		settings.FeatureFlags = make(map[string]bool)
	}
	
	settings.FeatureFlags[feature] = true
	
	return c.UpdateSettings(ctx, organizationID, settings)
}

// DisableFeature disables a feature for an organization
func (c *Client) DisableFeature(ctx context.Context, organizationID, feature string) (*types.Organization, error) {
	// Get current organization to preserve existing settings
	org, err := c.Get(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current organization: %w", err)
	}
	
	settings := org.Settings
	if settings == nil {
		settings = &types.OrganizationSettings{}
	}
	
	if settings.FeatureFlags == nil {
		settings.FeatureFlags = make(map[string]bool)
	}
	
	settings.FeatureFlags[feature] = false
	
	return c.UpdateSettings(ctx, organizationID, settings)
}