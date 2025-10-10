package services

import (
	"context"
	"strings"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"main-webbase/internal/models"
	repo "main-webbase/internal/repository"
)


func CanManagePolicy(userPolicies []models.Policy, target *models.Policy) error {
	partTrimmed := strings.Trim(target.OrgPrefix, "/")
	segs := []string{}

	if partTrimmed != "" {
		segs = strings.Split(partTrimmed, "/")
	}

	ancestors := []string{"/"}
	for i := 1; i <= len(segs); i++ {
		ancestor := "/" + strings.Join(segs[:i], "/")
		ancestors = append(ancestors, ancestor)
	}
	// ได้ ancestors = ["/", "/faculty", "/faculty/eng", "/faculty/eng/smo"] จาก "/faculty/eng/smo"
	
	for _, policy := range userPolicies {
		if !policy.Enabled {
			continue
		}
		hasAction := false
		for _, act := range policy.Actions {
			if act == "membership:assign" {
				hasAction = true
				break
			}
		}
		if !hasAction {
			continue
		}

		if policy.Scope == "exact" && policy.OrgPrefix == target.OrgPrefix {
			return nil
		}

		if policy.Scope == "subtree" {
			for _, anc := range ancestors {
				if policy.OrgPrefix == anc {
					return nil
				}
			}
		}
	}
	return errors.New("no permission to manage this policy")
}


func CanManageEvent(ctx context.Context, userPolicies []models.Policy, EventID string) error {
	eventID, err := bson.ObjectIDFromHex(EventID)
	if err != nil {
		return errors.New("invalid EventID")
	}

	event, err := repo.GetEventByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	org_node, err := repo.GetOrgByID(ctx, event.NodeID)
	if err != nil {
		return fmt.Errorf("org_node not found: %w", err)
	}
	
	for _, policy := range userPolicies {
		if !policy.Enabled {
			continue
		}
		hasAction := false
		for _, act := range policy.Actions {
			if act == "event:create" {
				hasAction = true
				break
			}
		}
		if !hasAction {
			continue
		}

		if policy.Scope == "exact" && policy.OrgPrefix == org_node.OrgPath {
			return nil
		}

		if policy.Scope == "subtree" {
			for _, anc := range org_node.Ancestors {
				if policy.OrgPrefix == anc {
					return nil
				}
			}
		}
	}

	return errors.New("no permission to manage this Event")
}
