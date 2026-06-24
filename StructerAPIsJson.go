package main

import (
	"fmt"
	"sync"
	"time"
)

// ------------------------------  S0
var globalDB *ConnectDB
var BotSilenceMap sync.Map // مخزن لحفظ العملاء الصامتين: المفتاح هو رقم العميل، والقيمة هي وقت انتهاء السكوت
// --------------------------------------      S1      -----------------------------------------
var infData SystemView

type SystemView struct {
	TotalRAM          int     `json:"total_ram"`
	FreeRAM           int     `json:"free_ram"`
	UsedRAM           int     `json:"used_ram"`
	AvailableRAM      int     `json:"available_ram"`
	CPULogicalThreads int     `json:"cpu_threads"`
	CPUPhysical       int     `json:"cpu_cores"`
	CPUPercent        float64 `json:"cpu_percent"`
}

type loginReq struct {
	ID       string `json:"id"`
	Passward string `json:"passward"`
}

type Location struct {
}

//###################################################################################################################################################
//###################################################################################################################################################
//###################################################################################################################################################
//###################################################################################################################################################
//###################################################################################################################################################
//###################################################################################################################################################
//###################################################################################################################################################
//###################################################################################################################################################
// ----------------------------------------------    S2    -----------------------------------
//                    TESTING THis   ||      يمكن حذفهم في حال وجود حل افضل
//                     ترتيب الحذف
//                  D1 =>   يمكن حذفه و إضافة تعديل بسرعه  وحل للمشكله بشكل سهل
//                  D2 =>    المشكله لا يعرف هل حلها سهل او لا او صعب
//                  D3 =>    صعوبه في حل المشكله و يجب التحقق و التدقيق   في حال حلها
//###################################################################################################################################################
//###################################################################################################################################################
//###################################################################################################################################################
//###################################################################################################################################################
//###################################################################################################################################################
//###################################################################################################################################################
//###################################################################################################################################################
//###################################################################################################################################################

// D1    يمكن حلها من خلال    .env    ------------------------------------------------------------------------------- Connect with DB
const (
	DBUser     = "postgres"
	DBPassword = "12345678"
	DBHost     = "localhost"
	DBPort     = "5432"
	DBName     = "testconnectwhats"
	LogLevel   = "DEBUG"
)

func GetConnStr() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		DBUser, DBPassword, DBHost, DBPort, DBName)
}

//      --------------------------------------------------------------------------------------------------------------------------------------------  Finish connect with DB

// #########################################################################################################################
// #########################################################################################################################
// ---------------------------------------- S3  For Server-Sent Events ( SSE )  Structure  -----------------------------------------------
// #########################################################################################################################
// #########################################################################################################################
// StreamManager يدير القنوات التي تنقل البيانات للفرونت إند
/*
type StreamManager struct {
	MessageChan chan string
}

// إنشاء نسخة واحدة عامة (Singleton) ليستخدمها الجميع
var Streamer = &StreamManager{
	MessageChan: make(chan string, 100), // سعة 100 لتجنب التعليق
}
*/
// StructerAPIsJson.go (تعديل السكشن S3)

type StreamManager struct {
	// خريطة تربط ID الشركة بمجموعة من القنوات (لدعم فتح أكثر من تبويب لنفس الشركة)
	CompanyStreams map[int][]chan string
	mu             sync.RWMutex // لضمان الـ Thread-safety عند الإضافة والحذف
}

var Streamer = &StreamManager{
	CompanyStreams: make(map[int][]chan string),
}

// دالة لإرسال رسالة لشركة محددة فقط
func (sm *StreamManager) BroadcastToCompany(companyID int, msg string) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if channels, ok := sm.CompanyStreams[companyID]; ok {
		for _, ch := range channels {
			// نستخدم select لتجنب تعليق السيرفر إذا كانت القناة ممتلئة
			select {
			case ch <- msg:
			default:
				// يمكن تسجيل log هنا أن القناة ممتلئة
			}
		}
	}
}

// #########################################################################################################################
// #########################################################################################################################
// ---------------------------------------- Finished  For Server-Sent Events ( SSE )  Structure  -----------------------------------------------
// #########################################################################################################################
// #########################################################################################################################

//	-----------------------------   ----   S4  ----   ---------------------------------------------------------
//
// for Create a new company
type NewCompany struct {
	// بيانات الشركة
	CompanyName  string `json:"company_name" `
	PackageID    string `json:"package_id" `
	CompanyPhone string `json:"company_phone" `

	// بيانات الموظف (المدير/الأدمن)
	EmployeeID    string `json:"employee_id" binding:"required,min=6"`
	EmployeeName  string `json:"employee_name" `
	EmployeePhone string `json:"employee_phone" `

	// بيانات حساب الدخول
	Username string `json:"username" `
	Password string `json:"password" `
}

// ---------------------------------------------   S5   -----------------------------------------------------------
type Package struct {
	ID              string  `json:"id" `               // الـ ID النصي (مثلاً: GOLD, VIP)
	Name            string  `json:"name" `             // اسم الباقة بالعربي
	Price           float64 `json:"price" `            // سعر الباقة
	MaxSendMsg      int     `json:"max_send_msg" `     // الحد الأقصى للرسائل
	MaxEmployee     int     `json:"max_employee" `     // الحد الأقصى للموظفين
	MaxBotSend      int     `json:"max_bot_send"`      // الحد الأقصى لرسائل البوت
	MaxAISend       int     `json:"max_ai_send"`       // الحد الأقصى لرسائل الذكاء الاصطناعي
	MaxEmployeeSend int     `json:"max_employee_send"` // الحد الأقصى لرسائل الموظفين
}

// ---------------------------------------   S6   ----------------------------------------------------------
type MessageChat struct {
	From         string `json:"from" `
	To           string `json:"to" `
	Msg          string `json:"msg" `
	Date         string `json:"date" `
	PhoneCompany string `json:"phonecompany" `
	IDCompany    int    `json:"idcompany" `

	IdEmployee string `json:"idemployee"`
	IsMe       bool   `json:"is_me"`
}

// -----------------------------------------------   S7   ----------------------------------------------------
type BlockBot struct {
	IDCompany   int    `json:"id_company" `
	PhoneClient string `json:"phoneclient"`
}

// ########################################################################################################
// S8    ----->  Struct for single api for add keyword from employee and struct has array about another struct
type EmployeeKeywordInput struct { // --------> Single json struct
	IDCompany    int    `json:"id_company" `
	EmpKeyword   string `json:"emp_keyword" `
	EmployeeName string `json:"employee_name" `
	Phone        string `json:"phone"`
	IsActive     bool   `json:"is_active"`
}

type AddEmployeesRequest struct { // ----> array for EmployeeKeywordInput struct
	Employees []EmployeeKeywordInput `json:"employees" `
}

//  -------------------------------------------   Finshed S8  ----------------------------------------------------------

// -----------------------------------   S9  ------------------------------------
type AppointmentSlots struct {
	EmpKeyword string `json:"emp_keyword" `
	Keyword    string `json:"keyword" `
	DayName    string `json:"dayname" `
	SoltTime   string `json:"solttime" `
	IsActive   bool   `json:"is_active"`
}

type AddAppointmentSlotsRequest struct {
	IDCompany int                `json:"id_company" `
	Slots     []AppointmentSlots `json:"slots" `
}

// -----------------------  Finished   S9  ------------------------------

// --------------------------   S10  for Add BotResponses Auto Responses  -----------------------
type BotResponses struct {
	Keyword  string `json:"keyword" `
	Response string `json:"response" `
	IsActive bool   `json:"isactive" `
}

type AddBotResponses struct {
	IDCompany int            `json:"id_company" `
	Slots     []BotResponses `json:"slots" `
}

//  -----------------------------------  Finished  --------------------------------------

// -----------------------------    S11    ---------------------------------------------------------
type NewEmployee struct {
	EmployeeID    string `json:"employee_id" binding:"required,min=6"`
	EmployeeName  string `json:"employee_name" `
	EmployeePhone string `json:"employee_phone" `
	Description   string `json:"description"`
	NameTeam      string `json:"nameteam"`
	// بيانات حساب الدخول
	Password string `json:"password" `
}

type AddMultieNewEmployee struct {
	IDCompany int           `json:"id_company" `
	Slots     []NewEmployee `json:"slots" `
}

// --------------------------------------------------    Finished          ----------------------------------------------

type GetTimeBooking struct {
	Keyword  string `json:"keyword"`
	DayName  string `json:"day_name"`
	SlotTime string `json:"slot_time"`
}

type GetAllTimeBooking struct {
	Slots []GetTimeBooking
}

type NameEmployee struct {
	Name    string
	Keyword string
}

//***********************************************************

type SlotDetail struct {
	EmpKeyword string `json:"emp_keyword"`
	DayName    string `json:"day_name"`
	SlotTime   string `json:"slot_time"`
}

// -------------------------------   S14  ----------------------------
//
//	Get Information Company
type GetInfoCompany struct {
	IDCompany    int     `json:"id_company" `
	NumEmployee  int     `json:"numemployee"`
	NumSendMsg   int     `json:"numsendmsg"`
	NameCompany  string  `json:"namecompany"`
	IdPackage    string  `json:"idpackage"`
	MAXSendMsg   int     `json:"maxsendmsg"`
	MAXEmployee  int     `json:"maxemployee"`
	Price        float32 `json:"price"`
	NamePackages string  `json:"namepackages`
}

//	-----------------------------    S15   ---------------------------------
//
// GETALLNumberPhone
type Get_NumberPhoneAndInfo struct {
	IDCompany  *int    `json:"id_company,omitempty"`
	FromPhone  *string `json:"from_phone,omitempty"`
	ToPhone    *string `json:"to_phone,omitempty"`
	Date       *string `json:"date,omitempty"`
	Message    *string `json:"message,omitempty"`
	IsMe       *bool   `json:"is_me,omitempty"`
	IdEmployee *string `json:"id_employee,omitempty"`
}

type Get_MoreNumberPhoneAndInfo struct {
	IDCompany int                      `json:"id_company" `
	Slots     []Get_NumberPhoneAndInfo `json:"slots,omitempty"`
}

type CustomDateTime struct {
	time.Time
}

// S16 ---> Struct لتمثيل بيانات الموظف عند استرجاعها
type EmployeeDetails struct {
	ID          string `json:"id"`
	IDCompany   int    `json:"id_company"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Description string `json:"description"` // يقابل العمود role
	NameTeam    string `json:"name_team"`   // يقابل العمود role_type
}

// S17 ---> Struct لطلب موظف واحد
type GetSingleEmployeeReq struct {
	IDCompany  int    `json:"id_company" `
	IDEmployee string `json:"id_employee" `
}

// S18 ---> Struct لطلب جميع الموظفين
type GetAllEmployeesReq struct {
	IDCompany int `json:"id_company" `
}

// S19 ---> Struct لتحديث بيانات الموظف
type UpdateEmployeeReq struct {
	IDEmployee  string  `json:"id_employee" `
	Name        *string `json:"name"`
	Phone       *string `json:"phone"`
	Description *string `json:"description"`
	NameTeam    *string `json:"name_team"`
	Passward    *string `json:"passward"`
}
type UpdateMultiEmployee struct {
	IDCompany int `json:"id_company" `
	Slots     []UpdateEmployeeReq
}

// S20 ------>  Struct for Admian (Devoloper system) the system  "Bassam Salamah Dham Alshammri"  SWE
type AdmianStructDetailsAllSystem struct {
	IdCompany      *int    `json:"id_company,omitempty"`
	NameCompany    *string `json:"name_company,omitempty"`
	IdPackage      *string `json:"id_package,omitempty"`
	PhoneCompany   *string `json:"phone_company,omitempty"`
	NumSendMsg     *int    `json:"num_send_msg,omitempty"`
	NumEmp         *int    `json:"num_emp,omitempty"`
	WhatsDeviceJID *string `json:"whats_device_jid,omitempty"`
	IdEmployee     *string `json:"id_employee,omitempty"`
	NameEmployee   *string `json:"name_employee,omitempty"`
	PhoneEmployee  *string `json:"phone_employee,omitempty"`
	Description    *string `json:"description,omitempty"`
	Role           *string `json:"role,omitempty"`
}

type AdmianMultieDetails struct {
	IdAdmian       string                         `json:"IdAdmian" `
	PaswaredAdmian string                         `json:"PaswaredAdmian" `
	AdmianSlots    []AdmianStructDetailsAllSystem `json:"AdmianSlots,omitempty"`
}

// S21 ---> Struct لجلب تفاصيل الحجوزات شاملة اسم الموظف
type GetAllBookingDetails struct {
	CustomerPhone string `json:"customer_phone"`
	EmployeeID    string `json:"employee_id"`
	NameEmployee  string `json:"name_employee"`
	DateBooking   string `json:"date_booking"`
	IsActive      bool   `json:"is_active"`
}

// الستركت العام للرد على الطلب
type GetMultiBooking struct {
	IDCompany int                    `json:"id_company" `
	Slots     []GetAllBookingDetails `json:"slots"`
}

// S22 ------>  Struct for OTP Verification
var otpStore sync.Map

type otpData struct {
	code      string
	expiresAt time.Time
}

// ------------------------------------------
//
//	AI Employee
type CompanyAIConfig struct {
	IDCompany       int    `json:"id_company" binding:"required"`
	AIName          string `json:"ai_name" binding:"required"`
	AIAge           string `json:"ai_age"`
	NameCompany     string `json:"name_company"`
	MoreInfoCompany string `json:"more_info_company"`
	MoreInfo        string `json:"more_info"`
}

//  ---------------------------------------------------------
//-------------------------------------------------------------
//  For Bot Is active or no

// CompanyRegistry هيكل بيانات يحوي الخريطة وقفل الأمان
type CompanyRegistry struct {
	mu        sync.RWMutex
	companies map[int]bool
}

// GlobalRegistry هو المتغير العام الذي ستستخدمه في كل ملفات المشروع
var GlobalRegistry *CompanyRegistry

// دالة init تعمل تلقائياً عند تشغيل المشروع لتهيئة الذاكرة
func init() {
	GlobalRegistry = &CompanyRegistry{
		companies: make(map[int]bool),
	}
}

// SetStatus دالة عامة لتحديث حالة الشركة (آمنة للتزامن)
func (r *CompanyRegistry) SetStatus(companyID int, isActive bool) {
	r.mu.Lock()         // قفل كامل للكتابة
	defer r.mu.Unlock() // فتح القفل فور الانتهاء
	r.companies[companyID] = isActive
}

// IsActive دالة عامة وفائقة السرعة للتحقق من حالة الشركة
func (r *CompanyRegistry) IsActive(companyID int) bool {
	r.mu.RLock()         // قفل خفيف مخصص للقراءة فقط (يسمح بآلاف القراءات المتزامنة)
	defer r.mu.RUnlock() // فتح القفل فور الانتهاء

	active, exists := r.companies[companyID]
	return exists && active
}
