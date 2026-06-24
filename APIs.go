package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, db *ConnectDB) {
	// for cross with web
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // يسمح لأي موقع بالوصول (مناسب للتطوير)
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// r1 --->
	r.POST("/admin/:admin_name/report", db.AdminGetAllSystemDataHandler)
	r.POST("/get-qr-admian/:id/:pass/getqr", GetQR_AdmianSystem)
	// r2 --->  Post for login
	r.POST("/login", LogIn)

	// r3 ---------->   SSE port  | مسار البورت المفتوح
	r.GET("/events", SSENetwork)
	// Finished  SSE port  | مسار البورت المفتوح   نهاية

	// Finish r4 --------> Connect with use QR  code whatsApp
	// r5  ---->  For insert package in DB
	r.POST("/add-package", db.AddPackageHandler)

	// r6 -----> Create new company
	r.POST("/register-company", db.RegisterCompanyHandler)
	//  r7 ----->  GET ALL Packages
	r.GET("/packages", db.GetAllPackagesHandler)

	/*
		// r5  ---->  For insert package in DB
		r.POST("/add-package", db.AddPackageHandler)
		//  r7 ----->  GET ALL Packages
		r.GET("/packages", db.GetAllPackagesHandler)
		//  r8  ----->  Add phone client to blocks bot
		r.POST("/block-client", db.BlockBotHandler)
		// r9 -----> Add keyword employee
		r.POST("/employee-keys", db.AddEmployeeKeywordsHandler)
		// r9 ---->  Get All keyword employee
		r.GET("/getall-employee-keys", db.GetAllEmployeeKeywordsHandler)
		// r10 -----> Add Appointment Slots for Booking
		r.POST("/add-appointment-slots", db.AddAppointmentSlotsHandler)
		// r11  ----->   For add Responses auto bot
		r.POST("/add-bot-responses", db.AddAutoBotResponses)
		//r12  ---->  For Add New Employee in the company
		r.POST("add-employee", db.AddNewEmployeeHandler)
		//  r13  ---->  For Get Info company
		r.POST("get-info-company", db.GetAllInfoCompany)
		//  r14  --->  Get All client Phone
		r.POST("/get-client-phones", db.AddGet_ClientPhones)
		// r15 --->  Get All info Chat
		r.POST("/get-allclient-chat", db.AddGet_ChatHistoryHandler)
		// r16 --->  Send And save msg from employee to client
		r.POST("/send-msg", db.AddEmployeeSendMessageHandler)
		// r17 ---> Get All Employees in Company
		r.POST("/get-all-employees", db.GetAllEmployeesHandler)
		// r18 ---> Get Single Employee Details
		r.POST("/get-employee-details", db.GetSingleEmployeeHandler)

		// r19 ---> Update Multiple Employees Data (Batch Update)
		r.POST("/update-multi-employees", db.UpdateMultiEmployeesHandler)
		// r20 -----> لوحة تحكم الأدمن الرئيسي (بسام)
		r.POST("/admin/master-report", db.AdminGetAllSystemDataHandler)
		// r21 -----> Get All Bookings for company
		r.POST("/get-all-bookings", db.GetAllBookingsHandler)
	*/

	authGroup := r.Group("/")
	authGroup.Use(AuthMiddleware())
	{

		//   r1 ---->  Get Information System
		authGroup.GET("/info", GetInfo)
		// r4 --------> Connect with use QR  code whatsApp
		authGroup.GET("/get-qr", GetQR)
		// مسارات الشركة والموظفين (المحمية)
		authGroup.POST("/set-ai-config", db.SetAIConfigHandler)
		authGroup.POST("/block-client", db.BlockBotHandler)
		authGroup.POST("/employee-keys", db.AddEmployeeKeywordsHandler)
		authGroup.GET("/getall-employee-keys", db.GetAllEmployeeKeywordsHandler)
		authGroup.POST("/add-appointment-slots", db.AddAppointmentSlotsHandler)
		authGroup.POST("/add-bot-responses", db.AddAutoBotResponses)
		authGroup.POST("/add-employee", db.AddNewEmployeeHandler) // تأكدت من إضافة / قبل الكلمة
		authGroup.POST("/get-info-company", db.GetAllInfoCompany)
		authGroup.POST("/get-client-phones", db.AddGet_ClientPhones)
		authGroup.POST("/get-allclient-chat", db.AddGet_ChatHistoryHandler)
		authGroup.POST("/send-msg", db.AddEmployeeSendMessageHandler)
		authGroup.POST("/get-all-employees", db.GetAllEmployeesHandler)
		authGroup.POST("/get-employee-details", db.GetSingleEmployeeHandler)
		authGroup.POST("/update-multi-employees", db.UpdateMultiEmployeesHandler)
		authGroup.POST("/get-all-bookings", db.GetAllBookingsHandler)

		// يمكنك حتى إنشاء مجموعة فرعية داخلية للأدمن فقط إذا أردت مستقبلاً
		authGroup.POST("/admin/master-report", db.AdminGetAllSystemDataHandler)
		//  For bot is active or no
		authGroup.POST("/api/v0/botisactive", db.IsBotActive)
	}

	// r22 -----> طلب إرسال رمز التحقق (OTP) عبر واتساب
	r.POST("/request-otp", db.SupInSystemHandlerOTP)
	// r23 -----> التحقق من صحة رمز الـ OTP المدخل
	r.POST("/verify-otp", db.VerifyHandler)

	r.StaticFS("/static_files", http.Dir("./data/web"))

	// خدمة الملفات الأساسية التي يطلبها Flutter في الـ Root
	r.StaticFile("/flutter_bootstrap.js", "./data/web/flutter_bootstrap.js")
	r.StaticFile("/manifest.json", "./data/web/manifest.json")
	r.StaticFile("/favicon.png", "./data/web/favicon.png")
	r.StaticFile("/main.dart.js", "./data/web/main.dart.js")
	r.StaticFile("/flutter_service_worker.js", "./data/web/flutter_service_worker.js")
	r.StaticFile("/version.json", "./data/web/version.json")

	// خدمة مجلدات الأصول
	r.Static("/assets", "./data/web/assets")
	r.Static("/canvaskit", "./data/web/canvaskit")

	// الصفحة الرئيسية
	r.GET("/", func(c *gin.Context) {
		c.File("./data/web/index.html")
	})

	// حل مشكلة الـ Refresh في Flutter (SPA)
	r.NoRoute(func(c *gin.Context) {
		// إذا كان الطلب ليس لـ API، أرسل index.html
		c.File("./data/web/index.html")
	})
}
