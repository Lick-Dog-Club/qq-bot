package util

func FuzzyPhone(phone string) string {
	if len(phone) == 11 {
		phone = phone[0:3] + "****" + phone[7:11]
	}
	return phone
}
