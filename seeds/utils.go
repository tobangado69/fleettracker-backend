package seeds

import (
	"fmt"
	"math/rand"
	"time"
)

// Indonesian names database
var (
	indonesianFirstNames = []string{
		"Ahmad", "Budi", "Candra", "Dewi", "Eko", "Fitri", "Gita", "Hadi",
		"Indra", "Joko", "Kartika", "Lestari", "Made", "Nia", "Oki", "Putri",
		"Raden", "Siti", "Taufik", "Umar", "Vina", "Wawan", "Yudi", "Zainal",
		"Agus", "Bambang", "Dian", "Erna", "Fajar", "Hari", "Imam", "Jaya",
	}

	indonesianLastNames = []string{
		"Santoso", "Wijaya", "Kusuma", "Pratama", "Saputra", "Permana", "Nugroho",
		"Sutanto", "Hidayat", "Raharjo", "Setiawan", "Wibowo", "Gunawan", "Susanto",
		"Kurniawan", "Supardi", "Hartono", "Atmaja", "Pranoto", "Indarto", "Budiman",
		"Suryanto", "Firmansyah", "Harahap", "Situmorang", "Siregar", "Nasution",
	}

	jakartaStreets = []string{
		"Jl. Sudirman No. 45", "Jl. Thamrin No. 23", "Jl. Gatot Subroto No. 67",
		"Jl. Rasuna Said No. 12", "Jl. Kuningan Barat No. 34", "Jl. HR Rasuna Said No. 89",
		"Jl. MT Haryono No. 56", "Jl. Casablanca No. 78", "Jl. Menteng Raya No. 90",
	}

	surabayaStreets = []string{
		"Jl. Darmo No. 123", "Jl. Basuki Rahmat No. 45", "Jl. Pemuda No. 67",
		"Jl. Ahmad Yani No. 89", "Jl. Diponegoro No. 34", "Jl. Raya Gubeng No. 56",
	}

	vehicleBrands = []string{
		"Toyota", "Mitsubishi", "Isuzu", "Hino", "Daihatsu", "Suzuki", "Honda",
	}

	vehicleModels = map[string][]string{
		"Toyota":     {"Avanza", "Innova", "Fortuner", "Hilux", "Dyna"},
		"Mitsubishi": {"L300", "Colt Diesel", "Pajero Sport", "Triton"},
		"Isuzu":      {"Elf", "Panther", "D-Max", "Giga"},
		"Hino":       {"Dutro", "Ranger"},
		"Daihatsu":   {"Gran Max", "Xenia", "Terios"},
		"Suzuki":     {"Carry", "Ertiga", "APV"},
		"Honda":      {"Brio", "Mobilio", "CR-V"},
	}
)

// GenerateIndonesianName creates a realistic Indonesian name
func GenerateIndonesianName() (firstName, lastName string) {
	firstName = indonesianFirstNames[rand.Intn(len(indonesianFirstNames))]
	lastName = indonesianLastNames[rand.Intn(len(indonesianLastNames))]
	return
}

// GenerateNPWP creates a valid NPWP format (Indonesian Tax ID)
func GenerateNPWP() string {
	// Format: XX.XXX.XXX.X-XXX.XXX
	return fmt.Sprintf("%02d.%03d.%03d.%d-%03d.%03d",
		rand.Intn(100),
		rand.Intn(1000),
		rand.Intn(1000),
		rand.Intn(10),
		rand.Intn(1000),
		rand.Intn(1000),
	)
}

// GenerateNIK creates a valid NIK format (Indonesian ID Number)
func GenerateNIK() string {
	// Format: PPKKSSDDMMYY#### (16 digits)
	// PP = Province, KK = City, SS = District, DDMMYY = Birth date
	province := 31 + rand.Intn(10) // Jakarta/Surabaya codes
	city := rand.Intn(100)
	district := rand.Intn(100)
	day := 1 + rand.Intn(28)
	month := 1 + rand.Intn(12)
	year := 85 + rand.Intn(20) // 1985-2005
	serial := rand.Intn(10000)

	return fmt.Sprintf("%02d%02d%02d%02d%02d%02d%04d",
		province, city, district, day, month, year, serial)
}

// GenerateSIM creates a valid SIM (Driver's License) number
func GenerateSIM() string {
	// Format: XXXX-XXXX-XXXX (12 digits)
	return fmt.Sprintf("%04d-%04d-%04d",
		rand.Intn(10000),
		rand.Intn(10000),
		rand.Intn(10000),
	)
}

// GenerateLicensePlate creates a realistic Indonesian license plate
func GenerateLicensePlate(region string) string {
	// Jakarta: B, Surabaya: L
	var prefix string
	if region == "Jakarta" {
		prefix = "B"
	} else {
		prefix = "L"
	}

	// Format: X 1234 ABC
	number := 1000 + rand.Intn(9000)
	letters := []rune{'A' + rune(rand.Intn(26)), 'A' + rune(rand.Intn(26)), 'A' + rune(rand.Intn(26))}

	return fmt.Sprintf("%s %d %c%c%c", prefix, number, letters[0], letters[1], letters[2])
}

// GeneratePhoneNumber creates a valid Indonesian phone number
func GeneratePhoneNumber(isMobile bool) string {
	if isMobile {
		// Mobile: +62 8xx-xxxx-xxxx
		provider := []string{"812", "813", "821", "822", "852", "853"}
		return fmt.Sprintf("+62 %s-%04d-%04d",
			provider[rand.Intn(len(provider))],
			rand.Intn(10000),
			rand.Intn(10000),
		)
	}
	// Landline: +62 21-xxxx-xxxx (Jakarta) or +62 31-xxxx-xxxx (Surabaya)
	areaCode := []string{"21", "31"}
	return fmt.Sprintf("+62 %s-%04d-%04d",
		areaCode[rand.Intn(len(areaCode))],
		rand.Intn(10000),
		rand.Intn(10000),
	)
}

// GetJakartaAddress returns a random Jakarta address
func GetJakartaAddress() (street, city, province, postalCode string) {
	street = jakartaStreets[rand.Intn(len(jakartaStreets))]
	city = "Jakarta Pusat"
	province = "DKI Jakarta"
	postalCode = fmt.Sprintf("101%02d", rand.Intn(100))
	return
}

// GetSurabayaAddress returns a random Surabaya address
func GetSurabayaAddress() (street, city, province, postalCode string) {
	street = surabayaStreets[rand.Intn(len(surabayaStreets))]
	city = "Surabaya"
	province = "Jawa Timur"
	postalCode = fmt.Sprintf("601%02d", rand.Intn(100))
	return
}

// GetRandomVehicleBrandModel returns a random vehicle brand and compatible model
func GetRandomVehicleBrandModel() (brand, model string) {
	brand = vehicleBrands[rand.Intn(len(vehicleBrands))]
	models := vehicleModels[brand]
	model = models[rand.Intn(len(models))]
	return
}

// RandomBool returns a random boolean
func RandomBool() bool {
	return rand.Intn(2) == 1
}

// RandomFloat generates a random float between min and max
func RandomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// RandomDate generates a random date within the last N days
func RandomDate(daysAgo int) time.Time {
	return time.Now().AddDate(0, 0, -rand.Intn(daysAgo))
}

// Helper functions for pointer conversions
func ptrString(s string) *string {
	return &s
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func ptrDate(t time.Time) *time.Time {
	return &t
}

// init initializes the random seed
func init() {
	// No need to call rand.Seed in Go 1.20+
	// The global random number generator is automatically seeded
}

