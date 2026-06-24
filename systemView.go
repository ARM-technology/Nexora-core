package main
import (
	"fmt"
"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"strings"
	"time"
)

func systemView() {  //     This for View System Info  CPU RAM ,... another that  |   هذه الميثود  لعرض بيانات و معلومات الجهاز بحيث يتم استعماله    و استعمال الموارد بعنايه و اداء عالي     
	v, _ := mem.VirtualMemory()
	// I use for View Space GiB  not GB 
	total := v.Total / 1073741824
	free := v.Free / 1073741824
	used := v.Used / 1073741824
	available := v.Available / 1073741824
	// I use for View Space GiB  not GB 

cpuLogicalThreads , _ :=cpu.Counts(true)
cpuPhysical,_:=cpu.Counts(false)
percent,_:=cpu.Percent(0,true)

// Here Save All info inside Struct |  هنا يتم الحفظ في هيكل بيانات لإستعماله في مختلف المشروع 
infData.TotalRAM = int(total)
infData.FreeRAM=int(free)
infData.UsedRAM=int(used)
infData.AvailableRAM=int(available)
infData.CPULogicalThreads=int(cpuLogicalThreads)
infData.CPUPhysical=int(cpuPhysical)
//  Finished save   |     نهاية الحفظ      

fmt.Printf("Total RAM : %v GiB \n Free RAM : %v GiB \n Used RAM : %v GiB \n Available RAM : %v GiB \n",total,free,used,available)
fmt.Printf(" CPU Logical Thread: %v \n CPU Physical : %v",cpuLogicalThreads,cpuPhysical)
fmt.Printf("\n CPU Percent: %v ",percent)
}

func getCPUPhysical() int {
	cpuPhysical,_:=cpu.Counts(false)
	return cpuPhysical
}

func getCPULogicalThreads() int {
	cpuLogicalThreads,_:=cpu.Counts(true)
	return cpuLogicalThreads
}




//#########################################################################################################################
//   Func for Select date     this use for booking auto 


// GetNextAppointmentDate يحسب التاريخ والوقت الكامل للموعد القادم
func GetNextAppointmentDate(dayName string, slotTime string) (string, error) {
	// تحويل أسماء الأيام العربية إلى إنجليزية للتعامل مع مكتبة time
	daysMap := map[string]time.Weekday{
		"الأحد": time.Sunday, "الاثنين": time.Monday, "الثلاثاء": time.Tuesday,
		"الأربعاء": time.Wednesday, "الخميس": time.Thursday, "الجمعة": time.Friday, "السبت": time.Saturday,
	}

	targetDay, ok := daysMap[dayName]
	if !ok {
		return "", fmt.Errorf("اسم اليوم غير صالح: %s", dayName)
	}

	// جلب الوقت الحالي وتصفير الثواني
	now := time.Now()
	
	// تحليل الوقت المرسل (يفترض تنسيق 15:04)
	t, err := time.Parse("15:04", slotTime)
	if err != nil {
		return "", fmt.Errorf("تنسيق الوقت غير صحيح: %s", slotTime)
	}

	// البحث عن التاريخ القادم لهذا اليوم
	daysUntil := int(targetDay - now.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}
	
	nextDate := now.AddDate(0, 0, daysUntil)
	// دمج التاريخ مع الوقت المطلوب
	finalDate := time.Date(nextDate.Year(), nextDate.Month(), nextDate.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)

	return finalDate.Format("2006-01-02 15:04:05"), nil
}

// ميثود مساعدة لتنظيف النص ومعالجة الهمزات (أ، إ، آ -> ا)
func normalizeDayName(day string) string {
	d := strings.TrimSpace(day)
	d = strings.ToLower(d)
	replacer := strings.NewReplacer("أ", "ا", "إ", "ا", "آ", "ا")
	return replacer.Replace(d)
}

// ميثود تحويل النص لرقم (الأحد=0 ... السبت=6)
func getWeekday(day string) int {
	switch day {
	case "الاحد", "sunday", "sun": return 0
	case "الاثنين", "monday", "mon": return 1
	case "الثلاثاء", "tuesday", "tue": return 2
	case "الاربعاء", "wednesday", "wed": return 3
	case "الخميس", "thursday", "thu": return 4
	case "الجمعة", "friday", "fri": return 5
	case "السبت", "saturday", "sat": return 6
	default: return -1
	}
}




func ToEmojiNumbers(input string) string {
	var builder strings.Builder

	for _, char := range input {
		switch char {
		case '0':
			builder.WriteString("0️⃣")
		case '1':
			builder.WriteString("1️⃣")
		case '2':
			builder.WriteString("2️⃣")
		case '3':
			builder.WriteString("3️⃣")
		case '4':
			builder.WriteString("4️⃣")
		case '5':
			builder.WriteString("5️⃣")
		case '6':
			builder.WriteString("6️⃣")
		case '7':
			builder.WriteString("7️⃣")
		case '8':
			builder.WriteString("8️⃣")
		case '9':
			builder.WriteString("9️⃣")
		default:
			builder.WriteRune(char)
		}
	}
	
	return builder.String()
}

func FromEmojiNumbers(input string) string {
	// نقوم بعمل خريطة للاستبدال: الإيموجي مقابل الرقم العادي
	replacer := strings.NewReplacer(
		"0️⃣", "0",
		"1️⃣", "1",
		"2️⃣", "2",
		"3️⃣", "3",
		"4️⃣", "4",
		"5️⃣", "5",
		"6️⃣", "6",
		"7️⃣", "7",
		"8️⃣", "8",
		"9️⃣", "9",
	)

	return replacer.Replace(input)
}



//  OTP Verification
func (db *ConnectDB) VerifyOTP(phone string, inputCode string) bool {
    // 1. البحث عن البيانات المرتبطة بالرقم
    value, exists := otpStore.Load(phone)
    if !exists {
        return false // الرقم ليس له كود مسجل
    }

    data := value.(otpData)

    // 2. التحقق من صلاحية الوقت (60 ثانية)
    if time.Now().After(data.expiresAt) {
        otpStore.Delete(phone) // احذف الكود المنتهي
        return false
    }

    // 3. مطابقة الكود
    if data.code == inputCode {
        otpStore.Delete(phone) // احذفه بعد النجاح لمنع استخدامه مرة أخرى
        return true
    }

    return false
}