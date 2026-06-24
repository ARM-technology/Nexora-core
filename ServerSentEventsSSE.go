package main

import (
	"io"
"fmt"
	"github.com/gin-gonic/gin"
)



// ServerSentEventsSSE.go

func SSENetwork(c *gin.Context) {
    // 1. الحصول على معرف الشركة من الطلب
    // نفترض أن العميل يرسله كـ Query Param: /sse?company_id=1
    companyIDStr := c.Query("company_id")
    if companyIDStr == "" {
        c.JSON(400, gin.H{"error": "company_id is required"})
        return
    }
    
    // تحويله لـ int (بما أن الـ Structs عندك تستخدم int للـ IDCompany)
    var companyID int
    fmt.Sscanf(companyIDStr, "%d", &companyID)

    // 2. إعداد الهيدرز
    w := c.Writer
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    // 3. إنشاء قناة خاصة لهذا الاتصال فقط
    clientChan := make(chan string, 10)

    // 4. تسجيل القناة في الـ Streamer
    Streamer.mu.Lock()
    Streamer.CompanyStreams[companyID] = append(Streamer.CompanyStreams[companyID], clientChan)
    Streamer.mu.Unlock()

    // 5. تنظيف القناة عند قطع الاتصال
    defer func() {
        Streamer.mu.Lock()
        // حذف القناة من مصفوفة قنوات الشركة
        channels := Streamer.CompanyStreams[companyID]
        for i, ch := range channels {
            if ch == clientChan {
                Streamer.CompanyStreams[companyID] = append(channels[:i], channels[i+1:]...)
                break
            }
        }
        Streamer.mu.Unlock()
        close(clientChan)
    }()

    // 6. الاستماع للإرسال
    c.Stream(func(w io.Writer) bool {
        select {
        case msg, ok := <-clientChan:
            if !ok { return false }
            c.SSEvent("message", msg)
            return true
        case <-c.Request.Context().Done():
            return false
        }
    })
}