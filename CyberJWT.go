package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)


func getSecretKey() []byte {
	// تحميل ملف الإعدادات المخصص System.env
	// ملاحظة هندسية: يفضل استدعاء Load مرة واحدة في main() للأداء العالي، 
	// ولكن وضعناها هنا لتلبية طلبك في تشغيل الميثود بشكل مستقل.
	_ = godotenv.Load("System.env")

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
	 	log.Fatalf("❌ خطأ أمني حرج: لم يتم العثور على JWT_SECRET. لا يمكن تشغيل السيرفر بدون مفتاح تشفير!")

	}

	return []byte(secret)
}

type AppClaims struct {
	IDCompany    int    `json:"id_company"`
	IDEmployee   string `json:"id_employee"`
	Role         string `json:"role"`
	NameEmployee string `json:"name_employee"`
	jwt.RegisteredClaims
}

func GenerateToken(idCompany int, idEmployee string, role string, nameEmployee string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &AppClaims{
		IDCompany:    idCompany,
		IDEmployee:   idEmployee,
		Role:         role,
		NameEmployee: nameEmployee,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ARMORA_SYSTEM",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecretKey())
}


// ---------------------------------------------------------
// 2. الجدار الناري (Middleware)
// ---------------------------------------------------------
// AuthMiddleware دالة التحقق بصيغتها المختصرة والعملية
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. استخراج التوكن من الهيدر (Bearer Token)
		authHeader := c.GetHeader("Authorization")
		tokenParts := strings.Split(authHeader, " ")

		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid or missing token"})
			c.Abort() // إلغاء المهمة فوراً
			return
		}

		// 2. التحقق من التوكن (استدعاء دالة التحقق لتقليل التكرار)
		claims, err := ValidateToken(tokenParts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
			c.Abort() // إلغاء المهمة فوراً
			return
		}

		// 3. حقن البيانات في الكونتيكست لتستخدمها في أي API
		c.Set("user_claims", claims)
		c.Next() // إكمال المهمة بنجاح
	}
}

// دالة التحقق المستقلة (تستخدمها الفانكشن أعلاه وفي أي مكان آخر بالمشروع)
func ValidateToken(tokenString string) (*AppClaims, error) {
	claims := &AppClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return getSecretKey(), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ---------------------------------------------------------
// 3. أدوات مساعدة للمبرمج (Helpers)
// ---------------------------------------------------------

// دالة لاستخراج ID الشركة بأمان تام داخل الـ Controller
// بدلاً من قراءته من الـ JSON القابل للتزوير
func GetSafeCompanyID(c *gin.Context) (int, error) {
	val, exists := c.Get("id_company")
	if !exists {
		return 0, errors.New("فشل استرجاع معرف الشركة من التوكن")
	}
	// تأكيد النوع (Type Assertion) لضمان أنه رقم
	id, ok := val.(int)
	if !ok {
		return 0, errors.New("بيانات الشركة في التوكن تالفة")
	}
	return id, nil
}