package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"

	"time"

	_ "github.com/jackc/pgx/v4/stdlib"

	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

/*
T1 => StartClientForUser
T2 => MyCustomEventHandler   ---->  This is Event for multi client
T3 => InitAllClients  ----> For search and connect with multi client use gorutine
//  D1 ====>  Testing Can Deleting This



*/

// ############################################################################################################################################
// ############################################################################################################################################
// ############################################################################################################################################
// ############################################################################################################################################
// ############################################################################################################################################
// ############################################################################################################################################

// Testing this for get all device in DB   |
// This is testing for connect  multi phone\account    and use Goroutine for that
//    هذا اختبار لأستعمال للأتصال بأكثر من رقم و مزامنتهم  و استعمال القو روتين لذالك
// is Success ? Yes is Success
//  هل نج ذالك ؟  نعم نجح

// ############################################################################################################################################
// ############################################################################################################################################
// ############################################################################################################################################
// ############################################################################################################################################

type ClientManager struct { // For managment Multi Client  |     للإدارة اكثر من عميل
	Clients map[string]*whatsmeow.Client
	mu      sync.RWMutex
}

var Manager = &ClientManager{ //  لتخزين كل عميل في سلايد لمزامنته |     For Save Client in Slid\array for use
	Clients: make(map[string]*whatsmeow.Client),
}

//	T1  ================================ use client
//
// نمرر الـ Device والـ Container كباراميترات
func StartClientForUser(deviceStore *store.Device) {
	// 1. استخدام معرف الجهاز كـ ID فريد
	userID := deviceStore.ID.ToNonAD().User

	// 2. إعداد سجلات خاصة بهذا العميل فقط
	clientLog := waLog.Stdout("Client-"+userID, "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// 3. تخزين العميل في الـ Manager للوصول له لاحقاً
	Manager.mu.Lock()
	Manager.Clients[userID] = client
	Manager.mu.Unlock()

	// 4. ربط الـ EventHandler بشكل محلي (Closure)
	// هذا يضمن أن الرسائل التي تصل لهذا العميل لا تختلط بغيره
	client.AddEventHandler(func(evt interface{}) {
		// يمكنك هنا تمرير userID للـ EventHandler الخاص بك
		MyCustomEventHandler(userID, evt)
	})

	// 5. الاتصال
	err := client.Connect()
	if err != nil {
		log.Printf("❌ فشل اتصال العميل %s: %v", userID, err)
		return
	}
	fmt.Printf("✅ العميل %s متصل الآن بنجاح\n", userID)
}

var LastResponseMap sync.Map // ====>  Testing Can Deleting This  D1
// T2 =====================  For Event
func MyCustomEventHandler(userID string, evt interface{}) {
	// 1. تحديد المنطقة الزمنية
	loc, _ := time.LoadLocation("Asia/Riyadh")

	switch v := evt.(type) {
	case *events.Message:
		// جلب الكلاينت للتحقق من وجوده
		Manager.mu.RLock()
		client, exists := Manager.Clients[userID]
		Manager.mu.RUnlock()

		if !exists || client == nil {
			return
		}

		// 2. تحديد الطرف الآخر (Logic الجاهز الخاص بك)
		var otherJID types.JID
		if v.Info.IsGroup { // ======>  هنا دالة مهمته ترجع اذا كانت الرساله من قروب ترجع true
			otherJID = v.Info.Sender.ToNonAD()
			log.Printf("This msg in group ! [%v]", otherJID)
		} else {
			if v.Info.IsFromMe {
				otherJID = v.Info.Chat.ToNonAD()
			} else {
				otherJID = v.Info.Sender.ToNonAD()
			}
		}

		if !v.Info.IsGroup { // هذا شرط اذا كانت الرساله قادمه من فرد ليس قروب != not group    فنفذ الاكواد هذه
			// 3. حل مشكلة الـ LID (Logic الجاهز الخاص بك)
			isLID := (otherJID.Server == types.HiddenUserServer) || strings.Contains(otherJID.String(), "@lid")
			if isLID {
				if !v.Info.IsGroup && !v.Info.IsFromMe {
					alt := v.Info.SenderAlt.ToNonAD()
					if alt.User != "" && alt.Server == types.DefaultUserServer {
						otherJID = alt
						isLID = false
					}
				}
				if isLID && !v.Info.IsGroup && v.Info.IsFromMe {
					alt := v.Info.RecipientAlt.ToNonAD()
					if alt.User != "" && alt.Server == types.DefaultUserServer {
						otherJID = alt
						isLID = false
					}
				}
				if isLID && client.Store != nil && client.Store.LIDs != nil {
					pn, err := client.Store.LIDs.GetPNForLID(context.Background(), otherJID)
					if err == nil && pn.User != "" {
						otherJID = pn.ToNonAD()
					}
				}
			}

			// 4. استخراج نص الرسالة
			text := v.Message.GetConversation()
			if text == "" && v.Message.GetExtendedTextMessage() != nil {
				text = v.Message.GetExtendedTextMessage().GetText()
			}

			// 5. تحديد المرسل والمستقبل
			from := otherJID.User
			to := userID
			if v.Info.IsFromMe {
				from = userID
				to = otherJID.User
			}

			timeStrMsg := v.Info.Timestamp.In(loc).Format("2006/01/02 | 15:04")

			// 6. إرسال للستريمر (واجهة المستخدم)
			companyID, err := globalDB.GetCompanyIDByPhone(userID)
			if err == nil {
				output := fmt.Sprintf("companyPhone[%s]time[%s] from [%s] to [%s]: %s", userID, timeStrMsg, from, to, text)

				// 🔥 التعديل هنا: نرسل للشركة المحددة فقط
				Streamer.BroadcastToCompany(companyID, output)
			} else {
				fmt.Printf("⚠️ تعذر إرسال SSE: لم يتم العثور على ID للشركة %s\n", userID)
			}

			// 7. تجهيز البيانات الأساسية في الستركت
			msgData := MessageChat{
				From:         from,
				To:           to,
				Msg:          text,
				Date:         timeStrMsg,
				IsMe:         v.Info.IsFromMe,
				PhoneCompany: userID,
				IDCompany:    companyID, // إسناد الـ ID الذي جلبناه
			}

			// نقل كل العمليات التي تتطلب وقت (DB + WhatsApp Send) إلى الـ Background
			go func(data MessageChat, isFromMe bool) {
				if globalDB == nil {
					return
				}

				// أ) جلب ID الشركة (استخدام Cache أو DB)
				companyID, err := globalDB.GetCompanyIDByPhone(data.PhoneCompany)
				if err != nil {
					fmt.Printf("⚠️ فشل جلب ID الشركة للرقم %s: %v\n", data.PhoneCompany, err)
					return
				}
				data.IDCompany = companyID

				// ب) حفظ الرسالة في قاعدة البيانات
				err = globalDB.SaveMessage(data)
				if err != nil {
					fmt.Printf("❌ خطأ أثناء حفظ الرسالة في DB: %v\n", err)
				}

				// ج) الرد الآلي (Bot Response)
				// فقط إذا كانت الرسالة واردة (ليست مني)
				if !isFromMe {
					// 1️⃣ تنظيف النص القادم (حذف المسافات)
					cleanText := strings.TrimSpace(data.Msg)

					// 2️⃣ هل طلب العميل السكوت؟ (أرسل رقم 1)
					if cleanText == "1" || cleanText == "١" {
						// نضعه في "قائمة السكوت" لمدة دقيقة واحدة من الآن
						expiry := time.Now().Add(1 * time.Minute)
						BotSilenceMap.Store(data.From, expiry)

						// اختياري: إرسال رسالة تأكيد للعميل
						// SendMesages(client, data.IDCompany, data.From, "تم إيقاف الرد الآلي لمدة دقيقة ✅")
						fmt.Printf("🔇 تم تفعيل وضع السكوت للعميل: %s\n", data.From)
						return // نخرج ولا نجعل البوت يرد على رقم 1 نفسه ردوداً أخرى
					}

					// 3️⃣ فحص هل العميل حالياً في وضع السكوت؟
					if until, locked := BotSilenceMap.Load(data.From); locked {
						// نتحقق هل الوقت المحدد للسكوت انتهى أم لا؟
						if time.Now().Before(until.(time.Time)) {
							fmt.Printf("⏳ البوت ساكت حالياً عن العميل %s (متبقي ثوانٍ)\n", data.From)
							return // العميل صامت، البوت لن يرد
						} else {
							// انتهت الدقيقة، نحذفه من الماب ليعود البوت للعمل
							BotSilenceMap.Delete(data.From)
						}
					}

					//   D1 =====> You can deleting this is for testing and dev || Start
					if lastResp, ok := LastResponseMap.Load(data.From); ok {
						if time.Since(lastResp.(time.Time)) < time.Second {
							return
						}
					}
					LastResponseMap.Store(data.From, time.Now())
					//  D1 =====> You can deleting this is for testing and dev  || End
					if GlobalRegistry.IsActive(data.IDCompany) {
						SendMesages(client, data.IDCompany, data.From, data.Msg)
					}
					// نستخدم دالة الإرسال التي صلحناها سابقاً

				}

			}(msgData, v.Info.IsFromMe)

			// 8. طباعة في الكونسول للمراقبة
			fmt.Printf("time [%s] 📩 from [%s] to [%s]: %s\n", timeStrMsg, from, to, text)
		}
	}
}

// T3    ======================== For all client search get and connect
func InitAllClients(container *sqlstore.Container) {
	ctx := context.Background()

	checkAndConnect := func() {
		devices, err := container.GetAllDevices(ctx)
		if err != nil {
			return
		}

		/*
		       for _, device := range devices {
		           // التعديل هنا: نرسل متغير واحد فقط (الجهاز)
		           go StartClientForUser(device)
		       }
		   }
		*/
		for _, device := range devices {
			if device.ID == nil {
				continue
			}

			// استخراج رقم الهاتف
			userID := device.ID.ToNonAD().User

			// 2. التحقق من وجود الرقم في قاعدة بياناتك الخاصة
			_, err := globalDB.GetCompanyIDByPhone(userID)

			if err != nil {
				// 🛑 الرقم غير موجود في نظامك -> تصفية شاملة من الـ DB
				log.Printf("🗑️ جاري حذف كافة بيانات الرقم %s من قاعدة البيانات لعدم وجود اشتراك نشط", userID)

				// حذف السجل والمفاتيح والجلسات من جداول SQL
				deleteErr := device.Delete(ctx)
				if deleteErr != nil {
					log.Printf("❌ خطأ أثناء الحذف الفيزيائي من DB: %v", deleteErr)
				} else {
					log.Printf("✅ تم تطهير قاعدة البيانات بنجاح من الرقم: %s", userID)
				}

				// ننتقل للرقم التالي ولا نقوم بتشغيله
				continue
			}

			// ✅ الرقم موجود ومصرح له -> ابدأ تشغيل العميل
			// ملاحظة: بدون ClientsMap سيحاول النظام تشغيل Goroutine جديد كل 20 ثانية
			// لذا يجب أن تتأكد داخل StartClientForUser أن العميل لا يعمل مسبقاً
			go StartClientForUser(device)
		}
	}

	checkAndConnect()

	go func() {
		ticker := time.NewTicker(25 * time.Minute) // ------->      Check All 20 minute
		for range ticker.C {
			checkAndConnect()
		}
	}()
}

// #################################################################################
//
//	T4 ---> Func for send msg and handule this msg
var NameEmp []NameEmployee

func SendMesages(cli *whatsmeow.Client, idCompany int, phone string, msg string) {
	ctx := context.Background()
	phoneJID := PhoneToJID(phone)
	cleanMsg := FromEmojiNumbers(strings.TrimSpace(msg))

	// 1. فحص الحظر
	isBlocked, _ := globalDB.IsClientBlocked(idCompany, phone)
	if isBlocked {
		log.Printf("🚫 الهاتف %s محظور، تم تجاهل الرسالة.\n", phone)
		return
	}
	globalDB.IncrementSendMsgCount(idCompany, 1) //  زيادة في عنصر   numsendmsg   في قاعدة البيانات
	var response string
	var err error

	// 2. محاولة جلب تفاصيل الموعد (هل الرسالة عبارة عن كود حجز؟)
	// نستخدم الميثود التي صممناها في الخطوة السابقة
	booking, err := globalDB.GetSlotDetailByKeyword(idCompany, cleanMsg, true)

	if err == nil {
		// ✅ إذا نجح الجلب، يعني أن الرسالة هي "كود حجز" فعلي
		// نستدعي الميثود الخاصة بك التي تعيد نص النجاح مباشرة
		response, err = globalDB.InsertBooking(idCompany, phone, booking.EmpKeyword, booking.DayName, booking.SlotTime)
		if err != nil {
			fmt.Println(err)
			response = "⚠️ عذراً، هذا الموعد لم يعد متاحاً، ربما سبقه شخص آخر بحجزه."
		}
	} else {
		// ❌ إذا لم يكن كود حجز، ننتقل لمعالجة الأوامر والردود الأخرى
		switch cleanMsg {
		case "2":
			response = HanduleBookingEmployee(idCompany)
		default:
			// البحث عن اسم موظف لعرض جدول مواعيده
			_, isEmployee := GetEmployeeNameByKeyword(cleanMsg)
			if isEmployee {
				response = GetAllBooking(idCompany, cleanMsg)
			} else {
				// البحث في الردود الآلية في قاعدة البيانات
				response, err = globalDB.GetBotResponse(idCompany, msg)
				// إذا لم يوجد رد، نرسل الرسالة الترحيبية الافتراضية (000000)
				if err != nil || response == "" || response == "عذراً، هذه الكلمة غير مسجلة لدينا." {
					response, _ = globalDB.GetBotResponse(idCompany, "000000")
				}
			}
		}
	}

	// 4. إرسال الرد النهائي للمستخدم
	if response != "" {
		_, err = cli.SendMessage(ctx, phoneJID, BuildTextMessage(response))
		if err != nil {
			fmt.Printf("❌ فشل إرسال الرسالة إلى %s: %v\n", phone, err)
		}
	}
}

// ################################################################################
//
//	T5  --->  Func for  chang string phone number to JID and return JID
func PhoneToJID(phone string) types.JID {
	return types.NewJID(phone, types.DefaultUserServer)
}

func BuildTextMessage(text string) *waProto.Message {
	return &waProto.Message{
		Conversation: proto.String(text),
	}
}

func HanduleBookingEmployee(idCompany int) string {
	// 1. جلب الموظفين من قاعدة البيانات
	allEmployees, err := globalDB.GetAllEmployeeKeywords(idCompany)
	if err != nil {
		log.Println("❌ خطأ:", err)
		return ""
	}

	// تصفير السلايس العام قبل تعبئته لضمان عدم تكرار البيانات القديمة
	NameEmp = []NameEmployee{}

	var messageBuilder strings.Builder
	messageBuilder.WriteString("📋 *طاقمنا المتميز:*\n\n")

	hasActive := false

	for _, emp := range allEmployees {
		if emp.IsActive {
			hasActive = true

			// --- إضافة الموظف للسلايس العام ---
			NameEmp = append(NameEmp, NameEmployee{
				Name:    emp.EmployeeName,
				Keyword: emp.EmpKeyword,
			})
			// ---------------------------------

			// تحويل الكلمة المفتاحية إلى إيموجي للرسالة
			emojiKeyword := ToEmojiNumbers(emp.EmpKeyword)

			// بناء سطر الرسالة
			line := fmt.Sprintf("👤 %s : %s\n", emp.EmployeeName, emojiKeyword)
			messageBuilder.WriteString(line)
		}
	}

	if !hasActive {
		return "نعتذر، لا يوجد موظفون متاحون حالياً."
	}

	return messageBuilder.String()
}

func GetEmployeeNameByKeyword(keyword string) (string, bool) {
	// نقوم بتنظيف الكود من أي مسافات أو إيموجي إذا لزم الأمر
	cleanKeyword := FromEmojiNumbers(strings.TrimSpace(keyword))

	for _, emp := range NameEmp {
		if emp.Keyword == cleanKeyword {
			return emp.Name, true // وجدنا الموظف، نرجع اسمه و true
		}
	}
	return "", false // لم نجد الموظف
}

func GetAllBooking(idCompany int, empKeyword string) string {
	nameEmployee, _ := GetEmployeeNameByKeyword(empKeyword)
	// جلب المواعيد باستخدام empKeyword لضمان الدقة
	slotsData, err := globalDB.GetEmployeeSlots(idCompany, empKeyword)
	if err != nil || len(slotsData.Slots) == 0 {
		return fmt.Sprintf("❌ نعتذر، لا توجد مواعيد متاحة حالياً للـ %s.", nameEmployee)
	}

	var msg strings.Builder
	// رأس الرسالة بشكل جذاب
	msg.WriteString(fmt.Sprintf("🗓️ *جدول مواعيد %s*\n", nameEmployee))
	msg.WriteString("ــــــــــــــــــــــــــــــــــــــــــــــــ\n\n")

	for _, s := range slotsData.Slots {
		// تحويل الكود إلى إيموجي (مثل 22221)
		emojiCode := ToEmojiNumbers(s.Keyword)

		// تنسيق السطر بشكل احترافي
		// مثال: 🔹 الأحد الساعة 05:15 PM  كود: 2️⃣2️⃣2️⃣2️⃣1️⃣
		line := fmt.Sprintf("🔹 *%s* الساعة %s  🏷️\n %s\n",
			s.DayName,
			s.SlotTime,
			emojiCode)

		msg.WriteString(line)
	}

	msg.WriteString("\nــــــــــــــــــــــــــــــــــــــــــــــــ\n")
	msg.WriteString("✅ *للحجز:* أرسل كود الموعد الموضح أعلاه.")

	return msg.String()
}

func (db *ConnectDB) SendMsgFromEmployeeToClient(companyID int, idEmployee string, toPhone string, msg string) error {
	// 1. إعداد Context بمهلة 10 ثوانٍ لحماية السيرفر من التعليق
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. تنظيف الرقم وتجهيزه بصيغة دولية (بدون مسافات)
	cleanPhone := strings.ReplaceAll(toPhone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "+", "")

	// 3. جلب رقم الشركة من قاعدة البيانات
	phoneCompany, err := db.GetPhoneByCompanyID(companyID)
	if err != nil || strings.TrimSpace(phoneCompany) == "" {
		return fmt.Errorf("❌ رقم الشركة غير موجود أو غير مرتبط")
	}

	// 4. الوصول لعميل الواتساب النشط
	Manager.mu.RLock()
	cli, exists := Manager.Clients[phoneCompany]
	Manager.mu.RUnlock()

	if !exists || cli == nil {
		return fmt.Errorf("⚠️ الواتساب غير مرتبط للشركة: %s", phoneCompany)
	}

	// 5. محاولة "تنشيط" الاتصال بالرقم المستهدف (Pre-fetch)
	// نرسل الرقم مع علامة "+" كما طلبت المكتبة في الصورة المرفقة
	phoneWithPlus := "+" + cleanPhone
	go func() {
		// استخدام context منفصل بسيط للتنشيط لكي لا يؤثر على سرعة الدالة الرئيسية
		actCtx, actCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer actCancel()
		cli.IsOnWhatsApp(actCtx, []string{phoneWithPlus})
	}()

	// 6. تجهيز الـ JID والإرسال
	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)

	_, err = cli.SendMessage(ctx, recipientJID, BuildTextMessage(msg))
	if err != nil {
		// محاولة أخيرة في حال فشل جلب قائمة الأجهزة (Common Fix)
		if strings.Contains(err.Error(), "device list") {
			_, err = cli.SendMessage(ctx, recipientJID, BuildTextMessage(msg))
		}

		if err != nil {
			return fmt.Errorf("❌ فشل إرسال الرسالة عبر واتساب: %v", err)
		}
	}

	// 7. حفظ الرسالة في قاعدة البيانات في الخلفية
	saveErr := db.AddMessageByEmployee(companyID, idEmployee, cleanPhone, msg)
	if saveErr != nil {
		fmt.Printf("⚠️ فشل حفظ الرسالة في السجل: %v\n", saveErr)
		return fmt.Errorf("❌ تم الإرسال لكن فشل الحفظ: %v", saveErr)
	}

	go globalDB.IncrementSendMsgCount(companyID, 1)

	return nil
}

func (db *ConnectDB) SendOTP(senderPhone string, recipientPhone string) (string, error) {
	// 1. توليد رمز OTP عشوائي (4 أرقام)
	rand.Seed(time.Now().UnixNano())
	otpCode := fmt.Sprintf("%04d", rand.Intn(10000))

	// 2. إعداد Context سريع (7 ثوانٍ)
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	// 3. تنظيف الأرقام
	cleanSender := strings.ReplaceAll(strings.ReplaceAll(senderPhone, " ", ""), "+", "")
	cleanRecipient := strings.ReplaceAll(strings.ReplaceAll(recipientPhone, " ", ""), "+", "")

	// 4. التأكد من اتصال العميل (Sender)
	Manager.mu.RLock()
	cli, exists := Manager.Clients[cleanSender]
	Manager.mu.RUnlock()

	if !exists || cli == nil {
		return "", fmt.Errorf("⚠️ جهاز الإرسال غير متصل")
	}

	// 5. تجهيز الرسالة
	messageBody := fmt.Sprintf("Your verification code is: %s", otpCode)
	recipientJID := types.NewJID(cleanRecipient, types.DefaultUserServer)

	// 6. الإرسال الفعلي
	_, err := cli.SendMessage(ctx, recipientJID, BuildTextMessage(messageBody))

	// محاولة ثانية في حال خطأ التزامن (Device List)
	if err != nil && strings.Contains(err.Error(), "device list") {
		_, err = cli.SendMessage(ctx, recipientJID, BuildTextMessage(messageBody))
	}

	if err != nil {
		return "", fmt.Errorf("❌ فشل الإرسال: %v", err)
	}

	// --- الإضافة الجديدة: حفظ الكود في الذاكرة لمدة 60 ثانية فقط ---
	otpStore.Store(cleanRecipient, otpData{
		code:      otpCode,
		expiresAt: time.Now().Add(60 * time.Second),
	})

	return otpCode, nil
}
