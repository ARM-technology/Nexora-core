package main

import (
	"context"
	"time"

	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func DeleteOldBookings(pool *pgxpool.Pool) error {
	// تحميل منطقة الرياض الزمنية
	loc, err := time.LoadLocation("Asia/Riyadh")
	if err != nil {
		// في حال فشل تحميل المنطقة الزمنية (مثلاً في بعض حاويات Docker)، نستخدم التوقيت العالمي كاحتياط
		log.Printf("⚠️ تحذير: فشل تحميل منطقة Asia/Riyadh، سيتم استخدام التوقيت المحلي للسيرفر: %v", err)
		loc = time.Local
	}

	// حساب الوقت قبل ساعتين من الآن بناءً على توقيت الرياض
	threshold := time.Now().In(loc).Add(-2 * time.Hour)

	// الاستعلام يحذف أي حجز تاريخه أقدم من (عتبة الساعتين الماضية)
	query := `DELETE FROM bookings WHERE appointment_date < $1`

	result, err := pool.Exec(context.Background(), query, threshold)
	if err != nil {
		log.Printf("❌ خطأ أثناء حذف الحجوزات القديمة: %v", err)
		return err
	}

	// طباعة عدد الصفوف المحذوفة للمراقبة في الـ Log
	if result.RowsAffected() > 0 {
		log.Printf("🧹 تنظيف: تم حذف %d حجوزات منتهية (أقدم من: %s)", result.RowsAffected(), threshold.Format("15:04:05"))
	}

	return nil
}
