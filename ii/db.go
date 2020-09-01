package ii

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type MsgInfo struct {
	Id    string
	Echo  string
	Off   int64
	Repto string
}

type Index struct {
	Hash     map[string]MsgInfo
	List     []string
	FileSize int64
}

type DB struct {
	Path string
	Idx  Index
	Sync sync.RWMutex
	Name string
}

func mkdir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
	}
}

func append_file(fn string, text string) error {
	f, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(text + "\n"); err != nil {
		return err
	}
	return nil
}

func (db *DB) Lock() bool {
	try := 16
	for try > 0 {
		if err := os.Mkdir(db.LockPath(), 0777); err == nil || os.IsExist(err) {
			return true
		}
		time.Sleep(time.Second)
		try -= 1
	}
	Error.Printf("Can not acquire lock for 16 seconds")
	return false
}

func (db *DB) Unlock() {
	os.Remove(db.LockPath())
}

func (db *DB) IndexPath() string {
	return fmt.Sprintf("%s.idx", db.Path)
}

func (db *DB) BundlePath() string {
	return fmt.Sprintf("%s", db.Path)
}

func (db *DB) LockPath() string {
	pat := strings.Replace(db.Path, "/", "_", -1)
	return fmt.Sprintf("%s/%s-bundle.lock", os.TempDir(), pat)
}

// var MaxMsgLen int = 128 * 1024 * 1024

func (db *DB) CreateIndex() error {
	db.Sync.Lock()
	defer db.Sync.Unlock()
	db.Lock()
	defer db.Unlock()

	return db._CreateIndex()
}
func file_lines(path string, fn func(string) bool) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	return f_lines(f, fn)
}

func f_lines(f *os.File, fn func(string) bool) error {
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		line = strings.TrimSuffix(line, "\n")
		if err == io.EOF {
			break
		}
		if !fn(line) {
			break
		}
	}
	// scanner := bufio.NewScanner(f)
	// scanner.Buffer(make([]byte, MaxMsgLen), MaxMsgLen)

	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	if !fn(line) {
	// 		break
	// 	}
	// }
	return nil
}

func (db *DB) _CreateIndex() error {
	fidx, err := os.OpenFile(db.IndexPath(), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fidx.Close()
	var off int64
	return file_lines(db.BundlePath(), func(line string) bool {
		msg, _ := DecodeBundle(line)
		if msg == nil {
			off += int64(len(line) + 1)
			return true
		}
		repto, _ := msg.Tag("repto")
		if repto != "" {
			repto = ":" + repto
		}
		fidx.WriteString(fmt.Sprintf("%s:%s:%d%s\n", msg.MsgId, msg.Echo, off, repto))
		off += int64(len(line) + 1)
		return true
	})
}
func (db *DB) _ReopenIndex() (*os.File, error) {
	err := db._CreateIndex()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(db.IndexPath())
	if err != nil {
		return nil, err
	}
	return file, nil
}
func (db *DB) LoadIndex() error {
	var Idx Index
	file, err := os.Open(db.IndexPath())
	if err != nil {
		db.Idx = Idx
		if os.IsNotExist(err) {
			file, err = db._ReopenIndex()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}
	fsize := info.Size()

	if db.Idx.Hash != nil { // already loaded
		if fsize > db.Idx.FileSize {
			Trace.Printf("Refreshing index file...")
			if _, err := file.Seek(0, 2); err != nil {
				return err
			}
			Idx = db.Idx
		} else if info.Size() < db.Idx.FileSize {
			Info.Printf("Index file truncated, rebuild inndex...")
			file, err = db._ReopenIndex()
			if err != nil {
				return err
			}
			defer file.Close()
		}
		return nil
	} else {
		Idx.Hash = make(map[string]MsgInfo)
	}
	var err2 error
	err = f_lines(file, func(line string) bool {
		info := strings.Split(line, ":")
		if len(info) < 3 {
			err2 = errors.New("Wrong format")
			return false
		}
		mi := MsgInfo{Id: info[0], Echo: info[1]}
		if _, err := fmt.Sscanf(info[2], "%d", &mi.Off); err != nil {
			err2 = errors.New("Wrong offset")
			return false
		}
		if len(info) > 3 {
			mi.Repto = info[3]
		}
		if _, ok := Idx.Hash[mi.Id]; !ok { // new msg
			Idx.List = append(Idx.List, mi.Id)
		}
		Idx.Hash[mi.Id] = mi
		return true
	})
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	Idx.FileSize = fsize
	db.Idx = Idx
	return nil
}

func (db *DB) _Lookup(Id string) *MsgInfo {
	if err := db.LoadIndex(); err != nil {
		return nil
	}
	info, ok := db.Idx.Hash[Id]
	if !ok {
		return nil
	}
	return &info
}

func (db *DB) Lookup(Id string) *MsgInfo {
	db.Sync.RLock()
	defer db.Sync.RUnlock()
	db.Lock()
	defer db.Unlock()

	return db._Lookup(Id)
}

func (db *DB) GetBundle(Id string) string {
	db.Sync.RLock()
	defer db.Sync.RUnlock()
	db.Lock()
	defer db.Unlock()

	info := db._Lookup(Id)
	if info == nil {
		Info.Printf("Can not find bundle: %s\n", Id)
		return ""
	}
	f, err := os.Open(db.BundlePath())
	if err != nil {
		Error.Printf("Can not open DB: %s\n", err)
		return ""
	}
	defer f.Close()
	_, err = f.Seek(info.Off, 0)
	if err != nil {
		Error.Printf("Can not seek DB: %s\n", err)
		return ""
	}
	var bundle string
	err = f_lines(f, func(line string) bool {
		bundle = line
		return false
	})
	if err != nil {
		Error.Printf("Can not get %s from DB: %s\n", Id, err)
		return ""
	}
	return bundle
}

func (db *DB) Get(Id string) *Msg {
	bundle := db.GetBundle(Id)
	if bundle == "" {
		return nil
	}
	m, err := DecodeBundle(bundle)
	if err != nil {
		Error.Printf("Can not decode bundle on get: %s\n", Id)
	}
	return m
}

type Query struct {
	Echo  string
	Repto string
	Start int
	Lim   int
}

func prependStr(x []string, y string) []string {
	x = append(x, "")
	copy(x[1:], x)
	x[0] = y
	return x
}

func (db *DB) Match(info MsgInfo, r Query) bool {
	if r.Echo != "" && r.Echo != info.Echo {
		return false
	}
	if r.Repto != "" && r.Repto != info.Repto {
		return false
	}
	return true
}

type Echo struct {
	Name  string
	Count int
}

func (db *DB) Echoes(names []string) []Echo {
	db.Sync.Lock()
	defer db.Sync.Unlock()
	db.Lock()
	defer db.Unlock()
	var list []Echo

	filter := make(map[string]bool)
	for _, n := range names {
		filter[n] = true
	}

	if err := db.LoadIndex(); err != nil {
		return list
	}

	hash := make(map[string]Echo)
	size := len(db.Idx.List)
	for i := 0; i < size; i++ {
		id := db.Idx.List[i]
		info := db.Idx.Hash[id]
		e := info.Echo
		if names != nil { // filter?
			if _, ok := filter[e]; !ok {
				continue
			}
		}
		if v, ok := hash[e]; ok {
			v.Count++
			hash[e] = v
		} else {
			hash[e] = Echo{Name: e, Count: 1}
		}
	}
	if names != nil {
		for _, v := range names {
			list = append(list, hash[v])
		}
	} else {
		for _, v := range hash {
			list = append(list, v)
		}
		sort.Slice(list, func(i, j int) bool {
			return list[i].Name < list[j].Name
		})
	}
	return list
}

func (db *DB) SelectIDS(r Query) []string {
	var Resp []string
	db.Sync.Lock()
	defer db.Sync.Unlock()
	db.Lock()
	defer db.Unlock()

	if err := db.LoadIndex(); err != nil {
		return Resp
	}
	size := len(db.Idx.List)
	if size == 0 {
		return Resp
	}
	if r.Start < 0 {
		start := 0
		for i := size - 1; i >= 0; i-- {
			id := db.Idx.List[i]
			if db.Match(db.Idx.Hash[id], r) {
				Resp = prependStr(Resp, id)
				start -= 1
				if start == r.Start {
					break
				}
			}
		}
		if r.Lim > 0 && len(Resp) > r.Lim {
			Resp = Resp[0:r.Lim]
		}
		return Resp
	}
	found := 0
	for i := r.Start; i < size; i++ {
		id := db.Idx.List[i]
		if db.Match(db.Idx.Hash[id], r) {
			Resp = append(Resp, id)
			found += 1
			if r.Lim > 0 && found == r.Lim {
				break
			}
		}
	}
	return Resp
}

func (db *DB) Store(m *Msg) error {
	return db._Store(m, false)
}

func (db *DB) Edit(m *Msg) error {
	return db._Store(m, true)
}

func (db *DB) _Store(m *Msg, edit bool) error {
	db.Sync.Lock()
	defer db.Sync.Unlock()
	db.Lock()
	defer db.Unlock()
	repto, _ := m.Tag("repto")
	if err := db.LoadIndex(); err != nil {
		return err
	}
	if _, ok := db.Idx.Hash[m.MsgId]; ok && !edit { // exist and not edit
		return errors.New("Already exists")
	}
	fi, err := os.Stat(db.BundlePath())
	var off int64
	if err == nil {
		off = fi.Size()
	}
	if err := append_file(db.BundlePath(), m.Encode()); err != nil {
		return err
	}

	if repto != "" {
		repto = ":" + repto
	}
	rec := fmt.Sprintf("%s:%s:%d%s", m.MsgId, m.Echo, off, repto)
	if err := append_file(db.IndexPath(), rec); err != nil {
		return err
	}
	// if _, ok := db.Idx.Hash[m.MsgId]; !ok { // new msg
	//  	db.Idx.List = append(db.Idx.List, m.MsgId)
	// }
	// mi := MsgInfo{Id: m.MsgId, Echo: m.Echo, Off: off, Repto: repto}
	// db.Idx.Hash[m.MsgId] = mi
	// db.Idx.FileSize += (int64(len(rec) + 1))
	return nil
}

func OpenDB(path string) *DB {
	var db DB
	db.Path = path
	info, err := os.Stat(filepath.Dir(path))
	if err != nil || !info.IsDir() {
		return nil
	}
	db.Name = "node"
	//	db.Idx = make(map[string]Index)
	return &db
}

type User struct {
	Id     int32
	Name   string
	Mail   string
	Secret string
	Tags   Tags
}

type UDB struct {
	Path    string
	Names   map[string]User
	Secrets map[string]string
	List    []string
	Sync    sync.Mutex
}

func IsUsername(u string) bool {
	return !strings.ContainsAny(u, ":\n\r\t") && len(u) <= 16
}

func MakeSecret(msg string) string {
	h := sha256.Sum256([]byte(msg))
	s := base64.URLEncoding.EncodeToString(h[:])
	return s[0:10]
}

func (db *UDB) Access(Secret string) bool {
	_, ok := db.Secrets[Secret]
	return ok
}

func (db *UDB) Name(Secret string) string {
	name, ok := db.Secrets[Secret]
	if ok {
		return name
	}
	Error.Printf("No user for secret: %s", Secret)
	return ""
}

func (db *UDB) Id(Secret string) int32 {
	name, ok := db.Secrets[Secret]
	if ok {
		v, ok := db.Names[name]
		if !ok {
			return -1
		}
		return v.Id
	}
	Error.Printf("No user for secret: %s", Secret)
	return -1
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func (db *UDB) Add(Name string, Mail string, Passwd string) error {
	db.Sync.Lock()
	defer db.Sync.Unlock()

	if _, ok := db.Names[Name]; ok {
		return errors.New("Already exists")
	}
	if !IsUsername(Name) {
		return errors.New("Wrong username")
	}
	if ! emailRegex.MatchString(Mail) {
		return errors.New("Wrong email")
	}
	var id int32 = 0
	for _, v := range db.Names {
		if v.Id > id {
			id = v.Id
		}
	}
	id++
	var u User
	u.Name = Name
	u.Mail = Mail
	u.Secret = MakeSecret(string(id) + Name + Passwd)
	u.Tags = NewTags("")
	db.List = append(db.List, u.Name)
	if err := append_file(db.Path, fmt.Sprintf("%d:%s:%s:%s:%s",
		id, Name, Mail, u.Secret, u.Tags.String())); err != nil {
		return err
	}
	return nil
}

func LoadUsers(path string) *UDB {
	var db UDB
	db.Path = path
	db.Names = make(map[string]User)
	db.Secrets = make(map[string]string)
	err := file_lines(path, func(line string) bool {
		a := strings.Split(line, ":")
		if len(a) < 4 {
			Error.Printf("Wrong entry in user DB: %s", line)
			return true
		}
		var u User
		var err error
		_, err = fmt.Sscanf(a[0], "%d", &u.Id)
		if err != nil {
			Error.Printf("Wrong ID in user DB: %s", a[0])
			return true
		}
		u.Name = a[1]
		u.Mail = a[2]
		u.Secret = a[3]
		u.Tags = NewTags(a[4])
		db.Names[u.Name] = u
		db.Secrets[u.Secret] = u.Name
		db.List = append(db.List, u.Name)
		return true
	})
	if err != nil {
		return nil
	}
	return &db
}
