package identity

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var _ json.Marshaler = &SystemAndUserAssignedList{}

type SystemAndUserAssignedList struct {
	Type        Type     `json:"type"`
	PrincipalId string   `json:"principalId"`
	TenantId    string   `json:"tenantId"`
	IdentityIds []string `json:"userAssignedIdentities"`
}

func (s *SystemAndUserAssignedList) MarshalJSON() ([]byte, error) {
	// we use a custom marshal function here since we can only send the Type / UserAssignedIdentities field
	identityType := TypeNone
	userAssignedIdentityIds := []string{}

	if s != nil {
		if s.Type == TypeSystemAssigned {
			identityType = TypeSystemAssigned
		}
		if s.Type == TypeSystemAssignedUserAssigned {
			identityType = TypeSystemAssignedUserAssigned
		}
		if s.Type == TypeUserAssigned {
			identityType = TypeUserAssigned
		}

		if identityType != TypeNone {
			userAssignedIdentityIds = s.IdentityIds
		}
	}

	out := map[string]interface{}{
		"type":                   string(identityType),
		"userAssignedIdentities": userAssignedIdentityIds,
	}
	return json.Marshal(out)
}

// ExpandSystemAndUserAssignedList expands the schema input into a SystemAndUserAssignedList struct
func ExpandSystemAndUserAssignedList(input []interface{}) (*SystemAndUserAssignedList, error) {
	identityType := TypeNone
	identityIds := make([]string, 0)

	if len(input) > 0 {
		raw := input[0].(map[string]interface{})
		typeRaw := raw["type"].(string)
		if typeRaw == string(TypeSystemAssigned) {
			identityType = TypeSystemAssigned
		}
		if typeRaw == string(TypeSystemAssignedUserAssigned) {
			identityType = TypeSystemAssignedUserAssigned
		}
		if typeRaw == string(TypeUserAssigned) {
			identityType = TypeUserAssigned
		}

		identityIdsRaw := raw["identity_ids"].(*schema.Set).List()
		for _, v := range identityIdsRaw {
			identityIds = append(identityIds, v.(string))
		}
	}

	if len(identityIds) > 0 && (identityType != TypeSystemAssignedUserAssigned && identityType != TypeUserAssigned) {
		return nil, fmt.Errorf("`identity_ids` can only be specified when `type` is set to %q or %q", string(TypeSystemAssignedUserAssigned), string(TypeUserAssigned))
	}

	return &SystemAndUserAssignedList{
		Type:        identityType,
		IdentityIds: identityIds,
	}, nil
}

// FlattenSystemAndUserAssignedList turns a SystemAndUserAssignedList into a []interface{}
func FlattenSystemAndUserAssignedList(input *SystemAndUserAssignedList) (*[]interface{}, error) {
	if input == nil {
		return &[]interface{}{}, nil
	}

	input.Type = normalizeType(input.Type)
	if input.Type != TypeSystemAssigned && input.Type != TypeSystemAssignedUserAssigned && input.Type != TypeUserAssigned {
		return &[]interface{}{}, nil
	}

	identityIds := make([]string, 0)
	for _, raw := range input.IdentityIds {
		id, err := commonids.ParseUserAssignedIdentityIDInsensitively(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing %q as a User Assigned Identity ID: %+v", raw, err)
		}
		identityIds = append(identityIds, id.ID())
	}

	return &[]interface{}{
		map[string]interface{}{
			"type":         string(input.Type),
			"identity_ids": identityIds,
			"principal_id": input.PrincipalId,
			"tenant_id":    input.TenantId,
		},
	}, nil
}
