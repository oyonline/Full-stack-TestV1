package dateUtils

import (
	"fmt"
	"math"
	"time"
)

func DateDiff(layout, endDate, startDate string) int {
	start, _ := time.Parse(layout, startDate)
	end, _ := time.Parse(layout, endDate)
	return int(end.Sub(start).Hours()/24) + 1
}

func DateAddDays(baseDate string, days int) string {
	layout := "2006-01-02"
	t, _ := time.Parse(layout, baseDate)
	newDate := t.AddDate(0, 0, days)
	return newDate.Format(layout)
}

func NowDate() string {
	return time.Now().Format(time.DateOnly)
}

func NowMonth() string {
	return time.Now().Format("2006-01")
}

func NowWeek() string {
	now := time.Now()
	year, week := now.ISOWeek()

	return fmt.Sprintf("%d-W%02d\n", year, week)
}

// CompareDates 比较两个日期字符串的大小
func CompareDates(date1, date2 string) int {
	// 解析输入的日期字符串
	d1, err := time.Parse(time.DateOnly, date1)
	if err != nil {
		return 0
	}
	d2, err := time.Parse(time.DateOnly, date2)
	if err != nil {
		return 0
	}

	// 比较两个日期
	if d1.Before(d2) {
		return -1 // date1 小于 date2
	} else if d1.After(d2) {
		return 1 // date1 大于 date2
	} else {
		return 0 // date1 等于 date2
	}
}

// GetNumOfYear 根据日期字符串和返回类型获取年份、第几天或周数
func GetNumOfYear(dateString string, returnType int) int {
	// 解析输入的日期字符串
	date, err := time.Parse(time.DateOnly, dateString)
	if err != nil {
		return 0 // 如果解析失败，返回错误
	}

	// 根据 returnType 返回对应的值
	switch returnType {
	case 0:
		// 返回本年第几日
		return date.YearDay()
	case 1:
		// 返回本月第几日
		return date.Day()
	case 2:
		// 返回年份
		return date.Year()
	case 3:
		// 返回月份
		return int(date.Month())
	case 4:
		// 返回周数
		_, week := date.ISOWeek()
		return week
	default:
		return 0
	}
}

var PossibleLayouts = []string{
	"2006-01-02",
	"2006年01月02日",
	"2006年1月2日",
	"01-02-2006",
	"02-01-2006",
	"01-02-06",
	"02-01-06",
	"2006/01/02",
	"01/02/2006",
	"02/01/2006",
	"01/02/06",
	"02/01/06",
	"2006.01.02",
	"01.02.2006",
	"02.01.2006",
	"01.02.06",
	"02.01.06",
	"2006-01-02 15:04",
	"2006-01-02 15:04:05",
	"2006/01/02 15:04",
	"2006/01/02 15:04:05",
	"01/02/06 15:04",
	"02/01/06 15:04:05",
	"01-02-06 15:04",
	"02-01-06 15:04:05",
	"2006-01-02T15:04:05.999",
	"2006-01-02 15:04:05.999",
}

func ParseDate(dateStr string, formatter string) string {
	if dateStr == "" {
		return ""
	}
	var parsedTime time.Time
	var err error
	for _, format := range PossibleLayouts {
		parsedTime, err = time.Parse(format, dateStr)
		if err == nil {
			return parsedTime.Format(formatter)
		}
	}
	return dateStr
}

// CalculateEndTime 计算结束时间
// startAt: 开始时间
// timeType: 1=小时, 2=天
// timeValue: 值
// 返回: 结束时间
func CalculateEndTime(startAt string, timeType, addValue int) string {
	// 1. 统一转换为需要增加的总工作小时数
	totalWorkHours := 0
	if timeType == 1 {
		totalWorkHours = addValue
	} else if timeType == 2 {
		totalWorkHours = addValue * 8 // 每天8小时
	} else {
		// 无效的 timeType，返回原时间或 panic
		return startAt
	}

	// 2. 初始化当前时间
	current, _ := time.Parse(time.DateTime, startAt)

	// 3. 逐小时增加，直到满足 totalWorkHours
	for i := 0; i < totalWorkHours; {
		// 增加一小时
		current = current.Add(time.Hour)

		// 获取当前小时所在的日期是星期几
		weekday := current.Weekday()

		// 4. 如果是周六或周日，跳过，不计入工作小时
		if weekday == time.Saturday || weekday == time.Sunday {
			continue // 不增加 i，继续循环，相当于这一个小时不算
		}

		// 5. 只有在工作日时，才计入工作小时
		i++
	}

	return current.Format(time.DateTime)
}

// RoundToHour 将时间四舍五入到最接近的小时 Exp: RoundToHour(dateString,format)
func RoundToHour(arg ...string) time.Time {
	// 获取当前时间的分钟和秒
	t := time.Now()
	if arg != nil && len(arg) > 1 {
		t, _ = time.Parse(arg[1], arg[0])
	}
	_, minc, sec := t.Clock()
	duration := time.Duration(minc)*time.Minute + time.Duration(sec)*time.Second

	// 判断是否需要向上取整到下一小时
	if duration >= 30*time.Minute {
		// 向上取整：进位到下一小时，然后归零分钟和秒
		t = t.Truncate(time.Hour).Add(time.Hour)
	} else {
		// 向下取整：直接截断到小时
		t = t.Truncate(time.Hour)
	}
	return t
}

// WorkDayAddHour 给一个时间点增加指定的工作小时数，只计算工作日 9:00-18:00 的时间
// 跳过周末（周六、周日）
func WorkDayAddHour(date time.Time, hours int) time.Time {
	// 如果小时数为 0，直接返回原时间
	if hours <= 0 {
		return date
	}
	current := date

	for hours > 0 {
		// 跳过周末
		for IsWeekend(current) {
			// 跳到下一天的 9:00
			current = current.Add(24 * time.Hour)
			current = time.Date(current.Year(), current.Month(), current.Day(), 9, 0, 0, 0, current.Location())
		}

		// 获取当天的工作时间边界
		workStart := time.Date(current.Year(), current.Month(), current.Day(), 9, 0, 0, 0, current.Location())
		workEnd := time.Date(current.Year(), current.Month(), current.Day(), 18, 0, 0, 0, current.Location())

		var availableHours int
		var newCurrent time.Time
		// 向后推（加时间）
		// 当前时间在工作时间内
		if current.After(workStart) && current.Before(workEnd) || current.Equal(workStart) {
			// 可用小时数 = 工作结束时间 - 当前时间（向下取整）
			availableHours = int(workEnd.Sub(current).Hours())
			if availableHours > hours {
				// 不需要跨天，直接加上小时
				newCurrent = current.Add(time.Duration(hours) * time.Hour)
				hours = 0
			} else {
				// 跨天，加上可用小时，然后跳到下一天
				newCurrent = workEnd
				hours -= availableHours
				// 跳到下一天 9:00
				newCurrent = newCurrent.Add(15*time.Hour + 24*time.Hour) // 加15小时到明天9点
			}
		} else if current.Before(workStart) {
			// 当前时间在工作开始之前，从当天9:00开始算
			newCurrent = workStart
			availableHours = 9
			if availableHours > hours {
				newCurrent = newCurrent.Add(time.Duration(hours) * time.Hour)
				hours = 0
			} else {
				newCurrent = workEnd
				hours -= availableHours
				newCurrent = newCurrent.Add(15*time.Hour + 24*time.Hour)
			}
		} else {
			// 当前时间在工作结束之后，从下一天9:00开始算
			newCurrent = workStart
			newCurrent = newCurrent.Add(24 * time.Hour) // 下一天
			availableHours = 9
			if availableHours > hours {
				newCurrent = newCurrent.Add(time.Duration(hours) * time.Hour)
				hours = 0
			} else {
				newCurrent = workEnd
				newCurrent = newCurrent.Add(24 * time.Hour)
				hours -= availableHours
			}
		}

		current = newCurrent
	}

	return current
}

// IsWeekend 判断是否为周末
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// WorkHoursBetween 计算两个时间之间的工作小时数（只在 09:00-18:00，跳过周末）
func WorkHoursBetween(start, end time.Time) float64 {
	if start.After(end) {
		return 0
	}

	location := start.Location()
	current := start
	totalHours := 0.0
	currentStart := time.Date(current.Year(), current.Month(), current.Day(), 9, 0, 0, 0, location)
	currentEnd := time.Date(current.Year(), current.Month(), current.Day(), 18, 0, 0, 0, location)

	if current.After(currentEnd) {
		current = current.Add(24 * time.Hour)
		current = time.Date(current.Year(), current.Month(), current.Day(), 9, 0, 0, 0, location)
	}
	if current.Before(currentStart) {
		current = time.Date(current.Year(), current.Month(), current.Day(), 9, 0, 0, 0, location)
	}
	for current.Before(end) || current.Equal(end) {
		// 跳过周末
		if IsWeekend(current) {
			// 跳到下一个工作日 09:00
			current = current.Add(24 * time.Hour)
			current = time.Date(current.Year(), current.Month(), current.Day(), 9, 0, 0, 0, location)
			continue
		}

		// 获取当天工作时间边界
		dayStart := time.Date(current.Year(), current.Month(), current.Day(), 9, 0, 0, 0, location)
		dayEnd := time.Date(current.Year(), current.Month(), current.Day(), 18, 0, 0, 0, location)

		// 确定当前时间段的起止
		segStart := current
		if segStart.Before(dayStart) {
			segStart = dayStart
		}

		segEnd := end
		if segEnd.After(dayEnd) {
			segEnd = dayEnd
		}
		if segEnd.Before(dayStart) {
			// 当天无交集，跳过
			current = current.Add(24 * time.Hour)
			current = time.Date(current.Year(), current.Month(), current.Day(), 9, 0, 0, 0, location)
			continue
		}

		// 计算当前段的工作时间
		if segStart.Before(segEnd) {
			totalHours += segEnd.Sub(segStart).Hours() - 1
		}

		// 跳到下一天
		current = current.Add(24 * time.Hour)
		current = time.Date(current.Year(), current.Month(), current.Day(), 9, 0, 0, 0, location)
	}

	return totalHours
}

// CalculateDelayHours 计算延期小时数（四舍五入）
func CalculateDelayHours(actualTime, estimatedTime, format string) int {
	if len(estimatedTime) == 10 {
		estimatedTime = estimatedTime + " 23:59:59"
	}
	if len(actualTime) == 10 {
		actualTime = actualTime + " 23:59:59"
	}
	actual, _ := time.Parse(format, actualTime)
	estimated, _ := time.Parse(format, estimatedTime)
	if !actual.After(estimated) {
		return 0 // 未延期
	}

	delayHours := WorkHoursBetween(estimated, actual)
	return int(math.Round(delayHours))
}

// AddMonths 在 YYYY-MM 格式的基础上增加 N 个月（内部辅助函数）
func AddMonths(yearMonth string, months int) string {
	t, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return yearMonth
	}
	return t.AddDate(0, months, 0).Format("2006-01")
}
