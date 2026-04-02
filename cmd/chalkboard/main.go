package main
import ("fmt";"log";"net/http";"os";"github.com/stockyard-dev/stockyard-chalkboard/internal/server";"github.com/stockyard-dev/stockyard-chalkboard/internal/store")
func main(){port:=os.Getenv("PORT");if port==""{port="8860"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./chalkboard-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("chalkboard: %v",err)};defer db.Close();srv:=server.New(db)
fmt.Printf("\n  Chalkboard — Self-hosted collaborative whiteboard\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n\n",port,port)
log.Printf("chalkboard: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
