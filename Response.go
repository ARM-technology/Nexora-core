package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

/*
📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖

	This is for search a func fast search  you can write this keyword

📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖📖

F0 ---> دالة جلب اختبار
F1 --> دالة تسجيل الدخول
F2  ---> دالة مخصصه لربط مع واتس اب من خلال QR
F3 ---> For add package | لإضافة باقات
F4 ---> Create new company
F5  ----->// GetAllPackagesHandler API لجلب كل الباقات
F6 ------>  for block bot to response client 10 minuts
F7  ---->   For   add keyword from Employee
*/
func GetInfo(c *gin.Context) { // F0
	c.JSON(http.StatusOK, infData)
}

func (db *ConnectDB) AdminGetAllSystemDataHandler(c *gin.Context) {
	var input NewCompany

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "تنسيق طلب غير صحيح",
		})
		return
	}

	err := godotenv.Load("System.env")
	if err != nil {
		fmt.Printf("Error loading System.env file")
	}

	nameAdmianID := os.Getenv("NAMEADMIAN")
	if nameAdmianID == "" {
		fmt.Printf("ADMIAN name not found in system")
	}
	passAdmian := os.Getenv("PASSADMIAN")
	if passAdmian == "" {
		fmt.Printf("Passward admian not found in system")
	}
	if input.EmployeeID != nameAdmianID || input.Password != passAdmian {
		c.JSON(401, gin.H{
			"status":  "unauthorized",
			"message": fmt.Errorf("Error Not true Username or Passward"),
		})
		return
	}
	err = RegisterNewCompany(
		db.Conn,
		input.CompanyName,
		input.PackageID,
		input.CompanyPhone,
		input.EmployeeID,
		input.EmployeeName,
		input.EmployeePhone,
		input.Username,
		input.Password,
	)
	if err != nil {
		c.JSON(402, gin.H{
			"status":  "unauthorized",
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"status":      "success",
		"namecompany": input.CompanyName,
	})
}

func GetQR_AdmianSystem(c *gin.Context) {
	var input NewCompany
	// 1. استقبال البيانات
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "تنسيق طلب غير صحيح"})
		return
	}

	// 2. التحقق من الأمان (تم إصلاح المنطق ليكون OR || كما اتفقنا سابقاً)
	_ = godotenv.Load("System.env")
	nameAdmianID := os.Getenv("NAMEADMIAN")
	passAdmian := os.Getenv("PASSADMIAN")

	// استخدمنا TrimSpace لتجنب مشاكل المسافات
	if strings.TrimSpace(input.EmployeeID) != nameAdmianID || strings.TrimSpace(input.Password) != passAdmian {
		c.JSON(401, gin.H{"status": "unauthorized", "message": "بيانات الدخول غير صحيحة"})
		return
	}

	// 3. توليد الـ QR (هنا كان الخطأ)
	// نستدعي دالة التوليد مباشرة وليس الـ Handler
	qrCode, err := GetWhatsAppQRString()
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "فشل توليد الـ QR: " + err.Error()})
		return
	}

	// 4. إرجاع النتيجة
	c.JSON(200, gin.H{
		"status":  "success",
		"qr_code": qrCode, // تأكد أن هذا المفتاح يطابق ما ينتظره Flutter
	})
}

func LogIn(c *gin.Context) { // F1 --->  for login
	var login loginReq // الستركت الخاص بك
	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	// استدعاء الميثود المحدث
	isValid, role, nameEmployee, companyID := globalDB.CheckUserLogin(login.ID, login.Passward)

	if isValid {
		token, err := GenerateToken(companyID, login.ID, role, nameEmployee)

		if err != nil {
			// في حال فشل توليد التوكن لسبب تقني
			c.JSON(500, gin.H{"success": false, "message": "Failed to generate security token"})
			return
		}

		c.JSON(200, gin.H{
			"success":       true,
			"token":         token,
			"type_role":     role,
			"id_company":    companyID, // إرجاع الـ ID للتطبيق
			"name_employee": nameEmployee,
		})
	} else {
		c.JSON(401, gin.H{"success": false, "message": "Login failed"})
	}
}

// ######################################################################
//
//	F2   GET QR For Connect with whatsApp
//
// ######################################################################
func GetQR(c *gin.Context) { // F2
	qrCode, err := GetWhatsAppQRString()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "فشل في توليد رمز QR",
			"details": err.Error(),
		})
		return
	}

	// إرسال الكود للفرونت إند
	c.JSON(http.StatusOK, gin.H{
		"qr_code": qrCode, // هذا الكود سيقوم الفرونت إند بتحويله لصورة QR
	})
}

// ########################################################################################
//
//	F3 ---> For add package | لإضافة باقات
//
// ########################################################################################
func (db *ConnectDB) AddPackageHandler(c *gin.Context) {
	var pkg Package

	// 1. قراءة بيانات الـ JSON والتحقق منها
	if err := c.ShouldBindJSON(&pkg); err != nil {
		c.JSON(400, gin.H{"error": "بيانات غير صحيحة", "details": err.Error()})
		return
	}

	// 2. استدعاء ميثود الإضافة اللي كتبتها أنت (AddPackage)
	// لاحظ نمرر db.Conn لأن دالتك تتوقعه
	err := AddPackage(db.Conn, pkg)
	if err != nil {
		c.JSON(500, gin.H{"error": "فشل حفظ الباقة في النظام"})
		return
	}

	// 3. الرد بالنجاح
	c.JSON(200, gin.H{
		"status":  "success",
		"message": "تمت إضافة باقة الاشتراك بنجاح",
	})
}

// ############################################################################################################
// F4 ---> Create new company
// ############################################################################################################
func (db *ConnectDB) RegisterCompanyHandler(c *gin.Context) {
	var input NewCompany

	// 1. فحص بيانات الـ JSON المرسلة
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "بيانات غير مكتملة", "details": err.Error()})
		return
	}

	// 2. استدعاء الميثود اللي كتبتها أنت في QueryDataBaseSQL.go
	// نلاحظ أن ميثودك تأخذ الباراميترات فرادى، لذا سنمررها من الـ Struct
	err := RegisterNewCompany(
		db.Conn,
		input.CompanyName,
		input.PackageID,
		input.CompanyPhone,
		input.EmployeeID,
		input.EmployeeName,
		input.EmployeePhone,
		input.Username,
		input.Password,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": "فشل إنشاء الشركة والمدير", "details": err.Error()})
		return
	}

	_, role, nameEmployee, companyID := globalDB.CheckUserLogin(input.EmployeeID, input.Password)
	token, err := GenerateToken(companyID, input.EmployeeID, role, nameEmployee)

	if err != nil {
		// في حال فشل توليد التوكن لسبب تقني
		c.JSON(500, gin.H{"success": false, "message": "Failed to generate security token"})
		return
	}

	// 3. الرد بالنجاح
	c.JSON(200, gin.H{
		"status":  "success",
		"token":   token,
		"message": "تم إنشاء الشركة وحساب المدير بنجاح",
	})
}

// #################################################################################################
//
//	F5  ----->// GetAllPackagesHandler API لجلب كل الباقات
func (db *ConnectDB) GetAllPackagesHandler(c *gin.Context) {
	packages, err := GetAllPackages(db.Conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, packages)
}

// ##################################################################################################
//
//	F6 ------>  for block bot to response client 10 minuts
func (db *ConnectDB) BlockBotHandler(c *gin.Context) {
	var body BlockBot
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	// ملاحظة: هنا نحتاج الـ ID الخاص بالشركة، يمكن جلبه من التوكن أو إرساله
	// سأفترض هنا وجود قيمة تجريبية أو جلبها من السياق

	err := db.BlockClientTemporarily(body.PhoneClient, body.IDCompany)
	if err != nil {
		c.JSON(500, gin.H{"error": "فشل حظر العميل: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "تم حظر العميل " + body.PhoneClient + " لمدة 10 دقائق",
	})
}

// ############################################################################################################################################
//
//	F7  ---->   For   add keyword from Employee
func (db *ConnectDB) AddEmployeeKeywordsHandler(c *gin.Context) {
	var req AddEmployeesRequest

	// فك شفرة الـ JSON والتأكد من صحة البيانات
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "تنسيق البيانات غير صحيح", "details": err.Error()})
		return
	}

	// إرسال المصفوفة كاملة للدالة التي كتبناها سابقاً
	err := db.AddEmployeeKeywords(req.Employees)
	if err != nil {
		c.JSON(500, gin.H{"error": "فشل في حفظ البيانات في قاعدة البيانات"})
		return
	}

	c.JSON(200, gin.H{"message": fmt.Sprintf("تم بنجاح إضافة/تحديث %d من الموظفين", len(req.Employees))})
}

// F8   --->  Get Info company
func (db *ConnectDB) GetAllInfoCompany(c *gin.Context) {
	var input GetInfoCompany
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "تنسيق البيانات غير صحيح", "details": err.Error()})
		return
	}

	companyDetails, err := db.GetCompanyDetailsByID(input.IDCompany)
	if err != nil {
		c.JSON(404, gin.H{"error": "فشل جلب بيانات الشركة", "details": err.Error()})
		return
	}
	c.JSON(200, companyDetails)
}

// F9  --->  Add New Employee into company system
func (db *ConnectDB) AddNewEmployeeHandler(c *gin.Context) {
	var input AddMultieNewEmployee

	// 1. استلام الـ JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "تنسيق البيانات غير صحيح", "details": err.Error()})
		return
	}

	// 2. استدعاء العملية الجماعية
	err := db.AddMultipleEmployeesBulk(input)
	if err != nil {
		// في حال فشل أي موظف، سيتراجع الـ Transaction عن الجميع ويرجع الخطأ هنا
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "فشلت عملية الإضافة الجماعية",
			"details": err.Error(),
		})
		return
	}

	// 3. نجاح كامل للعملية
	c.JSON(200, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("تم إضافة %d موظفين بنجاح للشركة رقم %d", len(input.Slots), input.IDCompany),
	})
}

// F10   Get All Employee in company has keyword
func (db *ConnectDB) GetAllEmployeeKeywordsHandler(c *gin.Context) {
	var idStr = c.Query("id_company")

	if idStr == "" {
		c.JSON(400, gin.H{"error": "يجب إرسال id_company"})
		return
	}
	var companyID int
	fmt.Sscanf(idStr, "%d", &companyID)
	employees, err := db.GetAllEmployeeKeywords(companyID)
	if err != nil {
		c.JSON(500, gin.H{"error": "فشل جلب البيانات: " + err.Error()})
		return
	}
	c.JSON(200, employees)
}

// F11  ---->  Get All phones send to company
func (db *ConnectDB) AddGet_ClientPhones(c *gin.Context) {
	var req Get_MoreNumberPhoneAndInfo

	// 1. فك تشفير البيانات القادمة (نحتاج فقط الـ IDCompany من الـ Request)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "تنسيق البيانات خطأ", "details": err.Error()})
		return
	}

	// 2. استدعاء ميثود قاعدة البيانات لجلب الأرقام (التي تعيد []string)
	phones, err := db.GetClientPhones(req.IDCompany)
	if err != nil {
		c.JSON(500, gin.H{"error": "فشل جلب البيانات من القاعدة", "details": err.Error()})
		return
	}

	// 3. تحويل []string إلى []Get_NumberPhoneAndInfo (ملء الـ Slots)
	// نستخدم make بسعة محددة مسبقاً لتحسين الأداء (High Performance)
	req.Slots = make([]Get_NumberPhoneAndInfo, 0, len(phones))

	for _, p := range phones {
		// بما أن FromPhone هو *string (Pointer)، نحتاج لأخذ عنوان المتغير
		phoneCopy := p
		req.Slots = append(req.Slots, Get_NumberPhoneAndInfo{
			FromPhone: &phoneCopy, // إسناد عنوان الرقم
		})
	}

	// 4. إرسال الاستجابة النهائية بنفس الستركت المطلوب
	c.JSON(200, req)
}

// F12 ----->  Get All Info Chat
func (db *ConnectDB) AddGet_ChatHistoryHandler(c *gin.Context) {
	var req struct {
		IDCompany   int    `json:"id_company" binding:"required"`
		ClientPhone string `json:"client_phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "بيانات ناقصة"})
		return
	}

	messages, err := db.GetChatHistory(req.IDCompany, req.ClientPhone)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// استخدام الستركت المطلوب لتوحيد الرد
	response := Get_MoreNumberPhoneAndInfo{
		IDCompany: req.IDCompany,
		Slots:     messages,
	}

	c.JSON(200, response)
}

// F13  ----> Send msg from employee and save this msg
func (db *ConnectDB) AddEmployeeSendMessageHandler(c *gin.Context) {
	// الستركت المطلوب استلامه فقط من الموبايل
	var req struct {
		IDCompany  int    `json:"id_company" binding:"required"`
		IDEmployee string `json:"id_employee" binding:"required"`
		ToPhone    string `json:"to_phone" binding:"required"`
		Message    string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "بيانات الطلب ناقصة"})
		return
	}

	// استدعاء الميثود الاحترافية
	err := db.SendMsgFromEmployeeToClient(req.IDCompany, req.IDEmployee, req.ToPhone, req.Message)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success", "message": "تم إرسال الرسالة وتسجيلها"})
}

// F14 ----> Handler لجلب جميع الموظفين
func (db *ConnectDB) GetAllEmployeesHandler(c *gin.Context) {
	var req GetAllEmployeesReq

	// استقبال JSON يحتوي على id_company
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "بيانات الطلب ناقصة", "details": err.Error()})
		return
	}

	employees, err := db.GetAllEmployeesByCompany(req.IDCompany)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"status": "success",
		"count":  len(employees),
		"data":   employees,
	})
}

// F15 ----> Handler لجلب موظف واحد
func (db *ConnectDB) GetSingleEmployeeHandler(c *gin.Context) {
	var req GetSingleEmployeeReq

	// استقبال JSON يحتوي على id_company و id_employee
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "بيانات الطلب ناقصة", "details": err.Error()})
		return
	}

	employee, err := db.GetEmployeeByID(req.IDCompany, req.IDEmployee)
	if err != nil {
		// نرجح خطأ 404 إذا لم يتم العثور عليه
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"status": "success",
		"data":   employee,
	})
}

// F17 ----> Handler للتحديث الجماعي للموظفين
func (db *ConnectDB) UpdateMultiEmployeesHandler(c *gin.Context) {
	var req UpdateMultiEmployee

	// 1. استقبال الـ JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "تنسيق البيانات غير صحيح", "details": err.Error()})
		return
	}

	// 2. التحقق من وجود بيانات
	if len(req.Slots) == 0 {
		c.JSON(400, gin.H{"error": "قائمة التعديلات فارغة!"})
		return
	}

	// 3. تنفيذ التعديل
	err := db.UpdateMultipleEmployees(req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 4. الرد بالنجاح
	c.JSON(200, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("تم تحديث بيانات %d موظف بنجاح", len(req.Slots)),
	})
}

// F18 ---> API Handler لجلب جميع الحجوزات
func (db *ConnectDB) GetAllBookingsHandler(c *gin.Context) {
	var req struct {
		IDCompany int `json:"id_company" binding:"required"`
	}

	// التحقق من صحة الـ JSON المرسل
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "يجب إرسال id_company بشكل صحيح",
		})
		return
	}

	// استدعاء ميثود قاعدة البيانات
	bookings, err := db.GetAllCompanyBookings(req.IDCompany)
	if err != nil {
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "فشل في جلب البيانات",
			"details": err.Error(),
		})
		return
	}

	// الرد بالبيانات
	c.JSON(200, gin.H{
		"status": "success",
		"data":   bookings,
	})
}

// F19 ----> Handler لإضافة المواعيد
func (db *ConnectDB) AddAppointmentSlotsHandler(c *gin.Context) {
	var req AddAppointmentSlotsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "تنسيق البيانات خطأ", "details": err.Error()})
		return
	}

	err := db.AddAppointmentSlots(req.IDCompany, req.Slots)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "success", "message": "تم حفظ المواعيد بنجاح"})
}

// F20 ----> Handler لإضافة الردود الآلية
func (db *ConnectDB) AddAutoBotResponses(c *gin.Context) {
	var req AddBotResponses

	// 1. فك التشفير والتحقق من الحقول الإجبارية (JSON Validation)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "تنسيق البيانات غير صحيح",
			"details": err.Error(),
		})
		return
	}

	// 2. استدعاء ميثود قاعدة البيانات (التي استخدمنا فيها Transaction)
	err := db.AddBotResponsesAutoResponses(req.IDCompany, req.Slots)
	if err != nil {
		// يفضل طباعة الخطأ في الكونسول للمطور (Log)
		fmt.Printf("Database Error: %v\n", err)

		c.JSON(500, gin.H{
			"status":  "error",
			"message": "فشل حفظ البيانات في قاعدة البيانات",
		})
		return
	}

	// 3. رد النجاح
	c.JSON(200, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("تم حفظ %d من الردود الآلية بنجاح", len(req.Slots)),
	})
}

func (db *ConnectDB) SupInSystemHandlerOTP(c *gin.Context) {
	var req struct {
		// غيرنا النوع لـ string ليتناسب مع أرقام الجوال التي تبدأ بـ 0 أو مفتاح دولي
		PhoneOTP string `json:"phoneotp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "يجب إرسال رقم الجوال بشكل صحيح",
		})
		return
	}

	// رقم النظام الذي سيرسل الرسالة (يجب أن يكون مسجلاً في Manager.Clients)
	senderSystemPhone := "966532131074"

	// استدعاء ميثود الإرسال والتخزين في الذاكرة (60 ثانية)
	_, err := db.SendOTP(senderSystemPhone, req.PhoneOTP)
	if err != nil {
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "فشل في إرسال الـ OTP: " + err.Error(),
		})
		return
	}

	// ملاحظة: في الإنتاج لا نرسل الـ otp في الـ JSON، لكن وضعته لك للتأكد أثناء البرمجة
	c.JSON(200, gin.H{
		"status":    "success",
		"message":   "تم إرسال رمز التحقق إلى واتساب، متاح لمدة 60 ثانية",
		"debug_otp": "", // احذفه في النسخة النهائية
	})
}

func (db *ConnectDB) VerifyHandler(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "البيانات مطلوبة"})
		return
	}

	if db.VerifyOTP(req.Phone, req.Code) {
		c.JSON(200, gin.H{"status": "success", "message": "تم التحقق بنجاح!"})
	} else {
		c.JSON(401, gin.H{"status": "error", "message": "الرمز خطأ أو انتهت صلاحية الـ 60 ثانية"})
	}
}

// ------------------------------------------
//
//	AI Employee
func (db *ConnectDB) SetAIConfigHandler(c *gin.Context) {
	var input CompanyAIConfig

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	if err := db.SaveOrUpdateAIConfig(input); err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "فشل حفظ إعدادات الذكاء الاصطناعي", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "تم تحديث هوية الموظف الذكي بنجاح 🤖",
	})
}

// ---------------------------------------------------------------
// Bot is Active

func (db *ConnectDB) IsBotActive(c *gin.Context) {
	var req struct {
		IdCompany int   `json:"idcompany" binding:"required"`
		IsActive  *bool `json:"isactive" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "البيانات مطلوبة"})
		return
	}

	GlobalRegistry.SetStatus(req.IdCompany, *req.IsActive)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": " Done ✅ Save ",
	})

}
