package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	encryption "bassamarmoraserver/Tools/Cryptography/Encryption"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type ConnectDB struct {
	Conn *pgxpool.Pool
}

func main() {
	/* connStr := GetConnStr()

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("❌ فشل إنشاء مجمع الاتصالات: %v", err)
	}
	defer pool.Close()
	connectDataBasesSQL := &ConnectDB{Conn: pool}
	globalDB = connectDataBasesSQL
	APIWhatsApp()
	*/

	//   For download file .env
	err := godotenv.Load("System.env")
	if err != nil {
		log.Fatalf("❌ فشل تحميل ملف البيئة: %v", err)
	}
	fmt.Println("ملف البيئة تم تحميله بنجاح")

	// for uri encryption
	keyURI := os.Getenv("MY_KMS_KEY_URI")
	encryption.InitCrypto(keyURI)
	// finished  uri encryption
	connStr := GetConnStr()

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("❌ فشل إنشاء مجمع الاتصالات: %v", err)
	}
	defer pool.Close()

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New(context.Background(), "pgx", connStr, dbLog)
	if err != nil {
		log.Fatalf("فشل إنشاء حاوية قاعدة البيانات: %v", err)
	}

	connectDataBasesSQL := &ConnectDB{Conn: pool}

	InitDatabase(connectDataBasesSQL.Conn) //  This Func for Initialization DataBase    | هذه الميثود مخصصه لتهيئة قاعدة البيانات

	// ----------------------------------------> Delete Old Bookings Every Hour This methods is testing
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			DeleteOldBookings(connectDataBasesSQL.Conn)
		}
	}()
	// ----------------------------------------> Delete Old Bookings Every Hour This methods is testing

	r := gin.Default()
	systemView()

	RegisterRoutes(r, connectDataBasesSQL)

	globalDB = connectDataBasesSQL //    تخزين في متغير عام

	InitAllClients(container)

	fmt.Println("localhost:8080")

	r.Run(":8080")

}
