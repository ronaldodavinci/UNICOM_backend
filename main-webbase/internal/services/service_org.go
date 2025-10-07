package services

import (
	"context"
	"errors"
	"strings"
	"time"
	"sort"

	"go.mongodb.org/mongo-driver/v2/bson"

	"main-webbase/database"
	"main-webbase/dto"
	"main-webbase/internal/models"
	repo "main-webbase/internal/repository"
)

func AllUserOrg(userID bson.ObjectID) ([]string, error) {
	collection_membership := database.DB.Collection("memberships")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection_membership.Find(ctx, bson.M{"user_id": userID, "active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	organize_set := map[string]struct{}{}

	for cursor.Next(ctx) {
		var user_org models.MembershipDoc
		if err := cursor.Decode(&user_org); err != nil {
			return nil, err
		}
		if user_org.OrgPath == "" {
			continue
		}
		organize_set[user_org.OrgPath] = struct{}{}

		parts := strings.Split(user_org.OrgPath, "/")
		for i := 1; i < len(parts); i++ {
			parent := strings.Join(parts[:i], "/")
			if parent != "" {
				organize_set[parent] = struct{}{}
			}
		}
	}

	// Flatten
	orgs := make([]string, 0, len(organize_set))
	for path := range organize_set {
		orgs = append(orgs, path)
	}
	return orgs, nil
}

func CreateOrgUnit(body dto.OrgUnitDTO, ctx context.Context) (*models.OrgUnitNode, error) {
	now := time.Now().UTC()

	// TrimSpace = เคลียร์ช่องว่างหน้า-หลัง
	body.ParentPath = strings.TrimSpace(body.ParentPath)
	body.Name = strings.TrimSpace(body.Name)
	body.Slug = strings.TrimSpace(body.Slug)
	body.Type = strings.TrimSpace(body.Type)

	if body.Name == "" || body.Slug == "" || body.Type == "" || body.ParentPath == "" {
		return nil, errors.New("name, slug, parent_path and type are required (if parent is root, enter '/' as parent path)")
	}

	slug := strings.ToLower(body.Slug)
	// Check parent path & orgpath correct 
	parent := body.ParentPath
	orgpath := ""
	if parent == "/" {
		orgpath = "/" + slug
	}	else {
		orgpath = parent + "/" + slug
	}

	if parent != "/" {
		parentNode, err := repo.FindByOrgPath(ctx, parent)
		if err != nil {
			return nil, errors.New("error connected to database")
		}
		if parentNode == nil {
			return nil, errors.New("parent path not found")
		}
	}

	duplicateNode, err := repo.FindByOrgPath(ctx, orgpath)
	if err != nil {
		return nil, errors.New("error connected to database")
	}
	if duplicateNode != nil {
		return nil, errors.New("this node already exists")
	}

	// Create Ancestors
	partTrimmed := strings.Trim(parent, "/")
	segs := []string{}
	if partTrimmed != "" {
		segs = strings.Split(partTrimmed, "/")
	}
	ancestors := []string{"/"}
	for i := 1; i <= len(segs); i++ {
		ancestor := "/" + strings.Join(segs[:i], "/")
		ancestors = append(ancestors, ancestor)
	}
	depth := len(segs) + 1

	// Create Shortname
	parentLeaf := ""
	shortname := ""
	if parent != "/" {
		parts := strings.Split(strings.Trim(parent, "/"), "/")
		parentLeaf = strings.ToUpper(parts[len(parts) - 1])
		shortname = strings.TrimSpace(parentLeaf + " • " + strings.ToUpper(body.Slug))

	}  else {
		shortname = strings.ToUpper(body.Slug)
	}

	// Complete
	node := &models.OrgUnitNode{
		ID:		 	bson.NewObjectID(),
		OrgPath: 	orgpath,
		ParentPath: parent,
		Ancestors: 	ancestors,
		Depth: 		depth,
		Name: 		body.Name,
		ShortName:  shortname,
		Slug: 		body.Slug,
		Type: 		body.Type,
		Status: 	"active",
		CreatedAt: 	now,
		UpdatedAt: 	now,
	}

	if err := repo.NodeCreate(ctx, *node); err != nil {
		return nil, errors.New("insert to DB failed")
	}

	return node, nil
}

func BuildOrgTree(ctx context.Context, query dto.OrgUnitTreeQuery) ([]*dto.OrgUnitTree, error) {
	orgUnits, err := repo.FindByPrefix(ctx, query.Start)
	if err != nil {
		return nil, err
	}

	if len(orgUnits) == 0 {
		return []*dto.OrgUnitTree{}, nil
	}

	nodes := map[string]*dto.OrgUnitTree{}
	links := map[string]string{}

	for _, orgunit := range orgUnits {
		node := &dto.OrgUnitTree{
			OrgPath: 	orgunit.OrgPath,
			Type:    	orgunit.Type,
			Label:   	orgunit.Name,
			ShortName:	orgunit.ShortName,
			Children:   []*dto.OrgUnitTree{},
			Sort:       orgunit.Depth,
		}
		nodes[orgunit.OrgPath] = node
		links[orgunit.OrgPath] = orgunit.ParentPath
	}

	var roots []*dto.OrgUnitTree
	for path, node := range nodes {
		parentPath := links[path]
		if parentPath == "" || nodes[parentPath] == nil {
			roots = append(roots, node)
		} else {
			nodes[parentPath].Children = append(nodes[parentPath].Children, node)
		}
	}

	var sortChildren func(list []*dto.OrgUnitTree)
	sortChildren = func(list []*dto.OrgUnitTree) {
		sort.Slice(list, func(i, j int) bool {
			return strings.ToLower(list[i].Label) < strings.ToLower(list[j].Label)
		})
		for _, c := range list {
			sortChildren(c.Children)
		}
	}
	sortChildren(roots)

	if query.Depth > 0 {
		var prune func(list []*dto.OrgUnitTree, level int)
		prune = func(list []*dto.OrgUnitTree, level int) {
			if level >= query.Depth {
				for _, n := range list {
					n.Children = nil
				}
				return
			}	
			for _, n := range list {
				prune(n.Children, level + 1)
			}
		}
		prune(roots, 1)
	}

	return roots, nil
}