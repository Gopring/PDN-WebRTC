// Package memory provides an in-memory database implementation.
package memory

import "github.com/hashicorp/go-memdb"

const (
	tblChannels    = "channels"
	tblUsers       = "users"
	tblConnections = "connections"
)

const (
	idxChannelID = "id"
	idxUserID    = "id"
	idxConnID    = "id"
	idxConnTo    = "to"
)

// schema is the schema of the memory database.
var schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		tblChannels: {
			Name: tblChannels,
			Indexes: map[string]*memdb.IndexSchema{
				idxChannelID: {
					Name:    idxChannelID,
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "ID"},
				},
			},
		},
		tblUsers: {
			Name: tblUsers,
			Indexes: map[string]*memdb.IndexSchema{
				idxUserID: {
					Name:   idxUserID,
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "ChannelID"},
							&memdb.StringFieldIndex{Field: "ID"},
						},
					},
				},
			},
		},
		tblConnections: {
			Name: tblConnections,
			Indexes: map[string]*memdb.IndexSchema{
				idxConnID: {
					Name:    idxConnID,
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "ID"},
				},
				idxConnTo: {
					Name:   idxConnTo,
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "ChannelID"},
							&memdb.StringFieldIndex{Field: "To"},
						},
					},
				},
			},
		},
	},
}
