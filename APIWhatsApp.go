package main

import (
	"bytes"

	"encoding/json"
	"fmt"

	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"


)

const (
	AccessToken   = "EAA3k7rl8wTsBQzSXLOkKlBbAZAiTjeWHXGjkMAiyDaVkGpzdOh66VdnVeeoyJccBbXHpZBkXUQ0ZCTZCi8yzhEZAFTfe764Eup9fZBf6dmpFfTs06Ixbn6DIdA9E5uFnjlpFCV1GvZB34EGlZCo0PtuEhvBUUZAOQ1QUpeDYtsCeJVU6wIPFHrRpIWazwI8XHXUVltsvWmZBq8H3YKaKnYv4IZCrXQRhylb7jtPipFHFYEaU5VPpBzjZBzZCcF4xO3vicAHBFZBEvq1tZBCvuyEJqv46q1c4m4v"
	PhoneNumberID = "1028432257012380"
	VerifyToken   = "Hamouri_Secret_123"
)

func APIWhatsApp() {




	r := gin.Default()

	// 1. مسارات الـ Webhook
	r.GET("/webhook", VerifyWebhook)
	r.POST("/webhook", HandleIncomingMessages)

	fmt.Println("🚀 السيرفر يعمل الآن بنظام الرد الآلي (Sandbox)...")
	r.Run(":8080")
}

// دالة التحقق من الـ Webhook عند الربط بميتا
func VerifyWebhook(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token == VerifyToken {
		fmt.Println("✅ تم التحقق من الـ Webhook بنجاح!")
		c.String(http.StatusOK, challenge)
		return
	}
	c.Status(http.StatusForbidden)
}

// دالة معالجة الرسائل الواردة والرد الفوري
func HandleIncomingMessages(c *gin.Context) {
	var body map[string]interface{}
	if err := c.BindJSON(&body); err != nil {
		return
	}

	go func(data map[string]interface{}) {
		defer func() { recover() }() 

		entry := data["entry"].([]interface{})[0].(map[string]interface{})
		changes := entry["changes"].([]interface{})[0].(map[string]interface{})
		value := changes["value"].(map[string]interface{})

		// استخراج بيانات Metadata الرقم الأمريكي (1555)
		metadata := value["metadata"].(map[string]interface{})
		displayPhone := metadata["display_phone_number"].(string)

		// فحص وجود رسائل
		if messages, ok := value["messages"].([]interface{}); ok {
			msg := messages[0].(map[string]interface{})
			
			msgID := msg["id"].(string) 
			fromPhone := msg["from"].(string) 
			timestampStr := msg["timestamp"].(string) 
			
			textBody := ""
			if textObj, ok := msg["text"].(map[string]interface{}); ok {
				textBody = textObj["body"].(string)
			}

			// تنسيق الوقت
			tsInt, _ := strconv.ParseInt(timestampStr, 10, 64)
			dateFormatted := time.Unix(tsInt, 0).Format("2006-01-02 15:04:05")

			// الطباعة بالصيغة الهندسية المطلوبة
			fmt.Printf("\nID[%s] Phone [%s] From[%s] To[%s] Date[%s] : %s\n", 
				msgID, displayPhone, fromPhone, displayPhone, dateFormatted, textBody)

			// الرد الآلي "Hi"
			var mesg string 
			fmt.Println(textBody)
			mesg, err := globalDB.GetBotResponse(1, textBody)
			fmt.Println(mesg)
			if err != nil {
				fmt.Printf("❌ فشل الإرسال: %v\n", err)
			} else {
				fmt.Println("✅ تم الإرسال بنجاح")
			}
				fmt.Printf("⏳ جاري إرسال رسالة 'Hi' إلى %s...\n", fromPhone)
			err = sendText(fromPhone, mesg)

			if err != nil {
				fmt.Printf("❌ فشل الإرسال: %v\n", err)
			} else {
				fmt.Println("✅ تم الإرسال بنجاح")
			}

		
		
			// تأخير بسيط لترتيب الـ Logs
			time.Sleep(100 * time.Millisecond)

			
		}
	}(body)

	c.Status(http.StatusOK)
}

// دالة إرسال الرسالة النصية عبر API ميتا
func sendText(to string, message string) error {
	apiURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", PhoneNumberID)

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text":              map[string]string{"body": message},
	}

	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return nil
}




/*


package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =====================================================================
// 🔒 ثوابت Meta API (تأكد من تجديد التوكن)
// =====================================================================
const (
	MetaAccessToken   = "EAA3k7rl8wTsBQqDdjvWkMpTA7o0JXRZBH8ZADH5ZCZB9Bvdzpc9QhOyHp7pwubKV6QrzCMshZAtmB2gfTpbBMKtQIiRTXkVuTJfZCbXJUf0a0oFWO7zcA5g8iQvPNiFka0XRJrW4T7gYbsFMkVkvZAcPmdt3ELtGaH0VFl8ZCO7M94qOEOeMd1szqKYMfAmrxIY3FAIk75FaZAZCNtzIm695ZCOzkr73T23unvYyJaTsL7OkQK00SAPfp02GzCAQR8NyhrvNJXCOJlUFM4ZC4HJYQd34gAxp"
	MetaPhoneNumberID = "1028432257012380"
	MetaVerifyToken   = "Hamouri_Secret_123"
)

type ConnectDB struct {
	Conn *pgxpool.Pool
}

func main() {
	connStr := GetConnStr()
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("❌ فشل إنشاء مجمع الاتصالات: %v", err)
	}
	defer pool.Close()

	connectDataBasesSQL := &ConnectDB{Conn: pool}

	// تهيئة الجداول
	err = InitDatabase(connectDataBasesSQL.Conn)
	if err != nil {
		log.Fatalf("❌ فشل تهيئة قاعدة البيانات: %v", err)
	}

	systemView()

	r := gin.Default()
	RegisterRoutes(r, connectDataBasesSQL)

	r.GET("/webhook", VerifyMetaWebhook)
	r.POST("/webhook", connectDataBasesSQL.HandleMetaIncomingMessages)

	fmt.Println("🚀 السيرفر يعمل الآن بنظام الـ WhatsApp Cloud API الرسمي...")
	r.Run(":8080")
}

// ---------------------------------------------------------------------
// 📥 دالة معالجة الرسائل الواردة والرد من جدول bot_responses
// ---------------------------------------------------------------------
func (db *ConnectDB) HandleMetaIncomingMessages(c *gin.Context) {
	var body map[string]interface{}
	if err := c.BindJSON(&body); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	go func(data map[string]interface{}) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("⚠️ Recovered from panic: %v\n", r)
			}
		}()

		// تفكيك JSON ميتا
		entries, ok := data["entry"].([]interface{})
		if !ok || len(entries) == 0 { return }
		entry, ok := entries[0].(map[string]interface{})
		if !ok { return }
		changes, ok := entry["changes"].([]interface{})
		if !ok || len(changes) == 0 { return }
		value, ok := changes[0].(map[string]interface{})["value"].(map[string]interface{})
		if !ok { return }

		if messages, hasMessages := value["messages"].([]interface{}); hasMessages {
			msg := messages[0].(map[string]interface{})
			msgID := msg["id"].(string)
			fromPhone := msg["from"].(string)
			
			textBody := ""
			if textObj, isText := msg["text"].(map[string]interface{}); isText {
				textBody = strings.TrimSpace(textObj["body"].(string))
			}

			fmt.Printf("\n📩 رسالة جديدة [ID: %s] | من: [%s] | النص: %s\n", msgID, fromPhone, textBody)

			// =================================================================
			// 🤖 منطق الرد الآلي من قاعدة البيانات
			// =================================================================
			
			// 1️⃣ البحث أولاً في جدول bot_responses (الردود العامة)
			botReply, err := db.GetBotResponse(1, textBody) // id_company = 1
			if err == nil && botReply != "" {
				fmt.Printf("✅ تم إيجاد رد في bot_responses للكلمة: %s\n", textBody)
				SendWhatsAppText(fromPhone, botReply)
				return
			}

			// 2️⃣ البحث ثانياً في جدول مواعيد الموظفين
			slots, err := db.GetEmployeeSlots(1, textBody)
			if err == nil && len(slots.Slots) > 0 {
				reply := fmt.Sprintf("🗓️ المواعيد المتاحة لـ (%s):\n", textBody)
				for _, s := range slots.Slots {
					reply += fmt.Sprintf("- يوم %s الساعة %s\n", s.DayName, s.SlotTime)
				}
				SendWhatsAppText(fromPhone, reply)
				return
			}

			// 3️⃣ الرد الافتراضي إذا لم يتطابق النص مع أي شيء
			fmt.Println("⚠️ لم يتم العثory على تطابق، إرسال الرد الافتراضي...")
			defaultMsg := "أهلاً بك، شكراً لتواصلك معنا. لم يتم التعرف على الكلمة، يرجى كتابة اسم الخدمة أو الموظف."
			SendWhatsAppText(fromPhone, defaultMsg)
		}
	}(body)

	c.Status(http.StatusOK)
}

// ---------------------------------------------------------------------
// 🔍 ميثود جلب الرد من جدول bot_responses
// ---------------------------------------------------------------------


// ---------------------------------------------------------------------
// 📤 دالة الإرسال عبر API ميتا الرسمي
// ---------------------------------------------------------------------
func SendWhatsAppText(to string, message string) error {
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

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		fmt.Printf("❌ Meta Error Details: %v\n", errorResponse)
		return fmt.Errorf("Meta API Error: %d", resp.StatusCode)
	}

	fmt.Println("✅ تم إرسال الرد للعميل بنجاح")
	return nil
}

func VerifyMetaWebhook(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token == MetaVerifyToken {
		fmt.Println("✅ Webhook Verified!")
		c.String(http.StatusOK, challenge)
		return
	}
	c.Status(http.StatusForbidden)
}



*/




/*
//  For Send PDF or Word

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
"net/textproto"
	"github.com/gin-gonic/gin"
)

// الثوابت الخاصة بك
const (
	AccessToken   = "EAA3k7rl8wTsBQiX5WtgWiDZBJdmYpbOZAyuwWqzlHIhqaX8twyj1YJOuZCZA6j4sNJurNTIpDLOJ8zsUGiqHvGicIxyxO4uGhZBUZBErcJUzWd7oolH5d4yxwuGy5t5V5rN0ooR6jqO5ZCU9We9quifwgYENVls63efff2ZAe1FK49FBXXDp7cYwMJlVbRqLH2WwHP7VKgrjxbWCD3kc8clT7lyLXRuOIZBBANrwWSowiWwb95FMeZCgZBLGQfIp5VkJNp25k5CYkbQBplBWpJUAGNps0VhAgZDZD"
	PhoneNumberID = "1028432257012380"
	VerifyToken   = "Hamouri_Secret_123"
)

func main() {
	r := gin.Default()

	r.GET("/webhook", VerifyWebhook)
	r.POST("/webhook", HandleIncomingMessages)

	fmt.Println("🚀 السيرفر يعمل الآن... جاهز لإرسال Mm.pdf")
	r.Run(":8080")
}

func VerifyWebhook(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token == VerifyToken {
		c.String(http.StatusOK, challenge)
		return
	}
	c.Status(http.StatusForbidden)
}

func HandleIncomingMessages(c *gin.Context) {
	var body map[string]interface{}
	if err := c.BindJSON(&body); err != nil {
		return
	}

	go func(data map[string]interface{}) {
		defer func() { recover() }()

		entry := data["entry"].([]interface{})[0].(map[string]interface{})
		changes := entry["changes"].([]interface{})[0].(map[string]interface{})
		value := changes["value"].(map[string]interface{})

		if messages, ok := value["messages"].([]interface{}); ok {
			msg := messages[0].(map[string]interface{})
			fromPhone := msg["from"].(string)
			timestampStr := msg["timestamp"].(string)

			tsInt, _ := strconv.ParseInt(timestampStr, 10, 64)
			dateFormatted := time.Unix(tsInt, 0).Format("2006-01-02 15:04:05")

			fmt.Printf("\n📩 رسالة جديدة من [%s] في [%s]\n", fromPhone, dateFormatted)

			// 1. رفع الملف أولاً للحصول على Media ID
			fmt.Println("⏳ جاري رفع الملف Mm.pdf إلى ميتا...")
			mediaID, err := uploadMedia("Mm.pdf")
			if err != nil {
				fmt.Printf("❌ فشل الرفع: %v\n", err)
				return
			}

			// 2. إرسال الملف باستخدام الـ ID
			fmt.Printf("⏳ جاري إرسال المستند (ID: %s) إلى %s...\n", mediaID, fromPhone)
			err = sendDocument(fromPhone, mediaID)
			
			if err != nil {
				fmt.Printf("❌ فشل الإرسال: %v\n", err)
			} else {
				fmt.Println("✅ تم إرسال ملف Mm.pdf بنجاح!")
			}
		}
	}(body)

	c.Status(http.StatusOK)
}

// دالة لرفع الملف من الـ Root إلى سيرفرات ميتا


// دالة الرفع المحدثة لتجاوز خطأ الـ Content-Type
func uploadMedia(filePath string) (string, error) {
	apiURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/media", PhoneNumberID)

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// --- التعديل هنا لضبط الـ Content-Type ليكون PDF ---
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filePath))
	h.Set("Content-Type", "application/pdf") // تحديد النوع صراحة
	part, err := writer.CreatePart(h)
	if err != nil {
		return "", err
	}
	io.Copy(part, file)

	writer.WriteField("messaging_product", "whatsapp")
	writer.Close()

	req, _ := http.NewRequest("POST", apiURL, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	if id, ok := res["id"].(string); ok {
		return id, nil
	}
	
	// لطباعة الخطأ كاملاً للفهم
	errorDetail, _ := json.Marshal(res)
	return "", fmt.Errorf("Meta Error: %s", string(errorDetail))
}

// دالة إرسال المستند باستخدام الـ Media ID
func sendDocument(to string, mediaID string) error {
	apiURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", PhoneNumberID)

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "document",
		"document": map[string]string{
			"id":       mediaID,
			"filename": "Mm.pdf",
			"caption":  "تفضل هذا هو الملف المطلوب",
		},
	}

	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return nil
}


*/












/*

package main

import (
	"fmt"
	"net/http"
)

func main() {
	// المسار الذي حددناه في رابط ngrok
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		// ميتا ترسل طلب GET للتحقق من السيرفر لأول مرة
		if r.Method == http.MethodGet {
			token := r.URL.Query().Get("hub.verify_token")
			challenge := r.URL.Query().Get("hub.challenge")

			if token == "my_crm_secret_2026" {
				fmt.Println("✅ تم التحقق من السيرفر بنجاح من قبل ميتا!")
				w.Write([]byte(challenge))
				return
			}
			w.WriteHeader(http.StatusForbidden)
			return
		}
		
		// هنا سيتم استقبال رسائل الواتساب الفعلية (POST) مستقبلاً
		fmt.Println("📩 وصلت رسالة جديدة أو إشعار من واتساب!")
	})

	fmt.Println("🚀 سيرفر الـ CRM يعمل الآن وينتظر ميتا على المنفذ 8080...")
	http.ListenAndServe(":8080", nil)
}





*/












/*

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv" // مكتبة لقراءة ملفات البيئة
	"google.golang.org/genai"
)

func main() {
	// 1. تحميل المتغيرات من ملف System.env تحديداً
	err := godotenv.Load("System.env")
	if err != nil {
		log.Fatal("تعذر تحميل ملف System.env")
	}

	ctx := context.Background()

	// 2. جلب مفتاح واجهة برمجة التطبيقات (API Key) من الملف
	apiKey := os.Getenv("GEMINI_API_KEY")

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	localPdfPath := "Mm.pdf"

	// 3. رفع الملف
	uploadConfig := &genai.UploadFileConfig{MIMEType: "application/pdf"}
	uploadedFile, err := client.Files.UploadFromPath(ctx, localPdfPath, uploadConfig)
	if err != nil {
		log.Fatal(err)
	}

	// 4. إعداد الطلب (الملف المرفوع والسؤال)
	promptParts := []*genai.Part{
		genai.NewPartFromURI(uploadedFile.URI, uploadedFile.MIMEType),
		genai.NewPartFromText(`
		You must generate a JSON data structure. The only required output must be the JSON structure itself without any additional text.

The system must be able to receive a file in PDF, Excel, Word, or any other format. It must search within the document and extract employee names and their appointments. Each employee must be assigned a unique 4-digit key (non-repeated). The appointments must then be added under the corresponding 4-digit key for each employee.

For example:

{
"employee": {
"Dr.Ahmed": "1111",
"Dr.Saed": "2212"
},
"Booking": {
"1111": {
"11111":{
"time":"20:45:00",
"day":"Sunday"
},
"11112":{
"time":"17:15:00",
"day":"Sunday"
}
},
"2212": {
"22121":{
"time":"10:00:00",
"day":"Monday"
}
}
}
}

The structure must follow exactly this format.

Do not include titles such as Doctor or Engineer unless the document explicitly states that the person is a doctor, engineer, etc.

The response language must be Arabic.

		`),
	}
	contents := []*genai.Content{
		genai.NewContentFromParts(promptParts, genai.RoleUser),
	}

	// 5. إرسال الطلب للنموذج واستقبال النتيجة
	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash-lite",
		contents,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.Text())
}

*/