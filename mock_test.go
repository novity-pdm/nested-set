package nestedset

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bluele/factory-go/factory"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func databaseURL() string {
	databaseURL := os.Getenv("DATABASE_URL")
	if len(databaseURL) == 0 {
		databaseURL = "postgres://postgres:postgres@localhost:5432/nested-set-test?sslmode=disable"
	}
	return databaseURL
}

var (
	memoryDB, _ = sql.Open("postgres", databaseURL())
	db          = newMock(memoryDB)

	ctx = context.TODO()
)
var clothing, mens, suits, slacks, jackets, womens, dresses, skirts, blouses, eveningGowns, sunDresses Category

type Category struct {
	ID            uuid.UUID `gorm:"primary_key; unique; type:uuid; column:id; default:uuid_generate_v4()" nestedset:"id"`
	Title         string
	UserID        int           `nestedset:"scope"`
	UserType      string        `nestedset:"scope"`
	ParentID      uuid.NullUUID `nestedset:"parent_id"`
	Rgt           int           `nestedset:"rgt"`
	Lft           int           `nestedset:"lft"`
	Depth         int           `nestedset:"depth"`
	ChildrenCount int           `nestedset:"children_count"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type SpecialItem struct {
	ItemID     uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" nestedset:"id"`
	Title      string
	Pid        uuid.NullUUID `nestedset:"parent_id"`
	Right      int           `nestedset:"rgt"`
	Left       int           `nestedset:"lft"`
	Depth1     int           `nestedset:"depth"`
	NodesCount int           `nestedset:"children_count"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func findNode(query *gorm.DB, id uuid.UUID) (category Category, err error) {
	err = query.Where("id=?", id).Find(&category).Error
	return
}

var CategoryFactory = factory.NewFactory(&Category{
	Title:    "Clothing",
	ParentID: uuid.NullUUID{Valid: false},
	UserType: "User",
	UserID:   999,
	Rgt:      1,
	Lft:      2,
	Depth:    0,
}).
	OnCreate(func(args factory.Args) error {
		return db.Create(args.Instance()).Error
	})

func newMock(_db *sql.DB) *gorm.DB {
	dir, _ := os.Getwd()
	os.MkdirAll(filepath.Join(dir, "./log"), 0777)
	logFile, err := os.OpenFile(filepath.Join(dir, "./log/test.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{
		DSN:                  databaseURL(),
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.New(log.New(logFile, "\n", 1), logger.Config{
			LogLevel: logger.Info,
		}),
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	return gormDB
}

func initData() {
	db.Exec("DROP TABLE IF EXISTS categories")
	db.Exec("DROP TABLE IF EXISTS special_items")
	err := db.AutoMigrate(
		&Category{},
		&SpecialItem{},
	)
	if err != nil {
		panic(err)
	}
	buildTestData()
}

func buildTestData() {
	clothing = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":         "Clothing",
		"Lft":           1,
		"Rgt":           22,
		"Depth":         0,
		"ChildrenCount": 2,
	}).(*Category)

	// Create a category in other group
	_ = CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":         "Clothing",
		"Lft":           1,
		"UserID":        98,
		"Rgt":           22,
		"Depth":         0,
		"ChildrenCount": 2,
	}).(*Category)

	mens = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":         "Men's",
		"ParentID":      uuid.NullUUID{Valid: true, UUID: clothing.ID},
		"Lft":           2,
		"Rgt":           9,
		"Depth":         1,
		"ChildrenCount": 1,
	}).(*Category)
	suits = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":         "Suits",
		"ParentID":      uuid.NullUUID{Valid: true, UUID: mens.ID},
		"Lft":           3,
		"Rgt":           8,
		"Depth":         2,
		"ChildrenCount": 2,
	}).(*Category)
	slacks = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Slacks",
		"ParentID": uuid.NullUUID{Valid: true, UUID: suits.ID},
		"Lft":      4,
		"Rgt":      5,
		"Depth":    3,
	}).(*Category)
	jackets = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Jackets",
		"ParentID": uuid.NullUUID{Valid: true, UUID: suits.ID},
		"Lft":      6,
		"Rgt":      7,
		"Depth":    3,
	}).(*Category)
	womens = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":         "Women's",
		"ParentID":      uuid.NullUUID{Valid: true, UUID: clothing.ID},
		"Lft":           10,
		"Rgt":           21,
		"Depth":         1,
		"ChildrenCount": 3,
	}).(*Category)
	dresses = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":         "Dresses",
		"ParentID":      uuid.NullUUID{Valid: true, UUID: womens.ID},
		"Lft":           11,
		"Rgt":           16,
		"Depth":         2,
		"ChildrenCount": 2,
	}).(*Category)
	eveningGowns = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Evening Gowns",
		"ParentID": uuid.NullUUID{Valid: true, UUID: dresses.ID},
		"Lft":      12,
		"Rgt":      13,
		"Depth":    3,
	}).(*Category)
	sunDresses = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Sun Dresses",
		"ParentID": uuid.NullUUID{Valid: true, UUID: dresses.ID},
		"Lft":      14,
		"Rgt":      15,
		"Depth":    3,
	}).(*Category)
	skirts = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Skirts",
		"ParentID": uuid.NullUUID{Valid: true, UUID: womens.ID},
		"Lft":      17,
		"Rgt":      18,
		"Depth":    2,
	}).(*Category)
	blouses = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Blouses",
		"ParentID": uuid.NullUUID{Valid: true, UUID: womens.ID},
		"Lft":      19,
		"Rgt":      20,
		"Depth":    2,
	}).(*Category)
}

func reloadCategories() {
	clothing, _ = findNode(db, clothing.ID)
	mens, _ = findNode(db, mens.ID)
	suits, _ = findNode(db, suits.ID)
	slacks, _ = findNode(db, slacks.ID)
	jackets, _ = findNode(db, jackets.ID)
	womens, _ = findNode(db, womens.ID)
	dresses, _ = findNode(db, dresses.ID)
	skirts, _ = findNode(db, skirts.ID)
	blouses, _ = findNode(db, blouses.ID)
	eveningGowns, _ = findNode(db, eveningGowns.ID)
	sunDresses, _ = findNode(db, sunDresses.ID)
}
