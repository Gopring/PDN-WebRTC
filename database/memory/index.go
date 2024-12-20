// Package memory provides an in-memory database implementation.
package memory

import "github.com/hashicorp/go-memdb"

const (
	tblChannels    = "channels"
	tblClients     = "clients"
	tblConnections = "connections"
)

const (
	idxChannelID       = "id"
	idxClientID        = "id"
	idxClientChannelID = "channel_id"
	idxConnID          = "id"
	idxConnTo          = "to"
	idxConnFrom        = "from"
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
		tblClients: {
			Name: tblClients,
			Indexes: map[string]*memdb.IndexSchema{
				idxClientID: {
					Name:   idxClientID,
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "ChannelID"},
							&memdb.StringFieldIndex{Field: "ID"},
						},
					},
				},
				idxClientChannelID: {
					Name:    idxClientChannelID,
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "ChannelID"},
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
				idxConnFrom: {
					Name:   idxConnFrom,
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "ChannelID"},
							&memdb.StringFieldIndex{Field: "From"},
						},
					},
				},
			},
		},
	},
}
