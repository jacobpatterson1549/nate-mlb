package db

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	firestoreDB struct {
		client               *firestore.Client
		sportTypeActiveYears map[SportType]int
		sportTypesByName     map[string]SportType
		sportTypeMap         SportTypeMap
		playerTypeMap        PlayerTypeMap
	}

	firestoreTX struct {
		db  *firestoreDB
		ops []firestoreTransactionOperation
	}

	firestoreTransactionOperation struct {
		name  string
		class firestoreTransactionOperationClass
		doc   *firestore.DocumentRef
		data  map[string]interface{}
		fc    *firestoreFriendChange
	}
	firestoreTransactionOperationClass int
	firestoreFriendChange              struct {
		class       firestoreFriendChangeClass
		sportType   SportType
		oldFriendID ID
		newFriendID ID
	}
	firestoreFriendChangeClass int
	firestoreTransactionReads  struct {
		sportTypePlayerDocs map[SportType][]*firestore.DocumentRef
	}

	firestoreFriend struct {
		DisplayOrder int `firestore:"display_order"`
	}
	firestorePlayer struct {
		DisplayOrder int        `firestore:"display_order"`
		PlayerType   PlayerType `firestore:"player_type"`
		FriendID     ID         `firestore:"friend_id"`
	}
	firestoreStat struct {
		EtlJSON      string     `firestore:"etl_json"`
		EtlTimestamp *time.Time `firestore:"etl_timestamp"`
	}
	firestoreAdminUser struct {
		HashedPassword string `firestore:"admin_password"`
	}
)

const (
	add firestoreTransactionOperationClass = iota + 1
	set
	del
	delPlayersWithFriendID = iota + 1
	setPlayersWithFriendID
	adminUsername              = "admin"
	firestoreContextTimeout    = 5 * time.Second
	firestoreFieldDisplayOrder = "display_order"
	firestoreFieldSportType    = "sport_type"
	firestoreFieldPlayerType   = "player_type"
	firestoreFieldFriendID     = "friend_id"
	firestoreFieldEtlTimestamp = "etl_timestamp"
	firestoreFieldEtlJSON      = "etl_json"
	firestoreFieldPassword     = "admin_password"
	firestoreFieldActiveYear   = "__active__"
)

func newFirestoreDB(projectID string) (*firestoreDB, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID) // do not timeout context - the client is used by the backend
	if err != nil {
		return nil, fmt.Errorf("creating firestore client: %w", err)
	}
	d := firestoreDB{
		client: client,
	}
	return &d, nil
}

func (d *firestoreDB) begin() (dbTX, error) {
	t := firestoreTX{
		db: d,
	}
	return &t, nil
}

func (t *firestoreTX) execute() error {
	if err := withFirestoreTimeoutContext(func(ctx context.Context) error {
		return t.db.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
			reads, err := t.makeReads(tx)
			if err != nil {
				return err
			}
			for _, op := range t.ops {
				if err := op.execute(ctx, tx, *reads); err != nil {
					return fmt.Errorf("%v: %w", op.name, err)
				}
			}
			return nil
		})
	}); err != nil {
		return fmt.Errorf("executing transaction: %w", err)
	}
	return nil
}

func (op firestoreTransactionOperation) execute(ctx context.Context, tx *firestore.Transaction, reads firestoreTransactionReads) error {
	switch op.class {
	case add:
		if err := tx.Create(op.doc, op.data); err != nil {
			return err
		}
	case set:
		var updates []firestore.Update
		for k, v := range op.data {
			u := firestore.Update{
				Path:  k,
				Value: v,
			}
			updates = append(updates, u)
		}
		if err := tx.Update(op.doc, updates); err != nil {
			return err
		}
	case del:
		if op.data != nil {
			return fmt.Errorf("cannot delete document data, can only delete the whole document")
		}
		if err := tx.Delete(op.doc); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown firestoreTransactionOperationClass: %v", op.class)
	}
	if op.fc != nil {
		if err := op.fc.updatePlayers(ctx, tx, reads); err != nil {
			return fmt.Errorf("updating players: %w", err)
		}
	}
	return nil
}

func (fc *firestoreFriendChange) updatePlayers(ctx context.Context, tx *firestore.Transaction, reads firestoreTransactionReads) error {
	playerDocs := reads.sportTypePlayerDocs[fc.sportType]
	for _, doc := range playerDocs {
		snap, err := doc.Get(ctx)
		if err != nil {
			return err
		}
		var p firestorePlayer
		if err := snap.DataTo(&p); err != nil {
			return err
		}
		if p.FriendID != fc.oldFriendID {
			continue
		}
		switch fc.class {
		case delPlayersWithFriendID:
			if err := tx.Delete(doc); err != nil {
				return err
			}
		case setPlayersWithFriendID:
			updates := []firestore.Update{
				{
					Path:  firestoreFieldFriendID,
					Value: fc.newFriendID,
				},
			}
			if err := tx.Update(doc, updates); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown update player operation: %v", fc.class)
		}
	}
	return nil
}

func (t firestoreTX) makeReads(tx *firestore.Transaction) (*firestoreTransactionReads, error) {
	reads := firestoreTransactionReads{
		sportTypePlayerDocs: make(map[SportType][]*firestore.DocumentRef),
	}
	for _, op := range t.ops {
		if op.fc != nil {
			if _, ok := reads.sportTypePlayerDocs[op.fc.sportType]; !ok {
				c, ok := t.db.playersCollection(op.fc.sportType)
				if !ok {
					return nil, fmt.Errorf("could not get players collection to update players")
				}
				var err error
				reads.sportTypePlayerDocs[op.fc.sportType], err = tx.DocumentRefs(c).GetAll()
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return &reads, nil
}

func withFirestoreTimeoutContext(f func(ctx context.Context) error) error {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, firestoreContextTimeout)
	defer cancelFunc()
	return f(ctx)
}

func (d firestoreDB) IsNotExist(err error) bool {
	return status.Code(err) == codes.NotFound
}

// ----- BEGIN COLLECTION / DOCUMENT helpers -----

func (d *firestoreDB) rootDocument() *firestore.DocumentRef {
	return d.client.Collection("services").Doc("nate-mlb")
}

func (d *firestoreDB) statsCollection() *firestore.CollectionRef {
	return d.rootDocument().Collection("stats")
}

func (d *firestoreDB) activeYearsDocument() *firestore.DocumentRef {
	return d.statsCollection().Doc("active-years")
}

func (d *firestoreDB) yearsCollection(st SportType) *firestore.CollectionRef {
	sportTypeName := d.sportTypeMap[st].Name
	return d.statsCollection().Doc(sportTypeName).Collection("years")
}

func (d *firestoreDB) activeYearDoc(st SportType) (_ *firestore.DocumentRef, ok bool) {
	activeYear, ok := d.sportTypeActiveYears[st]
	if !ok {
		return nil, false
	}
	year := strconv.Itoa(activeYear)
	return d.yearsCollection(st).Doc(year), true
}

func (d *firestoreDB) friendsCollection(st SportType) (_ *firestore.CollectionRef, ok bool) {
	doc, ok := d.activeYearDoc(st)
	if !ok {
		return nil, false
	}
	return doc.Collection("friends"), true
}

func (d *firestoreDB) playersCollection(st SportType) (_ *firestore.CollectionRef, ok bool) {
	doc, ok := d.activeYearDoc(st)
	if !ok {
		return nil, false
	}
	return doc.Collection("players"), true
}

// ----- BEGIN QUERY/ SINGLE-EXEC FUNCTIONS -----

func (d *firestoreDB) GetSportTypes() (SportTypeMap, error) {
	d.loadSportTypeMap()
	if err := d.loadSportTypesByName(); err != nil {
		return nil, err
	}
	if err := d.loadActiveYears(); err != nil {
		return nil, err
	}
	return d.sportTypeMap, nil
}

func (d *firestoreDB) loadSportTypeMap() {
	d.sportTypeMap = SportTypeMap{
		SportTypeMlb: SportTypeInfo{
			Name: "MLB",
			URL:  "mlb",
		},
		SportTypeNfl: SportTypeInfo{
			Name: "NFL",
			URL:  "nfl",
		},
	}
}

func (d *firestoreDB) loadSportTypesByName() error {
	d.sportTypesByName = make(map[string]SportType, len(d.sportTypeMap))
	for st, sti := range d.sportTypeMap {
		d.sportTypesByName[sti.Name] = st
	}
	if len(d.sportTypeMap) != len(d.sportTypesByName) {
		return fmt.Errorf("wanted sport types to have unique names: %v", d.sportTypeMap)
	}
	return nil
}

func (d *firestoreDB) loadActiveYears() error {
	doc := d.activeYearsDocument()
	if err := withFirestoreTimeoutContext(func(ctx context.Context) error {
		snap, err := doc.Get(ctx)
		if err != nil {
			if d.IsNotExist(err) {
				return d.initActiveYears(ctx, doc)
			}
			return err
		}
		data := snap.Data()
		d.sportTypeActiveYears = make(map[SportType]int, len(data))
		for name, year := range data {
			st, ok := d.sportTypesByName[name]
			if !ok {
				return fmt.Errorf("unknown sport type name: %v", name)
			}
			if year != nil {
				switch y := year.(type) {
				case int64:
					d.sportTypeActiveYears[st] = int(y)
				case int:
					d.sportTypeActiveYears[st] = y
				default:
					return fmt.Errorf("invalid active sport type year: %v", year)
				}
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("loading active years for sport types: % w", err)
	}
	return nil
}

func (d *firestoreDB) initActiveYears(ctx context.Context, doc *firestore.DocumentRef) error {
	data := make(map[string]interface{}, len(d.sportTypesByName))
	for stName := range d.sportTypesByName {
		data[stName] = nil
	}
	if _, err := doc.Create(ctx, data); err != nil {
		return err
	}
	return nil
}

func (d *firestoreDB) GetPlayerTypes() (PlayerTypeMap, error) {
	d.loadPlayerTypeMap()
	return d.playerTypeMap, nil
}

func (d *firestoreDB) loadPlayerTypeMap() {
	d.playerTypeMap = PlayerTypeMap{
		PlayerTypeMlbTeam: PlayerTypeInfo{
			SportType:    SportTypeMlb,
			Name:         "Teams",
			Description:  "Wins",
			ScoreType:    "Wins",
			DisplayOrder: 1,
		},
		PlayerTypeMlbHitter: PlayerTypeInfo{
			SportType:    SportTypeMlb,
			Name:         "Hitting",
			Description:  "Home Runs",
			ScoreType:    "HRs",
			DisplayOrder: 2,
		},
		PlayerTypeMlbPitcher: PlayerTypeInfo{
			SportType:    SportTypeMlb,
			Name:         "Pitching",
			Description:  "Wins",
			ScoreType:    "Wins",
			DisplayOrder: 3,
		},
		PlayerTypeNflTeam: PlayerTypeInfo{
			SportType:    SportTypeNfl,
			Name:         "Teams",
			Description:  "Wins",
			ScoreType:    "Wins",
			DisplayOrder: 4,
		},
		PlayerTypeNflQB: PlayerTypeInfo{
			SportType:    SportTypeNfl,
			Name:         "Quarterbacks",
			Description:  "Touchdown (passes+runs)",
			ScoreType:    "TDs",
			DisplayOrder: 5,
		},
		PlayerTypeNflMisc: PlayerTypeInfo{
			SportType:    SportTypeNfl,
			Name:         "Misc",
			Description:  "Touchdowns (RB/WR/TE) (Rushing/Receiving)",
			ScoreType:    "TDs",
			DisplayOrder: 6,
		},
	}
}

func (d *firestoreDB) GetYears(st SportType) ([]Year, error) {
	c := d.yearsCollection(st)
	var years []Year
	if err := withFirestoreTimeoutContext(func(ctx context.Context) error {
		snaps, err := c.Documents(ctx).GetAll()
		if err != nil {
			if d.IsNotExist(err) {
				return nil
			}
			return err
		}
		for _, snap := range snaps {
			i, err := strconv.Atoi(snap.Ref.ID)
			if err != nil {
				return fmt.Errorf("invalid year: %w", err)
			}
			y := Year{Value: i}
			if d.sportTypeActiveYears[st] == y.Value {
				y.Active = true
			}
			years = append(years, y)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("get years: % w", err)
	}
	sort.Slice(years, func(i, j int) bool {
		return years[i].Value < years[j].Value
	})
	return years, nil
}

func (d *firestoreDB) GetFriends(st SportType) ([]Friend, error) {
	c, ok := d.friendsCollection(st)
	if !ok {
		return nil, nil
	}
	var friends []Friend
	if err := withFirestoreTimeoutContext(func(ctx context.Context) error {
		snaps, err := c.Documents(ctx).GetAll()
		if err != nil {
			return err
		}
		var ff firestoreFriend
		for _, snap := range snaps {
			if err := snap.DataTo(&ff); err != nil {
				return err
			}
			f := Friend{
				ID:           ID(snap.Ref.ID),
				Name:         snap.Ref.ID,
				DisplayOrder: ff.DisplayOrder,
			}
			friends = append(friends, f)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("get friends: % w", err)
	}
	sort.Slice(friends, func(i, j int) bool {
		return friends[i].Name < friends[j].Name
	})
	return friends, nil
}

func (d *firestoreDB) GetPlayers(st SportType) ([]Player, error) {
	c, ok := d.playersCollection(st)
	if !ok {
		return nil, nil
	}
	var players []Player
	if err := withFirestoreTimeoutContext(func(ctx context.Context) error {
		snaps, err := c.Documents(ctx).GetAll()
		if err != nil {
			return err
		}
		var fp firestorePlayer
		for _, snap := range snaps {
			if err := snap.DataTo(&fp); err != nil {
				return err
			}
			sourceID, err := strconv.Atoi(snap.Ref.ID)
			if err != nil {
				return err
			}
			p := Player{
				ID:           ID(snap.Ref.ID),
				PlayerType:   PlayerType(fp.PlayerType),
				FriendID:     fp.FriendID,
				DisplayOrder: fp.DisplayOrder,
				SourceID:     SourceID(sourceID),
			}
			players = append(players, p)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("get players: % w", err)
	}
	sort.Slice(players, func(i, j int) bool {
		if players[i].PlayerType != players[j].PlayerType {
			return players[i].PlayerType < players[j].PlayerType
		}
		if players[i].FriendID != players[j].FriendID {
			return players[i].FriendID < players[j].FriendID
		}
		return players[i].DisplayOrder < players[j].DisplayOrder
	})
	return players, nil
}

func (d *firestoreDB) GetStat(st SportType) (*Stat, error) {
	doc, ok := d.activeYearDoc(st)
	if !ok {
		return nil, nil
	}
	var stat Stat
	if err := withFirestoreTimeoutContext(func(ctx context.Context) error {
		snap, err := doc.Get(ctx)
		if err != nil {
			return err
		}
		var fs firestoreStat
		if err := snap.DataTo(&fs); err != nil {
			return err
		}
		y, err := strconv.Atoi(snap.Ref.ID)
		if err != nil {
			return err
		}
		stat.Year = y
		stat.SportType = st
		b := []byte(fs.EtlJSON)
		stat.EtlJSON = &b
		stat.EtlTimestamp = fs.EtlTimestamp
		return nil
	}); err != nil {
		return nil, fmt.Errorf("get stat: % w", err)
	}
	return &stat, nil
}

func (d *firestoreDB) SetStat(stat Stat) error {
	doc, ok := d.activeYearDoc(stat.SportType)
	if !ok {
		return fmt.Errorf("no active year to set stat for")
	}
	m := map[string]interface{}{
		firestoreFieldEtlJSON:      string(*stat.EtlJSON),
		firestoreFieldEtlTimestamp: stat.EtlTimestamp,
	}
	if err := withFirestoreTimeoutContext(func(ctx context.Context) error {
		if _, err := doc.Set(ctx, m); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("set stat: %w", err)
	}
	return nil
}

func (d *firestoreDB) ClrStat(st SportType) error {
	stat := Stat{
		SportType: st,
		EtlJSON:   &[]byte{}, // TODO: yuck (make it a string)
	}
	if err := d.SetStat(stat); err != nil {
		return fmt.Errorf("clear stat: %w", err)
	}
	return nil
}

func (d *firestoreDB) GetUserPassword(username string) (string, error) {
	if username != adminUsername {
		return "", fmt.Errorf("cannot get username for %q", username)
	}
	doc := d.rootDocument()
	var fu firestoreAdminUser
	if err := withFirestoreTimeoutContext(func(ctx context.Context) error {
		snap, err := doc.Get(ctx)
		if err != nil {
			return err
		}
		if err := snap.DataTo(&fu); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return "", fmt.Errorf("get user password: %w", err)
	}
	return fu.HashedPassword, nil
}

func (d *firestoreDB) SetUserPassword(username, hashedPassword string) error {
	if username != adminUsername {
		return fmt.Errorf("cannot set username for %q", username)
	}
	m := map[string]interface{}{
		firestoreFieldPassword: hashedPassword,
	}
	doc := d.rootDocument()
	if err := withFirestoreTimeoutContext(func(ctx context.Context) error {
		if _, err := doc.Set(ctx, m); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("set user password: %w", err)
	}
	return nil
}

func (d *firestoreDB) AddUser(username, hashedPassword string) error {
	if err := d.SetUserPassword(username, hashedPassword); err != nil {
		return fmt.Errorf("add user: %w", err)
	}
	return nil
}

// ----- BEGIN TRANSACTION FUNCTIONS -----

func (t *firestoreTX) AddYear(st SportType, year int) {
	c := t.db.yearsCollection(st)
	y := strconv.Itoa(year)
	doc := c.Doc(y)
	data := map[string]interface{}{} // firestore does not like nil data
	op := firestoreTransactionOperation{
		name:  "add year",
		class: add,
		doc:   doc,
		data:  data,
	}
	t.ops = append(t.ops, op)
}

func (t *firestoreTX) DelYear(st SportType, year int) {
	c := t.db.yearsCollection(st)
	y := strconv.Itoa(year)
	doc := c.Doc(y)
	op := firestoreTransactionOperation{
		name:  "delete year",
		class: del,
		doc:   doc,
	}
	t.ops = append(t.ops, op)
}

func (t *firestoreTX) SetYearActive(st SportType, year int) {
	t.db.sportTypeActiveYears[st] = year
	doc := t.db.activeYearsDocument()
	sportTypeName := t.db.sportTypeMap[st].Name
	data := map[string]interface{}{
		sportTypeName: year,
	}
	op := firestoreTransactionOperation{
		name:  "set active year",
		class: set,
		doc:   doc,
		data:  data,
	}
	t.ops = append(t.ops, op)
}

func (t *firestoreTX) ClrYearActive(st SportType) {
	if _, ok := t.db.sportTypeActiveYears[st]; !ok {
		return
	}
	delete(t.db.sportTypeActiveYears, st)
	doc := t.db.activeYearsDocument()
	sportTypeName := t.db.sportTypeMap[st].Name
	data := map[string]interface{}{
		sportTypeName: nil,
	}
	op := firestoreTransactionOperation{
		name:  "clear active year",
		class: set,
		doc:   doc,
		data:  data,
	}
	t.ops = append(t.ops, op)
}

func (t *firestoreTX) AddFriend(st SportType, displayOrder int, name string) {
	c, ok := t.db.friendsCollection(st)
	if !ok {
		return
	}
	doc := c.Doc(name)
	data := map[string]interface{}{
		firestoreFieldDisplayOrder: displayOrder,
	}
	op := firestoreTransactionOperation{
		name:  "add friend",
		class: add,
		doc:   doc,
		data:  data,
	}
	t.ops = append(t.ops, op)
}

func (t *firestoreTX) SetFriend(st SportType, id ID, displayOrder int, name string) {
	c, ok := t.db.friendsCollection(st)
	if !ok {
		return
	}
	path := string(id)
	doc := c.Doc(path)
	data := map[string]interface{}{
		firestoreFieldDisplayOrder: displayOrder,
	}
	if id == ID(name) {
		// set id only
		op := firestoreTransactionOperation{
			name:  "set friend [display order]",
			class: set,
			doc:   doc,
			data:  data,
		}
		t.ops = append(t.ops, op)
	} else {
		// remove friend, add new friend, update friend names for players with old friendID
		op1 := firestoreTransactionOperation{
			name:  "set friend [delete old name]",
			class: del,
			doc:   doc,
		}
		doc2 := c.Doc(name)
		op2 := firestoreTransactionOperation{
			name:  "set friend [name]",
			class: add,
			doc:   doc2,
			data:  data,
			fc: &firestoreFriendChange{
				class:       setPlayersWithFriendID,
				sportType:   st,
				oldFriendID: id,
				newFriendID: ID(name),
			},
		}
		t.ops = append(t.ops, op1, op2)
	}
}

func (t *firestoreTX) DelFriend(st SportType, id ID) {
	c, ok := t.db.friendsCollection(st)
	if !ok {
		return
	}
	path := string(id)
	doc := c.Doc(path)
	// delete friend, delete players with old friendID
	op := firestoreTransactionOperation{
		name:  "delete friend",
		class: del,
		doc:   doc,
		fc: &firestoreFriendChange{
			class:       delPlayersWithFriendID,
			sportType:   st,
			oldFriendID: id,
		},
	}
	t.ops = append(t.ops, op)
}

func (t *firestoreTX) AddPlayer(st SportType, displayOrder int, pt PlayerType, sourceID SourceID, friendID ID) {
	c, ok := t.db.playersCollection(st)
	if !ok {
		return
	}
	path := strconv.Itoa(int(sourceID))
	doc := c.Doc(path)
	data := map[string]interface{}{
		firestoreFieldDisplayOrder: displayOrder,
		firestoreFieldPlayerType:   pt,
		firestoreFieldFriendID:     friendID,
	}
	op := firestoreTransactionOperation{
		name:  "add player",
		class: add,
		doc:   doc,
		data:  data,
	}
	t.ops = append(t.ops, op)
}

func (t *firestoreTX) SetPlayer(st SportType, id ID, displayOrder int) {
	c, ok := t.db.playersCollection(st)
	if !ok {
		return
	}
	path := string(id)
	doc := c.Doc(path)
	data := map[string]interface{}{
		firestoreFieldDisplayOrder: displayOrder,
	}
	op := firestoreTransactionOperation{
		name:  "set player",
		class: set,
		doc:   doc,
		data:  data,
	}
	t.ops = append(t.ops, op)
}

func (t *firestoreTX) DelPlayer(st SportType, id ID) {
	c, ok := t.db.playersCollection(st)
	if !ok {
		return
	}
	path := string(id)
	doc := c.Doc(path)
	op := firestoreTransactionOperation{
		name:  "delete player",
		class: del,
		doc:   doc,
	}
	t.ops = append(t.ops, op)
}
