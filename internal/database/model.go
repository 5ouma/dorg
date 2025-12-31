package database

import "gorm.io/gorm"

// Types
const (
	_ = iota
	RootType
	FolderRootType
	PageType
)

// LaunchPad is a LaunchPad struct
type LaunchPad struct {
	DB     *gorm.DB
	File   string
	Folder string

	Config Config

	rootPage    int
	confFolders []string
}

// Category CREATE TABLE categories (rowid INTEGER PRIMARY KEY ASC, uti VARCHAR)
type Category struct {
	ID  uint   `gorm:"column:rowid;primary_key"`
	UTI string `gorm:"column:uti"`
}

// Group CREATE TABLE groups (item_id INTEGER PRIMARY KEY, category_id INTEGER, title VARCHAR)
type Group struct {
	ID         int    `gorm:"column:item_id;primary_key"`
	CategoryID int    `gorm:"column:category_id;default:null"`
	Title      string `gorm:"column:title;default:null"`
}

// Item - CREATE TABLE items (rowid INTEGER PRIMARY KEY ASC, uuid VARCHAR, flags INTEGER, type INTEGER, parent_id INTEGER NOT NULL, ordering INTEGER)
type Item struct {
	ID       int    `gorm:"column:rowid;primary_key"`
	UUID     string `gorm:"column:uuid"`
	Flags    int    `gorm:"column:flags;default:null"`
	Type     int    `gorm:"column:type"`
	Group    Group  `gorm:"ForeignKey:ID"`
	ParentID int    `gorm:"not null;column:parent_id"`
	Ordering int    `gorm:"column:ordering"`
}

// DBInfo - CREATE TABLE dbinfo (key VARCHAR, value VARCHAR)
type DBInfo struct {
	Key   string
	Value string
}

// TableName set DBInfo's table name to be `dbinfo`
func (DBInfo) TableName() string {
	return "dbinfo"
}
