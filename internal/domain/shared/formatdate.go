package share

import (
	"fmt"
	"time"
)

// format date from yyyymmdd to yyy mmm dd, HH:MM
func FormatDate(dateStr string) string {
	// Map bulan ke nama bulan dalam bahasa Indonesia
	monthString := map[string]string{
		"01": "Jan",
		"02": "Feb",
		"03": "Mar",
		"04": "Apr",
		"05": "May",
		"06": "Jun",
		"07": "Jul",
		"08": "Aug",
		"09": "Sep",
		"10": "Oct",
		"11": "Nov",
		"12": "Dec",
	}

	// Parse string to date component
	year := dateStr[0:4]
	month := dateStr[4:6]
	day := dateStr[6:8]

	if len(dateStr) != 8 {
		hour := dateStr[8:10]
		minute := dateStr[10:12]

		// Format date to string
		formattedDate := fmt.Sprintf("%s %s %s, %s:%s",
			day,
			monthString[month],
			year,
			hour,
			minute,
		)
		return formattedDate
	}

	// Format date to string
	formattedDate := fmt.Sprintf("%s/%s/%s",
		day,
		monthString[month],
		year,
	)

	return formattedDate
}

// format date from yyyy-mm-dd to yyyymmddHHMM
func FormatToCompactDateTime(input string) (string, error) {
	parsedTime, err := time.Parse("2006-01-02", input)
	if err != nil {
		return "", fmt.Errorf("invalid input date format: %w", err)
	}
	return parsedTime.Format("20060102"), nil
}