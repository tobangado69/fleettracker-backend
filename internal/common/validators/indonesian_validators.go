package validators

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Indonesian validation errors
var (
	ErrInvalidNIK          = fmt.Errorf("NIK must be 16 digits")
	ErrInvalidNIKFormat    = fmt.Errorf("NIK contains invalid characters")
	ErrInvalidNIKChecksum  = fmt.Errorf("NIK checksum validation failed")
	ErrInvalidSIM          = fmt.Errorf("SIM number must be 12 digits")
	ErrInvalidSIMFormat    = fmt.Errorf("SIM contains invalid characters")
	ErrInvalidPlateNumber  = fmt.Errorf("invalid Indonesian license plate format")
	ErrInvalidKTP          = fmt.Errorf("KTP must be 16 digits")
	ErrInvalidNPWP         = fmt.Errorf("NPWP must be 15 digits")
	ErrInvalidNPWPFormat   = fmt.Errorf("NPWP format invalid")
	ErrInvalidPhoneNumber  = fmt.Errorf("invalid Indonesian phone number")
	ErrInvalidPostalCode   = fmt.Errorf("postal code must be 5 digits")
)

// ValidateNIK validates Indonesian National ID (Nomor Induk Kependudukan)
// Format: 16 digits (DDMMYYPPKKKKSSSS)
// DD = District, MM = Month, YY = Year, PP = Province, KKKK = Subdistrict, SSSS = Sequential
func ValidateNIK(nik string) error {
	// Remove any whitespace
	nik = strings.TrimSpace(nik)

	// Check length
	if len(nik) != 16 {
		return ErrInvalidNIK
	}

	// Check if all digits
	if !isNumeric(nik) {
		return ErrInvalidNIKFormat
	}

	// Validate district code (first 2 digits: 01-99)
	district, err := strconv.Atoi(nik[0:2])
	if err != nil || district < 1 || district > 99 {
		return fmt.Errorf("invalid district code in NIK")
	}

	// Validate month (positions 3-4: 01-12 for male, 41-52 for female)
	month, err := strconv.Atoi(nik[2:4])
	if err != nil {
		return fmt.Errorf("invalid month in NIK")
	}
	if (month < 1 || month > 12) && (month < 41 || month > 52) {
		return fmt.Errorf("invalid month code in NIK (must be 01-12 or 41-52)")
	}

	// Validate year (positions 5-6: 00-99)
	year, err := strconv.Atoi(nik[4:6])
	if err != nil || year < 0 || year > 99 {
		return fmt.Errorf("invalid year in NIK")
	}

	// Validate province code (positions 7-8: 01-99)
	province, err := strconv.Atoi(nik[6:8])
	if err != nil || province < 1 || province > 99 {
		return fmt.Errorf("invalid province code in NIK")
	}

	return nil
}

// ValidateSIM validates Indonesian Driver's License Number (SIM)
// Format: 12 digits
func ValidateSIM(sim string) error {
	// Remove any whitespace
	sim = strings.TrimSpace(sim)

	// Check length
	if len(sim) != 12 {
		return ErrInvalidSIM
	}

	// Check if all digits
	if !isNumeric(sim) {
		return ErrInvalidSIMFormat
	}

	return nil
}

// ValidatePlateNumber validates Indonesian license plate format
// Formats:
//   - "B 1234 ABC" (standard)
//   - "B 1234 AB" (2-letter suffix)
//   - "DK 1234 AB" (2-letter prefix for diplomatic)
//   - Accepts with or without spaces
func ValidatePlateNumber(plate string) error {
	// Remove extra whitespace and convert to uppercase
	plate = strings.ToUpper(strings.TrimSpace(plate))

	// Remove spaces for validation
	plateNoSpace := strings.ReplaceAll(plate, " ", "")

	// Pattern: 1-2 letters + 1-4 digits + 1-3 letters
	pattern := `^[A-Z]{1,2}\d{1,4}[A-Z]{1,3}$`
	matched, err := regexp.MatchString(pattern, plateNoSpace)
	if err != nil {
		return err
	}

	if !matched {
		return ErrInvalidPlateNumber
	}

	// Validate prefix (province codes)
	validPrefixes := map[string]bool{
		// Java
		"B": true, "D": true, "E": true, "F": true, "G": true, "H": true,
		"K": true, "L": true, "M": true, "N": true, "P": true, "R": true,
		"S": true, "T": true, "W": true, "Z": true, "AA": true, "AB": true,
		"AD": true, "AE": true, "AG": true,
		// Sumatra
		"BA": true, "BB": true, "BD": true, "BE": true, "BG": true, "BH": true,
		"BK": true, "BL": true, "BM": true, "BN": true, "BP": true, "BT": true,
		// Kalimantan
		"DA": true, "KB": true, "KH": true, "KT": true,
		// Sulawesi
		"DB": true, "DC": true, "DD": true, "DE": true, "DG": true, "DL": true,
		"DM": true, "DN": true, "DR": true, "DT": true,
		// Other islands
		"PA": true, "PB": true, // Papua
		"DK": true, // Bali
		"ED": true, "EA": true, "EB": true, // Nusa Tenggara
	}

	// Extract prefix (1 or 2 letters)
	prefix := ""
	if len(plateNoSpace) > 1 && isAlpha(string(plateNoSpace[1])) {
		prefix = plateNoSpace[0:2]
	} else {
		prefix = plateNoSpace[0:1]
	}

	if !validPrefixes[prefix] {
		return fmt.Errorf("invalid license plate prefix: %s", prefix)
	}

	return nil
}

// ValidateKTP validates Indonesian KTP (same as NIK in modern system)
func ValidateKTP(ktp string) error {
	return ValidateNIK(ktp)
}

// ValidateNPWP validates Indonesian Tax ID (Nomor Pokok Wajib Pajak)
// Format: 15 digits (XX.XXX.XXX.X-XXX.XXX)
// Or: XXXXXXXXXXXXXXX (without formatting)
func ValidateNPWP(npwp string) error {
	// Remove formatting (dots and dash)
	npwpClean := strings.ReplaceAll(npwp, ".", "")
	npwpClean = strings.ReplaceAll(npwpClean, "-", "")
	npwpClean = strings.TrimSpace(npwpClean)

	// Check length
	if len(npwpClean) != 15 {
		return ErrInvalidNPWP
	}

	// Check if all digits
	if !isNumeric(npwpClean) {
		return ErrInvalidNPWPFormat
	}

	return nil
}

// ValidatePhoneNumber validates Indonesian phone number
// Formats accepted:
//   - +62812345678 (international)
//   - 0812345678 (local)
//   - 812345678 (without prefix)
func ValidatePhoneNumber(phone string) error {
	// Remove spaces and dashes
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.TrimSpace(phone)

	// Pattern: +62 or 0 prefix, then 8-12 digits
	patterns := []string{
		`^\+628\d{8,11}$`,  // +628xxxxxxxxx
		`^08\d{8,11}$`,     // 08xxxxxxxxx
		`^628\d{8,11}$`,    // 628xxxxxxxxx (without +)
		`^8\d{8,11}$`,      // 8xxxxxxxxx
	}

	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, phone)
		if err != nil {
			return err
		}
		if matched {
			return nil
		}
	}

	return ErrInvalidPhoneNumber
}

// ValidatePostalCode validates Indonesian postal code (Kode Pos)
// Format: 5 digits
func ValidatePostalCode(postalCode string) error {
	postalCode = strings.TrimSpace(postalCode)

	if len(postalCode) != 5 {
		return ErrInvalidPostalCode
	}

	if !isNumeric(postalCode) {
		return ErrInvalidPostalCode
	}

	return nil
}

// FormatNPWP formats NPWP with proper formatting
// Input: XXXXXXXXXXXXXXX
// Output: XX.XXX.XXX.X-XXX.XXX
func FormatNPWP(npwp string) string {
	// Remove existing formatting
	clean := strings.ReplaceAll(npwp, ".", "")
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.TrimSpace(clean)

	if len(clean) != 15 {
		return npwp // Return original if invalid
	}

	return fmt.Sprintf("%s.%s.%s.%s-%s.%s",
		clean[0:2],
		clean[2:5],
		clean[5:8],
		clean[8:9],
		clean[9:12],
		clean[12:15],
	)
}

// FormatPhoneNumber formats Indonesian phone number to international format
// Output: +628xxxxxxxxx
func FormatPhoneNumber(phone string) string {
	// Remove formatting
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.TrimSpace(phone)

	// Convert to +62 format
	if strings.HasPrefix(phone, "+62") {
		return phone
	}
	if strings.HasPrefix(phone, "62") {
		return "+" + phone
	}
	if strings.HasPrefix(phone, "0") {
		return "+62" + phone[1:]
	}
	return "+62" + phone
}

// FormatPlateNumber formats Indonesian license plate with proper spacing
// Input: "B1234ABC" or "B 1234 ABC"
// Output: "B 1234 ABC"
func FormatPlateNumber(plate string) string {
	// Remove existing spaces and convert to uppercase
	plate = strings.ToUpper(strings.TrimSpace(plate))
	plateClean := strings.ReplaceAll(plate, " ", "")

	if len(plateClean) < 5 {
		return plate // Return original if too short
	}

	// Extract prefix (1-2 letters)
	prefixLen := 1
	if len(plateClean) > 1 && isAlpha(string(plateClean[1])) {
		prefixLen = 2
	}
	prefix := plateClean[0:prefixLen]

	// Extract number and suffix
	rest := plateClean[prefixLen:]
	
	// Find where numbers end
	numEnd := 0
	for i, char := range rest {
		if char < '0' || char > '9' {
			numEnd = i
			break
		}
	}

	if numEnd == 0 {
		return plate // Return original if no numbers found
	}

	number := rest[0:numEnd]
	suffix := rest[numEnd:]

	return fmt.Sprintf("%s %s %s", prefix, number, suffix)
}

// ValidateSIMType validates SIM type (A, B1, B2, C)
func ValidateSIMType(simType string) error {
	validTypes := map[string]bool{
		"A":  true, // Motorcycle
		"B1": true, // Car
		"B2": true, // Bus/Truck
		"C":  true, // Trailer/Heavy equipment
		"D":  true, // Special purpose
	}

	simType = strings.ToUpper(strings.TrimSpace(simType))
	if !validTypes[simType] {
		return fmt.Errorf("invalid SIM type: must be A, B1, B2, C, or D")
	}

	return nil
}

// ValidateVehicleType validates Indonesian vehicle type
func ValidateVehicleType(vehicleType string) error {
	validTypes := map[string]bool{
		"sedan":      true,
		"suv":        true,
		"mpv":        true,
		"pickup":     true,
		"van":        true,
		"truck":      true,
		"bus":        true,
		"motorcycle": true,
		"minibus":    true,
	}

	vehicleType = strings.ToLower(strings.TrimSpace(vehicleType))
	if !validTypes[vehicleType] {
		return fmt.Errorf("invalid vehicle type: %s", vehicleType)
	}

	return nil
}

// ValidateFuelType validates Indonesian fuel type
func ValidateFuelType(fuelType string) error {
	validTypes := map[string]bool{
		"pertalite": true, // RON 90
		"pertamax":  true, // RON 92
		"pertamax_turbo": true, // RON 98
		"solar":     true, // Diesel
		"biosolar":  true, // Bio Diesel
		"dexlite":   true, // Premium Diesel
		"pertamina_dex": true, // High-grade Diesel
		"lpg":       true, // Gas
		"cng":       true, // Compressed Natural Gas
	}

	fuelType = strings.ToLower(strings.TrimSpace(fuelType))
	if !validTypes[fuelType] {
		return fmt.Errorf("invalid fuel type: %s", fuelType)
	}

	return nil
}

// ValidateProvince validates Indonesian province name
func ValidateProvince(province string) error {
	provinces := map[string]bool{
		// Java
		"dki jakarta": true, "jawa barat": true, "jawa tengah": true,
		"jawa timur": true, "yogyakarta": true, "banten": true,
		// Sumatra
		"aceh": true, "sumatera utara": true, "sumatera barat": true,
		"riau": true, "jambi": true, "sumatera selatan": true,
		"bengkulu": true, "lampung": true, "kepulauan bangka belitung": true,
		"kepulauan riau": true,
		// Kalimantan
		"kalimantan barat": true, "kalimantan tengah": true, "kalimantan selatan": true,
		"kalimantan timur": true, "kalimantan utara": true,
		// Sulawesi
		"sulawesi utara": true, "sulawesi tengah": true, "sulawesi selatan": true,
		"sulawesi tenggara": true, "gorontalo": true, "sulawesi barat": true,
		// Other
		"bali": true, "nusa tenggara barat": true, "nusa tenggara timur": true,
		"maluku": true, "maluku utara": true, "papua": true, "papua barat": true,
		"papua tengah": true, "papua pegunungan": true, "papua selatan": true,
		"papua barat daya": true,
	}

	province = strings.ToLower(strings.TrimSpace(province))
	if !provinces[province] {
		return fmt.Errorf("invalid province: %s", province)
	}

	return nil
}

// ValidateCity validates major Indonesian city
func ValidateCity(city string) error {
	// This is a simplified version - you might want to add all cities
	majorCities := map[string]bool{
		"jakarta": true, "surabaya": true, "bandung": true, "medan": true,
		"semarang": true, "makassar": true, "palembang": true, "tangerang": true,
		"depok": true, "bekasi": true, "yogyakarta": true, "malang": true,
		"bogor": true, "batam": true, "pekanbaru": true, "bandar lampung": true,
	}

	city = strings.ToLower(strings.TrimSpace(city))
	
	// Allow any city name, just check it's not empty
	if city == "" {
		return fmt.Errorf("city cannot be empty")
	}

	// Optionally warn if not a major city (but don't reject)
	_ = majorCities

	return nil
}

// ValidateVIN validates Vehicle Identification Number
// 17 characters, alphanumeric (no I, O, Q to avoid confusion with 1, 0)
func ValidateVIN(vin string) error {
	vin = strings.ToUpper(strings.TrimSpace(vin))

	if len(vin) != 17 {
		return fmt.Errorf("VIN must be 17 characters")
	}

	// Check for invalid characters
	invalidChars := regexp.MustCompile(`[IOQ]`)
	if invalidChars.MatchString(vin) {
		return fmt.Errorf("VIN cannot contain I, O, or Q")
	}

	// Check if alphanumeric
	validVIN := regexp.MustCompile(`^[A-HJ-NPR-Z0-9]{17}$`)
	if !validVIN.MatchString(vin) {
		return fmt.Errorf("VIN contains invalid characters")
	}

	return nil
}

// ValidateSIUP validates Business License Number (Surat Izin Usaha Perdagangan)
// Format varies, but typically 13-20 alphanumeric characters
func ValidateSIUP(siup string) error {
	siup = strings.TrimSpace(siup)

	if len(siup) < 10 || len(siup) > 30 {
		return fmt.Errorf("SIUP must be 10-30 characters")
	}

	// Allow alphanumeric and some special characters (/, -)
	validSIUP := regexp.MustCompile(`^[A-Z0-9\-\/]+$`)
	if !validSIUP.MatchString(strings.ToUpper(siup)) {
		return fmt.Errorf("SIUP contains invalid characters")
	}

	return nil
}

// ValidateEmail validates email format (RFC 5322)
func ValidateEmail(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// RFC 5322 regex (simplified)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must be less than 128 characters")
	}

	// Check for at least one uppercase
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	return nil
}

// ValidateUsername validates username format
func ValidateUsername(username string) error {
	username = strings.ToLower(strings.TrimSpace(username))

	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters")
	}

	if len(username) > 30 {
		return fmt.Errorf("username must be less than 30 characters")
	}

	// Alphanumeric and underscore only
	validUsername := regexp.MustCompile(`^[a-z0-9_]+$`)
	if !validUsername.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, and underscores")
	}

	// Must start with letter
	if username[0] < 'a' || username[0] > 'z' {
		return fmt.Errorf("username must start with a letter")
	}

	return nil
}

// ValidateDate validates date format and range
func ValidateDate(dateStr string) (time.Time, error) {
	// Try multiple formats
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"02/01/2006",
		"02-01-2006",
	}

	var parsedDate time.Time
	var err error

	for _, format := range formats {
		parsedDate, err = time.Parse(format, dateStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format")
	}

	// Check if date is reasonable (not too far in past or future)
	now := time.Now()
	if parsedDate.After(now.AddDate(10, 0, 0)) {
		return time.Time{}, fmt.Errorf("date cannot be more than 10 years in the future")
	}
	if parsedDate.Before(now.AddDate(-100, 0, 0)) {
		return time.Time{}, fmt.Errorf("date cannot be more than 100 years in the past")
	}

	return parsedDate, nil
}

// Helper functions

// isNumeric checks if string contains only digits
func isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

// isAlpha checks if string contains only letters
func isAlpha(s string) bool {
	for _, char := range s {
		if (char < 'A' || char > 'Z') && (char < 'a' || char > 'z') {
			return false
		}
	}
	return true
}

// NormalizeNIK normalizes NIK format (removes spaces, dashes)
func NormalizeNIK(nik string) string {
	nik = strings.TrimSpace(nik)
	nik = strings.ReplaceAll(nik, " ", "")
	nik = strings.ReplaceAll(nik, "-", "")
	return nik
}

// NormalizePlateNumber normalizes license plate format
func NormalizePlateNumber(plate string) string {
	return FormatPlateNumber(plate)
}

