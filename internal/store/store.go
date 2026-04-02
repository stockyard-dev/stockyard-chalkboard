package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Board struct{
	ID string `json:"id"`
	Name string `json:"name"`
	Data string `json:"data"`
	Width int `json:"width"`
	Height int `json:"height"`
	BackgroundColor string `json:"background_color"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"chalkboard.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS boards(id TEXT PRIMARY KEY,name TEXT NOT NULL,data TEXT DEFAULT '{}',width INTEGER DEFAULT 1920,height INTEGER DEFAULT 1080,background_color TEXT DEFAULT '#1a1410',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *Board)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO boards(id,name,data,width,height,background_color,created_at)VALUES(?,?,?,?,?,?,?)`,e.ID,e.Name,e.Data,e.Width,e.Height,e.BackgroundColor,e.CreatedAt);return err}
func(d *DB)Get(id string)*Board{var e Board;if d.db.QueryRow(`SELECT id,name,data,width,height,background_color,created_at FROM boards WHERE id=?`,id).Scan(&e.ID,&e.Name,&e.Data,&e.Width,&e.Height,&e.BackgroundColor,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]Board{rows,_:=d.db.Query(`SELECT id,name,data,width,height,background_color,created_at FROM boards ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []Board;for rows.Next(){var e Board;rows.Scan(&e.ID,&e.Name,&e.Data,&e.Width,&e.Height,&e.BackgroundColor,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM boards WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM boards`).Scan(&n);return n}
