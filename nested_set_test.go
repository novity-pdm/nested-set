package nestedset

import (
	"context"
	"github.com/google/uuid"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/clause"
)

func TestReloadData(t *testing.T) {
	reloadCategories()
}

func TestNewNodeItem(t *testing.T) {
	sourceNodeID := uuid.New()

	source := Category{
		ID:            sourceNodeID,
		ParentID:      uuid.NullUUID{Valid: true, UUID: uuid.New()},
		Depth:         2,
		Rgt:           12,
		Lft:           32,
		UserType:      "User",
		UserID:        1000,
		ChildrenCount: 10,
	}
	tx, node, err := parseNode(context.Background(), db, source)
	assert.NoError(t, err)
	assert.Equal(t, source.ID, node.ID)
	assert.Equal(t, source.ParentID, node.ParentID)
	assert.Equal(t, source.Depth, node.Depth)
	assert.Equal(t, source.Lft, node.Lft)
	assert.Equal(t, source.Rgt, node.Rgt)
	assert.Equal(t, source.ChildrenCount, node.ChildrenCount)
	assert.Equal(t, "categories", node.TableName)
	stmt := tx.Statement
	stmt.Build(clause.Where{}.Name())
	assert.Equal(t, "WHERE user_id = $1 AND user_type = $2", stmt.SQL.String())

	tx, node, err = parseNode(context.Background(), db, &source)
	assert.NoError(t, err)
	assert.Equal(t, source.ID, node.ID)
	assert.Equal(t, source.ParentID, node.ParentID)
	assert.Equal(t, source.Depth, node.Depth)
	assert.Equal(t, source.Lft, node.Lft)
	assert.Equal(t, source.Rgt, node.Rgt)
	assert.Equal(t, source.ChildrenCount, node.ChildrenCount)
	assert.Equal(t, "categories", node.TableName)
	stmt = tx.Statement
	stmt.Build(clause.Where{}.Name())
	assert.Equal(t, "WHERE user_id = $1 AND user_type = $2", stmt.SQL.String())

	dbNames := node.DbNames
	assert.Equal(t, "id", dbNames["id"])
	assert.Equal(t, "parent_id", dbNames["parent_id"])
	assert.Equal(t, "depth", dbNames["depth"])
	assert.Equal(t, "rgt", dbNames["rgt"])
	assert.Equal(t, "lft", dbNames["lft"])
	assert.Equal(t, "children_count", dbNames["children_count"])

	specialItemID := uuid.New()

	// Test for difference column names
	specialItem := SpecialItem{
		ItemID:     specialItemID,
		Pid:        uuid.NullUUID{Valid: true, UUID: uuid.New()},
		Depth1:     2,
		Right:      10,
		Left:       1,
		NodesCount: 8,
	}
	tx, node, err = parseNode(context.Background(), db, specialItem)
	assert.NoError(t, err)
	assert.Equal(t, specialItem.ItemID, node.ID)
	assert.Equal(t, specialItem.Pid, node.ParentID)
	assert.Equal(t, specialItem.Depth1, node.Depth)
	assert.Equal(t, specialItem.Right, node.Rgt)
	assert.Equal(t, specialItem.Left, node.Lft)
	assert.Equal(t, specialItem.NodesCount, node.ChildrenCount)
	assert.Equal(t, "special_items", node.TableName)

	stmt = tx.Statement
	stmt.Build(clause.Where{}.Name())
	assert.Equal(t, "", stmt.SQL.String())

	dbNames = node.DbNames
	assert.Equal(t, "item_id", dbNames["id"])
	assert.Equal(t, "pid", dbNames["parent_id"])
	assert.Equal(t, "depth1", dbNames["depth"])
	assert.Equal(t, "right", dbNames["rgt"])
	assert.Equal(t, "left", dbNames["lft"])
	assert.Equal(t, "nodes_count", dbNames["children_count"])

	// formatSQL test
	assert.Equal(t, "item_id = ? AND left > right AND pid = ?, nodes_count = 1, depth1 = 0", formatSQL(":id = ? AND :lft > :rgt AND :parent_id = ?, :children_count = 1, :depth = 0", node))
}

func TestCreateSource(t *testing.T) {
	initData()

	c1 := Category{Title: "c1s"}
	Create(context.Background(), db, &c1, nil)
	assert.Equal(t, c1.Lft, 1)
	assert.Equal(t, c1.Rgt, 2)
	assert.Equal(t, c1.Depth, 0)

	cp := Category{Title: "cps"}
	Create(context.Background(), db, &cp, nil)
	assert.Equal(t, cp.Lft, 3)
	assert.Equal(t, cp.Rgt, 4)

	c2 := Category{Title: "c2s", UserType: "ux"}
	Create(context.Background(), db, &c2, nil)
	assert.Equal(t, c2.Lft, 1)
	assert.Equal(t, c2.Rgt, 2)

	c3 := Category{Title: "c3s", UserType: "ux"}
	Create(context.Background(), db, &c3, nil)
	assert.Equal(t, c3.Lft, 3)
	assert.Equal(t, c3.Rgt, 4)

	c4 := Category{Title: "c4s", UserType: "ux"}
	Create(context.Background(), db, &c4, &c2)
	assert.Equal(t, c4.Lft, 2)
	assert.Equal(t, c4.Rgt, 3)
	assert.Equal(t, c4.Depth, 1)

	// after insert a new node into c2
	db.Find(&c3)
	db.Find(&c2)
	assert.Equal(t, c3.Lft, 5)
	assert.Equal(t, c3.Rgt, 6)
	assert.Equal(t, c2.ChildrenCount, 1)
}

func TestDeleteSource(t *testing.T) {
	initData()

	c1 := Category{Title: "c1s"}
	Create(context.Background(), db, &c1, nil)

	cp := Category{Title: "cp"}
	Create(context.Background(), db, &cp, c1)

	c2 := Category{Title: "c2s"}
	Create(context.Background(), db, &c2, nil)

	db.First(&c1)
	Delete(context.Background(), db, &c1)

	db.Model(&c2).First(&c2)
	assert.Equal(t, c2.Lft, 1)
	assert.Equal(t, c2.Rgt, 2)
}

func TestMoveToRight(t *testing.T) {
	// case 1
	initData()

	if err := MoveTo(context.Background(), db, dresses, jackets, MoveDirectionRight); err != nil {
		t.Fatal(err)
	}

	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 2, uuid.NullUUID{Valid: false})
	assertNodeEqual(t, mens, 2, 15, 1, 1, uuid.NullUUID{UUID: clothing.ID, Valid: true})
	assertNodeEqual(t, suits, 3, 14, 2, 3, uuid.NullUUID{UUID: mens.ID, Valid: true})
	assertNodeEqual(t, slacks, 4, 5, 3, 0, uuid.NullUUID{UUID: suits.ID, Valid: true})
	assertNodeEqual(t, jackets, 6, 7, 3, 0, uuid.NullUUID{UUID: suits.ID, Valid: true})
	assertNodeEqual(t, dresses, 8, 13, 3, 2, uuid.NullUUID{UUID: suits.ID, Valid: true})
	assertNodeEqual(t, eveningGowns, 9, 10, 4, 0, uuid.NullUUID{UUID: dresses.ID, Valid: true})
	assertNodeEqual(t, sunDresses, 11, 12, 4, 0, uuid.NullUUID{UUID: dresses.ID, Valid: true})
	assertNodeEqual(t, womens, 16, 21, 1, 2, uuid.NullUUID{UUID: clothing.ID, Valid: true})
	assertNodeEqual(t, skirts, 17, 18, 2, 0, uuid.NullUUID{UUID: womens.ID, Valid: true})
	assertNodeEqual(t, blouses, 19, 20, 2, 0, uuid.NullUUID{UUID: womens.ID, Valid: true})

	// case 2
	initData()
	if err := MoveTo(context.Background(), db, suits, blouses, MoveDirectionRight); err != nil {
		t.Fatal(err)
	}

	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 2, uuid.NullUUID{Valid: false})
	assertNodeEqual(t, mens, 2, 3, 1, 0, uuid.NullUUID{UUID: clothing.ID, Valid: true})
	assertNodeEqual(t, womens, 4, 21, 1, 4, uuid.NullUUID{UUID: clothing.ID, Valid: true})
	assertNodeEqual(t, dresses, 5, 10, 2, 2, uuid.NullUUID{UUID: womens.ID, Valid: true})
	assertNodeEqual(t, eveningGowns, 6, 7, 3, 0, uuid.NullUUID{UUID: dresses.ID, Valid: true})
	assertNodeEqual(t, sunDresses, 8, 9, 3, 0, uuid.NullUUID{UUID: dresses.ID, Valid: true})
	assertNodeEqual(t, skirts, 11, 12, 2, 0, uuid.NullUUID{UUID: womens.ID, Valid: true})
	assertNodeEqual(t, blouses, 13, 14, 2, 0, uuid.NullUUID{UUID: womens.ID, Valid: true})
	assertNodeEqual(t, suits, 15, 20, 2, 2, uuid.NullUUID{UUID: womens.ID, Valid: true})
	assertNodeEqual(t, slacks, 16, 17, 3, 0, uuid.NullUUID{UUID: suits.ID, Valid: true})
	assertNodeEqual(t, jackets, 18, 19, 3, 0, uuid.NullUUID{UUID: suits.ID, Valid: true})
}

func TestMoveToLeft(t *testing.T) {
	// case 1
	initData()
	if err := MoveTo(context.Background(), db, dresses, jackets, MoveDirectionLeft); err != nil {
		t.Fatal(err)
	}
	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 2, uuid.NullUUID{Valid: false})
	assertNodeEqual(t, mens, 2, 15, 1, 1, uuid.NullUUID{UUID: clothing.ID, Valid: true})
	assertNodeEqual(t, suits, 3, 14, 2, 3, uuid.NullUUID{UUID: mens.ID, Valid: true})
	assertNodeEqual(t, slacks, 4, 5, 3, 0, uuid.NullUUID{UUID: suits.ID, Valid: true})
	assertNodeEqual(t, dresses, 6, 11, 3, 2, uuid.NullUUID{UUID: suits.ID, Valid: true})
	assertNodeEqual(t, eveningGowns, 7, 8, 4, 0, uuid.NullUUID{UUID: dresses.ID, Valid: true})
	assertNodeEqual(t, sunDresses, 9, 10, 4, 0, uuid.NullUUID{UUID: dresses.ID, Valid: true})
	assertNodeEqual(t, jackets, 12, 13, 3, 0, uuid.NullUUID{UUID: suits.ID, Valid: true})
	assertNodeEqual(t, womens, 16, 21, 1, 2, uuid.NullUUID{UUID: clothing.ID, Valid: true})
	assertNodeEqual(t, skirts, 17, 18, 2, 0, uuid.NullUUID{UUID: womens.ID, Valid: true})
	assertNodeEqual(t, blouses, 19, 20, 2, 0, uuid.NullUUID{UUID: womens.ID, Valid: true})

	// case 2
	initData()
	if err := MoveTo(context.Background(), db, suits, blouses, MoveDirectionLeft); err != nil {
		t.Fatal(err)
	}
	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 2, uuid.NullUUID{Valid: false})
	assertNodeEqual(t, mens, 2, 3, 1, 0, uuid.NullUUID{UUID: clothing.ID, Valid: true})
	assertNodeEqual(t, womens, 4, 21, 1, 4, uuid.NullUUID{UUID: clothing.ID, Valid: true})
	assertNodeEqual(t, dresses, 5, 10, 2, 2, uuid.NullUUID{UUID: womens.ID, Valid: true})
	assertNodeEqual(t, eveningGowns, 6, 7, 3, 0, uuid.NullUUID{UUID: dresses.ID, Valid: true})
	assertNodeEqual(t, sunDresses, 8, 9, 3, 0, uuid.NullUUID{UUID: dresses.ID, Valid: true})
	assertNodeEqual(t, skirts, 11, 12, 2, 0, uuid.NullUUID{UUID: womens.ID, Valid: true})
	assertNodeEqual(t, suits, 13, 18, 2, 2, uuid.NullUUID{UUID: womens.ID, Valid: true})
	assertNodeEqual(t, slacks, 14, 15, 3, 0, uuid.NullUUID{UUID: suits.ID, Valid: true})
	assertNodeEqual(t, jackets, 16, 17, 3, 0, uuid.NullUUID{UUID: suits.ID, Valid: true})
	assertNodeEqual(t, blouses, 19, 20, 2, 0, uuid.NullUUID{UUID: womens.ID, Valid: true})
}

func TestMoveToInner(t *testing.T) {
	// case 1
	initData()
	if err := MoveTo(context.Background(), db, mens, blouses, MoveDirectionInner); err != nil {
		t.Fatal(err)
	}
	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 1, uuid.NullUUID{Valid: false})
	assertNodeEqual(t, womens, 2, 21, 1, 3, uuid.NullUUID{UUID: clothing.ID, Valid: true})
	assertNodeEqual(t, dresses, 3, 8, 2, 2, uuid.NullUUID{UUID: womens.ID, Valid: true})
	assertNodeEqual(t, eveningGowns, 4, 5, 3, 0, uuid.NullUUID{UUID: dresses.ID, Valid: true})
	assertNodeEqual(t, sunDresses, 6, 7, 3, 0, uuid.NullUUID{UUID: dresses.ID, Valid: true})
	assertNodeEqual(t, skirts, 9, 10, 2, 0, uuid.NullUUID{UUID: womens.ID, Valid: true})
	assertNodeEqual(t, blouses, 11, 20, 2, 1, uuid.NullUUID{UUID: womens.ID, Valid: true})
	assertNodeEqual(t, mens, 12, 19, 3, 1, uuid.NullUUID{UUID: blouses.ID, Valid: true})
	assertNodeEqual(t, suits, 13, 18, 4, 2, uuid.NullUUID{UUID: mens.ID, Valid: true})
	assertNodeEqual(t, slacks, 14, 15, 5, 0, uuid.NullUUID{UUID: suits.ID, Valid: true})
	assertNodeEqual(t, jackets, 16, 17, 5, 0, uuid.NullUUID{UUID: suits.ID, Valid: true})

	// case 2
	initData()
	if err := MoveTo(context.Background(), db, skirts, slacks, MoveDirectionInner); err != nil {
		t.Fatal(err)
	}
	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 2, uuid.NullUUID{Valid: false})
	assertNodeEqual(t, mens, 2, 11, 1, 1, uuid.NullUUID{UUID: clothing.ID, Valid: true})
	assertNodeEqual(t, suits, 3, 10, 2, 2, uuid.NullUUID{UUID: mens.ID, Valid: true})
	assertNodeEqual(t, slacks, 4, 7, 3, 1, uuid.NullUUID{UUID: suits.ID, Valid: true})
	assertNodeEqual(t, skirts, 5, 6, 4, 0, uuid.NullUUID{UUID: slacks.ID, Valid: true})
	assertNodeEqual(t, jackets, 8, 9, 3, 0, uuid.NullUUID{UUID: suits.ID, Valid: true})
	assertNodeEqual(t, womens, 12, 21, 1, 2, uuid.NullUUID{UUID: clothing.ID, Valid: true})
	assertNodeEqual(t, dresses, 13, 18, 2, 2, uuid.NullUUID{UUID: womens.ID, Valid: true})
	assertNodeEqual(t, eveningGowns, 14, 15, 3, 0, uuid.NullUUID{UUID: dresses.ID, Valid: true})
	assertNodeEqual(t, sunDresses, 16, 17, 3, 0, uuid.NullUUID{UUID: dresses.ID, Valid: true})
	assertNodeEqual(t, blouses, 19, 20, 2, 0, uuid.NullUUID{UUID: womens.ID, Valid: true})
}

func TestMoveIsInvalid(t *testing.T) {
	initData()
	err := MoveTo(context.Background(), db, womens, dresses, MoveDirectionInner)
	assert.NotEmpty(t, err)
	reloadCategories()
	assertNodeEqual(t, womens, 10, 21, 1, 3, uuid.NullUUID{UUID: clothing.ID, Valid: true})

	err = MoveTo(context.Background(), db, womens, dresses, MoveDirectionLeft)
	assert.NotEmpty(t, err)
	reloadCategories()
	assertNodeEqual(t, womens, 10, 21, 1, 3, uuid.NullUUID{UUID: clothing.ID, Valid: true})

	err = MoveTo(context.Background(), db, womens, dresses, MoveDirectionRight)
	assert.NotEmpty(t, err)
	reloadCategories()
	assertNodeEqual(t, womens, 10, 21, 1, 3, uuid.NullUUID{UUID: clothing.ID, Valid: true})
}

func assertNodeEqual(t *testing.T, target Category, left, right, depth, childrenCount int, parentID uuid.NullUUID) {
	assert.Equal(t, target.Lft, left)
	assert.Equal(t, target.Rgt, right)
	assert.Equal(t, target.Depth, depth)
	assert.Equal(t, target.ChildrenCount, childrenCount)
	assert.Equal(t, target.ParentID, parentID)
}
