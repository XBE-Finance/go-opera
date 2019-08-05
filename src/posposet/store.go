package posposet

import (
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// Store is a poset persistent storage working over physical key-value database.
type Store struct {
	physicalDB kvdb.Database

	table struct {
		Checkpoint     kvdb.Database `table:"checkpoint_"`
		Frames         kvdb.Database `table:"frame_"`
		Blocks         kvdb.Database `table:"block_"`
		Event2Frame    kvdb.Database `table:"event2frame_"`
		Event2Block    kvdb.Database `table:"event2block_"`
		SuperFrames    kvdb.Database `table:"sframe_"`
		ConfirmedEvent kvdb.Database `table:"confirmed_"`
		VectorIndex    kvdb.Database `table:"vectors_"`
		Balances       state.Database
	}

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.Database) *Store {
	s := &Store{
		physicalDB: db,
		Instance:   logger.MakeInstance(),
	}

	kvdb.MigrateTables(&s.table, s.physicalDB)
	s.table.Balances = state.NewDatabase(
		s.physicalDB.NewTable([]byte("balance_")))

	return s
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	db := kvdb.NewMemDatabase()
	return NewStore(db)
}

// Close leaves underlying database.
func (s *Store) Close() {
	kvdb.MigrateTables(&s.table, nil)
	s.physicalDB.Close()
}

// ApplyGenesis stores initial state.
func (s *Store) ApplyGenesis(balances map[hash.Peer]inter.Stake) error {
	if balances == nil {
		return fmt.Errorf("balances shouldn't be nil")
	}

	sf0 := s.GetSuperFrame(0)
	if sf0 != nil {
		if sf0.Balances == genesisHash(balances) {
			return nil
		}
		return fmt.Errorf("other genesis has applied already")
	}

	sf := &superFrame{}

	cp := &checkpoint{
		SuperFrameN: 0,
		TotalCap:    0,
	}

	sf.Members = make(internal.Members, len(balances))

	genesis := s.StateDB(hash.Hash{})
	for addr, balance := range balances {
		if balance == 0 {
			return fmt.Errorf("balance shouldn't be zero")
		}

		genesis.SetBalance(hash.Peer(addr), balance)
		cp.TotalCap += balance

		sf.Members.Add(addr, balance)
	}
	sf.Members = sf.Members.Top()

	var err error
	sf.Balances, err = genesis.Commit(true)
	if err != nil {
		return err
	}

	s.SetSuperFrame(cp.SuperFrameN, sf)
	s.SetCheckpoint(cp)

	return nil
}

/*
 * Utils:
 */

func (s *Store) set(table kvdb.Database, key []byte, val interface{}) {
	buf, err := rlp.EncodeToBytes(val)
	if err != nil {
		s.Fatal(err)
	}

	if err := table.Put(key, buf); err != nil {
		s.Fatal(err)
	}
}

func (s *Store) get(table kvdb.Database, key []byte, to interface{}) interface{} {
	buf, err := table.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return nil
	}

	err = rlp.DecodeBytes(buf, to)
	if err != nil {
		s.Fatal(err)
	}
	return to
}

func (s *Store) has(table kvdb.Database, key []byte) bool {
	res, err := table.Has(key)
	if err != nil {
		s.Fatal(err)
	}
	return res
}
