package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Constants - تأكد من مطابقتها لبيانات حسابك في Meta
const (
	MetaAccessToken   = "EAA3k7rl8wTsBQiX5WtgWiDZBJdmYpbOZAyuwWqzlHIhqaX8twyj1YJOuZCZA6j4sNJurNTIpDLOJ8zsUGiqHvGicIxyxO4uGhZBUZBErcJUzWd7oolH5d4yxwuGy5t5V5rN0ooR6jqO5ZCU9We9quifwgYENVls63efff2ZAe1FK49FBXXDp7cYwMJlVbRqLH2WwHP7VKgrjxbWCD3kc8clT7lyLXRuOIZBBANrwWSowiWwb95FMeZCgZBLGQfIp5VkJNp25k5CYkbQBplBWpJUAGNps0VhAgZDZD"
	MetaPhoneNumberID = "1028432257012380"
)

// HandleMetaIncomingMessages: المسار المسؤول عن استقبال الويبهوك الرسمي
func (db *ConnectDB) HandleMetaIncomingMessages(c *gin.Context) {
	var body map[string]interface{}
	if err := c.BindJSON(&body); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// معالجة الرسالة في الخلفية (Goroutine)
	go func(data map[string]interface{}) {
		// استخراج النص والرقم من هيكل JSON المعقد لميتا
		entries, ok := data["entry"].([]interface{})
		if !ok || len(entries) == 0 {
			return
		}
		changes, ok := entries[0].(map[string]interface{})["changes"].([]interface{})
		if !ok || len(changes) == 0 {
			return
		}
		value, ok := changes[0].(map[string]interface{})["value"].(map[string]interface{})
		if !ok {
			return
		}

		if messages, hasMessages := value["messages"].([]interface{}); hasMessages {
			msg := messages[0].(map[string]interface{})
			fromPhone := msg["from"].(string)
			textBody := ""

			if textObj, isText := msg["text"].(map[string]interface{}); isText {
				textBody = strings.TrimSpace(textObj["body"].(string))
			}

			fmt.Printf("\n📩 [Official API] رسالة من: %s | النص: %s\n", fromPhone, textBody)

			// --- 🛠️ استدعاء دوالك الجاهزة مسبقاً (استخدام فقط) ---

			// 1. استخدام دالة جلب الرد الآلي
			botReply, err := db.GetBotResponse(1, textBody)
			if err == nil && botReply != "" {
				db.SendWhatsAppOfficial(fromPhone, botReply)
				return
			}

			// 2. استخدام دالة جلب مواعيد الموظف
			slots, err := db.GetEmployeeSlots(1, textBody)
			if err == nil && len(slots.Slots) > 0 {
				reply := fmt.Sprintf("🗓️ المواعيد المتاحة لـ (%s):\n", textBody)
				for _, s := range slots.Slots {
					reply += fmt.Sprintf("• %s: %s\n", s.DayName, s.SlotTime)
				}
				db.SendWhatsAppOfficial(fromPhone, reply)
				return
			}

			// 3. رد افتراضي في حال عدم وجود تطابق
			db.SendWhatsAppOfficial(fromPhone, "أهلاً بك، يرجى كتابة اسم الخدمة أو الموظف المطلوبة.")
		}
	}(body)

	c.Status(http.StatusOK)
}

// SendWhatsAppOfficial: دالة الإرسال عبر API ميتا الرسمي
func (db *ConnectDB) SendWhatsAppOfficial(to string, message string) error {
	apiURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", MetaPhoneNumberID)

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              "text",
		"text":              map[string]string{"body": message},
	}

	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+MetaAccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		fmt.Printf("❌ فشل الإرسال الرسمي، كود الخطأ: %d\n", resp.StatusCode)
		return fmt.Errorf("error status: %d", resp.StatusCode)
	}

	fmt.Println("✅ تم إرسال الرد الرسمي بنجاح.")
	return nil
}
