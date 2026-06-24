package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitDatabase تقوم ببناء الهيكل مع جعل ID الموظف نصي يدوي لا يقل عن 6 خانات

func InitDatabase(conn *pgxpool.Pool) error {
	query := `
	-- 1. جدول الباقات
	CREATE TABLE IF NOT EXISTS packages (
		id VARCHAR(50) PRIMARY KEY, 
		name VARCHAR(100) NOT NULL,
		price DECIMAL(10, 2) NOT NULL,
		max_send_msg INT NOT NULL,
		max_employee INT NOT NULL,
      max_bot_send INT NOT NULL,       -- الحد الأقصى للبوت
    max_ai_send INT NOT NULL,        -- الحد الأقصى للذكاء الاصطناعي
    max_employee_send INT NOT NULL   -- الحد الأقصى للموظفين
	);

	-- 2. جدول الشركات (تم إضافة UNIQUE لرقم الهاتف)
	CREATE TABLE IF NOT EXISTS companies (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		id_package VARCHAR(50) REFERENCES packages(id) ON DELETE SET NULL, 
		phone_number VARCHAR(20) UNIQUE NOT NULL, -- UNIQUE يمنع تكرار رقم الشركة
		number_employee INT DEFAULT 0,
	number_send_msg INT DEFAULT 0,
    number_bot_send INT DEFAULT 0,
    number_ai_send INT DEFAULT 0,
    number_employee_send INT DEFAULT 0
	);

    -- 1. دالة المنطق البرمجي للتحقق من الكوتا
CREATE OR REPLACE FUNCTION check_company_limits()
RETURNS TRIGGER AS $$
DECLARE
    pkg RECORD;
BEGIN
    -- جلب حدود الباقة المرتبطة بالشركة
    SELECT max_send_msg, max_bot_send, max_ai_send, max_employee_send 
    INTO pkg 
    FROM packages 
    WHERE id = NEW.id_package;

    -- التحقق من حد رسائل البوت
    IF NEW.number_bot_send > pkg.max_bot_send THEN
        RAISE EXCEPTION 'تجاوزت الشركة الحد الأقصى لرسائل البوت المسموح بها في الباقة (%)', pkg.max_bot_send;
    END IF;

    -- التحقق من حد رسائل الذكاء الاصطناعي
    IF NEW.number_ai_send > pkg.max_ai_send THEN
        RAISE EXCEPTION 'تجاوزت الشركة الحد الأقصى لرسائل الذكاء الاصطناعي المسموح بها في الباقة (%)', pkg.max_ai_send;
    END IF;

    -- التحقق من حد رسائل الموظفين
    IF NEW.number_employee_send > pkg.max_employee_send THEN
        RAISE EXCEPTION 'تجاوزت الشركة الحد الأقصى لرسائل الموظفين المسموح بها في الباقة (%)', pkg.max_employee_send;
    END IF;

    -- التحقق من الحد الإجمالي العام (اختياري حسب رغبتك)
    IF NEW.number_send_msg > pkg.max_send_msg THEN
        RAISE EXCEPTION 'تجاوزت الشركة الحد الإجمالي العام للرسائل (%)', pkg.max_send_msg;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 2. ربط الدالة بجدول الشركات ليتم تنفيذها قبل أي عملية تحديث (UPDATE)
DROP TRIGGER IF EXISTS trg_limit_company_usage ON companies;
CREATE TRIGGER trg_limit_company_usage
BEFORE UPDATE ON companies
FOR EACH ROW
EXECUTE FUNCTION check_company_limits();

	-- 3. جدول الموظفين
	CREATE TABLE IF NOT EXISTS employees (
		id VARCHAR(50) PRIMARY KEY CHECK (length(id) >= 6), 
		id_company INT REFERENCES companies(id) ON DELETE CASCADE,
		name VARCHAR(100) NOT NULL,
		phone VARCHAR(20),
		role VARCHAR(50), 
		role_type VARCHAR(20) NOT NULL CHECK (role_type IN ('ادمن', 'اداري', 'علاقة عملاء'))
	);

	-- 4. جدول المستخدمين
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		id_employe VARCHAR(50) UNIQUE REFERENCES employees(id) ON DELETE CASCADE, 
		username VARCHAR(50) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL
	);

	
	-- 5. جدول مفاتيح الأطباء (يجب أن ينشأ أولاً لأنه المرجع)
	CREATE TABLE IF NOT EXISTS employee_keywords (
    id_company INT REFERENCES companies(id) ON DELETE CASCADE,
    emp_keyword VARCHAR(50) NOT NULL,    -- الكود (مثل 1111)
    employee_name VARCHAR(100) NOT NULL, -- اسم الطبيب
    phone VARCHAR(20),                   
    is_active BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (id_company, emp_keyword) 
);



	-- 6. جدول سجل الرسائل (تم تحديثه ليطابق الهيكل الجديد MessageChat)
	CREATE TABLE IF NOT EXISTS msg_whats (
		id SERIAL PRIMARY KEY,
		id_company INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE, -- إلزامي
		phone_company VARCHAR(20) NOT NULL, -- إلزامي (رقم الشركة المشغلة)
		id_employee VARCHAR(50) REFERENCES employees(id), -- اختياري (NULL يعني نظام أو وارد)
		from_phone VARCHAR(20) NOT NULL, -- إلزامي
		to_phone VARCHAR(20) NOT NULL,   -- إلزامي
		message TEXT NOT NULL,           -- إلزامي (يدعم العربية والإيموجي)
		is_me BOOLEAN NOT NULL,          -- إلزامي (هل الرسالة صادرة من النظام؟)
		date_msg VARCHAR(50) NOT NULL,   -- إلزامي (التاريخ بتنسيق النص كما تطلبه)
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- وقت الإدخال الفعلي في السيرفر
	);

	-- 7. نظام الحظر المؤقت للبوت
CREATE TABLE IF NOT EXISTS bot_blocks (
        id SERIAL PRIMARY KEY,
        id_company INT REFERENCES companies(id) ON DELETE CASCADE,
        client_phone VARCHAR(20) UNIQUE NOT NULL, -- أضفنا UNIQUE هنا لحل مشكلة ON CONFLICT
        blocked_until TIMESTAMP NOT NULL
    );

	-- 8. جدول ردود البوت (Composite Primary Key)
CREATE TABLE IF NOT EXISTS bot_responses (
    id_company INT REFERENCES companies(id) ON DELETE CASCADE,
    keyword VARCHAR(255) NOT NULL, 
    response TEXT NOT NULL,        
    is_active BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (id_company, keyword) 
);

-- 9. جدول الحجوزات (مربوط بجدول مفاتيح الأطباء مباشرة)
CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    id_company INT NOT NULL,
    client_phone VARCHAR(20) NOT NULL,
    -- هذا الحقل سيخزن الكود (مثل 1111) المأخوذ من emp_keyword
    employee_id VARCHAR(50) NOT NULL, 
    appointment_date TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,

    -- الربط الصحيح بجدول مفاتيح الأطباء باستخدام المفتاح المركب
    CONSTRAINT fk_booking_employee_keyword 
        FOREIGN KEY (id_company, employee_id) 
        REFERENCES employee_keywords(id_company, emp_keyword) 
        ON DELETE CASCADE,

    -- لضمان عدم حجز نفس الطبيب في نفس الوقت والشركة
    UNIQUE(id_company, employee_id, appointment_date)
);


-- 10. جدول المواعيد المتاحة (مرتبط بكود الطبيب)
CREATE TABLE IF NOT EXISTS appointment_slots (
    id_company INT REFERENCES companies(id) ON DELETE CASCADE,
    emp_keyword VARCHAR(50) NOT NULL, -- الربط مع جدول الأطباء
    keyword VARCHAR(50) NOT NULL,     -- كود الموعد (مثل 1)
    day_name VARCHAR(20) NOT NULL,
    slot_time TIME NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (id_company, keyword),
    -- الربط الشرطي: التأكد أن كود الطبيب يتبع لنفس الشركة
    FOREIGN KEY (id_company, emp_keyword) 
        REFERENCES employee_keywords(id_company, emp_keyword) 
        ON DELETE CASCADE
);

-- 11. جدول معلومات الذكاء الاصطناعي الخاص بالشركة  AI Employee    ذكاء صناعي كموظف
CREATE TABLE IF NOT EXISTS company_ai_info (
    -- جعلنا المعرف هو نفسه معرف الشركة لضمان وجود ملف واحد فقط لكل شركة
    id_company INT PRIMARY KEY REFERENCES companies(id) ON DELETE CASCADE,
    ai_name VARCHAR(100) NOT NULL,          -- اسم البوت (مثلاً: مساعد ذكي)
    ai_age VARCHAR(50),                     -- عمر البوت أو هويته (اختياري)
    name_company VARCHAR(150),              -- اسم الشركة كما يريد البوت تقديمه
    more_info_company TEXT,                 -- نبذة عن نشاط الشركة للبوت
    more_info TEXT,                         -- تعليمات إضافية للبوت (System Prompt)
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


	`

	_, err := conn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("❌ فشل في بناء الجداول: %v", err)
	}

	fmt.Println("✅ تم بناء الهيكل بنجاح: دعم العربية، منع تكرار أرقام الشركات، وتحديث جدول الرسائل.")
	return nil
}

// DropAllTables تقوم بحذف جميع الجداول  من قاعدة البيانات
func DropAllTables(conn *pgxpool.Pool) error {
	// استخدمنا CASCADE لضمان حذف القيود والعلاقات المرتبطة تلقائياً
	query := `
    DROP TABLE IF EXISTS appointment_slots CASCADE;
    DROP TABLE IF EXISTS employee_keywords CASCADE;
    DROP TABLE IF EXISTS bot_responses CASCADE;
    DROP TABLE IF EXISTS bot_blocks CASCADE;
    DROP TABLE IF EXISTS msg_whats CASCADE;
    DROP TABLE IF EXISTS bookings CASCADE;
    DROP TABLE IF EXISTS users CASCADE;
    DROP TABLE IF EXISTS employees CASCADE;
    DROP TABLE IF EXISTS companies CASCADE;
    DROP TABLE IF EXISTS packages CASCADE;
	DROP TABLE IF EXISTS company_ai_info CASCADE; -- الجدول الجديد الذي أضفناه
    `

	// تنفيذ عملية الحذف
	_, err := conn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("❌ فشل في حذف الجداول: %v", err)
	}

	fmt.Println("⚠️ تم حذف جميع الجداول بنجاح من قاعدة البيانات")
	return nil
}
