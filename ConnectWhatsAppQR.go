package main

import (
	"context"
	"fmt"
	
	"sync"

	_ "github.com/jackc/pgx/v4/stdlib"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type WhatsAppManager struct {
    Sessions sync.Map 
}

// تعديل الدالة لتعيد كود الـ QR كـ string
func GetWhatsAppQRString() (string, error) {
    dolog := waLog.Stdout("Database", "DEBUG", true)
    globalContainer, err := sqlstore.New(context.Background(), "pgx", GetConnStr(), dolog)
    if err != nil {
        return "", err
    }
    
    deviceStore := globalContainer.NewDevice()
    client := whatsmeow.NewClient(deviceStore, dolog)
 
    // الحصول على قناة الـ QR
    qrChan, _ := client.GetQRChannel(context.Background())
    err = client.Connect()
    if err != nil {
        return "", err
    }

    // ننتظر أول كود QR يصل من القناة
    for evt := range qrChan {
        if evt.Event == "code" {
            return evt.Code, nil // نرجع الكود مباشرة
        }
    }
    return "", fmt.Errorf("timeout or error getting QR")
}




