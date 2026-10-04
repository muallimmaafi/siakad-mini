package service

// BatasSKS menentukan batas maksimal SKS per semester berdasarkan IPK terakhir.
func BatasSKS(ipk float64) int {
	switch {
	case ipk >= 3.00:
		return 24
	case ipk >= 2.50:
		return 21
	default:
		return 18
	}
}