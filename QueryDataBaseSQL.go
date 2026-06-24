package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Q0  ---> For Admian
func (db *ConnectDB) AdmainQueryGetAllInfoInSystem(req AdmianMultieDetails) (AdmianMultieDetails, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var finalReport AdmianMultieDetails

	// 1. التحقق من الهوية (بناءً على القيم المرسلة في req)
	if req.IdAdmian == "BassamSalamahAlshammri" && req.PaswaredAdmian == "BassamSalamahAdmianDev123$ARMTechAndERMTech0021392" {

		// نملأ بيانات الهوية في التقرير النهائي
		finalReport.IdAdmian = req.IdAdmian
		finalReport.PaswaredAdmian = "Authorized ✅"

		// 2. الكويري لجلب البيانات الشاملة
		query := `
			SELECT 
				c.id, c.name, c.id_package, c.phone_number,
				e.id, e.name, e.phone, e.role
			FROM companies c
			LEFT JOIN employees e ON c.id = e.id_company
			ORDER BY c.id`

		rows, err := db.Conn.Query(ctx, query)
		if err != nil {
			return finalReport, fmt.Errorf("❌ خطأ في الكويري: %v", err)
		}
		defer rows.Close()

		// 3. تعبئة السلوتس (المصفوفة الداخلية)
		for rows.Next() {
			var item AdmianStructDetailsAllSystem

			// Scan تستخدم العناوين (Pointers) الموجودة داخل الستركت أصلاً
			err := rows.Scan(
				&item.IdCompany,
				&item.NameCompany,
				&item.IdPackage,
				&item.PhoneCompany,
				&item.IdEmployee,
				&item.NameEmployee,
				&item.PhoneEmployee,
				&item.Role,
			)
			if err != nil {
				return finalReport, err
			}

			finalReport.AdmianSlots = append(finalReport.AdmianSlots, item)
		}

		return finalReport, nil
	}

	// في حال عدم تطابق البيانات
	return finalReport, fmt.Errorf("ERROR: Admin ID or Password is incorrect ‼️")
}

// Q1 --->  RegisterNewCompany ميثود شاملة لتسجيل شركة جديدة مع مديرها وحساب دخوله
func RegisterNewCompany(
	conn *pgxpool.Pool,
	companyName string, packageID string, companyPhone string, // بيانات الشركة
	empID string, empName string, empPhone string, // بيانات الموظف (الأدمن)
	username string, password string, // بيانات الحساب
) error {
	ctx := context.Background()

	// 1. بدء العملية (Transaction) لضمان "الكل أو لا شيء"
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("❌ فشل بدء العملية: %v", err)
	}
	// في حال حدوث خطأ، نقوم بالتراجع عن كل العمليات
	defer tx.Rollback(ctx)

	// 2. إدخال الشركة وجلب الـ ID التلقائي
	var companyID int
	companyQuery := `INSERT INTO companies (name, id_package, phone_number) VALUES ($1, $2, $3) RETURNING id`
	err = tx.QueryRow(ctx, companyQuery, companyName, packageID, companyPhone).Scan(&companyID)
	if err != nil {
		return fmt.Errorf("❌ فشل إنشاء الشركة: %v", err)
	}

	// 3. إدخال الموظف (الأدمن) باستخدام ID الموظف اليدوي
	// ملاحظة: الـ ID النصي سيخضع لقيد الـ 6 خانات الذي وضعناه في الجداول
	empQuery := `INSERT INTO employees (id, id_company, name, phone, role, role_type) 
	             VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.Exec(ctx, empQuery, empID, companyID, empName, empPhone, "مدير النظام", "ادمن")
	if err != nil {
		return fmt.Errorf("❌ فشل إنشاء الموظف (تأكد أن ID الموظف >= 6 خانات وغير مكرر): %v", err)
	}

	// 4. إنشاء حساب المستخدم تلقائياً
	userQuery := `INSERT INTO users (id_employe, username, password) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, userQuery, empID, username, password)
	if err != nil {
		return fmt.Errorf("❌ فشل إنشاء حساب المستخدم (ربما اسم المستخدم مكرر): %v", err)
	}

	// 5. اعتماد كل العمليات وحفظها نهائياً
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("❌ فشل حفظ البيانات نهائياً: %v", err)
	}

	fmt.Printf("✅ تم تسجيل الشركة [%s] والموظف [%s] بنجاح\n", companyName, empName)
	return nil
}

// Q2 ---->  func AddPackage  for add package in DB    |  لإضافة باقات لقاعدة البيانات
func AddPackage(conn *pgxpool.Pool, p Package) error {
	query := `INSERT INTO packages (id, name, price, max_send_msg, max_employee,max_bot_send,max_ai_send,max_employee_send) 
              VALUES ($1, $2, $3, $4, $5,$6,$7,$8)`

	_, err := conn.Exec(context.Background(), query, p.ID, p.Name, p.Price, p.MaxSendMsg, p.MaxEmployee, p.MaxBotSend, p.MaxAISend, p.MaxEmployeeSend)
	if err != nil {
		return fmt.Errorf("❌ فشل في إضافة الباقة: %v", err)
	}
	return nil
}

// Q3  -----> GetAllPackages جلب قائمة بجميع الباقات من قاعدة البيانات
func GetAllPackages(conn *pgxpool.Pool) ([]Package, error) {
	var packages []Package

	// الكويري لجلب كل الأعمدة من جدول الباقات
	query := `SELECT id, name, price, max_send_msg, max_employee FROM packages`

	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("❌ خطأ في جلب الباقات: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pkg Package
		err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Price, &pkg.MaxSendMsg, &pkg.MaxEmployee)
		if err != nil {
			return nil, fmt.Errorf("❌ خطأ في قراءة بيانات الباقة: %v", err)
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}

// ######################################################################################################################
//
//	Q4   ----->    Save Msg in DB
//
// SaveMessage يقوم بحفظ الرسالة في قاعدة البيانات
func (db *ConnectDB) SaveMessage(msg MessageChat) error {
	ctx := context.Background()
	// الاستعلام بناءً على هيكل الجدول في ملف useDataBaseSQL.go
	query := `
		INSERT INTO msg_whats (
			id_company, phone_company, from_phone, to_phone, 
			message, is_me, date_msg
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := db.Conn.Exec(ctx, query,
		msg.IDCompany,    // ID الشركة (رقمي)
		msg.PhoneCompany, // رقم هاتف الشركة المرتبط
		msg.From,         // من
		msg.To,           // إلى
		msg.Msg,          // نص الرسالة
		msg.IsMe,         // هل هي مني؟
		msg.Date,         // التاريخ بتنسيق نصي
	)
	return err
}

// ############################################################################################################
//
//	Q5
//
// GetCompanyIDByPhone ميثود سريعة لجلب الـ ID فقط بناءً على رقم الهاتف
func (db *ConnectDB) GetCompanyIDByPhone(phone string) (int, error) {
	var id int
	// استعلام مباشر وبسيط لاستهداف الـ ID فقط
	query := `SELECT id FROM companies WHERE phone_number = $1 LIMIT 1`

	err := db.Conn.QueryRow(context.Background(), query, phone).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// BlockClientTemporarily يضع العميل في قائمة الحظر لمدة 10 دقائق
func (db *ConnectDB) BlockClientTemporarily(phone string, companyID int) error {
	ctx := context.Background()

	// الاستعلام يحسب وقت الانتهاء الحالي + 10 دقائق
	query := `
		INSERT INTO bot_blocks (id_company, client_phone, blocked_until)
		VALUES ($1, $2, NOW() + INTERVAL '10 minutes')
		ON CONFLICT (client_phone) DO UPDATE 
		SET blocked_until = NOW() + INTERVAL '10 minutes';`

	_, err := db.Conn.Exec(ctx, query, companyID, phone)
	return err
}

// ######################################################################################
//
//	Q6 ---->    For Add Keyword from employee all employe in company has keyword       |   لإضافة مفتاح للموظف لكل موظف او طبيب في الشركه له مفتاح خاص به
func (db *ConnectDB) AddEmployeeKeywords(employees []EmployeeKeywordInput) error {
	// نستخدم تكرار لبناء الاستعلام أو نستخدم Batch من مكتبة pgx
	ctx := context.Background()

	for _, emp := range employees {
		query := `
            INSERT INTO employee_keywords (id_company, emp_keyword, employee_name, phone, is_active)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (id_company, emp_keyword) 
            DO UPDATE SET 
                employee_name = EXCLUDED.employee_name, 
                phone = EXCLUDED.phone, 
                is_active = EXCLUDED.is_active`

		_, err := db.Conn.Exec(ctx, query, emp.IDCompany, emp.EmpKeyword, emp.EmployeeName, emp.Phone, emp.IsActive)
		if err != nil {
			return err
		}
	}
	return nil
}

// Q7 -----> Insert Booking into DB
func (db *ConnectDB) InsertBooking(idCompany int, phone string, empKeyword string, dayName string, slotTime string) (string, error) {
	ctx := context.Background()

	// 1. حساب التاريخ والتأكد من عدم وجود خطأ
	fullAppointmentStr, err := GetNextAppointmentDate(dayName, slotTime)
	if err != nil {
		// في حال كان اسم اليوم أو الوقت خطأ، نرجع الخطأ فوراً
		return "", fmt.Errorf("⚠️ خطأ في تحديد الموعد: %v", err)
	}

	// 2. كويري الإدراج
	// ملاحظة هامة: تأكد أن empKeyword المرسل يطابق employee_id في جدول الموظفين
	query := `
        INSERT INTO bookings (id_company, client_phone, employee_id, appointment_date)
        VALUES ($1, $2, $3, $4)`

	_, err = db.Conn.Exec(ctx, query,
		idCompany,
		phone,
		empKeyword,
		fullAppointmentStr,
	)

	if err != nil {
		// فحص إذا كان الخطأ بسبب تكرار الموعد (Unique Constraint)
		if strings.Contains(err.Error(), "unique_constraint") || strings.Contains(err.Error(), "23505") {
			return "", fmt.Errorf("نعتذر، هذا الموعد تم حجزه مسبقاً")
		}
		return "", err
	}

	// 3. تنسيق رسالة النجاح
	vStr := fmt.Sprintf("تم الحجز بنجاح ✅\nبتاريخ [%s]", fullAppointmentStr)
	return vStr, nil
}

// Q8 ---> GetSlotDetailByKeyword جلب تفاصيل موعد محدد بناءً على الكلمة المفتاحية وحالة التفعيل
func (db *ConnectDB) GetSlotDetailByKeyword(idCompany int, keyword string, isActive bool) (SlotDetail, error) {
	ctx := context.Background()
	var detail SlotDetail

	// الكويري مع تحويل الوقت لتنسيق نصي لسهولة التعامل معه في الفرونت إند
	query := `
        SELECT emp_keyword, day_name, TO_CHAR(slot_time, 'HH24:MI') 
        FROM appointment_slots 
        WHERE id_company = $1 
          AND keyword = $2 
          AND is_active = $3
        LIMIT 1`

	err := db.Conn.QueryRow(ctx, query, idCompany, keyword, isActive).Scan(
		&detail.EmpKeyword,
		&detail.DayName,
		&detail.SlotTime,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return detail, fmt.Errorf("⚠️ لا يوجد موعد مفعل بهذه الكلمة المفتاحية")
		}
		return detail, err
	}

	return detail, nil
}

// Q9 ---> IncrementSendMsgCount يقوم بزيادة عداد الرسائل المرسلة للشركة بمقدار 1
func (db *ConnectDB) IncrementSendMsgCount(idCompany int, key int) error {
	ctx := context.Background()

	// استعلام التحديث باستخدام Atomic Increment لضمان دقة البيانات عند حدوث Concurrency
	switch key {
	case 1:
		query := `
		UPDATE companies 
		SET number_send_msg = number_send_msg + 1 
		WHERE id = $1`

		result, err := db.Conn.Exec(ctx, query, idCompany)
		if err != nil {
			return fmt.Errorf("❌ فشل تحديث عداد الرسائل: %v", err)
		}

		// التحقق مما إذا كان المعرف موجوداً بالفعل
		if result.RowsAffected() == 0 {
			return fmt.Errorf("⚠️ لم يتم العثور على شركة بهذا المعرف: %d", idCompany)
		}
		break
	case 2:
		query := `
	UPDATE companies 
	SET number_employee = number_employee + 1 
	WHERE id = $1`

		result, err := db.Conn.Exec(ctx, query, idCompany)
		if err != nil {
			return fmt.Errorf("❌ فشل تحديث عداد الرسائل: %v", err)
		}

		// التحقق مما إذا كان المعرف موجوداً بالفعل
		if result.RowsAffected() == 0 {
			return fmt.Errorf("⚠️ لم يتم العثور على شركة بهذا المعرف: %d", idCompany)
		}
		break

	}
	return nil
}

// Q10 ---->  For LogIn
func (db *ConnectDB) CheckUserLogin(username, password string) (bool, string, string, int) {
	// إضافة تايم أوت 5 ثواني للحماية من تعليق الكويري في الضغط العالي
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var roleType string
	var idCompany int
	var nameEmployee string
	// الكويري المحسنة لجلب نوع الصلاحية ومعرف الشركة في طلب واحد
	query := `
		SELECT e.role_type, e.id_company , e.name
		FROM users u
		INNER JOIN employees e ON u.id_employe = e.id
		WHERE u.username = $1 AND u.password = $2
		LIMIT 1`

	// تنفيذ الاستعلام وعمل Scan للمتغيرين
	err := db.Conn.QueryRow(ctx, query, username, password).Scan(&roleType, &idCompany, &nameEmployee)

	if err != nil {
		// حالة عدم وجود المستخدم (Wrong username or password)
		if err.Error() == "no rows in result set" || err.Error() == "scanning one row: no rows in result set" {
			return false, "", "", 0
		}
		// طباعة الأخطاء التقنية الأخرى (مثل انقطاع الاتصال)
		log.Printf("❌ Database Error in Login: %v", err)
		return false, "", "", 0
	}

	// نجاح تسجيل الدخول: نرجع true، الصلاحية، ومعرف الشركة
	return true, roleType, nameEmployee, idCompany
}

// Q11 ---->  For Get All INFO  for Company
func (db *ConnectDB) GetCompanyDetailsByID(id int) (GetInfoCompany, error) {
	var info GetInfoCompany

	// 1. تحديد مهلة زمنية (Timeout) للطلب لضمان كفاءة السيرفر
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
        SELECT 
            c.id, c.number_employee, c.number_send_msg, c.name, 
            c.id_package, p.max_send_msg, p.max_employee, p.price, p.name
        FROM companies c
        LEFT JOIN packages p ON c.id_package = p.id
        WHERE c.id = $1
    `

	// 2. التنفيذ وتعبئة البيانات (Scan)
	err := db.Conn.QueryRow(ctx, query, id).Scan(
		&info.IDCompany,
		&info.NumEmployee,
		&info.NumSendMsg,
		&info.NameCompany,
		&info.IdPackage,
		&info.MAXSendMsg,
		&info.MAXEmployee,
		&info.Price,
		&info.NamePackages,
	)

	if err != nil {
		return info, err
	}

	return info, nil
}

//	Q12  ----> GETALL number phone
//
// GetClientPhones جلب قائمة فريدة بأرقام هواتف العملاء الذين تواصلوا مع الشركة
func (db *ConnectDB) GetClientPhones(idCompany int) ([]string, error) {
	// استخدام context مع Timeout فكرة ممتازة للأداء العالي (High Performance)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// الكويري المحدثة بناءً على طلبك
	// ملاحظة هندسية: الكويري تبحث عن الرسائل الواردة فقط (is_me = FALSE)
	// لجلب أرقام العملاء حصراً وليس أرقام الشركة.
	query := `
		SELECT DISTINCT from_phone 
		FROM msg_whats 
		WHERE id_company = $1 AND is_me = FALSE`

	rows, err := db.Conn.Query(ctx, query, idCompany)
	if err != nil {
		return nil, fmt.Errorf("❌ فشل استعلام أرقام العملاء: %v", err)
	}
	defer rows.Close()

	// سعة مبدئية لتقليل عمليات تخصيص الذاكرة (Memory Allocation)
	phones := make([]string, 0, 100)

	for rows.Next() {
		var phone string
		if err := rows.Scan(&phone); err != nil {
			return nil, fmt.Errorf("❌ خطأ أثناء قراءة البيانات: %v", err)
		}
		phones = append(phones, phone)
	}

	// التحقق من سلامة البيانات بعد انتهاء الـ Loop
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return phones, nil
}

// Q13 --->  Get All Info from chat
func (db *ConnectDB) GetChatHistory(idCompany int, clientPhone string) ([]Get_NumberPhoneAndInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ترتيب الأعمدة في الـ SELECT يجب أن يطابق الـ Scan تماماً
	query := `
        SELECT from_phone, to_phone, message, is_me, date_msg, id_employee
        FROM msg_whats
        WHERE id_company = $1 
          AND (from_phone = $2 OR to_phone = $2)
        ORDER BY id ASC`

	rows, err := db.Conn.Query(ctx, query, idCompany, clientPhone)
	if err != nil {
		return nil, fmt.Errorf("❌ فشل الاستعلام: %v", err)
	}
	defer rows.Close()

	chat := make([]Get_NumberPhoneAndInfo, 0, 50) // سعة مبدئية للأداء العالي

	for rows.Next() {
		var m Get_NumberPhoneAndInfo
		// نقوم بعمل Scan مباشرة للأعمدة
		err := rows.Scan(
			&m.FromPhone,
			&m.ToPhone,
			&m.Message,
			&m.IsMe,
			&m.Date, // سيقرأ النص "2026/01/28 | 02:43" كما هو بدون مشاكل
			&m.IdEmployee,
		)
		if err != nil {
			return nil, fmt.Errorf("❌ خطأ في قراءة السطر: %v", err)
		}
		chat = append(chat, m)
	}

	return chat, nil
}

// Q14  ----> Save msg Send from employee to client
func (db *ConnectDB) AddMessageByEmployee(idCompany int, idEmployee string, toPhone string, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. توليد الوقت بالتنسيق الذي تعتمده: "2026/01/28 | 14:07"
	loc, _ := time.LoadLocation("Asia/Riyadh")
	nowStr := time.Now().In(loc).Format("2006/01/02 | 15:04")

	// 2. الكويري العبقرية: INSERT مع SELECT بنفس الوقت
	// نستخدم SELECT لجلب رقم هاتف الشركة من جدول companies بناءً على الـ id_company
	query := `
		INSERT INTO msg_whats (
			id_company, phone_company, id_employee, 
			from_phone, to_phone, message, 
			is_me, date_msg
		)
		SELECT 
			c.id,           -- id_company
			c.phone_number, -- phone_company (رقم الشركة من جدولها)
			$2,             -- id_employee
			c.phone_number, -- from_phone (بما أن الموظف أرسلها، فالمرسل هو رقم الشركة)
			$3,             -- to_phone (رقم العميل المستلم)
			$4,             -- message
			TRUE,           -- is_me (دائماً true لأنها صادرة من النظام/الموظف)
			$5              -- date_msg (الوقت المولد في السيرفر)
		FROM companies c
		WHERE c.id = $1
		LIMIT 1`

	result, err := db.Conn.Exec(ctx, query, idCompany, idEmployee, toPhone, message, nowStr)
	if err != nil {
		return fmt.Errorf("❌ خطأ أثناء الإضافة: %v", err)
	}

	// 3. التحقق هل تمت العملية؟ (لو الـ id_company غلط لن يتم إضافة شيء)
	if result.RowsAffected() == 0 {
		return fmt.Errorf("⚠️ لم يتم الإضافة، تأكد من صحة id_company")
	}

	return nil
}

// Q15 -----> GetAllEmployeesByCompany جلب كافة موظفي الشركة
func (db *ConnectDB) GetAllEmployeesByCompany(idCompany int) ([]UpdateEmployeeReq, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// نستخدم LEFT JOIN لجلب كلمة المرور من جدول المستخدمين
	// الاسم في جدول users هو password وفي الستركت Passward
	query := `
		SELECT e.id, e.name, e.phone, e.role, e.role_type, u.password
		FROM employees e
		LEFT JOIN users u ON e.id = u.id_employe
		WHERE e.id_company = $1`

	rows, err := db.Conn.Query(ctx, query, idCompany)
	if err != nil {
		return nil, fmt.Errorf("❌ خطأ في استعلام الموظفين: %v", err)
	}
	defer rows.Close()

	employees := make([]UpdateEmployeeReq, 0)

	for rows.Next() {
		// نحتاج متغيرات مؤقتة لاستقبال البيانات لأن الحقول في الستركت هي مؤشرات (Pointers)
		var id string
		var name, phone, description, nameTeam, passward *string

		// ملاحظة: الـ Scan في pgx يتعامل مع المؤشرات بشكل ممتاز
		err := rows.Scan(&id, &name, &phone, &description, &nameTeam, &passward)
		if err != nil {
			return nil, fmt.Errorf("❌ خطأ في قراءة السطر: %v", err)
		}

		employees = append(employees, UpdateEmployeeReq{
			IDEmployee:  id,
			Name:        name,
			Phone:       phone,
			Description: description,
			NameTeam:    nameTeam,
			Passward:    passward,
		})
	}

	return employees, nil
}

// Q16 -----> GetEmployeeByID جلب بيانات موظف محدد داخل الشركة
func (db *ConnectDB) GetEmployeeByID(idCompany int, empID string) (UpdateEmployeeReq, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var emp UpdateEmployeeReq
	var id string // لأن IDEmployee في الستركت string وليس مؤشر

	query := `
		SELECT e.id, e.name, e.phone, e.role, e.role_type, u.password
		FROM employees e
		LEFT JOIN users u ON e.id = u.id_employe
		WHERE e.id_company = $1 AND e.id = $2
		LIMIT 1`

	err := db.Conn.QueryRow(ctx, query, idCompany, empID).Scan(
		&id,
		&emp.Name,
		&emp.Phone,
		&emp.Description,
		&emp.NameTeam,
		&emp.Passward,
	)

	if err != nil {
		return emp, err
	}

	emp.IDEmployee = id
	return emp, nil
}

// Q18 -----> UpdateMultipleEmployees تحديث الموظفين وجدول المستخدمين دفعة واحدة
func (db *ConnectDB) UpdateMultipleEmployees(req UpdateMultiEmployee) error {
	ctx := context.Background()

	// 1. بدء الـ Transaction لضمان تزامن التحديث بين الجدولين
	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("❌ فشل بدء الـ Transaction: %v", err)
	}
	defer tx.Rollback(ctx) // التراجع التلقائي في حال حدوث خطأ قبل الـ Commit

	for _, emp := range req.Slots {
		// --- أ: تحديث جدول الموظفين (employees) ---
		updates := []string{}
		args := []interface{}{}
		argIdx := 1

		if emp.Name != nil {
			updates = append(updates, fmt.Sprintf("name = $%d", argIdx))
			args = append(args, *emp.Name)
			argIdx++
		}
		if emp.Phone != nil {
			updates = append(updates, fmt.Sprintf("phone = $%d", argIdx))
			args = append(args, *emp.Phone)
			argIdx++
		}
		if emp.Description != nil {
			updates = append(updates, fmt.Sprintf("role = $%d", argIdx))
			args = append(args, *emp.Description)
			argIdx++
		}
		if emp.NameTeam != nil {
			updates = append(updates, fmt.Sprintf("role_type = $%d", argIdx))
			args = append(args, *emp.NameTeam)
			argIdx++
		}

		if len(updates) > 0 {
			query := fmt.Sprintf("UPDATE employees SET %s WHERE id_company = $%d AND id = $%d",
				strings.Join(updates, ", "), argIdx, argIdx+1)

			finalArgs := append(args, req.IDCompany, emp.IDEmployee)
			_, err := tx.Exec(ctx, query, finalArgs...)
			if err != nil {
				return fmt.Errorf("❌ خطأ في تحديث بيانات الموظف %s: %v", emp.IDEmployee, err)
			}
		}

		// --- ب: تحديث كلمة المرور في جدول (users) ---
		// يتم الدخول هنا فقط إذا أرسل المستخدم قيمة في حقل Passward
		if emp.Passward != nil {
			// نستخدم id_employe للربط (وهو يطابق IDEmployee)
			userQuery := `UPDATE users SET password = $1 WHERE id_employe = $2`
			_, err := tx.Exec(ctx, userQuery, *emp.Passward, emp.IDEmployee)
			if err != nil {
				return fmt.Errorf("❌ فشل تحديث كلمة المرور للموظف %s في جدول users: %v", emp.IDEmployee, err)
			}
		}
	}

	// 2. اعتماد كافة التغييرات نهائياً في الجدوليين
	return tx.Commit(ctx)
}

// Q19 ---> نسخة معدلة ومضمونة لجلب الحجوزات مع بيانات الموظف
func (db *ConnectDB) GetAllCompanyBookings(idCompany int) (GetMultiBooking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var response GetMultiBooking
	response.IDCompany = idCompany
	// نнициаليزيشن للمصفوفة لتجنب رجوع null في الـ JSON إذا كانت فارغة
	response.Slots = make([]GetAllBookingDetails, 0)

	query := `
        SELECT 
            b.client_phone, 
            b.employee_id, 
            COALESCE(ek.employee_name, 'موظف غير معروف') as employee_name, 
            TO_CHAR(b.appointment_date, 'YYYY-MM-DD HH24:MI') as appointment_date, 
            b.is_active
        FROM bookings b
        LEFT JOIN employee_keywords ek 
            ON b.id_company = ek.id_company 
            AND b.employee_id = ek.emp_keyword
        WHERE b.id_company = $1
        ORDER BY b.appointment_date DESC`

	rows, err := db.Conn.Query(ctx, query, idCompany)
	if err != nil {
		return response, fmt.Errorf("database_error: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var detail GetAllBookingDetails
		err := rows.Scan(
			&detail.CustomerPhone,
			&detail.EmployeeID,
			&detail.NameEmployee,
			&detail.DateBooking,
			&detail.IsActive,
		)
		if err != nil {
			return response, err
		}
		response.Slots = append(response.Slots, detail)
	}

	return response, nil
}

// AddMultipleEmployeesBulk إضافة مجموعة موظفين في عملية واحدة (Bulk Insert)
func (db *ConnectDB) AddMultipleEmployeesBulk(input AddMultieNewEmployee) error {
	ctx := context.Background()

	// 1. بدء العملية الجماعية
	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("❌ فشل بدء العملية: %v", err)
	}
	defer tx.Rollback(ctx)

	// 2. التحقق من وجود الشركة (مرة واحدة فقط للكل)
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1)`
	err = tx.QueryRow(ctx, checkQuery, input.IDCompany).Scan(&exists)
	if err != nil {
		return fmt.Errorf("❌ خطأ أثناء التحقق من وجود الشركة: %v", err)
	}
	if !exists {
		return fmt.Errorf("⚠️ الشركة رقم %d غير موجودة", input.IDCompany)
	}

	// 3. الدوران حول الموظفين وإضافتهم داخل نفس الـ Transaction
	for _, emp := range input.Slots {
		if emp.Description == "" || emp.NameTeam == "" {
			emp.Description = "Nothing"
		}

		if emp.NameTeam != "ادمن" && emp.NameTeam != "اداري" && emp.NameTeam != "علاقة عملاء" {
			// إذا كانت فارغة أو "Nothing" أو أي قيمة أخرى خربوطية، نحولها للأقل صلاحيات
			emp.NameTeam = "علاقة عملاء"
		}
		// أ) إضافة الموظف
		empQuery := `INSERT INTO employees (id, id_company, name, phone, role, role_type) 
                     VALUES ($1, $2, $3, $4, $5, $6)`
		_, err = tx.Exec(ctx, empQuery, emp.EmployeeID, input.IDCompany, emp.EmployeeName, emp.EmployeePhone, emp.Description, emp.NameTeam)
		if err != nil {
			return fmt.Errorf("❌ فشل إضافة الموظف %s: %v", emp.EmployeeName, err)
		}

		// ب) إضافة حساب المستخدم
		userQuery := `INSERT INTO users (id_employe, username, password) VALUES ($1, $2, $3)`
		_, err = tx.Exec(ctx, userQuery, emp.EmployeeID, emp.EmployeeID, emp.Password)
		if err != nil {
			return fmt.Errorf("❌ فشل إنشاء حساب للموظف %s (ربما اليوزر مكرر): %v", emp.EmployeeName, err)
		}
		db.IncrementSendMsgCount(input.IDCompany, 2)
	}

	// 4. تثبيت كل العمليات دفعة واحدة
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("❌ فشل حفظ البيانات النهائية: %v", err)
	}

	return nil
}

func (db *ConnectDB) GetAllEmployeeKeywords(companyID int) ([]EmployeeKeywordInput, error) {
	ctx := context.Background()
	var employees []EmployeeKeywordInput

	// الترتيب: 1.id_company, 2.emp_keyword, 3.employee_name...
	query := `SELECT id_company, emp_keyword, employee_name, phone, is_active 
              FROM employee_keywords 
              WHERE id_company = $1`

	rows, err := db.Conn.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var emp EmployeeKeywordInput

		err := rows.Scan(
			&emp.IDCompany,
			&emp.EmpKeyword,
			&emp.EmployeeName,
			&emp.Phone,
			&emp.IsActive,
		)
		if err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}
	return employees, nil
}

func (db *ConnectDB) AddAppointmentSlots(idCompany int, slots []AppointmentSlots) error {
	ctx := context.Background()

	for _, slot := range slots {
		query := `
            INSERT INTO appointment_slots (id_company, emp_keyword, keyword, day_name, slot_time, is_active)
            VALUES ($1, $2, $3, $4, $5, $6)
            ON CONFLICT (id_company, keyword) 
            DO UPDATE SET 
                emp_keyword = EXCLUDED.emp_keyword,
                day_name = EXCLUDED.day_name,
                slot_time = EXCLUDED.slot_time,
                is_active = EXCLUDED.is_active`
		_, err := db.Conn.Exec(ctx, query, idCompany, slot.EmpKeyword, slot.Keyword, slot.DayName, slot.SoltTime, slot.IsActive)
		if err != nil {
			return fmt.Errorf("خطأ في الموعد %s: %v", slot.Keyword, err)
		}

	}
	return nil
}

func (db *ConnectDB) AddBotResponsesAutoResponses(idCompany int, slots []BotResponses) error {
	ctx := context.Background()

	// 1. بدء عملية (Transaction) لضمان السرعة والأمان
	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("❌ فشل بدء العملية: %v", err)
	}
	// في حال حدوث خطأ مفاجئ، يتم التراجع عن كل شيء
	defer tx.Rollback(ctx)

	for _, slot := range slots {
		query := `
            INSERT INTO bot_responses (id_company, keyword, response, is_active)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (id_company, keyword) 
            DO UPDATE SET 
                response = EXCLUDED.response,
                is_active = EXCLUDED.is_active` // حذفنا تحديث keyword لأنه غير ضروري هنا

		_, err := tx.Exec(ctx, query, idCompany, slot.Keyword, slot.Response, slot.IsActive)
		if err != nil {
			return fmt.Errorf("❌ خطأ في الكلمة (%s): %v", slot.Keyword, err)
		}
	}

	// 2. اعتماد العملية نهائياً في قاعدة البيانات
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("❌ فشل حفظ البيانات النهائية: %v", err)
	}

	return nil
}

func (db *ConnectDB) GetBotResponse(idCompany int, keyword string) (string, error) {
	ctx := context.Background()

	var response string
	var isActive bool

	// استعلام لجلب الرد وحالة التفعيل بناءً على الشركة والكلمة المفتاحية
	query := `
        SELECT response, is_active 
        FROM bot_responses 
        WHERE id_company = $1 AND keyword = $2
        LIMIT 1`

	err := db.Conn.QueryRow(ctx, query, idCompany, keyword).Scan(&response, &isActive)

	if err != nil {
		// إذا لم توجد الكلمة أصلاً في الجدول
		if err.Error() == "no rows in result set" {

			requestAIemployee, err := db.GetAIConfigByCompanyID(idCompany)
			requestAIemployee.SendMsg = keyword
			if err != nil {
				return "عذراً، هذه الكلمة غير مسجلة لدينا.", nil
			}

		}
		return "", fmt.Errorf("❌ خطأ أثناء البحث عن الرد: %v", err)
	}

	// المنطق الذي طلبته: التأكد من حالة التفعيل
	if !isActive {
		return "غير متوفر حالياً", nil
	}

	// إذا كانت فعالة، نرجع الرد المسجل في الجدول
	return response, nil
}

// GetBlockedClients تجلب قائمة بكل الأرقام المحظورة لشركة معينة
func (db *ConnectDB) GetBlockedClients(idCompany int) ([]string, error) {
	ctx := context.Background()
	var phones []string

	// الاستعلام يجلب الأرقام التي لا يزال وقت حظرها مستقبلياً
	query := `
		SELECT client_phone 
		FROM bot_blocks 
		WHERE id_company = $1 AND blocked_until > NOW()`

	rows, err := db.Conn.Query(ctx, query, idCompany)
	if err != nil {
		return nil, fmt.Errorf("❌ خطأ أثناء جلب قائمة المحظورين: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var phone string
		if err := rows.Scan(&phone); err != nil {
			continue
		}
		phones = append(phones, phone)
	}

	return phones, nil
}

// IsClientBlocked تتحقق ما إذا كان رقم هاتف معين محظوراً حالياً لشركة محددة
func (db *ConnectDB) IsClientBlocked(idCompany int, clientPhone string) (bool, error) {
	ctx := context.Background()
	var exists bool

	// الاستعلام يبحث عن وجود الرقم بشرط أن تاريخ الحظر (blocked_until) لم ينتهِ بعد
	query := `
		SELECT EXISTS (
			SELECT 1 FROM bot_blocks 
			WHERE id_company = $1 AND client_phone = $2 AND blocked_until > NOW()
		);`

	err := db.Conn.QueryRow(ctx, query, idCompany, clientPhone).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("❌ خطأ أثناء فحص حالة الحظر: %v", err)
	}

	return exists, nil
}

// UnblockClient تقوم بحذف رقم معين من جدول المحظورين لشركة محددة
func (db *ConnectDB) UnblockClient(idCompany int, phone string) error {
	ctx := context.Background()

	query := `DELETE FROM bot_blocks WHERE id_company = $1 AND client_phone = $2`

	result, err := db.Conn.Exec(ctx, query, idCompany, phone)
	if err != nil {
		return fmt.Errorf("❌ خطأ أثناء فك الحظر: %v", err)
	}

	// التحقق إذا كان الرقم موجوداً أصلاً وتم حذفه
	if result.RowsAffected() == 0 {
		return fmt.Errorf("⚠️ الرقم غير موجود في قائمة المحظورين")
	}

	fmt.Printf("✅ تم فك الحظر عن الرقم %s بنجاح\n", phone)
	return nil
}

func (db *ConnectDB) GetEmployeeSlots(idCompany int, empKeyword string) (GetAllTimeBooking, error) {
	ctx := context.Background()
	var allSlots GetAllTimeBooking // الستركت العام

	// الكويري المطلوبة
	query := `
        SELECT keyword, day_name, TO_CHAR(slot_time, 'HH24:MI') as slot_time
        FROM appointment_slots
        WHERE id_company = $1 
          AND emp_keyword = $2 
          AND is_active = TRUE
        ORDER BY day_name, slot_time ASC`

	rows, err := db.Conn.Query(ctx, query, idCompany, empKeyword)
	if err != nil {
		return allSlots, err
	}
	defer rows.Close()

	for rows.Next() {
		var slot GetTimeBooking
		// مسح البيانات ووضعها في الستركت الصغير
		err := rows.Scan(
			&slot.Keyword,
			&slot.DayName,
			&slot.SlotTime,
		)
		if err != nil {
			return allSlots, err
		}
		// إضافة الستركت الصغير إلى مصفوفة الستركت الكبير
		allSlots.Slots = append(allSlots.Slots, slot)
	}

	return allSlots, nil
}

func (db *ConnectDB) GetPhoneByCompanyID(companyID int) (string, error) {
	var phone string
	// استعلام مباشر وبسيط لاستهداف الـ ID فقط
	query := `SELECT phone_number FROM companies WHERE id = $1 LIMIT 1`

	err := db.Conn.QueryRow(context.Background(), query, companyID).Scan(&phone)
	if err != nil {
		return "", err
	}
	return phone, nil
}

// ------------------------------------------
//
//	AI Employee
func (db *ConnectDB) SaveOrUpdateAIConfig(config CompanyAIConfig) error {
	query := `
	INSERT INTO company_ai_info (id_company, ai_name, ai_age, name_company, more_info_company, more_info, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	ON CONFLICT (id_company) 
	DO UPDATE SET 
		ai_name = EXCLUDED.ai_name,
		ai_age = EXCLUDED.ai_age,
		name_company = EXCLUDED.name_company,
		more_info_company = EXCLUDED.more_info_company,
		more_info = EXCLUDED.more_info,
		updated_at = CURRENT_TIMESTAMP;`

	_, err := db.Conn.Exec(context.Background(), query,
		config.IDCompany,
		config.AIName,
		config.AIAge,
		config.NameCompany,
		config.MoreInfoCompany,
		config.MoreInfo,
	)
	return err
}

func (db *ConnectDB) GetAIConfigByCompanyID(companyID int) (Info_AI, error) {
	var aiInfo Info_AI
	ctx := context.Background()

	// الكويري لجلب البيانات بناءً على الـ ID
	query := `
        SELECT ai_name, ai_age, name_company, more_info_company, more_info 
        FROM company_ai_info 
        WHERE id_company = $1
    `

	// تنفيذ الكويري وعمل Scan مباشرة للستركت
	err := db.Conn.QueryRow(ctx, query, companyID).Scan(
		&aiInfo.Name,
		&aiInfo.Age,
		&aiInfo.NameCompany,
		&aiInfo.MoreInfoCompany,
		&aiInfo.MoreInfo,
	)

	if err != nil {
		return aiInfo, err
	}

	return aiInfo, nil
}
